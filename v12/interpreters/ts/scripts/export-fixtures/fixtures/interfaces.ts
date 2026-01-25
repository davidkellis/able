import type { Fixture } from "../../types";
import { applyIndexDiagnostics, applyIndexDispatch } from "./interfaces/apply_index";
import { collisionFixture } from "./interfaces/collisions";
import { dynamicInterfaceCollectionsFixture } from "./interfaces/dynamic_collections";
import { importedImplDispatchFixture } from "./interfaces/imported_impl_dispatch";
import {
  customHashEqFixture,
  kernelInterfaceFixture,
  primitiveHashingFixture,
} from "./interfaces/hashing";

const interfacesFixtures: Fixture[] = [
  dynamicInterfaceCollectionsFixture,
  importedImplDispatchFixture,
  applyIndexDispatch,
  applyIndexDiagnostics,
  primitiveHashingFixture,
  kernelInterfaceFixture,
  customHashEqFixture,
  collisionFixture,
];

export default interfacesFixtures;
