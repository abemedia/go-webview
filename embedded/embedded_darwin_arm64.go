package embedded

import _ "embed"

const name = "webview.dylib"

//go:embed include/webview_darwin_arm64.dylib
var lib []byte
