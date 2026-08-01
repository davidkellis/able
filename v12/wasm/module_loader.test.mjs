import assert from "node:assert/strict";

import { loadStaticModuleClosure } from "./module_loader.mjs";

await testDependencyOrder();
await testRootPrecedence();
await testDynamicImportRejection();
await testCycleRejection();
await testProviderPolicyFailure();
process.stdout.write("module loader tests passed\n");

async function testDependencyOrder() {
  const closure = await loadStaticModuleClosure({
    entryPath: "/order/main.able",
    parseSource: parseJSONModule,
    sourceProvider: mapSourceProvider({
      "/order/main.able": moduleAst("app", [staticImport("math")], []),
      "/order/math.able": moduleAst("math", [staticImport("dep")], []),
      "/order/dep.able": moduleAst("dep", [], []),
    }),
  });

  assert.deepEqual(
    closure.setupModules.map(({ origin }) => origin),
    ["/order/dep.able", "/order/math.able"],
  );
  assert.equal(closure.entry.origin, "/order/main.able");
}

async function testRootPrecedence() {
  const closure = await loadStaticModuleClosure({
    entryPath: "/app/main.able",
    moduleRoots: ["/first", "/second"],
    parseSource: parseJSONModule,
    sourceProvider: mapSourceProvider({
      "/app/main.able": moduleAst("app", [staticImport("dep")], []),
      "/first/dep.able": moduleAst("dep", [], [{ from: "first" }]),
      "/second/dep.able": moduleAst("dep", [], [{ from: "second" }]),
    }),
  });

  assert.equal(closure.setupModules[0].origin, "/first/dep.able");
}

async function testDynamicImportRejection() {
  await assert.rejects(
    () => loadStaticModuleClosure({
      entryPath: "/dynamic/main.able",
      parseSource: parseJSONModule,
      sourceProvider: mapSourceProvider({
        "/dynamic/main.able": moduleAst("app", [{ ...staticImport("dep"), type: "DynImportStatement" }], []),
      }),
    }),
    /dynamic import.*unavailable/,
  );
}

async function testCycleRejection() {
  await assert.rejects(
    () => loadStaticModuleClosure({
      entryPath: "/cycle/main.able",
      parseSource: parseJSONModule,
      sourceProvider: mapSourceProvider({
        "/cycle/main.able": moduleAst("main", [staticImport("dep")], []),
        "/cycle/dep.able": moduleAst("dep", [staticImport("main")], []),
      }),
    }),
    /static import cycle: main\.able -> dep\.able -> main\.able/,
  );
}

async function testProviderPolicyFailure() {
  await assert.rejects(
    () => loadStaticModuleClosure({
      entryPath: "/policy/main.able",
      parseSource: parseJSONModule,
      sourceProvider: {
        async readSource(origin) {
          if (origin === "/policy/main.able") {
            return JSON.stringify(moduleAst("app", [staticImport("private")], []));
          }
          throw new Error("access denied by host policy");
        },
      },
    }),
    /read \/policy\/private\.able: access denied by host policy/,
  );
}

function mapSourceProvider(modules) {
  const sources = new Map(
    Object.entries(modules).map(([origin, module]) => [origin, JSON.stringify(module)]),
  );
  return {
    async readSource(origin) {
      return sources.get(origin) ?? null;
    },
  };
}

function parseJSONModule(source) {
  return JSON.parse(source);
}

function moduleAst(packageName, imports, body) {
  return {
    type: "Module",
    package: {
      type: "PackageStatement",
      namePath: [{ type: "Identifier", name: packageName }],
    },
    imports,
    body,
  };
}

function staticImport(packageName) {
  return {
    type: "ImportStatement",
    packagePath: packageName.split(".").map((name) => ({ type: "Identifier", name })),
    isWildcard: false,
  };
}
