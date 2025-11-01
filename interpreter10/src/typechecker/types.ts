import type * as AST from "../ast";

export type PrimitiveName =
  | "bool"
  | "char"
  | "string"
  | "i32"
  | "f64"
  | "nil"
  | "void";

export type TypeInfo =
  | { kind: "unknown" }
  | { kind: "primitive"; name: PrimitiveName }
  | {
      kind: "struct";
      name: string;
      typeArguments: TypeInfo[];
      definition?: AST.StructDefinition;
    }
  | {
      kind: "interface";
      name: string;
      typeArguments: TypeInfo[];
      definition?: AST.InterfaceDefinition;
    }
  | {
      kind: "function";
      parameters: TypeInfo[];
      returnType: TypeInfo;
      generics?: AST.GenericParameter[];
    }
  | { kind: "nullable"; inner: TypeInfo }
  | { kind: "result"; inner: TypeInfo }
  | { kind: "union"; members: TypeInfo[] };

export type InferenceMap = Map<AST.Node, TypeInfo>;

export const unknownType: TypeInfo = { kind: "unknown" };

export function primitiveType(name: PrimitiveName): TypeInfo {
  return { kind: "primitive", name };
}

export function isUnknown(type: TypeInfo | undefined | null): boolean {
  return !type || type.kind === "unknown";
}

export function isBoolean(type: TypeInfo): boolean {
  return type.kind === "primitive" && type.name === "bool";
}

export function isNumeric(type: TypeInfo): boolean {
  return type.kind === "primitive" && (type.name === "i32" || type.name === "f64");
}

export function describe(type: TypeInfo): string {
  return formatType(type);
}

export function formatType(type: TypeInfo): string {
  switch (type.kind) {
    case "unknown":
      return "Unknown";
    case "primitive":
      return type.name;
    case "struct": {
      const args = Array.isArray(type.typeArguments) ? type.typeArguments.map(formatType).filter(Boolean) : [];
      return args.length > 0 ? [type.name, ...args].join(" ") : type.name;
    }
    case "interface": {
      const args = Array.isArray(type.typeArguments) ? type.typeArguments.map(formatType).filter(Boolean) : [];
      return args.length > 0 ? [type.name, ...args].join(" ") : type.name;
    }
    case "function": {
      const params = type.parameters.map((param) => formatType(param)).join(", ");
      const returnType = formatType(type.returnType);
      return `fn(${params}) -> ${returnType}`;
    }
    case "nullable":
      return `${formatType(type.inner)}?`;
    case "result":
      return `Result ${formatType(type.inner)}`;
    case "union": {
      const members = type.members.map((member) => formatType(member));
      return members.join(" | ");
    }
    default:
      return "Unknown";
  }
}
