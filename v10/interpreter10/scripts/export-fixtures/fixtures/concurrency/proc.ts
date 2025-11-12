import type { Fixture } from "../../../types";
import { procCoreFixtures } from "./proc_core";
import { procSchedulingFixtures } from "./proc_scheduling";

const procConcurrencyFixtures: Fixture[] = [
  ...procCoreFixtures,
  ...procSchedulingFixtures,
];

export default procConcurrencyFixtures;
