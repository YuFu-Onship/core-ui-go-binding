/* cgo 导出的 Go 回调原型。每个 cgo 文件的 preamble 互相独立，
   所以统一放在这个头里，由需要的文件 include。 */
#ifndef COREUI_GO_CALLBACKS_H
#define COREUI_GO_CALLBACKS_H

#include <ui_core.h>

#ifdef __cplusplus
extern "C" {
#endif

void goCoreUiClick(UiWidget w, void *userdata);
void goCoreUiWindowClose(UiWindow w, void *userdata);
int goCoreUiWindowCloseRequest(UiWindow w, void *userdata);
void goCoreUiPageWatch(UiPage p, char *name, char *json, void *userdata);
void goCoreUiInvoke(void *userdata);

/* goCoreUiPageWatch 的形参是 cgo 生成的 char*，而 UiPageWatchCallback 要求
   const char*；这个 static inline 负责适配签名。 */
static inline void coreui_page_watch_cb(UiPage p, const char *name,
                                        const char *json, void *userdata) {
  goCoreUiPageWatch(p, (char *)name, (char *)json, userdata);
}

/* cgo 里直接给 C 函数指针赋值很啰嗦，这两个 static inline 把导出函数
   和真正的 C API 粘在一起。 */
static inline void coreui_page_watch(UiPage p, const char *name,
                                     void *userdata) {
  ui_page_watch(p, name, coreui_page_watch_cb, userdata);
}

static inline void coreui_invoke_sync(UiWindow win, void *userdata) {
  ui_window_invoke_sync(win, goCoreUiInvoke, userdata);
}

#ifdef __cplusplus
}
#endif

#endif /* COREUI_GO_CALLBACKS_H */
