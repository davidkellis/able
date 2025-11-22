import type { Fixture } from "../../../types";
import { procSchedulingPart1 } from "./proc_scheduling_part1";
import { procSchedulingPart2 } from "./proc_scheduling_part2";

export const procSchedulingFixtures: Fixture[] = [
  ...procSchedulingPart1,
  ...procSchedulingPart2,
];

export default procSchedulingFixtures;
