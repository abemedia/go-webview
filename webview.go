package webview

import (
	"encoding/json"
	"errors"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

// Hints are used to configure window sizing and resizing.
type Hint int

const (
	HintNone Hint = iota
	HintFixed
	HintMin
	HintMax
)

// WebView describes the common interface for the embedded browser window.
type WebView interface {
	Run()
	Terminate()
	Dispatch(f func())
	Destroy()
	Window() unsafe.Pointer
	SetTitle(title string)
	SetSize(w, h int, hint Hint)
	Navigate(url string)
	SetHtml(html string)
	Init(js string)
	Eval(js string)
	Bind(name string, f any) error
	Unbind(name string) error
}

// webview holds a handle to the native webview instance.
type webview struct {
	handle uintptr
}

// We assume you have assembly stubs named c_webview_* for each of these:
func c_webview_create(debug, wnd uintptr) uintptr
func c_webview_destroy(handle uintptr)
func c_webview_run(handle uintptr)
func c_webview_terminate(handle uintptr)
func c_webview_dispatch(handle, fnptr, userdata uintptr)
func c_webview_get_window(handle uintptr) uintptr
func c_webview_set_title(handle, title uintptr)
func c_webview_set_size(handle, width, height, hint uintptr)
func c_webview_navigate(handle, url uintptr)
func c_webview_set_html(handle, html uintptr)
func c_webview_init(handle, js uintptr)
func c_webview_eval(handle, js uintptr)
func c_webview_bind(handle, name, fnptr, ctx uintptr)
func c_webview_unbind(handle, name uintptr)
func c_webview_return(handle, seq, status, result uintptr)

// If you truly need malloc/free from the C runtime (and your static library calls them):
func c_malloc(size uintptr) uintptr
func c_free(ptr uintptr)

// Provide zero or actual function-pointer addresses (see the callback caveat above).
var (
	dispatchCallbackPtr uintptr = 0
	bindingCallbackPtr  uintptr = 0
)

// For queued dispatch calls from other goroutines
var (
	dispatchMu      sync.Mutex
	dispatchMap     = make(map[uintptr]func())
	dispatchCounter uintptr
)

// For bound functions
type bindingEntry struct {
	fn func(id, req string) (any, error)
	w  uintptr
}

var (
	bindMu         sync.Mutex
	bindingMap     = make(map[uintptr]bindingEntry)
	boundNames     = make(map[string]uintptr)
	bindingCounter uintptr
)

// boolToInt is a helper to convert Go bool to uintptr 0/1
func boolToInt(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}

// New creates a new webview, debugging off/on, with its own native window.
func New(debug bool) WebView {
	return NewWindow(debug, nil)
}

// NewWindow creates a new webview. If `window` is non-nil, the library
// may embed the webview in the given native window handle.
func NewWindow(debug bool, window unsafe.Pointer) WebView {
	handle := c_webview_create(boolToInt(debug), uintptr(window))
	if handle == 0 {
		return nil // creation failed
	}
	return &webview{handle: handle}
}

func (w *webview) Run() {
	c_webview_run(w.handle)
}

func (w *webview) Terminate() {
	c_webview_terminate(w.handle)
}

func (w *webview) Dispatch(f func()) {
	// Enqueue a function and pass an index to native code
	dispatchMu.Lock()
	idx := dispatchCounter
	dispatchCounter++
	dispatchMap[idx] = f
	dispatchMu.Unlock()

	// This will pass zero as the callback pointer if dispatchCallbackPtr=0
	c_webview_dispatch(w.handle, dispatchCallbackPtr, idx)
}

func (w *webview) Destroy() {
	c_webview_destroy(w.handle)
}

func (w *webview) Window() unsafe.Pointer {
	r := c_webview_get_window(w.handle)
	return unsafe.Pointer(r)
}

func (w *webview) SetTitle(title string) {
	b, ptr := goStringToCString(title)
	c_webview_set_title(w.handle, ptr)
	runtime.KeepAlive(b) // ensure b isn't GC'd
}

func (w *webview) SetSize(width, height int, hint Hint) {
	c_webview_set_size(w.handle, uintptr(width), uintptr(height), uintptr(hint))
}

func (w *webview) Navigate(url string) {
	b, ptr := goStringToCString(url)
	c_webview_navigate(w.handle, ptr)
	runtime.KeepAlive(b)
}

func (w *webview) SetHtml(html string) {
	b, ptr := goStringToCString(html)
	c_webview_set_html(w.handle, ptr)
	runtime.KeepAlive(b)
}

func (w *webview) Init(js string) {
	b, ptr := goStringToCString(js)
	c_webview_init(w.handle, ptr)
	runtime.KeepAlive(b)
}

func (w *webview) Eval(js string) {
	b, ptr := goStringToCString(js)
	c_webview_eval(w.handle, ptr)
	runtime.KeepAlive(b)
}

// Bind registers a Go function for JS calls via "webview_bind" in the native lib.
func (w *webview) Bind(name string, f any) error {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return errors.New("Bind error: only functions can be bound")
	}
	outCount := v.Type().NumOut()
	if outCount > 2 {
		return errors.New("Bind error: function may only return (value), (error), or (value, error)")
	}
	bindMu.Lock()
	if _, exists := boundNames[name]; exists {
		bindMu.Unlock()
		return errors.New("Bind error: name already bound")
	}

	// Build a wrapper that decodes JSON, calls v.Call(...), returns JSON or error
	funcType := v.Type()
	isVariadic := funcType.IsVariadic()
	numIn := funcType.NumIn()
	wrapper := func(id, req string) (any, error) {
		var rawArgs []json.RawMessage
		if err := json.Unmarshal([]byte(req), &rawArgs); err != nil {
			return nil, err
		}
		if (!isVariadic && len(rawArgs) != numIn) || (isVariadic && len(rawArgs) < numIn-1) {
			return nil, errors.New("argument count mismatch")
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

		// interpret results: (value?), (error?)
		var ret any
		var err error
		switch outCount {
		case 0:
			ret, err = nil, nil
		case 1:
			// could be just a value, or just an error
			if funcType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				// it's an error
				if e := results[0].Interface(); e != nil {
					err = e.(error)
				}
			} else {
				// it's a value
				ret = results[0].Interface()
			}
		case 2:
			ret = results[0].Interface()
			if e := results[1].Interface(); e != nil {
				err = e.(error)
			}
		}
		return ret, err
	}

	// We store the wrapper in bindingMap
	var contextKey uintptr
	// If your static library calls c_malloc or c_free, you can try them here.
	// Otherwise, just use an incrementing counter:
	contextKey = bindingCounter
	bindingCounter++

	bindingMap[contextKey] = bindingEntry{w: w.handle, fn: wrapper}
	boundNames[name] = contextKey
	bindMu.Unlock()

	bb, namePtr := goStringToCString(name)
	// We'll pass zero for `bindingCallbackPtr`, so the library won't actually call back
	// unless it gracefully handles a NULL function pointer.
	c_webview_bind(w.handle, namePtr, bindingCallbackPtr, contextKey)
	runtime.KeepAlive(bb)
	return nil
}

func (w *webview) Unbind(name string) error {
	bindMu.Lock()
	ctx, ok := boundNames[name]
	if !ok {
		bindMu.Unlock()
		return errors.New("Unbind error: name not found")
	}
	delete(boundNames, name)
	delete(bindingMap, ctx)
	bindMu.Unlock()

	bb, namePtr := goStringToCString(name)
	c_webview_unbind(w.handle, namePtr)
	runtime.KeepAlive(bb)
	return nil
}

//-------------------------------------------------------------------
// Helper functions to handle string conversions in a no-cgo environment.
//-------------------------------------------------------------------

func goStringToCString(s string) ([]byte, uintptr) {
	b := append([]byte(s), 0) // NUL-terminate
	return b, uintptr(unsafe.Pointer(&b[0]))
}

func cStringToGo(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	// find the length
	var length int
	for {
		if *(*byte)(unsafe.Pointer(ptr + uintptr(length))) == 0 {
			break
		}
		length++
	}
	if length == 0 {
		return ""
	}
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = *(*byte)(unsafe.Pointer(ptr + uintptr(i)))
	}
	return string(buf)
}
