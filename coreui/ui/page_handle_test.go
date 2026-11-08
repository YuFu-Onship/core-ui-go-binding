package ui

import (
	"fmt"
	"testing"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
)

// Regression: ui_page_root() used to insert a fresh strong reference into the
// global handle table on every call. App.render() calls it on every Rebuild
// (via SetRoot), so each add/delete in a list pinned the entire previous
// widget tree until process shutdown. Keep exactly one handle per page and
// release it in ui_page_destroy().
func TestPageRebuildReleasesRootHandles(t *testing.T) {
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	var rows string
	for i := 0; i < 20; i++ {
		rows += fmt.Sprintf("<div><text>row %d</text><button>x</button></div>", i)
	}
	src := "<template><div>" + rows + "</div></template>"

	base := coreui.DebugHandleCount()
	for i := 0; i < 30; i++ {
		p, err := coreui.LoadPageString(src)
		if err != nil {
			t.Fatalf("load %d: %v", i, err)
		}
		h1 := p.Root().Handle()
		if h1 == 0 {
			t.Fatalf("root handle missing at iteration %d", i)
		}
		if h2 := p.Root().Handle(); h2 != h1 {
			t.Fatalf("ui_page_root must reuse the handle: first=%d second=%d", h1, h2)
		}
		p.Destroy()
		if n := coreui.DebugHandleCount(); n > base {
			t.Fatalf("page %d leaked handles: base=%d now=%d", i, base, n)
		}
	}
}
