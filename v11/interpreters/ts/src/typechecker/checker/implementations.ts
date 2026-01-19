export type { ImplementationContext } from "./implementation-context";
export { collectImplementationDefinition, collectMethodsDefinition } from "./implementation-collection";
export {
  ambiguousImplementationDetail,
  enforceFunctionConstraints,
  lookupMethodSetsForCall,
  matchImplementationTarget,
  typeImplementsInterface,
} from "./implementation-constraints";
