import type { Fixture } from "../../types";
import channelConcurrencyFixtures from "./concurrency/channel";
import mutexConcurrencyFixtures from "./concurrency/mutex";
import procConcurrencyFixtures from "./concurrency/proc";

const concurrencyFixtures: Fixture[] = [
  ...procConcurrencyFixtures,
  ...channelConcurrencyFixtures,
  ...mutexConcurrencyFixtures,
];

export default concurrencyFixtures;
