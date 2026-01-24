import type { Fixture } from "../../types";
import channelConcurrencyFixtures from "./concurrency/channel";
import channelProgressFixtures from "./concurrency/channel_progress";
import awaitFixtures from "./concurrency/await";
import mutexConcurrencyFixtures from "./concurrency/mutex";
import futureConcurrencyFixtures from "./concurrency/future";
import futureMemoizationFixtures from "./concurrency/future_memoization";
import futureReentrancyFixtures from "./concurrency/future_reentrancy";
import executorDiagnosticsFixtures from "./concurrency/executor_diagnostics";

const concurrencyFixtures: Fixture[] = [
  ...futureConcurrencyFixtures,
  ...futureMemoizationFixtures,
  ...futureReentrancyFixtures,
  ...executorDiagnosticsFixtures,
  ...channelConcurrencyFixtures,
  ...channelProgressFixtures,
  ...awaitFixtures,
  ...mutexConcurrencyFixtures,
];

export default concurrencyFixtures;
