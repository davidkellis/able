import fs from "node:fs/promises";

// createNodeSourceProvider adapts the portable source-provider contract to
// Node's filesystem. The resolver itself deliberately has no Node dependency,
// so a browser can provide approved virtual source paths instead.
export function createNodeSourceProvider() {
  return {
    async readSource(origin) {
      try {
        return await fs.readFile(origin, "utf8");
      } catch (err) {
        if (err?.code === "ENOENT" || err?.code === "ENOTDIR") {
          return null;
        }
        throw err;
      }
    },
  };
}
