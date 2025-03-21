package webview_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/abemedia/webview"
	_ "github.com/abemedia/webview/embedded"
	"golang.design/x/mainthread"
)

// Needed to ensure that the tests run on the main thread.
func TestMain(m *testing.M) {
	mainthread.Init(func() {
		os.Exit(m.Run())
	})
}

func TestWebview(t *testing.T) {
	mainthread.Call(func() {
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
			if runtime.GOOS == "windows" {
				w.Dispatch(w.Terminate)
				w.Dispatch(w.Destroy)
			} else {
				w.Terminate()
			}
		})
		if err != nil {
			t.Fatal(err)
		}

		w.Run()
		if runtime.GOOS != "windows" {
			w.Destroy()
		}

		if !got {
			t.Fatal("got is false; want true")
		}
	})
}
