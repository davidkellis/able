import { execFileSync, spawn } from "node:child_process";
import { once } from "node:events";
import fsSync from "node:fs";
import fs from "node:fs/promises";
import http from "node:http";
import net from "node:net";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const wasmPath = path.join(__dirname, "ablewasm.wasm");
const wasmExecPath = resolveWasmExecPath();
const treeSitterPackage = path.join(__dirname, "node_modules", "web-tree-sitter");
const treeSitterJSPath = path.join(treeSitterPackage, "tree-sitter.js");
const treeSitterWasmPath = path.join(treeSitterPackage, "tree-sitter.wasm");
const ableGrammarWasmPath = path.join(__dirname, "../parser/tree-sitter-able/tree-sitter-able.wasm");
const server = await startServer();
const driverPort = await reservePort();
const driver = spawn("geckodriver", ["--port", String(driverPort)], {
  stdio: ["ignore", "pipe", "pipe"],
});
let driverLog = "";
driver.stdout.on("data", (chunk) => { driverLog += chunk; });
driver.stderr.on("data", (chunk) => { driverLog += chunk; });
let sessionID = "";

try {
  const driverURL = `http://127.0.0.1:${driverPort}`;
  await waitForDriver(driverURL);
  const session = await webdriver(driverURL, "/session", {
    method: "POST",
    body: {
      capabilities: {
        alwaysMatch: {
          browserName: "firefox",
          "moz:firefoxOptions": { args: ["-headless"] },
        },
      },
    },
  });
  sessionID = session.sessionId;
  await webdriver(driverURL, `/session/${sessionID}/url`, {
    method: "POST",
    body: { url: server.url("/browser_smoke.html") },
  });
  const detail = await waitForBrowserResult(driverURL, sessionID);
  process.stdout.write(`browser smoke passed: ${detail}\n`);
} finally {
  if (sessionID) {
    await webdriver(`http://127.0.0.1:${driverPort}`, `/session/${sessionID}`, { method: "DELETE" }).catch(() => {});
  }
  await stopProcess(driver);
  await server.close();
}

async function startServer() {
  await fs.access(wasmPath);
  await fs.access(treeSitterJSPath);
  await fs.access(treeSitterWasmPath);
  await fs.access(ableGrammarWasmPath);
  const files = new Map([
    ["/browser_smoke.html", path.join(__dirname, "browser_smoke.html")],
    ["/browser_smoke.mjs", path.join(__dirname, "browser_smoke.mjs")],
    ["/source_request.mjs", path.join(__dirname, "source_request.mjs")],
    ["/module_loader.mjs", path.join(__dirname, "module_loader.mjs")],
    ["/ast_adapter.mjs", path.join(__dirname, "ast_adapter.mjs")],
    ["/ablewasm.wasm", wasmPath],
    ["/wasm_exec.js", wasmExecPath],
    ["/vendor/web-tree-sitter.js", treeSitterJSPath],
    ["/vendor/tree-sitter.wasm", treeSitterWasmPath],
    ["/tree-sitter-able.wasm", ableGrammarWasmPath],
    ["/workspace/main.able", path.join(__dirname, "samples/modules/main.able")],
    ["/approved-stdlib/math.able", path.join(__dirname, "samples/modules/math.able")],
    ["/approved-stdlib/dep.able", path.join(__dirname, "samples/modules/dep.able")],
  ]);
  const httpServer = http.createServer(async (request, response) => {
    const pathname = new URL(request.url, "http://localhost").pathname;
    const file = files.get(pathname);
    if (!file) {
      response.writeHead(404).end("not found");
      return;
    }
    try {
      response.writeHead(200, { "content-type": contentType(pathname), "cache-control": "no-store" });
      response.end(await fs.readFile(file));
    } catch (err) {
      response.writeHead(500).end(String(err));
    }
  });
  httpServer.listen(0, "127.0.0.1");
  await once(httpServer, "listening");
  const address = httpServer.address();
  if (!address || typeof address === "string") {
    throw new Error("browser smoke server did not expose a TCP address");
  }
  return {
    url(pathname) {
      return `http://127.0.0.1:${address.port}${pathname}`;
    },
    async close() {
      httpServer.close();
      await once(httpServer, "close");
    },
  };
}

function contentType(pathname) {
  if (pathname.endsWith(".html")) return "text/html; charset=utf-8";
  if (pathname.endsWith(".mjs") || pathname.endsWith(".js")) return "text/javascript; charset=utf-8";
  if (pathname.endsWith(".wasm")) return "application/wasm";
  return "application/octet-stream";
}

async function reservePort() {
  const listener = net.createServer();
  listener.listen(0, "127.0.0.1");
  await once(listener, "listening");
  const address = listener.address();
  listener.close();
  await once(listener, "close");
  if (!address || typeof address === "string") {
    throw new Error("unable to reserve geckodriver port");
  }
  return address.port;
}

async function waitForDriver(driverURL, timeoutMs = 10000) {
  const deadline = Date.now() + timeoutMs;
  let lastError = "";
  while (Date.now() < deadline) {
    try {
      const response = await fetch(`${driverURL}/status`);
      if (response.ok) return;
      lastError = `${response.status} ${response.statusText}`;
    } catch (err) {
      lastError = err instanceof Error ? err.message : String(err);
    }
    await delay(50);
  }
  throw new Error(`geckodriver did not start: ${lastError || driverLog}`);
}

async function waitForBrowserResult(driverURL, sessionID, timeoutMs = 20000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const status = await execute(driverURL, sessionID, "return document.body.dataset.status;");
    const detail = await execute(driverURL, sessionID, "return document.body.dataset.detail;");
    if (status === "pass") return detail;
    if (status === "fail") throw new Error(`browser smoke failed: ${detail}`);
    await delay(50);
  }
  throw new Error("timed out waiting for browser smoke result");
}

async function execute(driverURL, sessionID, script) {
  return webdriver(driverURL, `/session/${sessionID}/execute/sync`, {
    method: "POST",
    body: { script, args: [] },
  });
}

async function webdriver(driverURL, pathname, { method, body } = {}) {
  const response = await fetch(`${driverURL}${pathname}`, {
    method,
    headers: body ? { "content-type": "application/json; charset=utf-8" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok || payload.value?.error) {
    throw new Error(`WebDriver ${method ?? "GET"} ${pathname}: ${payload.value?.message ?? response.statusText}`);
  }
  return payload.value;
}

function resolveWasmExecPath() {
  const goRoot = execFileSync("go", ["env", "GOROOT"], { encoding: "utf8" }).trim();
  for (const candidate of [
    path.join(goRoot, "lib", "wasm", "wasm_exec.js"),
    path.join(goRoot, "misc", "wasm", "wasm_exec.js"),
  ]) {
    try {
      fsSync.accessSync(candidate, fsSync.constants.R_OK);
      return candidate;
    } catch {
      // Try the next Go layout.
    }
  }
  throw new Error(`unable to locate wasm_exec.js under GOROOT=${goRoot}`);
}

async function stopProcess(process) {
  if (process.exitCode !== null || process.signalCode !== null) return;
  process.kill("SIGTERM");
  await Promise.race([
    once(process, "exit"),
    delay(3000).then(() => process.kill("SIGKILL")),
  ]);
}

function delay(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
