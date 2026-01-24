import type { Fixture } from "../../../types";
import { futureCoreFixtures } from "./future_core";
import { futureSchedulingFixtures } from "./future_scheduling";

const futureConcurrencyFixtures: Fixture[] = [
  ...futureCoreFixtures,
  ...futureSchedulingFixtures,
];

export default futureConcurrencyFixtures;
