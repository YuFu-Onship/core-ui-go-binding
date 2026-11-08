#!/usr/bin/env python3
"""Send a Windows toast notification without third-party Python packages.

The toast is rendered by the Windows Runtime notification API. Python writes a
small PowerShell helper (Windows PowerShell 5.1 can load WinRT types directly),
runs it, and deletes it. No pip install is required.

Examples:
    python scripts/notify.py --title "Core UI" --line "请拖动窗口测试"
    python scripts/notify.py --title "Core UI" --line "第一行" --line "第二行"

Exit code is non-zero if the toast could not be delivered; the error text is
printed to stderr.
"""

from __future__ import annotations

import argparse
import os
import subprocess
import sys
import tempfile
from typing import Sequence


def _ps_quote(value: str) -> str:
    """Single-quoted PowerShell string literal."""
    return "'" + value.replace("'", "''") + "'"


def _build_ps1(title: str, lines: Sequence[str], app_id: str) -> str:
    padded = list(lines[:3]) + [""] * (3 - min(len(lines), 3))
    return f"""\
$ErrorActionPreference = 'Stop'
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$template = [Windows.UI.Notifications.ToastTemplateType]::ToastText04
$xml = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent($template)
$nodes = $xml.GetElementsByTagName('text')
$values = @({_ps_quote(title)}, {_ps_quote(padded[0])}, {_ps_quote(padded[1])}, {_ps_quote(padded[2])})
for ($i = 0; $i -lt $values.Length; $i++) {{
    if ($values[$i].Length -gt 0) {{
        $nodes.Item($i).AppendChild($xml.CreateTextNode($values[$i])) | Out-Null
    }}
}}
$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier({_ps_quote(app_id)})
$notifier.Show($toast)
"""


def notify(title: str, lines: Sequence[str], app_id: str = "Core UI Go") -> None:
    if sys.platform != "win32":
        raise RuntimeError("Windows toast notifications are only available on Windows")

    script = _build_ps1(title, lines, app_id)
    fd, path = tempfile.mkstemp(prefix="coreui-toast-", suffix=".ps1",
                                dir=os.environ.get("TEMP") or None)
    os.close(fd)
    try:
        # PowerShell 5.1 needs a BOM to decode Chinese text in a .ps1 file.
        with open(path, "w", encoding="utf-8-sig", newline="\r\n") as f:
            f.write(script)
        result = subprocess.run(
            ["powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass",
             "-File", path],
            capture_output=True, text=True, timeout=30,
        )
        if result.returncode != 0:
            detail = (result.stderr or result.stdout or "").strip()
            raise RuntimeError(f"PowerShell toast failed: {detail}")
    finally:
        try:
            os.unlink(path)
        except OSError:
            pass


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Send a Windows toast notification")
    parser.add_argument("--title", default="Core UI", help="notification title")
    parser.add_argument("--line", action="append", default=[],
                        help="body line (repeat up to 3 times)")
    parser.add_argument("--app-id", default="Core UI Go",
                        help="AUMID / notification sender name")
    args = parser.parse_args(argv)
    try:
        notify(args.title, args.line, args.app_id)
    except Exception as exc:  # noqa: BLE001 - CLI boundary
        print(f"notify: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
