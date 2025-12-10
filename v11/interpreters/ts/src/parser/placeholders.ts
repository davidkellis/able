import * as AST from "../ast";
import { annotateExpressionNode, MapperError, sliceText, type Node } from "./shared";

export function parsePlaceholderExpression(node: Node, source: string) {
  const raw = sliceText(node, source).trim();
  if (raw === "@") {
    return annotateExpressionNode(AST.placeholderExpression(1), node);
  }
  if (raw.startsWith("@")) {
    const value = raw.slice(1);
    if (value === "") {
      return annotateExpressionNode(AST.placeholderExpression(1), node);
    }
    const index = Number.parseInt(value, 10);
    if (!Number.isInteger(index) || index <= 0) {
      throw new MapperError(`parser: invalid placeholder index ${raw}`);
    }
    return annotateExpressionNode(AST.placeholderExpression(index), node);
  }
  throw new MapperError(`parser: unsupported placeholder token ${raw}`);
}
