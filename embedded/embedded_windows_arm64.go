package embedded

import _ "embed"

const name = "webview.dll"

//go:embed include/webview_windows_arm64.dll
var lib []byte
