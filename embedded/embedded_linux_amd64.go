package embedded

import _ "embed"

const name = "webview.so"

//go:embed include/webview_linux_amd64.so
var lib []byte
