package webview

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Hints are used to configure window sizing and resizing
type Hint int

const (
	// Width and height are default size
	HintNone Hint = iota

	// Window size can not be changed by a user
	HintFixed

	// Width and height are minimum bounds
	HintMin

	// Width and height are maximum bounds
	HintMax
)

type WebView interface {
	// Run runs the main loop until it's terminated. After this function exits -
	// you must destroy the webview.
	Run()

	// Terminate stops the main loop. It is safe to call this function from
	// a background thread.
	Terminate()

	// Dispatch posts a function to be executed on the main thread. You normally
	// do not need to call this function, unless you want to tweak the native
	// window.
	Dispatch(f func())

	// Destroy destroys a webview and closes the native window.
	Destroy()

	// Window returns a native window handle pointer. When using GTK backend the
	// pointer is GtkWindow pointer, when using Cocoa backend the pointer is
	// NSWindow pointer, when using Win32 backend the pointer is HWND pointer.
	Window() unsafe.Pointer

	// SetTitle updates the title of the native window. Must be called from the UI
	// thread.
	SetTitle(title string)

	// SetSize updates native window size. See Hint constants.
	SetSize(w int, h int, hint Hint)

	// Navigate navigates webview to the given URL. URL may be a properly encoded data.
	// URI. Examples:
	// w.Navigate("https://github.com/webview/webview")
	// w.Navigate("data:text/html,%3Ch1%3EHello%3C%2Fh1%3E")
	// w.Navigate("data:text/html;base64,PGgxPkhlbGxvPC9oMT4=")
	Navigate(url string)

	// SetHtml sets the webview HTML directly.
	// Example: w.SetHtml(w, "<h1>Hello</h1>");
	SetHtml(html string)

	// Init injects JavaScript code at the initialization of the new page. Every
	// time the webview will open a the new page - this initialization code will
	// be executed. It is guaranteed that code is executed before window.onload.
	Init(js string)

	// Eval evaluates arbitrary JavaScript code. Evaluation happens asynchronously,
	// also the result of the expression is ignored. Use RPC bindings if you want
	// to receive notifications about the results of the evaluation.
	Eval(js string)

	// Bind binds a callback function so that it will appear under the given name
	// as a global JavaScript function. Internally it uses webview_init().
	// Callback receives a request string and a user-provided argument pointer.
	// Request string is a JSON array of all the arguments passed to the
	// JavaScript function.
	//
	// f must be a function
	// f must return either value and error or just error
	Bind(name string, f interface{}) error

	// Removes a callback that was previously set by Bind.
	Unbind(name string) error
}

// webviewImpl is a concrete implementation of WebView using native library calls.
type webviewImpl struct {
	handle uintptr
}

// Global once to load native library symbols
var loadOnce sync.Once

// Function pointers for native library functions
var (
	pCreate    uintptr
	pDestroy   uintptr
	pRun       uintptr
	pTerminate uintptr
	pDispatch  uintptr
	pGetWindow uintptr
	pSetTitle  uintptr
	pSetSize   uintptr
	pNavigate  uintptr
	pSetHtml   uintptr
	pInit      uintptr
	pEval      uintptr
	pBind      uintptr
	pUnbind    uintptr
	pReturn    uintptr
)

// Pointers for libc malloc/free (for context allocation)
var (
	cMalloc uintptr
	cFree   uintptr
)

// Callback function pointers
var (
	dispatchCallbackPtr uintptr
	bindingCallbackPtr  uintptr
)

// Global state for managing dispatched functions and bound callbacks
var (
	dispatchMu      sync.Mutex
	dispatchMap     = make(map[uintptr]func())
	dispatchCounter uintptr

	bindMu         sync.Mutex
	bindingMap     = make(map[uintptr]bindingEntry)
	boundNames     = make(map[string]uintptr)
	bindingCounter uintptr
)

// bindingEntry stores a bound callback and associated webview handle.
type bindingEntry struct {
	w  uintptr
	fn func(id, req string) (interface{}, error)
}

// Ensure main GUI thread is locked (some platforms require UI on main thread).
func init() {
	runtime.LockOSThread()
}

// boolToInt converts a bool to 0 or 1 (int) for passing to C functions.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ensureLibLoaded loads the shared library and resolves symbols (runs once).
func ensureLibLoaded() {
	loadOnce.Do(func() {
		fname := os.TempDir() + "/webview.so"

		libHandle, err := purego.Dlopen(fname, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			panic("webview: failed to load native library: " + err.Error())
		}
		if libHandle == 0 {
			panic("webview: native library not loaded")
		}
		// Resolve all required symbols from the library
		pCreate = mustLoadSymbol(libHandle, "webview_create")
		pDestroy = mustLoadSymbol(libHandle, "webview_destroy")
		pRun = mustLoadSymbol(libHandle, "webview_run")
		pTerminate = mustLoadSymbol(libHandle, "webview_terminate")
		pDispatch = mustLoadSymbol(libHandle, "webview_dispatch")
		pGetWindow = mustLoadSymbol(libHandle, "webview_get_window")
		pSetTitle = mustLoadSymbol(libHandle, "webview_set_title")
		pSetSize = mustLoadSymbol(libHandle, "webview_set_size")
		pNavigate = mustLoadSymbol(libHandle, "webview_navigate")
		pSetHtml = mustLoadSymbol(libHandle, "webview_set_html")
		pInit = mustLoadSymbol(libHandle, "webview_init")
		pEval = mustLoadSymbol(libHandle, "webview_eval")
		pBind = mustLoadSymbol(libHandle, "webview_bind")
		pUnbind = mustLoadSymbol(libHandle, "webview_unbind")
		pReturn = mustLoadSymbol(libHandle, "webview_return")

		// Attempt to load standard malloc/free for context allocation
		cMalloc, _ = purego.Dlsym(purego.RTLD_DEFAULT, "malloc")
		cFree, _ = purego.Dlsym(purego.RTLD_DEFAULT, "free")
		if cMalloc == 0 || cFree == 0 {
			if runtime.GOOS == "windows" {
				for _, lib := range []string{"msvcrt.dll", "ucrtbase.dll"} {
					if handle, err := purego.Dlopen(lib, purego.RTLD_NOW|purego.RTLD_GLOBAL); err == nil {
						cMalloc, _ = purego.Dlsym(handle, "malloc")
						cFree, _ = purego.Dlsym(handle, "free")
						// Do not Dlclose here; keep CRT loaded
						break
					}
				}
			}
		}

		// Create C-callable callbacks for dispatch and binding
		dispatchCallbackPtr = purego.NewCallback(dispatchCallback)
		bindingCallbackPtr = purego.NewCallback(bindingCallback)
	})
}

// mustLoadSymbol looks up a symbol and panics if not found.
func mustLoadSymbol(lib uintptr, name string) uintptr {
	ptr, err := purego.Dlsym(lib, name)
	if err != nil || ptr == 0 {
		panic("webview: failed to load symbol " + name + ": " + err.Error())
	}
	return ptr
}

// New creates a new webview window. If debug is true, developer tools are enabled.
func New(debug bool) WebView {
	return NewWindow(debug, nil)
}

// NewWindow creates a new webview. If window is non-nil, the webview will be embedded into the given native window.
func NewWindow(debug bool, window unsafe.Pointer) WebView {
	ensureLibLoaded()
	debugVal := boolToInt(debug)
	r1, _, _ := purego.SyscallN(pCreate, uintptr(debugVal), uintptr(window))
	if r1 == 0 {
		return nil // creation failed (e.g., missing webview runtime)
	}
	return &webviewImpl{handle: r1}
}

// Run starts the webview event loop and blocks until the window is closed.
func (w *webviewImpl) Run() {
	_, _, _ = purego.SyscallN(pRun, w.handle)
}

// Terminate stops the webview event loop, causing Run to return.
func (w *webviewImpl) Terminate() {
	_, _, _ = purego.SyscallN(pTerminate, w.handle)
}

// Dispatch posts a function to be executed on the main UI thread of the webview.
func (w *webviewImpl) Dispatch(f func()) {
	dispatchMu.Lock()
	idx := dispatchCounter
	dispatchCounter++
	dispatchMap[idx] = f
	dispatchMu.Unlock()
	_, _, _ = purego.SyscallN(pDispatch, w.handle, dispatchCallbackPtr, uintptr(idx))
}

// Destroy closes the webview and releases associated resources.
func (w *webviewImpl) Destroy() {
	_, _, _ = purego.SyscallN(pDestroy, w.handle)
}

// Window returns the native window handle of the webview (e.g., HWND, *GtkWindow, *NSWindow).
func (w *webviewImpl) Window() unsafe.Pointer {
	r1, _, _ := purego.SyscallN(pGetWindow, w.handle)
	return unsafe.Pointer(r1)
}

// SetTitle sets the title of the webview window.
func (w *webviewImpl) SetTitle(title string) {
	cs, ptr := goStringToCString(title)
	_, _, _ = purego.SyscallN(pSetTitle, w.handle, ptr)
	runtime.KeepAlive(cs)
}

// SetSize resizes the webview window to width x height, using the given Hint.
func (w *webviewImpl) SetSize(width int, height int, hint Hint) {
	_, _, _ = purego.SyscallN(pSetSize, w.handle, uintptr(width), uintptr(height), uintptr(hint))
}

// Navigate navigates the webview to the specified URL.
func (w *webviewImpl) Navigate(url string) {
	cs, ptr := goStringToCString(url)
	_, _, _ = purego.SyscallN(pNavigate, w.handle, ptr)
	runtime.KeepAlive(cs)
}

// SetHtml sets the webview content to the given HTML string.
func (w *webviewImpl) SetHtml(html string) {
	cs, ptr := goStringToCString(html)
	_, _, _ = purego.SyscallN(pSetHtml, w.handle, ptr)
	runtime.KeepAlive(cs)
}

// Init injects a JavaScript snippet into the webview to run before other scripts.
func (w *webviewImpl) Init(js string) {
	cs, ptr := goStringToCString(js)
	_, _, _ = purego.SyscallN(pInit, w.handle, ptr)
	runtime.KeepAlive(cs)
}

// Eval evaluates a JavaScript code string in the context of the current page.
func (w *webviewImpl) Eval(js string) {
	cs, ptr := goStringToCString(js)
	_, _, _ = purego.SyscallN(pEval, w.handle, ptr)
	runtime.KeepAlive(cs)
}

// Bind registers a Go function so it can be called from JavaScript as `window.name()`.
func (w *webviewImpl) Bind(name string, f interface{}) error {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return errors.New("webview: Bind requires a function")
	}
	if outCount := v.Type().NumOut(); outCount > 2 {
		return errors.New("webview: bound function may only return a value or a value and an error")
	}
	bindMu.Lock()
	if _, exists := boundNames[name]; exists {
		bindMu.Unlock()
		return errors.New("webview: function name already bound")
	}
	// Create a wrapper that decodes JSON and calls the Go function.
	funcType := v.Type()
	wrapper := func(id, req string) (interface{}, error) {
		var rawArgs []json.RawMessage
		if err := json.Unmarshal([]byte(req), &rawArgs); err != nil {
			return nil, err
		}
		isVariadic := funcType.IsVariadic()
		numIn := funcType.NumIn()
		if (!isVariadic && len(rawArgs) != numIn) || (isVariadic && len(rawArgs) < numIn-1) {
			return nil, errors.New("webview: argument count mismatch")
		}
		args := make([]reflect.Value, len(rawArgs))
		for i := range rawArgs {
			var argVal reflect.Value
			if isVariadic && i >= numIn-1 {
				argVal = reflect.New(funcType.In(numIn - 1).Elem())
			} else {
				argVal = reflect.New(funcType.In(i))
			}
			if err := json.Unmarshal(rawArgs[i], argVal.Interface()); err != nil {
				return nil, err
			}
			args[i] = argVal.Elem()
		}
		results := v.Call(args)
		// Handle function results (value, error) combinations
		var ret interface{}
		var err error
		switch outCount := v.Type().NumOut(); outCount {
		case 0:
			ret, err = nil, nil
		case 1:
			if funcType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				// Only error returned
				if resErr := results[0].Interface(); resErr != nil {
					err = resErr.(error)
				}
				ret = nil
			} else {
				// Only value returned
				ret = results[0].Interface()
				err = nil
			}
		case 2:
			// Value and error returned
			if results[1].Interface() != nil {
				err = results[1].Interface().(error)
			}
			ret = results[0].Interface()
		}
		return ret, err
	}
	// Use allocated context pointer if available, otherwise fallback to index key
	var contextKey uintptr
	if cMalloc != 0 {
		size := unsafe.Sizeof(uintptr(0)) * 2
		r1, _, _ := purego.SyscallN(cMalloc, uintptr(size))
		if r1 == 0 {
			bindMu.Unlock()
			return errors.New("webview: failed to allocate context")
		}
		contextKey = r1
	} else {
		contextKey = bindingCounter
		bindingCounter++
	}
	bindingMap[contextKey] = bindingEntry{w: w.handle, fn: wrapper}
	boundNames[name] = contextKey
	bindMu.Unlock()
	cs, namePtr := goStringToCString(name)
	_, _, _ = purego.SyscallN(pBind, w.handle, namePtr, bindingCallbackPtr, contextKey)
	runtime.KeepAlive(cs)
	return nil
}

// Unbind removes a previously bound function so it is no longer callable from JavaScript.
func (w *webviewImpl) Unbind(name string) error {
	bindMu.Lock()
	contextKey, exists := boundNames[name]
	if !exists {
		bindMu.Unlock()
		return errors.New("webview: function name not bound")
	}
	delete(boundNames, name)
	delete(bindingMap, contextKey)
	bindMu.Unlock()
	cs, namePtr := goStringToCString(name)
	_, _, _ = purego.SyscallN(pUnbind, w.handle, namePtr)
	runtime.KeepAlive(cs)
	return nil
}

// dispatchCallback executes a function posted with Dispatch on the main thread.
func dispatchCallback(_, arg uintptr) {
	dispatchMu.Lock()
	fn := dispatchMap[arg]
	delete(dispatchMap, arg)
	dispatchMu.Unlock()
	if fn != nil {
		fn()
	}
}

// bindingCallback is invoked by the native webview when a bound JS function is called.
func bindingCallback(seqPtr, reqPtr, arg uintptr) {
	bindMu.Lock()
	entry, ok := bindingMap[arg]
	bindMu.Unlock()
	if !ok {
		return
	}
	id := cStringToGo(seqPtr)
	req := cStringToGo(reqPtr)
	resultValue, err := entry.fn(id, req)
	status := 0
	var resultJSON string
	if err != nil {
		status = -1
		errMsg := err.Error()
		if data, e := json.Marshal(errMsg); e == nil {
			resultJSON = string(data)
		} else {
			resultJSON = "\"" + errMsg + "\""
		}
	} else {
		if data, e := json.Marshal(resultValue); e == nil {
			resultJSON = string(data)
		} else {
			status = -1
			msg := e.Error()
			if data, e2 := json.Marshal(msg); e2 == nil {
				resultJSON = string(data)
			} else {
				resultJSON = "\"" + msg + "\""
			}
		}
	}
	cs, resultPtr := goStringToCString(resultJSON)
	_, _, _ = purego.SyscallN(pReturn, entry.w, seqPtr, uintptr(status), resultPtr)
	runtime.KeepAlive(cs)
}

// goStringToCString converts a Go string to a NUL-terminated C string byte slice and returns the slice and pointer.
func goStringToCString(s string) ([]byte, uintptr) {
	b := append([]byte(s), 0)
	return b, uintptr(unsafe.Pointer(&b[0]))
}

// cStringToGo converts a NUL-terminated C string at ptr to a Go string.
func cStringToGo(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	// Determine length up to NUL terminator
	length := 0
	for {
		if *(*byte)(unsafe.Pointer(ptr + uintptr(length))) == 0 {
			break
		}
		length++
	}
	if length == 0 {
		return ""
	}
	// Copy bytes into a Go string
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = *(*byte)(unsafe.Pointer(ptr + uintptr(i)))
	}
	return string(buf)
}
