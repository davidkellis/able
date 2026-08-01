import { buildSourceEvaluationRequest } from "./source_request.mjs";
import { parseSourceToAstModule } from "./ast_adapter.mjs";

const resultNode = document.querySelector("#result");

async function main() {
  const output = installHostOutput();
  const evaluate = await startEvaluator();
  const parser = await createBrowserAbleParser();
  try {
    await verifyMode(evaluate, output, parser, "treewalker");
    await verifyMode(evaluate, output, parser, "bytecode");
    finish("pass", "treewalker and bytecode parsed Able sources and returned 42 with host output");
  } finally {
    parser.delete();
  }
}

function installHostOutput() {
  const stdout = [];
  const stderr = [];
  globalThis.able_host = {
    write_stdout(message) {
      stdout.push(String(message));
    },
    write_stderr(message) {
      stderr.push(String(message));
    },
  };
  return {
    reset() {
      stdout.length = 0;
      stderr.length = 0;
    },
    stdout() {
      return stdout.join("");
    },
    stderr() {
      return stderr.join("");
    },
  };
}

async function startEvaluator() {
  if (typeof globalThis.Go !== "function") {
    throw new Error("wasm_exec.js did not define Go");
  }
  const go = new globalThis.Go();
  const response = await fetch("/ablewasm.wasm");
  if (!response.ok) {
    throw new Error(`fetch ablewasm.wasm: ${response.status} ${response.statusText}`);
  }
  const { instance } = await WebAssembly.instantiateStreaming(response, go.importObject);
  // ablewasm installs its JavaScript callback then remains alive for future
  // host requests. Do not await this promise: it resolves only on shutdown.
  go.run(instance);
  await waitForEvaluator();
  return globalThis.__able_eval_request_json;
}

async function waitForEvaluator(timeoutMs = 5000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    if (typeof globalThis.__able_eval_request_json === "function") {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
  throw new Error("timed out waiting for __able_eval_request_json");
}

async function createBrowserAbleParser() {
  const { Language, Parser } = await loadWebTreeSitter();
  await Parser.init({
    locateFile(scriptName) {
      if (scriptName === "tree-sitter.wasm") {
        return "/vendor/tree-sitter.wasm";
      }
      return scriptName;
    },
  });
  const language = await Language.load("/tree-sitter-able.wasm");
  const parser = new Parser();
  parser.setLanguage(language);
  return parser;
}

async function loadWebTreeSitter() {
  // Go's browser wasm_exec.js defines process for its runtime shims but does
  // not provide Node's process.versions. web-tree-sitter probes that nested
  // property to choose its Node file loader, so make the browser shim explicit
  // before dynamically evaluating the package.
  if (globalThis.process && !globalThis.process.versions) {
    globalThis.process.versions = {};
  }
  return import("web-tree-sitter");
}

async function verifyMode(evaluate, output, parser, execMode) {
  output.reset();
  const request = await buildSourceEvaluationRequest({
    entryPath: "/workspace/main.able",
    moduleRoots: ["/approved-stdlib"],
    execMode,
    sourceProvider: browserSourceProvider(),
    parseSource(source) {
      return parseSourceToAstModule(parser, source);
    },
  });
  const response = JSON.parse(evaluate(JSON.stringify(request)));
  if (!response.ok || response.result !== "42") {
    throw new Error(`${execMode} response mismatch: ${JSON.stringify(response)}`);
  }
  if (output.stdout() !== "42\n" || output.stderr() !== "") {
    throw new Error(`${execMode} host output mismatch: ${JSON.stringify({ stdout: output.stdout(), stderr: output.stderr() })}`);
  }
}

function browserSourceProvider() {
  return {
    async readSource(origin) {
      const response = await fetch(origin);
      if (response.status === 404) {
        return null;
      }
      if (!response.ok) {
        throw new Error(`fetch ${origin}: ${response.status} ${response.statusText}`);
      }
      return response.text();
    },
  };
}

function finish(status, detail) {
  document.body.dataset.status = status;
  document.body.dataset.detail = detail;
  resultNode.textContent = detail;
}

main().catch((err) => {
  const detail = err instanceof Error
    ? [err.message, err.stack].filter(Boolean).join("\n")
    : String(err);
  finish("fail", detail);
});
