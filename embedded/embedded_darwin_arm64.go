package embedded

import _ "embed"

const name = "webview.dylib"

//go:embed darwin_arm64/libwebview.dylib
var lib []byte
