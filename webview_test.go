package webview_test

import (
	"testing"

	"github.com/abemedia/webview"
	_ "github.com/abemedia/webview/embedded"
)

func TestWebview(t *testing.T) {
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
				window.onload = function() { run(); }
			</script>
		</html>
	)`)

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
}
