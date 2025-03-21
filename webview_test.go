package webview_test

import (
	"testing"

	"github.com/abemedia/webview"
	_ "github.com/abemedia/webview/embedded"
)

func TestWebview(t *testing.T) {
	var got bool

	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("Hello")
	w.Bind("run", func() {
		got = true
	})
	w.Bind("quit", func() {
		w.Terminate()
	})
	w.SetHtml(`<!doctype html>
		<html>
			<body>hello</body>
			<script>
				window.onload = function() {
					run();
					quit();
				};
			</script>
		</html>
	)`)
	w.Run()

	if !got {
		t.Fatal("got is false; want true")
	}
}
