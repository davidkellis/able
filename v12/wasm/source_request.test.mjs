import assert from "node:assert/strict";

import { buildSourceEvaluationRequest } from "./source_request.mjs";

await testVirtualSourceMapRequest();
await testParserFailureIncludesOrigin();
process.stdout.write("source request tests passed\n");

async function testVirtualSourceMapRequest() {
  const parsedOrigins = [];
  const request = await buildSourceEvaluationRequest({
    entryPath: "/workspace/app/main.able",
    moduleRoots: ["/approved-stdlib"],
    execMode: "bytecode",
    sourceProvider: mapSourceProvider({
      "/workspace/app/main.able": moduleAst("app", [staticImport("math")], [{ value: "entry" }]),
      "/approved-stdlib/math.able": moduleAst("math", [staticImport("dep")], [{ value: "math" }]),
      "/approved-stdlib/dep.able": moduleAst("dep", [], [{ value: "dep" }]),
    }),
    parseSource(source, origin) {
      parsedOrigins.push(origin);
      return JSON.parse(source);
    },
  });

  assert.equal(request.execMode, "bytecode");
  assert.equal(request.entryOrigin, "/workspace/app/main.able");
  assert.deepEqual(
    request.setupModules.map(({ origin }) => origin),
    ["/approved-stdlib/dep.able", "/approved-stdlib/math.able"],
  );
  assert.equal(request.module.package.namePath[0].name, "app");
  assert.deepEqual(parsedOrigins, [
    "/workspace/app/main.able",
    "/approved-stdlib/math.able",
    "/approved-stdlib/dep.able",
  ]);
}

async function testParserFailureIncludesOrigin() {
  await assert.rejects(
    () => buildSourceEvaluationRequest({
      entryPath: "/workspace/bad.able",
      sourceProvider: mapSourceProvider({
        "/workspace/bad.able": "not a valid virtual source",
      }),
      parseSource() {
        throw new Error("host parser rejected source");
      },
    }),
    /parse \/workspace\/bad\.able: host parser rejected source/,
  );
}

function mapSourceProvider(modules) {
  const sources = new Map(
    Object.entries(modules).map(([origin, module]) => [
      origin,
      typeof module === "string" ? module : JSON.stringify(module),
    ]),
  );
  return {
    async readSource(origin) {
      return sources.get(origin) ?? null;
    },
  };
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
