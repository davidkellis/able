import type { Fixture } from "../../types";
import channelConcurrencyFixtures from "./concurrency/channel";
import mutexConcurrencyFixtures from "./concurrency/mutex";
import procConcurrencyFixtures from "./concurrency/proc";
import procMemoizationFixtures from "./concurrency/proc_memoization";
import procReentrancyFixtures from "./concurrency/proc_reentrancy";

const concurrencyFixtures: Fixture[] = [
  ...procConcurrencyFixtures,
  ...procMemoizationFixtures,
  ...procReentrancyFixtures,
  ...channelConcurrencyFixtures,
  ...mutexConcurrencyFixtures,
];

export default concurrencyFixtures;
