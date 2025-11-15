import type { IntegerPrimitive, PrimitiveName } from "./types";

export type IntegerTypeInfo = {
  name: IntegerPrimitive;
  bits: number;
  signed: boolean;
  min: bigint;
  max: bigint;
};

const signedInfo = (name: IntegerPrimitive, bits: number): IntegerTypeInfo => {
  const max = (1n << BigInt(bits - 1)) - 1n;
  const min = -(1n << BigInt(bits - 1));
  return { name, bits, signed: true, min, max };
};

const unsignedInfo = (name: IntegerPrimitive, bits: number): IntegerTypeInfo => {
  const max = (1n << BigInt(bits)) - 1n;
  return { name, bits, signed: false, min: 0n, max };
};

const INTEGER_INFO: Record<IntegerPrimitive, IntegerTypeInfo> = {
  i8: signedInfo("i8", 8),
  i16: signedInfo("i16", 16),
  i32: signedInfo("i32", 32),
  i64: signedInfo("i64", 64),
  i128: signedInfo("i128", 128),
  u8: unsignedInfo("u8", 8),
  u16: unsignedInfo("u16", 16),
  u32: unsignedInfo("u32", 32),
  u64: unsignedInfo("u64", 64),
  u128: unsignedInfo("u128", 128),
};

const SIGNED_SEQUENCE: IntegerPrimitive[] = ["i8", "i16", "i32", "i64", "i128"];
const UNSIGNED_SEQUENCE: IntegerPrimitive[] = ["u8", "u16", "u32", "u64", "u128"];

export function hasIntegerBounds(name: PrimitiveName): name is IntegerPrimitive {
  return Object.prototype.hasOwnProperty.call(INTEGER_INFO, name);
}

export function getIntegerTypeInfo(name: PrimitiveName | null | undefined): IntegerTypeInfo | null {
  if (!name || !hasIntegerBounds(name)) {
    return null;
  }
  return INTEGER_INFO[name];
}

export function integerBounds(name: IntegerPrimitive): { min: bigint; max: bigint } {
  const info = INTEGER_INFO[name];
  return { min: info.min, max: info.max };
}

export function signedIntegerInfos(): IntegerTypeInfo[] {
  return SIGNED_SEQUENCE.map((name) => INTEGER_INFO[name]);
}

export function unsignedIntegerInfos(): IntegerTypeInfo[] {
  return UNSIGNED_SEQUENCE.map((name) => INTEGER_INFO[name]);
}

export function findSmallestSigned(bits: number): IntegerTypeInfo | null {
  for (const name of SIGNED_SEQUENCE) {
    const info = INTEGER_INFO[name];
    if (info.bits >= bits) {
      return info;
    }
  }
  return null;
}

export function findSmallestUnsigned(bits: number): IntegerTypeInfo | null {
  for (const name of UNSIGNED_SEQUENCE) {
    const info = INTEGER_INFO[name];
    if (info.bits >= bits) {
      return info;
    }
  }
  return null;
}

export function widestUnsignedInfo(infos: IntegerTypeInfo[]): IntegerTypeInfo | null {
  let current: IntegerTypeInfo | null = null;
  for (const info of infos) {
    if (!info || info.signed) continue;
    if (!current || info.bits > current.bits) {
      current = info;
    }
  }
  return current;
}
