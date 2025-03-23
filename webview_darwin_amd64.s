//go:build darwin && amd64
// +build darwin,amd64

#include "textflag.h"

// On macOS AMD64, the calling convention is similar to Linux's SysV,
// but typically each symbol has a leading underscore if you inspect with `nm`.

TEXT ·c_webview_create(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    // If the real symbol is `_webview_create`, CALL _webview_create(SB)
    // If it's just `webview_create`, remove the underscore.
    CALL _webview_create(SB)
    RET

TEXT ·c_webview_destroy(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL _webview_destroy(SB)
    RET

TEXT ·c_webview_run(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL _webview_run(SB)
    RET

TEXT ·c_webview_terminate(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL _webview_terminate(SB)
    RET

TEXT ·c_webview_dispatch(SB), NOSPLIT, $0-24
    MOVQ DI, RDI
    MOVQ SI, RSI
    MOVQ DX, RDX
    CALL _webview_dispatch(SB)
    RET

TEXT ·c_webview_get_window(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL _webview_get_window(SB)
    RET

TEXT ·c_webview_set_title(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL _webview_set_title(SB)
    RET

TEXT ·c_webview_set_size(SB), NOSPLIT, $0-32
    MOVQ DI, RDI
    MOVQ SI, RSI
    MOVQ DX, RDX
    MOVQ CX, RCX
    CALL _webview_set_size(SB)
    RET

TEXT ·c_webview_navigate(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL _webview_navigate(SB)
    RET

TEXT ·c_webview_set_html(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL _webview_set_html(SB)
    RET

TEXT ·c_webview_init(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL _webview_init(SB)
    RET

TEXT ·c_webview_eval(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL _webview_eval(SB)
    RET

TEXT ·c_webview_bind(SB), NOSPLIT, $0-32
    MOVQ DI, RDI
    MOVQ SI, RSI
    MOVQ DX, RDX
    MOVQ CX, RCX
    CALL _webview_bind(SB)
    RET

TEXT ·c_webview_unbind(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL _webview_unbind(SB)
    RET

TEXT ·c_webview_return(SB), NOSPLIT, $0-32
    MOVQ DI, RDI
    MOVQ SI, RSI
    MOVQ DX, RDX
    MOVQ CX, RCX
    CALL _webview_return(SB)
    RET

TEXT ·c_malloc(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL _malloc(SB)
    RET

TEXT ·c_free(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL _free(SB)
    RET
