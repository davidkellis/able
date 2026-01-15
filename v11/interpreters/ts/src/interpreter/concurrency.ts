import type { Interpreter } from "./index";
import { applyConcurrencyAwait } from "./concurrency_await";
import { applyConcurrencyBuiltins } from "./concurrency_builtins";
import { applyConcurrencyProcFuture } from "./concurrency_proc_future";
import { applyConcurrencyScheduler } from "./concurrency_scheduler";

export function applyConcurrencyAugmentations(cls: typeof Interpreter): void {
  applyConcurrencyBuiltins(cls);
  applyConcurrencyScheduler(cls);
  applyConcurrencyProcFuture(cls);
  applyConcurrencyAwait(cls);
}
