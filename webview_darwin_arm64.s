//go:build darwin && arm64
// +build darwin,arm64

#include "textflag.h"

// macOS ARM64 uses a calling convention similar to iOS arm64, typically with underscore prefixes.
// The first 8 int/pointer params in x0..x7, return in x0.

// c_webview_create(debug, wnd)
TEXT ·c_webview_create(SB), NOSPLIT, $0-16
    // If nm shows the symbol as `___webview_create` or something else, adapt as needed.
    BL _webview_create<>(SB)
    RET

TEXT ·c_webview_destroy(SB), NOSPLIT, $0-8
    BL _webview_destroy<>(SB)
    RET

TEXT ·c_webview_run(SB), NOSPLIT, $0-8
    BL _webview_run<>(SB)
    RET

TEXT ·c_webview_terminate(SB), NOSPLIT, $0-8
    BL _webview_terminate<>(SB)
    RET

TEXT ·c_webview_dispatch(SB), NOSPLIT, $0-24
    BL _webview_dispatch<>(SB)
    RET

TEXT ·c_webview_get_window(SB), NOSPLIT, $0-8
    BL _webview_get_window<>(SB)
    RET

TEXT ·c_webview_set_title(SB), NOSPLIT, $0-16
    BL _webview_set_title<>(SB)
    RET

TEXT ·c_webview_set_size(SB), NOSPLIT, $0-32
    BL _webview_set_size<>(SB)
    RET

TEXT ·c_webview_navigate(SB), NOSPLIT, $0-16
    BL _webview_navigate<>(SB)
    RET

TEXT ·c_webview_set_html(SB), NOSPLIT, $0-16
    BL _webview_set_html<>(SB)
    RET

TEXT ·c_webview_init(SB), NOSPLIT, $0-16
    BL _webview_init<>(SB)
    RET

TEXT ·c_webview_eval(SB), NOSPLIT, $0-16
    BL _webview_eval<>(SB)
    RET

TEXT ·c_webview_bind(SB), NOSPLIT, $0-32
    BL _webview_bind<>(SB)
    RET

TEXT ·c_webview_unbind(SB), NOSPLIT, $0-16
    BL _webview_unbind<>(SB)
    RET

TEXT ·c_webview_return(SB), NOSPLIT, $0-32
    BL _webview_return<>(SB)
    RET

TEXT ·c_malloc(SB), NOSPLIT, $0-8
    BL _malloc<>(SB)
    RET

TEXT ·c_free(SB), NOSPLIT, $0-8
    BL _free<>(SB)
    RET
