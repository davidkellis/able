import { promises as fs } from "node:fs";
import { getTreeSitterParser } from "./src/parser/tree-sitter-loader";

async function main(): Promise<void> {
  const target = process.argv[2];
  if (!target) {
    console.error("usage: bun run tmp_inspect.ts <path>");
    process.exit(1);
  }
  const source = await fs.readFile(target, "utf8");
  const parser = await getTreeSitterParser();
  const tree = parser.parse(source);
  console.log(`hasError=${tree.rootNode.hasError}`);
  if (tree.rootNode.hasError) {
    const errors = tree.rootNode.descendantsOfType("ERROR");
    if (errors.length === 0) {
      console.log(tree.rootNode.toString());
    }
    const missing = tree.rootNode.descendantsOfType([]).filter(node => node.isMissing?.());
    if (missing.length > 0) {
      console.log("Missing nodes:");
      for (const m of missing) {
        const { row, column } = m.startPosition;
        console.log(`  ${m.type} at ${row + 1}:${column + 1}`);
      }
    }
    for (const err of errors) {
      const { row, column } = err.startPosition;
      const contextStart = Math.max(0, err.startIndex - 20);
      const contextEnd = Math.min(source.length, err.endIndex + 20);
      const context = source.slice(contextStart, contextEnd).replace(/\n/g, "\\n");
      console.log(`ERROR @ ${row + 1}:${column + 1} -> ${err.toString()}`);
      console.log(`  context: "${context}"`);
    }
  }
}

await main();
