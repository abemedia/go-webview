//go:build linux && amd64
// +build linux,amd64

#include "textflag.h"

// c_webview_create(debug, wnd) => returns uintptr
TEXT ·c_webview_create(SB), NOSPLIT, $0-16
    MOVQ DI, RDI         // 1st arg
    MOVQ SI, RSI         // 2nd arg
    CALL webview_create(SB)
    RET

TEXT ·c_webview_destroy(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL webview_destroy(SB)
    RET

TEXT ·c_webview_run(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL webview_run(SB)
    RET

TEXT ·c_webview_terminate(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL webview_terminate(SB)
    RET

TEXT ·c_webview_dispatch(SB), NOSPLIT, $0-24
    MOVQ DI, RDI  // handle
    MOVQ SI, RSI  // callback fn ptr
    MOVQ DX, RDX  // user data
    CALL webview_dispatch(SB)
    RET

TEXT ·c_webview_get_window(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL webview_get_window(SB)
    RET

TEXT ·c_webview_set_title(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL webview_set_title(SB)
    RET

TEXT ·c_webview_set_size(SB), NOSPLIT, $0-32
    MOVQ DI, RDI  // handle
    MOVQ SI, RSI  // width
    MOVQ DX, RDX  // height
    MOVQ CX, RCX  // hint
    CALL webview_set_size(SB)
    RET

TEXT ·c_webview_navigate(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL webview_navigate(SB)
    RET

TEXT ·c_webview_set_html(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL webview_set_html(SB)
    RET

TEXT ·c_webview_init(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL webview_init(SB)
    RET

TEXT ·c_webview_eval(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL webview_eval(SB)
    RET

TEXT ·c_webview_bind(SB), NOSPLIT, $0-32
    MOVQ DI, RDI  // handle
    MOVQ SI, RSI  // name
    MOVQ DX, RDX  // fnCallback
    MOVQ CX, RCX  // userData
    CALL webview_bind(SB)
    RET

TEXT ·c_webview_unbind(SB), NOSPLIT, $0-16
    MOVQ DI, RDI
    MOVQ SI, RSI
    CALL webview_unbind(SB)
    RET

TEXT ·c_webview_return(SB), NOSPLIT, $0-32
    MOVQ DI, RDI  // handle
    MOVQ SI, RSI  // seq
    MOVQ DX, RDX  // status
    MOVQ CX, RCX  // result
    CALL webview_return(SB)
    RET

// Optional if you need a c_malloc / c_free
TEXT ·c_malloc(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL malloc(SB)
    RET

TEXT ·c_free(SB), NOSPLIT, $0-8
    MOVQ DI, RDI
    CALL free(SB)
    RET
