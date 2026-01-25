export type ExecutorTask = () => void;

export interface Executor {
  schedule(task: ExecutorTask): void;
  ensureTick(): void;
  flush(limit?: number): void;
  pendingTasks?(): number;
}

type QueueOptions = {
  maxSteps?: number;
};

export class CooperativeExecutor implements Executor {
  private readonly maxSteps: number;
  private queue: ExecutorTask[] = [];
  private scheduled = false;
  private active = false;

  constructor(options: QueueOptions = {}) {
    this.maxSteps = options.maxSteps ?? 1024;
  }

  schedule(task: ExecutorTask): void {
    this.queue.push(task);
    this.ensureTick();
  }

  ensureTick(): void {
    if (this.scheduled || this.active) return;
    this.scheduled = true;
    const runner = () => this.flush();
    if (typeof queueMicrotask === "function") {
      queueMicrotask(runner);
    } else if (typeof setTimeout === "function") {
      setTimeout(runner, 0);
    } else {
      runner();
    }
  }

  flush(limit?: number): void {
    const wasActive = this.active;
    const stepLimit = limit ?? this.maxSteps;
    if (!wasActive) {
      this.active = true;
      this.scheduled = false;
    }
    let steps = 0;
    while (this.queue.length > 0 && steps < stepLimit) {
      const task = this.queue.shift()!;
      task();
      steps += 1;
    }
    if (!wasActive) {
      this.active = false;
      if (this.queue.length > 0) {
        this.ensureTick();
      }
    }
  }

  pendingTasks(): number {
    return this.queue.length;
  }
}
