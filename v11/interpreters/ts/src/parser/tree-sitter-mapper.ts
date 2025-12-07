import * as AST from "../ast";
import type { Expression, FunctionCall, Identifier, ImportStatement, Module, Statement } from "../ast";
import {
  MapperError,
  Node,
  annotate,
  clearMapperOrigin,
  createParseContext,
  inheritMetadata,
  isIgnorableNode,
  parseIdentifier,
  pruneUndefined,
  setActiveParseContext,
  setMapperOrigin,
} from "./shared";
import { registerDefinitionParsers } from "./definitions";
import { registerExpressionParsers } from "./expressions";
import { registerImportParsers } from "./imports";
import { registerPatternParsers } from "./patterns";
import { registerStatementParsers } from "./statements";
import { registerTypeParsers } from "./types";

export function mapSourceFile(root: Node, source: string, origin?: string): Module {
  if (!root) {
    throw new MapperError("parser: missing root node");
  }
  if (root.type !== "source_file") {
    throw new MapperError(`parser: unexpected root node ${root.type}`);
  }
  if ((root as unknown as { hasError?: boolean }).hasError) {
    throw new MapperError("parser: syntax errors present");
  }

  const context = createParseContext(source);
  registerImportParsers(context);
  registerTypeParsers(context);
  registerDefinitionParsers(context);
  registerPatternParsers(context);
  registerStatementParsers(context);
  registerExpressionParsers(context);

  setActiveParseContext(context);
  setMapperOrigin(origin);

  try {
    let packageStmt: AST.PackageStatement | undefined;
    const imports: ImportStatement[] = [];
    const body: Statement[] = [];

    for (let i = 0; i < root.namedChildCount; i++) {
      const node = root.namedChild(i);
      if (!node || isIgnorableNode(node)) continue;
      switch (node.type) {
        case "package_statement":
          packageStmt = context.parsePackageStatement(node);
          break;
        case "import_statement": {
          const kindNode = node.childForFieldName("kind");
          if (!kindNode) {
            throw new MapperError("parser: import missing kind");
          }
          const pathNode = node.childForFieldName("path");
          const path = context.parseQualifiedIdentifier(pathNode);
          const clauseNode = node.childForFieldName("clause");
          const clause = context.parseImportClause(clauseNode);
          const aliasNode = node.childForFieldName("alias");
          const alias: Identifier | undefined = aliasNode ? parseIdentifier(aliasNode, context.source) : undefined;

          if (alias && clause.selectors && clause.selectors.length > 0) {
            throw new MapperError("parser: alias cannot be combined with selectors");
          }
          if (alias && clause.isWildcard) {
            throw new MapperError("parser: alias cannot be combined with wildcard imports");
          }
          if (kindNode.type === "import") {
            imports.push(
              annotate(AST.importStatement(path, clause.isWildcard, clause.selectors, alias), node) as ImportStatement,
            );
          } else if (kindNode.type === "dynimport") {
            body.push(annotate(AST.dynImportStatement(path, clause.isWildcard, clause.selectors, alias), node));
          } else {
            throw new MapperError(`parser: unsupported import kind ${kindNode.type}`);
          }
          break;
        }
        default: {
          if (!node.isNamed) continue;
          const stmt = context.parseStatement(node);
          if (!stmt) {
            throw new MapperError(`parser: unsupported top-level node ${node.type}`);
          }
          if (stmt.type === "LambdaExpression" && body.length > 0) {
            const prev = body[body.length - 1];
            if (prev.type === "AssignmentExpression") {
              const rhs = prev.right;
              if (rhs.type === "FunctionCall") {
                rhs.arguments.push(stmt);
                rhs.isTrailingLambda = true;
                continue;
              }
              if ((rhs as Expression).type) {
                const call = inheritMetadata(AST.functionCall(rhs as Expression, [], undefined, true), rhs as Expression, stmt);
                call.arguments.push(stmt);
                prev.right = call;
                continue;
              }
            }
            if (prev.type === "FunctionCall") {
              const call = prev as FunctionCall;
              if (call.arguments.length === 0 || call.arguments[call.arguments.length - 1] !== stmt) {
                call.arguments.push(stmt);
              }
              call.isTrailingLambda = true;
              continue;
            }
            if ((prev as Expression).type) {
              const call = inheritMetadata(
                AST.functionCall(prev as Expression, [], undefined, true),
                prev as Expression,
                stmt,
              );
              call.arguments.push(stmt);
              body[body.length - 1] = call;
              continue;
            }
          }
          body.push(stmt);
          break;
        }
      }
    }

    const module = annotate(AST.module(body, imports, packageStmt), root);
    return pruneUndefined(module);
  } finally {
    clearMapperOrigin();
    setActiveParseContext(undefined);
  }
}
