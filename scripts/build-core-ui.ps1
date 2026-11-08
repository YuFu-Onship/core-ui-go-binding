<#
.SYNOPSIS
    构建 core-ui C++ 运行时，并把 cgo 需要的头文件 / 导入库 / DLL 复制到 third_party/coreui。

.DESCRIPTION
    core-ui 官方只支持 MSVC / clang-cl 工具链，所以这里用 Visual Studio 生成器编译
    core-ui.dll，再用 gendef + dlltool 生成一份 MinGW 能吃的导入库
    (libcore-ui.dll.a)，供 cgo 链接。

.PARAMETER Config
    Release（默认）或 Debug。

.PARAMETER Clean
    先删掉 build/core-ui 再配置。

.PARAMETER DynamicCRT
    用 /MD 动态 CRT（需要目标机器装 VC 运行时）。默认 /MT，交付更干净。
#>
param(
    [ValidateSet('Release', 'Debug')][string]$Config = 'Release',
    [switch]$Clean,
    [switch]$DynamicCRT
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$src = Join-Path $root 'core-ui'
$build = Join-Path $root 'build/core-ui'
$deps = Join-Path $root 'third_party/coreui'

if (-not (Test-Path (Join-Path $src 'CMakeLists.txt'))) {
    throw "找不到 core-ui 源码：$src"
}

function Get-VSGenerator {
    $vswhere = Join-Path ${env:ProgramFiles(x86)} 'Microsoft Visual Studio/Installer/vswhere.exe'
    if (-not (Test-Path $vswhere)) {
        throw 'vswhere.exe 不存在：需要安装 Visual Studio (含“使用 C++ 的桌面开发”)'
    }
    $version = & $vswhere -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationVersion
    if (-not $version) {
        throw '没有找到带 MSVC x64 工具的 Visual Studio 实例'
    }
    $major = [int]($version.Split('.')[0])
    switch ($major) {
        { $_ -ge 18 } { return 'Visual Studio 18 2026' }
        17 { return 'Visual Studio 17 2022' }
        16 { return 'Visual Studio 16 2019' }
        default { throw "未适配的 Visual Studio 主版本：$major" }
    }
}

if ($Clean -and (Test-Path $build)) {
    Remove-Item -Recurse -Force $build
}

$generator = Get-VSGenerator
$cmakeArgs = @(
    '-S', $src, '-B', $build,
    '-G', $generator, '-A', 'x64',
    '-DUI_CORE_BUILD_STATIC_ARCHIVE=OFF',
    '-DUI_CORE_BUILD_INTERNAL_TESTS=OFF',
    "-DUI_CORE_MSVC_STATIC_CRT=$(if ($DynamicCRT) { 'OFF' } else { 'ON' })"
)
Write-Host "==> cmake 配置 ($generator / $Config)"
& cmake @cmakeArgs
if ($LASTEXITCODE -ne 0) { throw 'cmake 配置失败' }

Write-Host '==> 编译 core-ui.dll'
& cmake --build $build --config $Config --target core-ui
if ($LASTEXITCODE -ne 0) { throw '编译失败' }

$outDir = Join-Path $build $Config
$dll = Join-Path $outDir 'core-ui.dll'
$lib = Join-Path $outDir 'core-ui.lib'
if (-not (Test-Path $dll)) { throw "找不到产物：$dll" }

$incDir = Join-Path $deps 'include'
$binDir = Join-Path $deps 'bin'
$libDir = Join-Path $deps 'lib'
New-Item -ItemType Directory -Force -Path $incDir, $binDir, $libDir | Out-Null

Copy-Item (Join-Path $src 'include/ui_core.h') (Join-Path $incDir 'ui_core.h') -Force
Copy-Item $dll (Join-Path $binDir 'core-ui.dll') -Force
if (Test-Path $lib) { Copy-Item $lib (Join-Path $libDir 'core-ui.lib') -Force }

Write-Host '==> 生成 MinGW 导入库 (libcore-ui.dll.a)'
$gendef = (Get-Command gendef -ErrorAction SilentlyContinue).Source
$dlltool = (Get-Command dlltool -ErrorAction SilentlyContinue).Source
if (-not $gendef -or -not $dlltool) {
    throw '找不到 gendef / dlltool，需要 MSYS2（mintty 里 pacman -S mingw-w64-ucrt-x86_64-binutils mingw-w64-ucrt-x86_64-tools-git）'
}
Push-Location $binDir
try {
    # gendef 把进度信息打到 stderr，Windows PowerShell 会当成 NativeCommandError，
    # 这里临时放宽 $ErrorActionPreference 再吞掉输出。
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    & $gendef core-ui.dll 2>&1 | Out-Null
    $ErrorActionPreference = $prevEap
    & $dlltool -d core-ui.def -D core-ui.dll -l (Join-Path $libDir 'libcore-ui.dll.a')
    if ($LASTEXITCODE -ne 0) { throw 'dlltool 失败' }
} finally {
    Pop-Location
}

Write-Host "==> 完成：$binDir/core-ui.dll"
