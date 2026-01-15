import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = findRepoRoot(__dirname);

export function repoRoot(): string {
  return REPO_ROOT;
}

export function normalizeRepoRelativePath(filePath: string): string {
  if (!filePath) return "";
  const absolute = path.isAbsolute(filePath) ? filePath : path.resolve(REPO_ROOT, filePath);
  const relative = path.relative(REPO_ROOT, absolute);
  return relative.split(path.sep).join("/");
}

function findRepoRoot(start: string): string {
  let dir = start;
  for (let i = 0; i < 12; i += 1) {
    if (existsSync(path.join(dir, ".git"))) {
      return dir;
    }
    const parent = path.dirname(dir);
    if (parent === dir) {
      break;
    }
    dir = parent;
  }
  return path.resolve(__dirname, "../..");
}
