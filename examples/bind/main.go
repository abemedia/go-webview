package main

import (
	"github.com/abemedia/go-webview"
	_ "github.com/abemedia/go-webview/embedded"
)

const html = `
<div>
  <button id="increment">+</button>
  <button id="decrement">-</button>
  <span>Counter: <span id="counterResult">0</span></span>
</div>
<script type="module">
  const getElements = ids => Object.assign({}, ...ids.map(id => ({ [id]: document.getElementById(id) })));
  const ui = getElements(["increment", "decrement", "counterResult"]);
  ui.increment.addEventListener("click", async () => {
    ui.counterResult.textContent = await window.count(1);
  });
  ui.decrement.addEventListener("click", async () => {
    ui.counterResult.textContent = await window.count(-1);
  });
</script>
`

func main() {
	var count int64

	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("Bind Example")
	w.SetSize(480, 320, webview.HintNone)

	// Synchronous binding for count
	err := w.Bind("count", func(delta int64) int64 {
		count += delta
		return count
	})
	if err != nil {
		panic(err)
	}

	w.SetHtml(html)
	w.Run()
}
