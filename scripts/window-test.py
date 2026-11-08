#!/usr/bin/env python3
"""Launch the gallery and ask the user (via a Windows toast) to run the
window-drag / ghost-window checks manually.

Why manual: edge-dragging is a precise, timing-sensitive interaction. Synthetic
mouse_event sequences are unreliable and can move the wrong HWND. This script
starts the real build, then sends a native Windows notification with the exact
checks and the information the developer needs back.

Usage:
    python scripts/window-test.py
    python scripts/window-test.py --no-launch
"""

from __future__ import annotations

import argparse
import os
import platform
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

NOTES = """\
# Core UI 窗口测试记录

## 本次构建已做的修复

- 系统标题栏/边框跟随 `ui.Theme` 变深/变浅（DWM immersive dark + caption/border colors）。
- 系统边框窗口不再临时加/删 `WS_EX_LAYERED`（避免 DWM 重建边框留下空白副本）。
- 纯 Go 层默认使用 Gio 风格 bitblt 交换链（DISCARD / 1 buffer / STRETCH），
  窗口拖动/缩放时更稳；`COREUI_FLIP_SWAPCHAIN=1` 可切回 flip 对比。
- `build-app.ps1` 默认 GUI 子系统；控制台子系统运行时也会隐藏“Windows 为本进程新建的”控制台。
  因此如果这次仍出现空白窗口，重点确认它是不是 `conhost.exe` / `ConsoleWindowClass`。

## 环境
- Windows: {windows}
- Python: {python}
- 构建: {exe}
- 合成模式: {mode}
- DPI/缩放: 请在 PowerShell 运行
  `Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.Screen]::PrimaryScreen.Bounds`
  或直接记录 Windows 设置里的缩放百分比。

## 测试步骤

1. 打开 gallery，切到左侧「窗口与主题」页。
2. 点「切换深/浅色」，观察窗口标题栏和外边框是否跟着变。
   - 深色 App + 浅色 Windows 时，标题栏不应保持白色。
3. 拖动标题栏左右移动；再拖动窗口左/右/上/下边缘改变大小。
   - 观察是否出现白色条/白边/白底（拖动的瞬间也算）。
   - 记录白边位置：标题栏 / 1px 外框 / 客户区内侧 / 窗口外侧。
4. 关掉 gallery，重新双击 `dist\\gallery.exe` 启动若干次（至少 3~5 次）。
   - 是否出现一个额外的空白「经典/Aero」窗口？
   - 它出现在主窗口旁边、后面还是屏幕角落？
   - 它有标题文字吗？有图标吗？最小化/最大化/关闭按钮可用吗？
   - 它会不会自己消失？大概多久？
   - 出现时打开任务管理器：多出来的是 `conhost.exe` / 控制台窗口，还是
     `gallery.exe` 自己的第二个窗口？
5. 如果还出现白边或空白窗口，请截图（整屏，不要只截窗口），并保留：
   - 进程名和 PID
   - 是否从资源管理器双击启动，还是从终端启动
   - 当时是深色还是浅色主题
   - 任务管理器里多出来的进程名（`conhost.exe` / `gallery.exe` / 其它）

## 需要回报给开发者的信息

- Windows 版本/内部版本号（`winver`）和显示缩放。
- 启动方式：资源管理器双击 / PowerShell / 终端 / 脚本。
- 白边：是否仍出现、位置、深色/浅色哪种主题、拖动还是缩放时出现。
- 空白经典窗口：是否仍出现、截图、出现时机、任务管理器里的进程。
- 复现概率（几次里出现一次）。

## 相关文件

- 程序：`dist\\gallery.exe`
- 日志/截图建议放：`dist\\`
- 生成通知的脚本：`scripts\\notify.py`、`scripts\\window-test.py`
"""


def send_toast(title: str, lines: list[str]) -> None:
    from notify import notify
    notify(title, lines)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--no-launch", action="store_true",
                        help="only write notes and send the toast")
    parser.add_argument("--exe", default=str(ROOT / "dist" / "gallery.exe"))
    parser.add_argument("--flip", action="store_true",
                        help="launch with COREUI_FLIP_SWAPCHAIN=1 (core-ui flip model)")
    parser.add_argument("--bitblt", action="store_true",
                        help="compatibility flag; bitblt is now the default")
    args = parser.parse_args(argv)

    exe = Path(args.exe)
    env = os.environ.copy()
    flip = args.flip and not args.bitblt
    if flip:
        env["COREUI_FLIP_SWAPCHAIN"] = "1"
    else:
        env.pop("COREUI_FLIP_SWAPCHAIN", None)
    mode = "flip (FLIP_SEQUENTIAL)" if flip else "bitblt (Gio 风格 COPY/DISCARD)"
    if not args.no_launch and exe.is_file():
        # Detached: the test script should exit immediately and leave the app
        # running so the user can drag it.
        subprocess.Popen([str(exe)], cwd=str(exe.parent), env=env,
                         creationflags=getattr(subprocess, "DETACHED_PROCESS", 0))
    elif not args.no_launch:
        print(f"window-test: missing {exe}; build it first with "
              f"scripts/build-app.ps1 -Target ./examples/gallery -Out dist/gallery.exe",
              file=sys.stderr)

    notes = NOTES.format(
        windows=platform.platform(),
        python=sys.version.split()[0],
        exe=exe,
        mode=mode,
    )
    notes_path = ROOT / "dist" / "window-test-notes.md"
    notes_path.parent.mkdir(parents=True, exist_ok=True)
    notes_path.write_text(notes, encoding="utf-8")

    try:
        send_toast(
            "Core UI 窗口测试（请手动拖动）",
            [
                f"当前合成模式：{mode}。请拖动标题栏和左/右边缘（拖远一点），观察白边/旧帧。",
                "再重启 3~5 次，看是否还有空白 Aero 窗口；截图并记录多出来的进程名。",
                f"详细步骤与需要的信息：{notes_path}",
            ],
        )
    except Exception as exc:  # noqa: BLE001 - CLI boundary
        print(f"window-test: toast failed: {exc}", file=sys.stderr)

    print(notes)
    print(f"\nNotes written to: {notes_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
