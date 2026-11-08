"""用指定编译器对 core-ui 的全部 .cpp 做语法检查（-fsyntax-only），统计失败面。

用法:
    python scripts/_try_syntax.py g++  [extra flags...]
    python scripts/_try_syntax.py clang++
"""
import os
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CORE = os.path.join(ROOT, "core-ui")
SRC = os.path.join(CORE, "src")
QJS = os.path.join(ROOT, "build", "core-ui", "_deps", "quickjs-src")
LUNASVG = os.path.join(CORE, "third_party", "lunasvg", "include")

DEFINES = [
    "-DUNICODE",
    "-D_UNICODE",
    "-DWIN32_LEAN_AND_MEAN",
    "-DNOMINMAX",
    "-D_WIN32_WINNT=0x0A00",
    "-DWINVER=0x0A00",
    "-DNTDDI_VERSION=0x0A000004",
    "-D_CRT_SECURE_NO_WARNINGS",
    "-D_SCL_SECURE_NO_WARNINGS",
]

INCLUDES = [
    "-I" + os.path.join(CORE, "include"),
    "-I" + os.path.join(CORE, "src"),
    "-I" + os.path.join(SRC, "ui"),
    "-I" + QJS,
    "-I" + LUNASVG,
]


def main():
    cc = sys.argv[1] if len(sys.argv) > 1 else "g++"
    extra = sys.argv[2:]
    files = []
    for dirpath, _dirs, names in os.walk(SRC):
        for name in names:
            if name.endswith(".cpp"):
                files.append(os.path.join(dirpath, name))
    files.sort()

    base = [cc, "-std=c++17", "-fsyntax-only"] + DEFINES + INCLUDES + extra
    failures = []
    for i, path in enumerate(files, 1):
        cmd = base + [path]
        proc = subprocess.run(cmd, capture_output=True, text=True, errors="replace")
        rel = os.path.relpath(path, CORE)
        if proc.returncode == 0:
            print(f"[{i:3}/{len(files)}] OK      {rel}")
        else:
            errs = [l for l in proc.stderr.splitlines() if " error: " in l]
            first = errs[0].split(" error: ", 1)[-1][:180] if errs else "(unknown)"
            failures.append((rel, len(errs), first))
            print(f"[{i:3}/{len(files)}] FAIL    {rel}  ({len(errs)} errors) {first}")

    print()
    print(f"== {cc}: {len(files) - len(failures)}/{len(files)} files OK, {len(failures)} failed")
    for rel, n, first in failures:
        print(f"   {rel}: {n} errors -> {first}")


if __name__ == "__main__":
    main()
