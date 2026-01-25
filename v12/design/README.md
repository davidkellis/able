# v12 Design Notes

v12 is Go-first (tree-walker + bytecode). The TypeScript interpreter was removed
from the active toolchain. Some design notes still reference TypeScript/Bun
workflows for historical context; treat those references as archived unless a
future non-Go runtime is revived.
