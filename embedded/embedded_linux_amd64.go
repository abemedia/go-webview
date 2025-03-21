package embedded

import _ "embed"

const name = "webview.so"

//go:embed linux_amd64/libwebview.so
var lib []byte
