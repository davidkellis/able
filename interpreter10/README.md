# interpreter10

TypeScript AST for Able v10, built with Bun.

To install dependencies:

```bash
bun install
```

Typecheck:

```bash
bun run typecheck
```

This project was created using `bun init` in bun v1.2.19. [Bun](https://bun.sh/) is a fast all-in-one JavaScript runtime.

Exports:

- `index.ts` re-exports all AST types and builders from `src/ast.ts` as `AST`.
