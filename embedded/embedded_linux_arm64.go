package embedded

import _ "embed"

const name = "webview.so"

//go:embed include/webview_linux_arm64.so
var lib []byte
