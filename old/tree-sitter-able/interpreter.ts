import * as TreeSitter from "tree-sitter";
// import { AbleParser } from "./parser";
import { AbleParser } from "./parser";

// Runtime value type for Able primitives
type AbleValue =
  | { kind: "i8"; value: number }
  | { kind: "i16"; value: number }
  | { kind: "i32"; value: number }
  | { kind: "i64"; value: bigint }
  | { kind: "i128"; value: bigint }
  | { kind: "u8"; value: number }
  | { kind: "u16"; value: number }
  | { kind: "u32"; value: number }
  | { kind: "u64"; value: bigint }
  | { kind: "u128"; value: bigint }
  | { kind: "f32"; value: number }
  | { kind: "f64"; value: number }
  | { kind: "string"; value: string }
  | { kind: "bool"; value: boolean }
  | { kind: "char"; value: string } // Single Unicode char as string for simplicity
  | { kind: "nil"; value: null };

export class AbleInterpreter {
  private parser: AbleParser;

  constructor() {
    this.parser = new AbleParser();
  }

  // Interpret a source string and return the evaluated values
  interpret(source: string): AbleValue[] {
    const tree = this.parser.parse(source);
    const results: AbleValue[] = [];
    const cursor = tree.walk();

    // Traverse top-level statements
    if (cursor.gotoFirstChild()) {
      do {
        if (cursor.nodeType === "expression_statement") {
          cursor.gotoFirstChild(); // Move to the expression
          const value = this.evaluateExpression(cursor.currentNode);
          if (value) results.push(value);
          cursor.gotoParent(); // Back to expression_statement
        }
      } while (cursor.gotoNextSibling());
    }

    return results;
  }

  private evaluateExpression(node: TreeSitter.SyntaxNode): AbleValue | null {
    switch (node.type) {
      case "integer_literal":
        return this.parseIntegerLiteral(node.text);
      case "float_literal":
        return this.parseFloatLiteral(node.text);
      case "string_literal":
        return { kind: "string", value: node.text.slice(1, -1) }; // Remove quotes
      case "boolean_literal":
        return { kind: "bool", value: node.text === "true" };
      case "char_literal":
        return { kind: "char", value: node.text.slice(1, -1) }; // Remove quotes
      case "nil_literal":
        return { kind: "nil", value: null };
      default:
        console.warn(`Unsupported expression type: ${node.type}`);
        return null;
    }
  }

  private parseIntegerLiteral(text: string): AbleValue {
    const suffixMatch = text.match(/(i|u)(8|16|32|64|128)$/);
    const baseText = suffixMatch ? text.slice(0, -suffixMatch[0].length) : text;
    const suffix = suffixMatch ? suffixMatch[0] : "i32";

    let value: number | bigint;
    if (baseText.startsWith("0x") || baseText.startsWith("-0x")) {
      value = BigInt(baseText.replace(/_/g, "")); // Hex
    } else if (baseText.startsWith("0o") || baseText.startsWith("-0o")) {
      value = BigInt(baseText.replace(/_/g, "").replace("0o", "").replace("-0o", "-")); // Octal
    } else if (baseText.startsWith("0b") || baseText.startsWith("-0b")) {
      value = BigInt(baseText.replace(/_/g, "").replace("0b", "").replace("-0b", "-")); // Binary
    } else {
      value = baseText.includes("_")
        ? BigInt(baseText.replace(/_/g, ""))
        : parseInt(baseText, 10);
    }

    const kind = suffix || (value >= 0 ? "u32" : "i32"); // Default to i32 for signed, u32 for unsigned if no suffix
    if (["i64", "u64", "i128", "u128"].includes(kind)) {
      return {
        kind: kind as "i64" | "u64" | "i128" | "u128",
        value: BigInt(value),
      };
    }
    return {
      kind: kind as "i8" | "i16" | "i32" | "u8" | "u16" | "u32",
      value: Number(value),
    };
  }

  private parseFloatLiteral(text: string): AbleValue {
    const suffixMatch = text.match(/_(f32|f64)$/);
    const baseText = suffixMatch ? text.slice(0, -suffixMatch[0].length) : text;
    const value = parseFloat(baseText.replace(/_/g, ""));
    const kind = suffixMatch ? (suffixMatch[1] as "f32" | "f64") : "f64"; // Default to f64
    return { kind, value };
  }
}

export default AbleInterpreter;
