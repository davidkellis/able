import { collectModuleSearchPaths } from './scripts/module-search-paths.ts';
import { discoverRoot, discoverRootForPath, indexSourceFiles } from './scripts/module-utils.ts';

const entryPath = process.env.ENTRY!;
const searchPaths = collectModuleSearchPaths({
  ableModulePathsEnv: process.env.ABLE_MODULE_PATHS ?? '',
  ablePathEnv: '',
  ableStdLibEnv: '',
  cwd: '/home/david/sync/projects/able',
  probeStdlibFrom: ['/home/david/sync/projects/able', __dirname],
});
console.log('searchPaths', searchPaths);

const pkgIndex = new Map();
const origins = new Map();

function ensureNamespaceAllowed(root, allowSkip) {
  if (root.rootName === 'able' && !root.isStdlib) {
    const msg = `reserved able.* at ${root.rootDir}`;
    if (allowSkip) {
      console.warn(msg);
      return false;
    }
    throw new Error(msg);
  }
  return true;
}

function registerPackages(packages, root) {
  for (const [name, loc] of packages.entries()) {
    if (origins.has(name)) throw new Error(`collision ${name}`);
    origins.set(name, { root: root.rootDir, isStdlib: root.isStdlib });
    pkgIndex.set(name, loc);
  }
}

const entryRoot = await discoverRoot(entryPath);
console.log('entryRoot', entryRoot);
const { packages: entryPkgs } = await indexSourceFiles(entryRoot.rootDir, entryRoot.rootName);
console.log('entry packages', [...entryPkgs.keys()]);
registerPackages(entryPkgs, { ...entryRoot, isStdlib: false });

for (const sp of searchPaths) {
  const discovered = await discoverRootForPath(sp.path);
  const root = { rootDir: discovered.abs, rootName: discovered.rootName, isStdlib: Boolean(sp.isStdlib) };
  if (!ensureNamespaceAllowed(root, true)) continue;
  const { packages } = await indexSourceFiles(root.rootDir, root.rootName);
  console.log('root', root, 'packages', [...packages.keys()]);
  registerPackages(packages, root);
}

console.log('pkgIndex keys', [...pkgIndex.keys()]);
