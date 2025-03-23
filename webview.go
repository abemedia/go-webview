package webview

import (
	_ "embed"
	"encoding/json"
	"errors"
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

func init() {
	// Ensure that main.main is called from the main thread.
	runtime.LockOSThread()
}

// Hints are used to configure window sizing and resizing.
type Hint uintptr

const (
	// Width and height are default size.
	HintNone Hint = iota

	// Width and height are minimum bounds.
	HintMin

	// Width and height are maximum bounds.
	HintMax

	// Window size can not be changed by a user.
	HintFixed
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
	SetSize(w, h int, hint Hint)

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
	Bind(name string, f any) error

	// Removes a callback that was previously set by Bind.
	Unbind(name string) error
}

// New calls NewWindow to create a new window and a new webview instance. If debug
// is non-zero - developer tools will be enabled (if the platform supports them).
func New(debug bool) WebView { return NewWindow(debug, nil) }

// NewWindow creates a new webview instance. If debug is non-zero - developer
// tools will be enabled (if the platform supports them). Window parameter can be
// a pointer to the native window handle. If it's non-null - then child WebView is
// embedded into the given parent window. Otherwise a new window is created.
// Depending on the platform, a GtkWindow, NSWindow or HWND pointer can be passed
// here.
func NewWindow(debug bool, window unsafe.Pointer) WebView {
	load()
	w := webviewCreate(debug, window)
	if w == nil {
		panic("webview: failed to create window")
	}
	return &webview{w: w}
}

type webviewError int

func (e webviewError) Error() string {
	switch e {
	case -5:
		return "missing dependency"
	case -4:
		return "operation canceled"
	case -3:
		return "invalid state"
	case -2:
		return "invalid argument"
	case -1:
		return "unspecified error"
	case 0:
		return "no error"
	case 1:
		return "duplicate"
	case 2:
		return "not found"
	default:
		return "unknown error"
	}
}

// webview is a concrete implementation of WebView using native library calls.
type webview struct {
	w unsafe.Pointer
}

// Global once to load native library symbols.
var loadOnce sync.Once

// Function pointers for native library functions.
var (
	webviewCreate    func(debug bool, window unsafe.Pointer) unsafe.Pointer
	webviewDestroy   func(window unsafe.Pointer)
	webviewRun       func(window unsafe.Pointer)
	webviewTerminate func(window unsafe.Pointer)
	webviewDispatch  func(window unsafe.Pointer, fn, arg uintptr)
	webviewGetWindow func(window unsafe.Pointer) unsafe.Pointer
	webviewSetTitle  func(window unsafe.Pointer, title string) webviewError
	webviewSetSize   func(webview unsafe.Pointer, width, height, hints int32) webviewError
	webviewNavigate  func(window unsafe.Pointer, url string) webviewError
	webviewSetHtml   func(window unsafe.Pointer, html string) webviewError
	webviewInit      func(window unsafe.Pointer, js string) webviewError
	webviewEval      func(window unsafe.Pointer, js string) webviewError
	webviewBind      func(window unsafe.Pointer, name string, fn, arg uintptr) webviewError
	webviewUnbind    func(window unsafe.Pointer, name string) webviewError
	webviewReturn    func(window unsafe.Pointer, id string, status int, result string) webviewError
)

// Pointer for libc malloc (for context allocation).
var cMalloc uintptr

// Callback function pointers.
var (
	dispatchCallbackPtr uintptr
	bindingCallbackPtr  uintptr
)

// Global state for managing dispatched functions and bound callbacks.
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
	fn func(id, req string) (any, error)
	w  unsafe.Pointer
}

func (w *webview) Run() {
	webviewRun(w.w)
}

func (w *webview) Terminate() {
	webviewTerminate(w.w)
}

func (w *webview) Dispatch(f func()) {
	dispatchMu.Lock()
	idx := dispatchCounter
	dispatchCounter++
	dispatchMap[idx] = f
	dispatchMu.Unlock()
	webviewDispatch(w.w, dispatchCallbackPtr, idx)
}

func (w *webview) Destroy() {
	webviewDestroy(w.w)
}

func (w *webview) Window() unsafe.Pointer {
	return webviewGetWindow(w.w)
}

func (w *webview) SetTitle(title string) {
	err := webviewSetTitle(w.w, title)
	if err != 0 {
		panic(err)
	}
}

func (w *webview) SetSize(width, height int, hint Hint) {
	err := webviewSetSize(w.w, int32(width), int32(height), int32(hint))
	if err != 0 {
		panic(err)
	}
}

func (w *webview) Navigate(url string) {
	err := webviewNavigate(w.w, url)
	if err != 0 {
		panic(err)
	}
}

func (w *webview) SetHtml(html string) {
	err := webviewSetHtml(w.w, html)
	if err != 0 {
		panic(err)
	}
}

func (w *webview) Init(js string) {
	err := webviewInit(w.w, js)
	if err != 0 {
		panic(err)
	}
}

func (w *webview) Eval(js string) {
	err := webviewEval(w.w, js)
	if err != 0 {
		panic(err)
	}
}

//nolint:gocognit,cyclop,funlen
func (w *webview) Bind(name string, f any) error {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return errors.New("only functions can be bound")
	}
	if outCount := v.Type().NumOut(); outCount > 2 {
		return errors.New("function may only return a value or a value+error")
	}
	bindMu.Lock()
	if _, exists := boundNames[name]; exists {
		bindMu.Unlock()
		return errors.New("webview: function name already bound")
	}
	// Create a wrapper that decodes JSON and calls the Go function.
	funcType := v.Type()
	wrapper := func(id, req string) (any, error) {
		var rawArgs []json.RawMessage
		if err := json.Unmarshal([]byte(req), &rawArgs); err != nil {
			return nil, err
		}
		isVariadic := funcType.IsVariadic()
		numIn := funcType.NumIn()
		if (!isVariadic && len(rawArgs) != numIn) || (isVariadic && len(rawArgs) < numIn-1) {
			return nil, errors.New("function arguments mismatch")
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
		var ret any
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
		r1, _, _ := purego.SyscallN(cMalloc, size)
		if r1 == 0 {
			bindMu.Unlock()
			return errors.New("webview: failed to allocate context")
		}
		contextKey = r1
	} else {
		contextKey = bindingCounter
		bindingCounter++
	}
	bindingMap[contextKey] = bindingEntry{w: w.w, fn: wrapper}
	boundNames[name] = contextKey
	bindMu.Unlock()
	err := webviewBind(w.w, name, bindingCallbackPtr, contextKey)
	if err != 0 {
		return err
	}
	return nil
}

func (w *webview) Unbind(name string) error {
	bindMu.Lock()
	contextKey, exists := boundNames[name]
	if !exists {
		bindMu.Unlock()
		return errors.New("webview: function name not bound")
	}
	delete(boundNames, name)
	delete(bindingMap, contextKey)
	bindMu.Unlock()
	err := webviewUnbind(w.w, name)
	if err != 0 {
		return err
	}
	return nil
}

// dispatchCallback executes a function posted with Dispatch on the main thread.
func dispatchCallback(_ unsafe.Pointer, arg uintptr) uintptr {
	dispatchMu.Lock()
	fn := dispatchMap[arg]
	delete(dispatchMap, arg)
	dispatchMu.Unlock()
	if fn != nil {
		fn()
	}
	return 0
}

// bindingCallback is invoked by the native webview when a bound JS function is called.
func bindingCallback(seqPtr, reqPtr, arg uintptr) uintptr {
	bindMu.Lock()
	entry, ok := bindingMap[arg]
	bindMu.Unlock()
	if !ok {
		return 0
	}
	id := cStringToGo(seqPtr)
	req := cStringToGo(reqPtr)
	resultValue, err := entry.fn(id, req)
	status := 0
	var resultJSON string
	if err != nil { //nolint:nestif
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
	webviewReturn(entry.w, id, status, resultJSON)
	return 0
}

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
