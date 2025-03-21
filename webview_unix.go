//go:build darwin || linux
// +build darwin linux

package webview

import (
	_ "embed"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ebitengine/purego"
)

func load() {
	loadOnce.Do(func() {
		var name string
		var paths []string
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)
		switch runtime.GOOS {
		case "linux":
			name = "libwebview.so"
			paths = []string{
				os.Getenv("WEBVIEW_PATH"),
				execDir,
			}
		case "darwin":
			name = "libwebview.dylib"
			paths = []string{
				os.Getenv("WEBVIEW_PATH"),
				execDir,
				filepath.Join(execDir, "..", "Frameworks"),
			}
		}

		fname := "libwebview.dylib"
		for _, v := range paths {
			fn := filepath.Join(v, name)
			if _, err := os.Stat(fn); err == nil {
				fname = fn
				break
			}
		}

		libHandle, err := purego.Dlopen(fname, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
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
