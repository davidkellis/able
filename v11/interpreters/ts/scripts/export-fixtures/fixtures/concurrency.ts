import type { Fixture } from "../../types";
import channelConcurrencyFixtures from "./concurrency/channel";
import channelProgressFixtures from "./concurrency/channel_progress";
import awaitFixtures from "./concurrency/await";
import mutexConcurrencyFixtures from "./concurrency/mutex";
import procConcurrencyFixtures from "./concurrency/proc";
import procMemoizationFixtures from "./concurrency/proc_memoization";
import procReentrancyFixtures from "./concurrency/proc_reentrancy";
import executorDiagnosticsFixtures from "./concurrency/executor_diagnostics";

const concurrencyFixtures: Fixture[] = [
  ...procConcurrencyFixtures,
  ...procMemoizationFixtures,
  ...procReentrancyFixtures,
  ...executorDiagnosticsFixtures,
  ...channelConcurrencyFixtures,
  ...channelProgressFixtures,
  ...awaitFixtures,
  ...mutexConcurrencyFixtures,
];

export default concurrencyFixtures;
