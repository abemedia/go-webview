//go:build windows && amd64
// +build windows,amd64

#include "textflag.h"

// Windows x64 uses the "Microsoft x64" calling convention:
//   - RCX, RDX, R8, R9 for the first four arguments
//   - Return in RAX
// Symbol naming typically has no underscore prefix unless your .lib/.a is built otherwise.
// If your library uses e.g. `webview_create@8`, you'd have to call that exact label.

TEXT ·c_webview_create(SB), NOSPLIT, $0-16
    // The first argument from Go is in RCX, second in RDX, etc. 
    CALL webview_create(SB)  // or "CALL webview_create@16" if that's how it is exported
    RET

TEXT ·c_webview_destroy(SB), NOSPLIT, $0-8
    CALL webview_destroy(SB)
    RET

TEXT ·c_webview_run(SB), NOSPLIT, $0-8
    CALL webview_run(SB)
    RET

TEXT ·c_webview_terminate(SB), NOSPLIT, $0-8
    CALL webview_terminate(SB)
    RET

TEXT ·c_webview_dispatch(SB), NOSPLIT, $0-24
    CALL webview_dispatch(SB)
    RET

TEXT ·c_webview_get_window(SB), NOSPLIT, $0-8
    CALL webview_get_window(SB)
    RET

TEXT ·c_webview_set_title(SB), NOSPLIT, $0-16
    CALL webview_set_title(SB)
    RET

TEXT ·c_webview_set_size(SB), NOSPLIT, $0-32
    CALL webview_set_size(SB)
    RET

TEXT ·c_webview_navigate(SB), NOSPLIT, $0-16
    CALL webview_navigate(SB)
    RET

TEXT ·c_webview_set_html(SB), NOSPLIT, $0-16
    CALL webview_set_html(SB)
    RET

TEXT ·c_webview_init(SB), NOSPLIT, $0-16
    CALL webview_init(SB)
    RET

TEXT ·c_webview_eval(SB), NOSPLIT, $0-16
    CALL webview_eval(SB)
    RET

TEXT ·c_webview_bind(SB), NOSPLIT, $0-32
    CALL webview_bind(SB)
    RET

TEXT ·c_webview_unbind(SB), NOSPLIT, $0-16
    CALL webview_unbind(SB)
    RET

TEXT ·c_webview_return(SB), NOSPLIT, $0-32
    CALL webview_return(SB)
    RET

TEXT ·c_malloc(SB), NOSPLIT, $0-8
    CALL malloc(SB)
    RET

TEXT ·c_free(SB), NOSPLIT, $0-8
    CALL free(SB)
    RET
