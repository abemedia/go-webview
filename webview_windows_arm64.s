//go:build windows && arm64
// +build windows,arm64

#include "textflag.h"

// Windows on ARM64 uses a different calling convention, 
// generally the first 8 integer args in x0..x7, return in x0, 
// but with some MS-specific details. 
// We'll assume the library exports e.g. "webview_create", no leading underscore.

TEXT ·c_webview_create(SB), NOSPLIT, $0-16
    BL webview_create<>(SB)
    RET

TEXT ·c_webview_destroy(SB), NOSPLIT, $0-8
    BL webview_destroy<>(SB)
    RET

TEXT ·c_webview_run(SB), NOSPLIT, $0-8
    BL webview_run<>(SB)
    RET

TEXT ·c_webview_terminate(SB), NOSPLIT, $0-8
    BL webview_terminate<>(SB)
    RET

TEXT ·c_webview_dispatch(SB), NOSPLIT, $0-24
    BL webview_dispatch<>(SB)
    RET

TEXT ·c_webview_get_window(SB), NOSPLIT, $0-8
    BL webview_get_window<>(SB)
    RET

TEXT ·c_webview_set_title(SB), NOSPLIT, $0-16
    BL webview_set_title<>(SB)
    RET

TEXT ·c_webview_set_size(SB), NOSPLIT, $0-32
    BL webview_set_size<>(SB)
    RET

TEXT ·c_webview_navigate(SB), NOSPLIT, $0-16
    BL webview_navigate<>(SB)
    RET

TEXT ·c_webview_set_html(SB), NOSPLIT, $0-16
    BL webview_set_html<>(SB)
    RET

TEXT ·c_webview_init(SB), NOSPLIT, $0-16
    BL webview_init<>(SB)
    RET

TEXT ·c_webview_eval(SB), NOSPLIT, $0-16
    BL webview_eval<>(SB)
    RET

TEXT ·c_webview_bind(SB), NOSPLIT, $0-32
    BL webview_bind<>(SB)
    RET

TEXT ·c_webview_unbind(SB), NOSPLIT, $0-16
    BL webview_unbind<>(SB)
    RET

TEXT ·c_webview_return(SB), NOSPLIT, $0-32
    BL webview_return<>(SB)
    RET

TEXT ·c_malloc(SB), NOSPLIT, $0-8
    BL malloc<>(SB)
    RET

TEXT ·c_free(SB), NOSPLIT, $0-8
    BL free<>(SB)
    RET
