package webview_test

import (
	"runtime"
	"testing"

	"github.com/abemedia/webview"
	_ "github.com/abemedia/webview/embedded"
)

// Needed to ensure that the tests run on the main thread.
func init() {
	runtime.UnlockOSThread()
}

func TestWebview(t *testing.T) {
	runtime.LockOSThread()

	var got bool

	w := webview.New(false)
	w.SetTitle("Hello")
	w.SetSize(800, 600, webview.HintNone)
	w.SetHtml(`<!doctype html>
		<html>
			<script>
				window.onload = function() { run(); };
			</script>
		</html>`)

	err := w.Bind("run", func() {
		got = true
		w.Terminate()
		w.Dispatch(w.Destroy)
	})
	if err != nil {
		t.Fatal(err)
	}

	w.Run()

	if !got {
		t.Fatal("got is false; want true")
	}
}
