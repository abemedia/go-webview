package webview

import "syscall"

func load() {
	libHandle, err := syscall.LoadLibrary("webview.dll")
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
	for _, lib := range []string{"msvcrt.dll", "ucrtbase.dll"} {
		if handle, err := syscall.LoadLibrary(lib); err == nil {
			cMalloc, _ = syscall.GetProcAddress(handle, "malloc")
			// Do not Dlclose here; keep CRT loaded
			break
		}
	}
}

// mustLoadSymbol looks up a symbol and panics if not found.
func mustLoadSymbol(lib syscall.Handle, name string) uintptr {
	ptr, err := syscall.GetProcAddress(lib, name)
	if err != nil || ptr == 0 {
		panic("webview: failed to load symbol " + name + ": " + err.Error())
	}
	return ptr
}
