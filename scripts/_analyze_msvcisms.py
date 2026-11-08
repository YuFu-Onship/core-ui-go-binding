import os
import re
import sys

root = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "core-ui")
patterns = {
    "__uuidof": r"__uuidof",
    "__super": r"\b__super\b",
    "__try/__except": r"\b__(try|except|finally)\b",
    "#pragma comment": r"#pragma\s+comment",
    "WRL": r"Microsoft::WRL|#include\s*<wrl",
    "C++/WinRT": r"winrt/|::winrt",
    "_MSC_VER": r"_MSC_VER",
    "__declspec": r"__declspec",
    "_countof": r"\b_countof\b",
    "#import": r"^\s*#import\s",
    "d2d1_3.h": r"<d2d1_3\.h>|<d2d1svg\.h>",
    "dwrite_3.h": r"<dwrite_3\.h>",
    "restrict-ms": r"\b__restrict\b",
    "intrin": r"__intrin|#include\s*<intrin\.h>",
    "SecureZeroMemory": r"\bSecureZeroMemory\b",
    "__int64": r"\b__int64\b",
}

counts = {k: [0, set()] for k in patterns}
files = 0
for dirpath, _dirnames, filenames in os.walk(os.path.join(root, "src")):
    for name in filenames:
        if not name.endswith((".cpp", ".h", ".c")):
            continue
        path = os.path.join(dirpath, name)
        files += 1
        try:
            text = open(path, encoding="utf-8", errors="replace").read()
        except OSError:
            continue
        rel = os.path.relpath(path, root)
        for key, pat in patterns.items():
            found = re.findall(pat, text, re.MULTILINE)
            if found:
                counts[key][0] += len(found)
                counts[key][1].add(rel)

print(f"scanned {files} source files under core-ui/src")
print()
for key, (n, where) in counts.items():
    if n == 0:
        print(f"{key:16} 0")
        continue
    sample = ", ".join(sorted(where)[:6])
    more = "" if len(where) <= 6 else f" (+{len(where) - 6} more files)"
    print(f"{key:16} {n:4} occurrences in {len(where)} files: {sample}{more}")
