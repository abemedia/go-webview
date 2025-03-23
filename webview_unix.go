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
		var fname string
		var paths []string
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)
		switch runtime.GOOS {
		case "windows":
			fname = "webview.dll"
		case "linux":
			fname = "libwebview.so"
			paths = []string{
				os.Getenv("WEBVIEW_PATH"),
				execDir,
			}
		case "darwin":
			fname = "libwebview.dylib"
			paths = []string{
				os.Getenv("WEBVIEW_PATH"),
				execDir,
				filepath.Join(execDir, "..", "Frameworks"),
			}
		}

		for _, v := range paths {
			fn := filepath.Join(v, fname)
			if _, err := os.Stat(fn); err == nil {
				fname = fn
				break
			}
		}

		libHandle, err := purego.Dlopen(fname, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			panic("webview: failed to load native library: " + err.Error())
		}
		if libHandle == 0 {
			panic("webview: native library not loaded")
		}
		// Resolve all required symbols from the library
		purego.RegisterLibFunc(&webviewCreate, libHandle, "webview_create")
		purego.RegisterLibFunc(&webviewDestroy, libHandle, "webview_destroy")
		purego.RegisterLibFunc(&webviewRun, libHandle, "webview_run")
		purego.RegisterLibFunc(&webviewTerminate, libHandle, "webview_terminate")
		purego.RegisterLibFunc(&webviewGetWindow, libHandle, "webview_get_window")
		purego.RegisterLibFunc(&webviewNavigate, libHandle, "webview_navigate")
		purego.RegisterLibFunc(&webviewSetTitle, libHandle, "webview_set_title")
		purego.RegisterLibFunc(&webviewSetHtml, libHandle, "webview_dispatch")
		purego.RegisterLibFunc(&webviewDispatch, libHandle, "webview_dispatch")
		purego.RegisterLibFunc(&webviewSetSize, libHandle, "webview_set_size")
		purego.RegisterLibFunc(&webviewInit, libHandle, "webview_init")
		purego.RegisterLibFunc(&webviewEval, libHandle, "webview_eval")
		purego.RegisterLibFunc(&webviewBind, libHandle, "webview_bind")
		purego.RegisterLibFunc(&webviewUnbind, libHandle, "webview_unbind")
		purego.RegisterLibFunc(&webviewReturn, libHandle, "webview_return")

		// Attempt to load standard malloc/free for context allocation
		cMalloc, _ = purego.Dlsym(purego.RTLD_DEFAULT, "malloc")

		// Create C-callable callbacks for dispatch and binding
		dispatchCallbackPtr = purego.NewCallback(dispatchCallback)
		bindingCallbackPtr = purego.NewCallback(bindingCallback)
	})
}
