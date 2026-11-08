# Core UI Go Gallery

纯 Go 声明式 UI 的用法画廊。左侧切换页面，底部“设置”集中放置主题、窗口和信息。

## 运行

```powershell
scripts/build-app.ps1 -Target ./examples/gallery -Out dist/gallery.exe
dist\gallery.exe
```

## 页面

| 页面 | 内容 |
|------|------|
| 总览 | 当前时间，每秒更新 |
| 布局 | Row / Column / flex / padding / gap / 对齐 / Card |
| 文本与按钮 | Text、Primary / 普通 / 禁用按钮、状态绑定 |
| 输入控件 | Input、Password、TextArea、Check、Toggle、Slider、Progress、搜索 |
| 选择与条件 | Select、分段按钮、WhenInt、When / Unless、v-show |
| 列表与滚动 | ScrollView、Each、添加/批量添加/删除、Rebuild |
| 事件与诊断 | OnClick、OnDblClick、OnChange、OnEvent("wheel") |
| 颜色 Token | Fg / Bg / Card / Accent、color.RGBA、透明色 |
| 设置 | 明暗主题、最大化/还原/最小化、SetTitle、DPI、DLL 版本 |

页面切换使用一个 `IntState` + `ui.WhenInt`；只有列表结构变化时才调用
`app.Rebuild()`。

## 窗口合成模式

默认使用 bitblt（DISCARD / 1 buffer / STRETCH）。需要对比 core-ui 原有的
flip 模式时：

```powershell
$env:COREUI_FLIP_SWAPCHAIN = "1"
dist\gallery.exe
```

当前模式可在“设置 → 信息”中查看。

## 已知边界

- 列表增删使用整页 `Rebuild`，会重建页面，滚动位置回到顶部。
- ChipsInput 尚未接入内联 IME 预编辑。
- 原生最小化/最大化/还原动画是否播放取决于 Windows 系统设置。
