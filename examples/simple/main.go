package main

import (
	"github.com/abemedia/webview"
	_ "github.com/abemedia/webview/embedded"
)

func main() {
	w := webview.New(true)
	w.SetTitle("Hello")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("https://google.com")
	w.Run()
	w.Destroy()
}
