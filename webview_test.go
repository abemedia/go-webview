package webview_test

import (
	"os"
	"testing"

	"github.com/abemedia/webview"
	"golang.design/x/mainthread"
)

// TestMain is needed to run tests in the main thread.
func TestMain(m *testing.M) {
	mainthread.Init(func() { os.Exit(m.Run()) })
}

func TestWebview(t *testing.T) {
	mainthread.Call(func() {
		var got bool

		w := webview.New(true)
		if w == nil {
			t.Fatal("failed to create webview")
		}

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
		})
		if err != nil {
			t.Fatal(err)
		}

		w.Run()
		w.Destroy()

		if !got {
			t.Fatal("got is false; want true")
		}
	})
}
