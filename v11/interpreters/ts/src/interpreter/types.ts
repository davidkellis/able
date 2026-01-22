import * as AST from "../ast";
import type { Interpreter } from "./index";
import type { TypeImplementsInterfaceOptions } from "./types/helpers";
import { applyTypeFormatAugmentations } from "./types/format";
import { applyTypePrimitiveAugmentations } from "./types/primitives";
import { applyTypeStructAugmentations } from "./types/structs";
import { applyTypeUnionAugmentations } from "./types/unions";
import type { RuntimeValue } from "./values";

declare module "./index" {
  interface Interpreter {
    expandTypeAliases(t: AST.TypeExpression, seen?: Set<string>, seenNodes?: Set<AST.TypeExpression>): AST.TypeExpression;
    typeExpressionToString(t: AST.TypeExpression, seen?: Set<string>, seenNodes?: Set<AST.TypeExpression>, depth?: number): string;
    parseTypeExpression(t: AST.TypeExpression): { name: string; typeArgs: AST.TypeExpression[] } | null;
    typeExpressionsEqual(a: AST.TypeExpression, b: AST.TypeExpression): boolean;
    cloneTypeExpression(t: AST.TypeExpression): AST.TypeExpression;
    typeExpressionFromValue(value: RuntimeValue, seen?: WeakSet<object>): AST.TypeExpression | null;
    matchesType(t: AST.TypeExpression, v: RuntimeValue): boolean;
    getTypeNameForValue(value: RuntimeValue): string | null;
    typeImplementsInterface(typeName: string, interfaceName: string, opts?: AST.TypeExpression[] | TypeImplementsInterfaceOptions): boolean;
    coerceValueToType(typeExpr: AST.TypeExpression | undefined, value: RuntimeValue): RuntimeValue;
    castValueToType(typeExpr: AST.TypeExpression, value: RuntimeValue): RuntimeValue;
    toInterfaceValue(interfaceName: string, rawValue: RuntimeValue, interfaceArgs?: AST.TypeExpression[]): RuntimeValue;
  }
}

export function applyTypesAugmentations(cls: typeof Interpreter): void {
  applyTypeFormatAugmentations(cls);
  applyTypePrimitiveAugmentations(cls);
  applyTypeStructAugmentations(cls);
  applyTypeUnionAugmentations(cls);
}
