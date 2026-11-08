# Core UI Go Gallery

纯 Go 声明式界面层的用法画廊。每个左侧页面只演示一类能力，方便把问题缩小到
布局、输入、滚动、IME、窗口或颜色中的某一层，而不是在一个大 demo 里猜。

## 运行

```powershell
scripts/build-app.ps1 -Target ./examples/gallery -Out dist/gallery.exe
dist\gallery.exe
```

## 页面索引

| 页面 | 主要覆盖 | 适合排查 |
|------|----------|----------|
| 总览 | 全局状态、快速跳转 | 先确认问题是否只在某一页 |
| 布局 Layout | Row/Column、flex、padding/gap、对齐、Card、Divider | 尺寸/间距/挤压/铺满问题 |
| 文本与按钮 | H1-H3、Fg 层级、Primary/Disabled/Tooltip、OnDblClick | 文本渲染、按钮事件 |
| 输入控件 | Input、Password、TextArea、Check、Toggle、Slider、Progress、v-model、Enter 搜索 | 输入回写、Watch、字符长度、OnSubmit |
| 选择控件 | Select、分段按钮、WhenInt 分支 | 下拉选择、页内 v-if 切换 |
| 列表与滚动 | 两个独立 ScrollView、Each、Rebuild、滚动条 gutter | 滚动条位置、快速删除/Rebuild 崩溃 |
| 条件与 Rebuild | When/Unless/ShowWhen/HiddenWhen、状态持久化 | v-if vs v-show、重建后状态 |
| 事件与诊断 | OnClick/OnDblClick/OnChange/OnEvent("wheel")、原始参数 | 事件分发、JSON 参数、事件总线 |
| IME 输入法 | 内联 preedit、提交、取消、失焦后 IME context 检查、TextArea | 输入法黑框、候选窗位置、状态回写、点空白后仍弹候选窗 |
| 窗口与主题 | 明暗主题、Maximize/Restore/Minimize、SetTitle、DPI、flip/bitblt 合成模式 | 最大化布局、原生动画、主题刷新、拖动白边 |
| 颜色 Token | Fg/Bg/Card/Accent token、color.RGBA/NRGBA、RGB/RGBA | 颜色写错、透明色、token 查找 |

## 快速定位问题

1. **先切页面**：如果只在某一页出现，先看该页用到的 API，不要从整个 app 找。
2. **布局问题**：到「布局 Layout」对比 flex/padding/gap；再到「窗口与主题」最大化，
   确认布局是否重新铺满。
3. **输入/状态问题**：到「输入控件」检查 `v-model` 是否回写、`OnChange` 是否触发。
4. **搜索/回车提交**：到「输入控件」的「行内搜索」卡片，输入关键字后按 Enter，
   检查结果行是否更新；按钮「搜索」走同一个 Go 回调。
5. **滚动/列表问题**：到「列表与滚动」；
   - 滚动条应该在内容右侧的 gutter 内；
   - 快速连点「删除」用于复现 Rebuild / 旧控件指针问题；
   - 如果崩溃，记下右侧日志最后几条和操作顺序。
6. **IME 问题**：到「IME 输入法」；
   - 先用真实输入法输入；
   - 也可以点「聚焦输入框」后用「模拟预编辑」→「提交/取消」复现；
   - 预编辑阶段 Go 状态不应变化，提交后才更新一次。
7. **事件问题**：到「事件与诊断」查看 `lastEvent` / `lastArgs`，确认事件名和原始
   JSON 参数是否与预期一致。
8. **颜色问题**：到「颜色 Token」对照 token 值和 `image/color` 转换结果。

## 代码结构

- `main.go`：应用状态、页面路由、侧边栏、主题切换。
- `pages.go`：每个页面的视图构建函数。
- 页面切换使用一个 `IntState` + `ui.WhenInt`，不调用 `Rebuild`；
  只有列表结构变化（增删行）才调用 `app.Rebuild()`。

新增页面：

1. 在 `main.go` 的 `pageXxx` 常量里加索引；
2. 在 `pageTitles` 里加标题；
3. 在 `pages.go` 写 `func (g *gallery) xxxPage() *ui.Node`；
4. 在 `setupView()` 里加一条 `ui.WhenInt(g.page, pageXxx, g.xxxPage())`。

## 窗口合成模式

默认使用 Gio 风格的 bitblt（DISCARD / 1 buffer / STRETCH）。需要对比
core-ui 原来的 flip 交换链时：

```powershell
$env:COREUI_FLIP_SWAPCHAIN = "1"
dist\gallery.exe
```

「窗口与主题」页会显示当前是 `flip / FLIP_SEQUENTIAL` 还是
`bitblt / DISCARD（Gio 风格）`。用这一页对比拖动标题栏、拖动边缘缩放时
是否还有白边/旧帧。

## 已知边界

- 列表的增删仍然使用整页 `Rebuild`（状态值会保留，但列表滚动位置会回到顶部）。
  如果后续需要保留滚动位置，需要把列表改成 `v-for` + JSON state，或给 ScrollView
  增加滚动位置持久化。
- ChipsInput 还没有接入内联 IME 组合，使用它时仍会走系统默认组合窗。
- 最小化后需要从任务栏恢复；原生最小化/最大化/关闭动画遵循 Windows 的
  “在最小化和最大化时动画显示窗口”系统设置。
