<#
.SYNOPSIS
    用 go build 编译一个 core-ui 应用，并把 core-ui.dll 复制到 exe 旁边。

.DESCRIPTION
    cgo 链接时用的是 third_party/coreui/lib/libcore-ui.dll.a，运行时 exe 仍然
    需要同目录（或 PATH 里）的 core-ui.dll。

.PARAMETER Target
    Go 包路径，默认 examples/hello。

.PARAMETER Out
    输出 exe 路径，默认 dist/hello.exe。

.PARAMETER Gui
    兼容旧参数；现在默认就是 GUI 子系统。

.PARAMETER Console
    保留控制台子系统（调试输出用）。不加时默认 -H=windowsgui，
    避免双击运行额外弹出一个空的控制台窗口。
#>
param(
    [string]$Target = './examples/hello',
    [string]$Out = 'dist/hello.exe',
    [switch]$Gui,
    [switch]$Console
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$dll = Join-Path $root 'third_party/coreui/bin/core-ui.dll'
if (-not (Test-Path $dll)) {
    throw "缺少 $dll，先运行 scripts/build-core-ui.ps1"
}

$outPath = Join-Path $root $Out
$outDir = Split-Path -Parent $outPath
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

Push-Location $root
try {
    Write-Host "==> go build $Target -> $Out"
    if ($Console) {
        & go build -o $outPath $Target
    } else {
        & go build -ldflags '-H=windowsgui' -o $outPath $Target
    }
    if ($LASTEXITCODE -ne 0) { throw 'go build 失败' }
} finally {
    Pop-Location
}

Copy-Item $dll (Join-Path $outDir 'core-ui.dll') -Force
Write-Host "==> 完成：$outPath（同目录已放 core-ui.dll）"
