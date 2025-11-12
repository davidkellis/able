import type { Fixture } from "../../types";
import basicsFixtures from "./basics";
import expressionsFixtures from "./expressions";
import functionsFixtures from "./functions";
import stringsFixtures from "./strings";
import matchFixtures from "./match";
import controlFixtures from "./control";
import patternsFixtures from "./patterns";
import structsFixtures from "./structs";
import privacyFixtures from "./privacy";
import importsFixtures from "./imports";
import concurrencyFixtures from "./concurrency";
import stdlibFixtures from "./stdlib";
import typesFixtures from "./types";
import errorsFixtures from "./errors";
import interfacesFixtures from "./interfaces";

export const fixtures: Fixture[] = [
  ...basicsFixtures,
  ...expressionsFixtures,
  ...functionsFixtures,
  ...stringsFixtures,
  ...matchFixtures,
  ...controlFixtures,
  ...patternsFixtures,
  ...structsFixtures,
  ...privacyFixtures,
  ...importsFixtures,
  ...concurrencyFixtures,
  ...stdlibFixtures,
  ...typesFixtures,
  ...interfacesFixtures,
  ...errorsFixtures,
];
