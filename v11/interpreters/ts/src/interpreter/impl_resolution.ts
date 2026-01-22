import * as AST from "../ast";
import type { Environment } from "./environment";
import type { Interpreter } from "./index";
import type { ConstraintSpec, ImplMethodEntry, RuntimeValue } from "./values";
import { applyImplCandidateAugmentations } from "./impl_resolution/candidates";
import { applyImplConstraintAugmentations } from "./impl_resolution/constraints";
import { applyImplDefaultAugmentations } from "./impl_resolution/defaults";
import { applyImplSpecificityAugmentations } from "./impl_resolution/specificity";

declare module "./index" {
  interface Interpreter {
    enforceGenericConstraintsIfAny(funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall): void;
    collectConstraintSpecs(generics?: AST.GenericParameter[], where?: AST.WhereClauseConstraint[]): ConstraintSpec[];
    mapTypeArguments(generics: AST.GenericParameter[] | undefined, provided: AST.TypeExpression[] | undefined, context: string): Map<string, AST.TypeExpression>;
    enforceConstraintSpecs(constraints: ConstraintSpec[], typeArgMap: Map<string, AST.TypeExpression>, context: string): void;
    ensureTypeSatisfiesInterface(typeInfo: { name: string; typeArgs: AST.TypeExpression[] }, interfaceType: AST.TypeExpression, context: string, visited: Set<string>): void;
    inferTypeArgumentsFromCall(funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall, args: RuntimeValue[]): void;
    bindTypeArgumentsIfAny(funcNode: AST.FunctionDefinition | AST.LambdaExpression, call: AST.FunctionCall, env: Environment): void;
    collectInterfaceConstraintExpressions(typeExpr: AST.TypeExpression, memo?: Set<string>): AST.TypeExpression[];
    findMethod(
      typeName: string,
      methodName: string,
      opts?: {
        typeArgs?: AST.TypeExpression[];
        interfaceArgs?: AST.TypeExpression[];
        typeArgMap?: Map<string, AST.TypeExpression>;
        interfaceName?: string;
        includeInherent?: boolean;
      },
    ): Extract<RuntimeValue, { kind: "function" | "function_overload" }> | null;
    resolveInterfaceImplementation(
      typeName: string,
      interfaceName: string,
      opts?: { typeArgs?: AST.TypeExpression[]; interfaceArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression> },
    ): { ok: boolean; error?: Error };
    compareMethodMatches(
      a: { entry: ImplMethodEntry; bindings: Map<string, AST.TypeExpression>; constraints: ConstraintSpec[]; isConcreteTarget: boolean; score: number; method?: Extract<RuntimeValue, { kind: "function" | "function_overload" }> },
      b: { entry: ImplMethodEntry; bindings: Map<string, AST.TypeExpression>; constraints: ConstraintSpec[]; isConcreteTarget: boolean; score: number; method?: Extract<RuntimeValue, { kind: "function" | "function_overload" }> },
    ): number;
    buildConstraintKeySet(constraints: ConstraintSpec[]): Set<string>;
    isConstraintSuperset(a: Set<string>, b: Set<string>): boolean;
    isProperSubset(a: string[], b: string[]): boolean;
    matchImplEntry(entry: ImplMethodEntry, opts?: { typeArgs?: AST.TypeExpression[]; interfaceArgs?: AST.TypeExpression[]; typeArgMap?: Map<string, AST.TypeExpression>; subjectType?: AST.TypeExpression }): Map<string, AST.TypeExpression> | null;
    matchTypeExpressionTemplate(template: AST.TypeExpression, actual: AST.TypeExpression, genericNames: Set<string>, bindings: Map<string, AST.TypeExpression>): boolean;
    expandImplementationTargetVariants(target: AST.TypeExpression): Array<{ typeName: string; argTemplates: AST.TypeExpression[]; signature: string }>;
    measureTemplateSpecificity(t: AST.TypeExpression, genericNames: Set<string>): number;
    attachDefaultInterfaceMethods(
      imp: AST.ImplementationDefinition,
      funcs: Map<string, Extract<RuntimeValue, { kind: "function" | "function_overload" }>>,
    ): void;
    buildSelfTypePatternBindings(
      iface: AST.InterfaceDefinition,
      targetType: AST.TypeExpression,
    ): Map<string, AST.TypeExpression>;
    createDefaultMethodFunction(
      sig: AST.FunctionSignature,
      env: Environment,
      targetType: AST.TypeExpression,
      typeBindings?: Map<string, AST.TypeExpression>,
    ): Extract<RuntimeValue, { kind: "function" }> | null;
    substituteSelfTypeExpression(t: AST.TypeExpression | undefined, target: AST.TypeExpression): AST.TypeExpression | undefined;
    substituteSelfInPattern(pattern: AST.Pattern, target: AST.TypeExpression): AST.Pattern;
  }
}

export function applyImplResolutionAugmentations(cls: typeof Interpreter): void {
  applyImplConstraintAugmentations(cls);
  applyImplCandidateAugmentations(cls);
  applyImplSpecificityAugmentations(cls);
  applyImplDefaultAugmentations(cls);
}
