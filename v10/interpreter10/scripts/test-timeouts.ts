export function startRunTimeout(
  label: string,
  warnMs = 60_000,
  killMs = 120_000,
): () => void {
  const warnTimer = setTimeout(() => {
    console.warn(`[${label}] still running after ${warnMs / 1000}s`);
  }, warnMs);
  warnTimer.unref?.();

  const killTimer = setTimeout(() => {
    console.error(`[${label}] exceeded ${killMs / 1000}s; forcing exit`);
    process.exit(1);
  }, killMs);

  return () => {
    clearTimeout(warnTimer);
    clearTimeout(killTimer);
  };
}
