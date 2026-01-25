import type { Interpreter } from "../index";
import type { ImplMethodEntry } from "../values";

export function formatAmbiguousImplementationError(
  interp: Interpreter,
  interfaceName: string,
  typeName: string,
  contenders: Array<{ entry: ImplMethodEntry }>,
): string {
  const detail = Array.from(new Set(
    contenders.map(c => `impl ${c.entry.def.interfaceName.name} for ${interp.typeExpressionToString(c.entry.def.targetType)}`),
  )).join(", ");
  return `ambiguous implementations of ${interfaceName} for ${typeName}: ${detail}`;
}
