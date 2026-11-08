param([string]$ProcessName = "hello", [string]$Out = "shot.png")
Add-Type -AssemblyName System.Drawing
if (-not ("WinCap" -as [type])) {
Add-Type @"
using System;
using System.Runtime.InteropServices;
public class WinCap {
  [StructLayout(LayoutKind.Sequential)] public struct RECT { public int Left, Top, Right, Bottom; }
  [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out RECT r);
  [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr h);
  [DllImport("user32.dll")] public static extern bool PrintWindow(IntPtr h, IntPtr dc, uint flags);
}
"@
}
$p = Get-Process $ProcessName -ErrorAction Stop
[WinCap]::SetForegroundWindow($p.MainWindowHandle) | Out-Null
Start-Sleep -Milliseconds 250
$r = New-Object WinCap+RECT
[WinCap]::GetWindowRect($p.MainWindowHandle, [ref]$r) | Out-Null
$w = $r.Right - $r.Left
$h = $r.Bottom - $r.Top
$bmp = New-Object System.Drawing.Bitmap $w, $h
$g = [System.Drawing.Graphics]::FromImage($bmp)

# 优先 PrintWindow：即使窗口被别的窗口遮挡也能拿到自己的画面。
# PW_RENDERFULLCONTENT(2) 会包含 DWM 合成内容（深色标题栏等）。
$dc = $g.GetHdc()
$ok = [WinCap]::PrintWindow($p.MainWindowHandle, $dc, 2)
$g.ReleaseHdc($dc)
if (-not $ok) {
    $g.CopyFromScreen($r.Left, $r.Top, 0, 0, $bmp.Size)
}
$bmp.Save($Out, [System.Drawing.Imaging.ImageFormat]::Png)
$g.Dispose()
$bmp.Dispose()
Write-Output "saved $Out ($w x $h)"
