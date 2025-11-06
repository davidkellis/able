import * as AST from "./ast";
import { Environment } from "./environment";
import type {
  AbleArray,
  AbleError,
  AbleFunction,
  AbleImplementationDefinition,
  AbleInterfaceDefinition,
  AbleIterator,
  AbleMethodsCollection,
  AblePrimitive,
  AbleRange,
  AbleStructDefinition,
  AbleStructInstance,
  AbleUnionDefinition,
  AbleValue,
} from "./runtime";
import {
  IteratorEnd,
  createArrayIterator,
  createError,
  createRangeIterator,
  hasKind,
  isAbleArray,
  isAbleFunction,
  isAblePrimitive,
  isAbleRange,
  isAbleStructDefinition,
  isAbleStructInstance,
} from "./runtime";
import { evaluateForLoop, evaluateWhileLoop } from "./evaluate-loops";
import { isTruthy, valueToString } from "./value-utils";
import { BreakSignal, RaiseSignal, ReturnSignal } from "./signals";
import { evaluateNode } from "./evaluate-dispatcher";
import { evaluateBinaryExpression } from "./evaluate-binary";
import { bindMethod, evaluateFunctionCall, executeFunction, findMethod } from "./evaluate-call";
import { evaluateMatchExpression, matchPattern } from "./evaluate-match";

// --- Interpreter ---

export class Interpreter {
  private globalEnv: Environment = new Environment();

  // Store builtins separately for potential import resolution
  private builtins: Map<string, AbleValue> = new Map();

  // Storage for runtime definitions
  // Key: Interface Name (string)
  private interfaces: Map<string, AbleInterfaceDefinition> = new Map();
  // Key: Type Name (string) -> Interface Name (string) -> Implementation
  private implementations: Map<string, Map<string, AbleImplementationDefinition>> = new Map();
  // Key: Type Name (string) -> Methods Collection
  private inherentMethods: Map<string, AbleMethodsCollection> = new Map();
  // Key: Union Name (string) -> Union Definition
  private unions: Map<string, AbleUnionDefinition> = new Map();

  constructor() {
    this.defineBuiltins();
    // Add builtins to global scope by default for now
    this.builtins.forEach((value, name) => {
      this.globalEnv.define(name, value);
    });
  }

  private defineBuiltins(): void {
    // Define builtins in the 'builtins' map
    // Example: print function (assuming it belongs to a conceptual 'io' module)
    const printFunc = {
      kind: "function",
      node: null as any, // Mark as native
      closureEnv: this.globalEnv, // Native functions don't capture user env
      apply: (args: AbleValue[]) => {
        console.log(...args.map((v) => this.valueToString(v)));
        return { kind: "nil", value: null };
      },
    } as any;
    this.builtins.set("io.print", printFunc); // Use a qualified name internally
    this.builtins.set("print", printFunc); // Also provide unqualified for direct access/simplicity for now

    // Define other builtins similarly with qualified names if applicable
    this.builtins.set("math.sqrt", {
      /* ... native sqrt ... */
    } as any);
    this.builtins.set("math.pow", {
      /* ... native pow ... */
    } as any);
    // ... other builtins ...
    this.builtins.set("divide", {
      /* ... native divide potentially raising error ... */
    } as any); // Example global builtin
  }

  interpretModule(moduleNode: AST.Module): void {
    // 1. Handle Package Declaration (if any) - Placeholder
    if (moduleNode.package) {
      // Correct access using 'path' based on AST definition
      console.log(`Interpreting package: ${moduleNode.package.namePath.map((id) => id.name).join(".")}`);
    }

    // 2. Handle Imports
    const moduleEnv = new Environment(this.globalEnv);
    this.processImports(moduleNode.imports, moduleEnv);

    // 3. Evaluate Definitions First
    const definitionTypes = new Set(["FunctionDefinition", "StructDefinition", "UnionDefinition", "InterfaceDefinition", "ImplementationDefinition", "MethodsDefinition"]);
    const definitions = moduleNode.body.filter((stmt) => definitionTypes.has(stmt.type));
    const otherStatements = moduleNode.body.filter((stmt) => !definitionTypes.has(stmt.type));

    try {
      // Evaluate all definitions in the module environment
      for (const def of definitions) {
        // We know these are definition types due to the filter above
        this.evaluate(def, moduleEnv);
      }

      // --- Evaluate Top-Level Assignments --- // MODIFIED: Evaluate in moduleEnv
      for (const stmt of otherStatements) {
        if (stmt.type === "AssignmentExpression" && stmt.operator === ":=") {
          this.evaluate(stmt, moduleEnv); // Evaluate in the module environment
        }
      }

      // 4. Find 'main' function (optional)
      let mainFunc: AbleValue | undefined;
      try {
        mainFunc = moduleEnv.get("main"); // Look in moduleEnv
      } catch (e) {
        // 'main' not defined is okay
      }

      // 5. Execute 'main' OR evaluate remaining top-level statements
      if (mainFunc && isAbleFunction(mainFunc)) {
        console.log("--- Running main function ---"); // Indicate main execution
        this.executeFunction(mainFunc, [], moduleEnv); // Call main with moduleEnv as call site env
        console.log("--- main function finished ---");
      } else {
        if (mainFunc) {
          console.warn("Warning: 'main' was found but is not a function.");
        }
        // If no main, evaluate other top-level statements sequentially
        console.log("--- Evaluating top-level statements ---"); // Indicate script-like execution
        for (const stmt of otherStatements) {
          this.evaluate(stmt, moduleEnv); // Evaluate for side effects or results
        }
        console.log("--- Top-level statements finished ---");
      }
    } catch (error) {
      // Catch runtime errors from the interpreter itself (including uncaught signals)
      // Removed duplicated definition check logic from here
      if (error instanceof RaiseSignal) {
        console.error("Uncaught Exception:", this.valueToString(error.value));
      } else if (error instanceof ReturnSignal || error instanceof BreakSignal) {
        console.error("Interpreter Error: Unexpected return/break at module level.");
      } else if (error instanceof Error) {
        // Catch standard JS errors too
        console.error("Interpreter Runtime Error:", error.message, error.stack);
      } else {
        console.error("Unknown Interpreter Error:", error);
      }
    }
  }

  // Basic import processing - needs expansion for actual module loading
  private processImports(imports: AST.ImportStatement[], environment: Environment): void {
    for (const imp of imports) {
      // Correct access using 'path' based on AST definition
      const modulePath = imp.packagePath.map((id) => id.name).join(".");
      // For now, only handle selective import of builtins like io.print
      if (imp.selectors) {
        for (const selector of imp.selectors) {
          // Correct access using 'name' based on AST definition
          const originalName = selector.name.name;
          const alias = selector.alias?.name ?? originalName;
          const qualifiedName = `${modulePath}.${originalName}`;

          // Check builtins map
          if (this.builtins.has(qualifiedName)) {
            environment.define(alias, this.builtins.get(qualifiedName)!);
          } else if (this.builtins.has(originalName) && modulePath === "core") {
            // Allow importing core builtins unqualified?
            environment.define(alias, this.builtins.get(originalName)!);
          } else {
            console.warn(`Interpreter Warning: Cannot find imported item '${qualifiedName}' (or '${originalName}').`);
          }
        }
      } else if (imp.alias) {
        // import module as alias - TBD: Need module objects
        console.warn(`Interpreter Warning: Aliased module import 'import ${modulePath} as ${imp.alias.name}' not fully implemented.`);
      } else {
        // import module; - TBD: Need module objects
        console.warn(`Interpreter Warning: Module import 'import ${modulePath}' not fully implemented.`);
      }
    }
  }

  // Evaluates statements, returning the last value. Throws signals.
  private evaluateStatements(statements: AST.Statement[], environment: Environment): AbleValue {
    let lastValue: AbleValue = { kind: "nil", value: null }; // Default result if block is empty or ends with non-expression

    for (const stmt of statements) {
      // Evaluate each statement. Signals are thrown and caught by callers.
      lastValue = this.evaluate(stmt, environment);

      // If the statement was a definition, the block's value shouldn't become nil
      // Reset lastValue to nil only if it wasn't a definition? This seems complex.
      // Let's stick to the simpler rule: the value of the last evaluated statement/expression counts.
      // Definitions evaluate to nil, so a block ending in a definition returns nil.
    }
    return lastValue; // Return the value of the last statement/expression evaluated
  }

  evaluate = evaluateNode;

  // --- Specific Evaluators ---

  private evaluateBlockExpression(node: AST.BlockExpression, environment: Environment): AbleValue {
    const blockEnv = new Environment(environment); // Create new scope for the block
    // evaluateStatements returns the value of the last statement/expression
    // Signals (Return, Raise, Break) are thrown by evaluateStatements/evaluate
    try {
      return this.evaluateStatements(node.body, blockEnv);
    } catch (signal) {
      // Block expressions themselves don't handle signals, they propagate them up.
      // RescueExpression is the primary mechanism for catching RaiseSignal.
      throw signal;
    }
  }

  private evaluateUnaryExpression(node: AST.UnaryExpression, environment: Environment): AbleValue {
    const operand = this.evaluate(node.operand, environment);

    switch (node.operator) {
      case "-":
        // Add checks for other numeric kinds (i8, i16, etc.)
        if (hasKind(operand, "i32") && typeof operand.value === "number") return { kind: "i32", value: -operand.value };
        if (hasKind(operand, "f64") && typeof operand.value === "number") return { kind: "f64", value: -operand.value };
        if (hasKind(operand, "i64") && typeof operand.value === "bigint") return { kind: "i64", value: -operand.value };
        // Add other bigint types (i128, u64, u128 - though negation might change type for unsigned)
        throw new Error(`Interpreter Error: Unary '-' not supported for type ${operand?.kind ?? typeof operand}`);
      case "!":
        if (hasKind(operand, "bool")) {
          return { kind: "bool", value: !operand.value };
        }
        throw new Error(`Interpreter Error: Unary '!' not supported for type ${operand?.kind ?? typeof operand}`);
      case "~":
        // Add checks for all integer kinds
        if (hasKind(operand, "i32") && typeof operand.value === "number") return { kind: "i32", value: ~operand.value };
        if (hasKind(operand, "i64") && typeof operand.value === "bigint") return { kind: "i64", value: ~operand.value };
        // Add other integer types
        throw new Error(`Interpreter Error: Unary '~' not supported for type ${operand?.kind ?? typeof operand}`);
    }
    throw new Error(`Interpreter Error: Unknown unary operator ${node.operator}`);
  }

  private evaluateBinaryExpression = evaluateBinaryExpression;

  private evaluateAssignmentExpression(node: AST.AssignmentExpression, environment: Environment): AbleValue {
    // For compound assignment, evaluate LHS first to know the target
    let currentLhsValue: AbleValue | undefined = undefined;
    let lhsNodeForCompound: AST.Expression | undefined = undefined; // Store LHS node for compound eval if needed

    if (node.operator !== ":=" && node.operator !== "=") {
      if (node.left.type === "Identifier") {
        currentLhsValue = environment.get(node.left.name);
        lhsNodeForCompound = node.left; // Store the identifier node
      } else if (node.left.type === "MemberAccessExpression") {
        // Evaluate object and get current member value *before* evaluating RHS
        currentLhsValue = this.evaluateMemberAccess(node.left, environment);
        lhsNodeForCompound = node.left; // Store the member access node
      } else {
        throw new Error(`Interpreter Error: Invalid LHS for compound assignment.`);
      }
    }

    let valueToAssign = this.evaluate(node.right, environment);

    // Perform operation for compound assignment
    if (currentLhsValue && lhsNodeForCompound && node.operator !== ":=" && node.operator !== "=") {
      const binaryOp = node.operator.slice(0, -1); // e.g., '+' from '+='

      // --- Re-implement compound logic directly using values ---
      const leftVal = currentLhsValue;
      const rightVal = valueToAssign; // RHS value already evaluated
      const op = binaryOp;

      // This logic needs to mirror evaluateBinaryExpression carefully based on op, left.kind, right.kind
      if (op === "+" && leftVal.kind === "i32" && rightVal.kind === "i32") {
        valueToAssign = { kind: "i32", value: leftVal.value + rightVal.value };
      } else if (op === "+" && leftVal.kind === "string" && rightVal.kind === "string") {
        valueToAssign = { kind: "string", value: leftVal.value + rightVal.value };
      }
      // Add many more cases for other operators and type combinations...
      else {
        throw new Error(
          `Interpreter Error: Compound operator ${node.operator} not fully implemented for ${leftVal?.kind ?? typeof leftVal} and ${rightVal?.kind ?? typeof rightVal}`
        );
      }
      // --- End re-implementation ---
    }

    // Perform the assignment/definition
    if (node.left.type === "Identifier") {
      const name = node.left.name;
      if (node.operator === ":=") {
        environment.define(name, valueToAssign);
      } else {
        // '=', '+=' etc.
        environment.assign(name, valueToAssign);
      }
    } else if (node.left.type === "MemberAccessExpression") {
      if (node.operator === ":=") {
        throw new Error(`Interpreter Error: Cannot use ':=' with member access.`);
      }
      const obj = this.evaluate(node.left.object, environment);
      const member = node.left.member;

      if (isAbleStructInstance(obj)) {
        if (member.type === "Identifier") {
          // Named field access
          if (!(obj.values instanceof Map)) throw new Error("Interpreter Error: Expected named fields map for struct instance.");
          // Check field exists before setting
          if (!obj.definition.definitionNode.fields.some((f) => f.name?.name === member.name)) {
            throw new Error(`Interpreter Error: Struct '${obj.definition.name}' has no field named '${member.name}'.`);
          }
          // Use type guard
          if (obj.values instanceof Map) {
            obj.values.set(member.name, valueToAssign);
          } else {
            // This case should be prevented by the earlier check, but satisfies TS
            throw new Error("Internal Interpreter Error: Struct instance value map type mismatch.");
          }
        } else {
          // Positional field access (IntegerLiteral)
          if (!Array.isArray(obj.values)) throw new Error("Interpreter Error: Expected positional fields array for struct instance.");
          const index = Number(member.value); // Assuming integer literal for index
          if (index < 0 || index >= obj.values.length) {
            throw new Error(`Interpreter Error: Index ${index} out of bounds for struct '${obj.definition.name}'.`);
          }
          obj.values[index] = valueToAssign; // Use valueToAssign
        }
      } else if (isAbleArray(obj)) {
        // Handle array mutation obj.elements[index] = valueToAssign;
        if (member.type !== "IntegerLiteral") throw new Error("Interpreter Error: Array index must be an integer literal.");
        const index = Number(member.value);
        if (index < 0 || index >= obj.elements.length) {
          throw new Error(`Interpreter Error: Array index ${index} out of bounds (length ${obj.elements.length}).`);
        }
        obj.elements[index] = valueToAssign;
      } else {
        throw new Error(`Interpreter Error: Cannot assign to member of type ${obj?.kind ?? typeof obj}.`);
      }
    } else {
      // Handle destructuring assignment
      if (node.left.type === "StructPattern" || node.left.type === "ArrayPattern") {
        if (node.operator !== ":=" && node.operator !== "=") {
          throw new Error(`Interpreter Error: Compound assignment not supported with destructuring patterns.`);
        }
        this.evaluatePatternAssignment(node.left, valueToAssign, environment, node.operator === ":=");
        return valueToAssign;
      }

      // Destructuring assignment (:= or =)
      if (node.operator !== ":=" && node.operator !== "=") {
        throw new Error(`Interpreter Error: Compound assignment not supported with destructuring patterns.`);
      }
      this.evaluatePatternAssignment(node.left, valueToAssign, environment, node.operator === ":=");
    }

    return valueToAssign; // Assignment expression evaluates to the assigned value
  }

  // Helper for destructuring assignment (recursive)
  private evaluatePatternAssignment(pattern: AST.Pattern, value: AbleValue, environment: Environment, isDeclaration: boolean): void {
    switch (pattern.type) {
      case "Identifier":
        if (isDeclaration) {
          environment.define(pattern.name, value);
        } else {
          // Try assigning, will throw if not found
          environment.assign(pattern.name, value);
        }
        break;
      case "WildcardPattern":
        // Do nothing, ignore the value
        break;
      case "LiteralPattern": {
        const patternVal = this.evaluate(pattern.literal, environment);
        // Ensure both are primitives before comparing .value
        if ("value" in value && "value" in patternVal) {
          // TODO: Implement proper deep equality check based on Eq interface later
          if (value.kind !== patternVal.kind || value.value !== patternVal.value) {
            throw new Error(`Interpreter Error: Pattern mismatch. Expected literal ${this.valueToString(patternVal)}, got ${this.valueToString(value)}.`);
          }
        } else if (hasKind(value, "nil") && hasKind(patternVal, "nil")) {
          // Specifically allow nil to match nil literal
        } else {
          throw new Error(`Interpreter Error: Cannot match literal pattern ${this.valueToString(patternVal)} against non-primitive value ${this.valueToString(value)}.`);
        }
        break;
      }
      case "StructPattern": {
        // Use type guard
        if (!isAbleStructInstance(value)) {
           throw new Error(`Interpreter Error: Cannot destructure non-struct value (got ${value?.kind ?? typeof value}) with a struct pattern.`);
        }
        // Optional: Check value.definition.name against pattern.structType?.name
        if (pattern.structType && value.definition.name !== pattern.structType.name) {
          throw new Error(`Interpreter Error: Struct type mismatch. Expected ${pattern.structType.name}, got ${value.definition.name}.`);
        }

        if (pattern.isPositional) {
            if (!Array.isArray(value.values)) throw new Error("Interpreter Error: Expected positional struct values.");
            if (pattern.fields.length !== value.values.length) {
              throw new Error(`Interpreter Error: Pattern field count (${pattern.fields.length}) does not match struct field count (${value.values.length}).`);
            }
            for (let i = 0; i < pattern.fields.length; i++) {
              // Ensure field pattern and value exist before recursing
              const fieldPatternNode = pattern.fields[i];
              const fieldValue = value.values[i];
              if (!fieldPatternNode?.pattern || fieldValue === undefined) {
                 throw new Error(`Internal Interpreter Error: Missing pattern or value at index ${i} during positional struct assignment.`);
              }
              this.evaluatePatternAssignment(fieldPatternNode.pattern, fieldValue, environment, isDeclaration);
            }
        } else {
          // Named fields
          if (!(value.values instanceof Map)) throw new Error("Interpreter Error: Expected named struct values map.");
          const matchedFields = new Set<string>();
          for (const fieldPattern of pattern.fields) {
            if (!fieldPattern.fieldName) throw new Error("Interpreter Error: Missing field name in named struct pattern.");
            const fieldName = fieldPattern.fieldName.name;
            if (!value.values.has(fieldName)) {
              throw new Error(`Interpreter Error: Struct instance does not have field '${fieldName}' for destructuring.`);
            }
            const fieldValue = value.values.get(fieldName)!;
            this.evaluatePatternAssignment(fieldPattern.pattern, fieldValue, environment, isDeclaration);
            matchedFields.add(fieldName);
          }
          // Optional: Check for extra fields if pattern should be exhaustive? Spec doesn't say.
        }
        break;
      }
      case "ArrayPattern": {
        // Use type guard
        if (!isAbleArray(value)) {
          throw new Error(`Interpreter Error: Cannot destructure non-array value (got ${value?.kind ?? typeof value}) with an array pattern.`);
        }
        const minLen = pattern.elements.length;
        const hasRest = !!pattern.restPattern;
        if (!hasRest && value.elements.length !== minLen) {
          throw new Error(`Interpreter Error: Array pattern length (${minLen}) does not match value length (${value.elements.length}).`);
        }
        if (value.elements.length < minLen) {
          throw new Error(`Interpreter Error: Array value length (${value.elements.length}) is less than pattern length (${minLen}).`);
        }
        // Match fixed elements
        for (let i = 0; i < minLen; i++) {
          // Ensure element pattern and value exist before recursing
          const elemPattern = pattern.elements[i];
          const elemValue = value.elements[i];
           if (!elemPattern || elemValue === undefined) {
                 throw new Error(`Internal Interpreter Error: Missing pattern or value at index ${i} during array assignment.`);
           }
          this.evaluatePatternAssignment(elemPattern, elemValue, environment, isDeclaration);
        }
        // Match rest element
        if (hasRest && pattern.restPattern) {
          // Ensure restPattern is not undefined
          const restValue: AbleArray = { kind: "array", elements: value.elements.slice(minLen) };
          if (pattern.restPattern.type === "Identifier") {
            this.evaluatePatternAssignment(pattern.restPattern, restValue, environment, isDeclaration);
          } // Ignore if rest is Wildcard (already handled by hasRest check)
        }
        break;
      }
      default:
        // Use type assertion for exhaustive check
        const _exhaustiveCheck: never = pattern;
        throw new Error(`Interpreter Error: Unsupported pattern type for assignment: ${(_exhaustiveCheck as any).type}`);
    }
  }

  private evaluateFunctionDefinition(node: AST.FunctionDefinition, environment: Environment): void {
    const func: AbleFunction = {
      kind: "function",
      node: node,
      closureEnv: environment, // Capture the current environment
    };
    environment.define(node.id.name, func);
  }

  private evaluateStructDefinition(node: AST.StructDefinition, environment: Environment): void {
    const structDef: AbleStructDefinition = {
      kind: "struct_definition",
      name: node.id.name,
      definitionNode: node,
    };
    environment.define(node.id.name, structDef);
  }

  private evaluateStructLiteral(node: AST.StructLiteral, environment: Environment): AbleStructInstance {
    const structDefVal = node.structType ? environment.get(node.structType.name) : null;
    // TODO: Infer struct type if node.structType is null (requires type checking info)
    if (!structDefVal || !isAbleStructDefinition(structDefVal)) {
      throw new Error(`Interpreter Error: Cannot instantiate unknown or non-struct type '${node.structType?.name}'.`);
    }
    // const structDef = structDefVal as AbleStructDefinition; // No longer need cast
    const structDef = structDefVal;

    let instanceValues: AbleValue[] | Map<string, AbleValue>;

    if (node.isPositional) {
      if (structDef.definitionNode.kind !== "positional") {
        throw new Error(`Interpreter Error: Cannot use positional literal for non-positional struct '${structDef.name}'.`);
      }
      if (node.fields.length !== structDef.definitionNode.fields.length) {
        throw new Error(`Interpreter Error: Positional literal field count (${node.fields.length}) does not match struct definition (${structDef.definitionNode.fields.length}).`);
      }
      instanceValues = node.fields.map((field) => this.evaluate(field.value, environment));
    } else {
      // Named fields
      if (structDef.definitionNode.kind !== "named") {
        throw new Error(`Interpreter Error: Cannot use named literal for non-named struct '${structDef.name}'.`);
      }
      instanceValues = new Map<string, AbleValue>();
      const providedFields = new Set<string>();

      // Handle functional update source first
      if (node.functionalUpdateSource) {
        const sourceVal = this.evaluate(node.functionalUpdateSource, environment);
        if (!isAbleStructInstance(sourceVal) || sourceVal.definition !== structDef) {
          throw new Error(`Interpreter Error: Functional update source must be an instance of the same struct '${structDef.name}'.`);
        }
        // Ensure source has named fields (Map) before iterating
        if (!(sourceVal.values instanceof Map)) {
          throw new Error("Interpreter Error: Functional update source must have named fields.");
        }
        // Now safe to iterate
        sourceVal.values.forEach((val, key) => (instanceValues as Map<string, AbleValue>).set(key, val)); // Copy source values
      }

      // Apply explicit fields
      for (const field of node.fields) {
        // Handle shorthand { name } where field.name is set but field.value might be the same identifier
        const fieldName = field.name?.name;
        if (!fieldName) throw new Error("Interpreter Error: Missing field name in named struct literal initializer (should not happen if parser validates).");
        if (providedFields.has(fieldName)) throw new Error(`Interpreter Error: Field '${fieldName}' provided more than once in struct literal.`);

        const value = this.evaluate(field.value, environment);
        instanceValues.set(fieldName, value);
        providedFields.add(fieldName);
      }

      // Check if all required fields are present (only needed if no functional update source)
      if (!node.functionalUpdateSource) {
        for (const defField of structDef.definitionNode.fields) {
          if (defField.name && !instanceValues.has(defField.name.name)) {
            throw new Error(`Interpreter Error: Missing field '${defField.name.name}' in struct literal for '${structDef.name}'.`);
          }
        }
      }
    }

    return {
      kind: "struct_instance",
      definition: structDef,
      values: instanceValues,
    };
  }

  private evaluateMemberAccess(node: AST.MemberAccessExpression, environment: Environment): AbleValue {
    const objectExpr = node.object; // Get the AST node for the object part
    // console.log(`[evaluateMemberAccess] Evaluating object expression:`, objectExpr); // REMOVED DEBUG LOG
    const object = this.evaluate(objectExpr, environment);
    // console.log(`[evaluateMemberAccess] Object evaluated to:`, object); // REMOVED DEBUG LOG

    // Check if object is undefined BEFORE trying to access .kind
    if (object === undefined) {
      // console.error("[evaluateMemberAccess] CRITICAL ERROR: Object evaluated to undefined. AST:", objectExpr); // REMOVED DEBUG LOG
      // Optionally, try to inspect the environment here
      // console.log("[evaluateMemberAccess] Environment keys:", Array.from((environment as any).values.keys()));
      throw new Error("Internal Interpreter Error: Object evaluated to undefined during member access.");
    }

    const memberName = node.member.type === "Identifier" ? node.member.name : node.member.value.toString();
    // console.log(`[evaluateMemberAccess] Accessing member '${memberName}' on object kind '${object.kind}'`);

    // --- Instance Member/Method Access ---
    if (isAbleStructInstance(object)) {
      const member = node.member;
      if (member.type === "Identifier") {
        // Named field access
        if (!(object.values instanceof Map)) throw new Error(`Interpreter Error: Expected named fields map for struct instance '${object.definition.name}'.`);
        const fieldName = member.name;
        if (object.values.has(fieldName)) {
          const fieldValue = object.values.get(fieldName)!;
          // console.log(`[evaluateMemberAccess] Found field '${fieldName}', returning value kind '${fieldValue.kind}'`); // REMOVED DEBUG LOG
          return fieldValue;
        } else {
          // --- Method Call Check ---
          // console.log(`[evaluateMemberAccess] Field '${fieldName}' not found, checking for methods...`); // REMOVED DEBUG LOG
          const method = this.findMethod(object, fieldName);
          if (method) {
            // console.log(`[evaluateMemberAccess] Found method '${fieldName}', returning bound method.`); // REMOVED DEBUG LOG
            // Return a bound method (closure) that includes 'self'
            return this.bindMethod(object, method);
          }
          // --- End Method Call Check ---
          // console.error(`[evaluateMemberAccess] Error: Struct '${object.definition.name}' has no field or method named '${fieldName}'.`); // REMOVED DEBUG LOG
          throw new Error(`Interpreter Error: Struct '${object.definition.name}' has no field or method named '${fieldName}'.`);
        } // <-- Corrected closing brace
      } else {
        // Positional field access (IntegerLiteral)
        if (!Array.isArray(object.values)) throw new Error(`Interpreter Error: Expected positional fields array for struct instance '${object.definition.name}'.`);
        const index = Number(member.value); // Assuming integer literal for index
        if (index < 0 || index >= object.values.length) {
          // console.error(`[evaluateMemberAccess] Error: Index ${index} out of bounds for struct '${object.definition.name}'.`); // REMOVED DEBUG LOG
          throw new Error(`Interpreter Error: Index ${index} out of bounds for struct '${object.definition.name}'.`);
        }
        const positionalValue = object.values[index];
        // Add check for undefined
        if (positionalValue === undefined) {
            throw new Error(`Internal Interpreter Error: Undefined positional value at index ${index} for struct ${object.definition.name}`);
        }
        return positionalValue;
      }
    } else if (isAbleArray(object)) {
      // Handle array indexing
      if (node.member.type !== "IntegerLiteral") throw new Error("Interpreter Error: Array index must be an integer literal.");
      const index = Number(node.member.value);
      if (index < 0 || index >= object.elements.length) {
        // console.error(`[evaluateMemberAccess] Error: Array index ${index} out of bounds (length ${object.elements.length}).`); // REMOVED DEBUG LOG
        throw new Error(`Interpreter Error: Array index ${index} out of bounds (length ${object.elements.length}).`);
      }
      const arrayElement = object.elements[index];
       // Add check for undefined
       if (arrayElement === undefined) {
            throw new Error(`Internal Interpreter Error: Undefined array element at index ${index}`);
       }
      return arrayElement;
    }

    // --- Static Method Access ---
    if (isAbleStructDefinition(object)) {
        if (node.member.type === "Identifier") {
            const staticMethodName = node.member.name;
            const typeName = object.name;
            const inherent = this.inherentMethods.get(typeName);
            if (inherent && inherent.methods.has(staticMethodName)) {
                const staticMethod = inherent.methods.get(staticMethodName)!;
                // TODO: Verify this method is actually static (e.g., no self param?) - relies on correct definition for now
                return staticMethod; // Return the unbound function
            }
            throw new Error(`Interpreter Error: No static method named '${staticMethodName}' found on struct '${typeName}'.`);
        } else {
            throw new Error(`Interpreter Error: Cannot access static member using index.`);
        }
    }

    // --- Instance Method Check for other types (Array, String etc.) ---
    // Check inherent methods first for other types like Array, string
    if (node.member.type === "Identifier") {
        const methodName = node.member.name;
        const method = this.findMethod(object, methodName); // findMethod handles inherent lookups
        if (method) {
          return this.bindMethod(object, method); // Bind self for instance methods
        }
    }
    // --- End Instance Method Check ---

    // If none of the above matched
    throw new Error(`Interpreter Error: Cannot access member '${memberName}' on type ${object?.kind ?? typeof object}.`);
  }

  private evaluateFunctionCall = evaluateFunctionCall;

  private executeFunction = executeFunction;

  private evaluateIfExpression(node: AST.IfExpression, environment: Environment): AbleValue {
    // Evaluate the main 'if' condition
    const ifCondVal = this.evaluate(node.ifCondition, environment);
    if (this.isTruthy(ifCondVal)) {
      // Evaluate body, might throw signal
      return this.evaluate(node.ifBody, environment);
    }

    // Evaluate 'or' clauses
    for (const orClause of node.orClauses) {
      if (orClause.condition) {
        // It's an 'or condition {}'
        const orCondVal = this.evaluate(orClause.condition, environment);
        if (this.isTruthy(orCondVal)) {
          // Evaluate body, might throw signal
          return this.evaluate(orClause.body, environment);
        }
      } else {
        // It's the final 'or {}' (else)
        // Evaluate body, might throw signal
        return this.evaluate(orClause.body, environment);
      }
    }

    // If no conditions were met and no final 'or {}' exists
    return { kind: "nil", value: null };
  }

  private evaluateStringInterpolation(node: AST.StringInterpolation, environment: Environment): AbleValue {
    let result = "";
    for (const part of node.parts) {
      if (part.type === "StringLiteral") {
        result += part.value;
      } else {
        // Evaluate the expression part and convert to string
        const value = this.evaluate(part, environment);
        // TODO: Use Display interface if available
        result += this.valueToString(value);
      }
    }
    return { kind: "string", value: result };
  }

  private evaluateLambdaExpression(node: AST.LambdaExpression, environment: Environment): AbleFunction {
    // Lambdas are just anonymous functions, capture current environment
    const func: AbleFunction = {
      kind: "function",
      node: node,
      closureEnv: environment,
    };
    return func;
  }

  private evaluateRangeExpression(node: AST.RangeExpression, environment: Environment): AbleValue { // Return AbleValue, not AbleRange
    const startVal = this.evaluate(node.start, environment);
    const endVal = this.evaluate(node.end, environment);

    // Add type guards before accessing .value
    if (isAblePrimitive(startVal) && isAblePrimitive(endVal)) {
      // Basic validation - ensure both are numbers or both are bigints
      if (!((typeof startVal.value === "number" && typeof endVal.value === "number") || (typeof startVal.value === "bigint" && typeof endVal.value === "bigint"))) {
        throw new Error(`Interpreter Error: Range boundaries must be both numbers or both bigints. Got ${startVal.kind} and ${endVal.kind}.`);
      }

      // Return an AbleRange *value* object
      return {
        kind: "range", // Add kind property
        start: startVal.value as number | bigint, // Cast is safe due to check above
        end: endVal.value as number | bigint,
        inclusive: node.inclusive,
      } as AbleRange; // Explicitly cast to the interface type
    } else {
      // Use kind property safely after checking if they are primitives
      const startKind = isAblePrimitive(startVal) ? startVal.kind : typeof startVal;
      const endKind = isAblePrimitive(endVal) ? endVal.kind : typeof endVal;
      throw new Error(`Interpreter Error: Range boundaries must be primitive types. Got ${startKind} and ${endKind}.`);
    }
  }

  private evaluateWhileLoop = evaluateWhileLoop;

  private evaluateBreakStatement(node: AST.BreakStatement, environment: Environment): never {
    const value = this.evaluate(node.value, environment);
    // For now, the label isn't used for matching as breakpoint isn't implemented,
    // but we include it in the signal as per the AST/spec.
    throw new BreakSignal(node.label.name, value);
  }

  private evaluateForLoop = evaluateForLoop;

  private evaluateInterfaceDefinition(node: AST.InterfaceDefinition, environment: Environment): void {
    // TODO: Handle generics, where clauses, base interfaces, privacy
    if (this.interfaces.has(node.id.name)) {
      // Allow redefinition for now? Or error?
      console.warn(`Interpreter Warning: Redefining interface '${node.id.name}'.`);
    }
    const ifaceDef: AbleInterfaceDefinition = {
      kind: "interface_definition",
      name: node.id.name,
      definitionNode: node,
    };
    // Store globally for now
    this.interfaces.set(node.id.name, ifaceDef);
    // Also define in current env? Maybe not needed if lookup is global. Let's skip for now.
    // environment.define(node.id.name, ifaceDef);
  }

  private evaluateImplementationDefinition(node: AST.ImplementationDefinition, environment: Environment): void {
    // TODO: Handle generics (<T>), interface args ([A]), where clauses, named impls
    // 1. Find the interface definition
    const ifaceName = node.interfaceName.name;
    const ifaceDef = this.interfaces.get(ifaceName);
    if (!ifaceDef) {
      throw new Error(`Interpreter Error: Cannot implement unknown interface '${ifaceName}'.`);
    }

    // 2. Determine the target type name (simplistic for now)
    // TODO: Handle complex target types (Array T, generics etc.)
    let targetTypeName: string | null = null;
    if (node.targetType.type === "SimpleTypeExpression") {
      targetTypeName = node.targetType.name.name;
    } else {
      // For now, only support simple type targets
      throw new Error(`Interpreter Error: Implementation target type evaluation not fully implemented for ${node.targetType.type}.`);
    }
    if (!targetTypeName) {
      throw new Error(`Interpreter Error: Could not determine target type name for implementation.`);
    }

    // 3. Create runtime method closures
    const methodsMap = new Map<string, AbleFunction>();
    for (const funcDef of node.definitions) {
      const method: AbleFunction = {
        kind: "function",
        node: funcDef,
        closureEnv: environment, // Capture impl block's environment
      };
      methodsMap.set(funcDef.id.name, method);
    }

    // 4. Create and store the implementation definition
    const implDef: AbleImplementationDefinition = {
      kind: "implementation_definition",
      implNode: node,
      interfaceDef: ifaceDef,
      methods: methodsMap,
      closureEnv: environment,
    };

    if (!this.implementations.has(targetTypeName)) {
      this.implementations.set(targetTypeName, new Map());
    }
    const typeImpls = this.implementations.get(targetTypeName)!;

    // TODO: Handle overlapping implementations (named impls, specificity). For now, last one wins.
    if (typeImpls.has(ifaceName) && !node.implName) {
      // Only warn if not a named impl potentially overwriting default
      console.warn(`Interpreter Warning: Overwriting existing implementation of '${ifaceName}' for type '${targetTypeName}'.`);
    }
    // Store using interface name (and potentially implName later)
    typeImpls.set(ifaceName, implDef); // TODO: Use node.implName if present? Needs map structure change.
  }

  private evaluateMethodsDefinition(node: AST.MethodsDefinition, environment: Environment): void {
    // 1. Determine the target type name (simplistic for now)
    let targetTypeName: string | null = null;
    if (node.targetType.type === "SimpleTypeExpression") {
      targetTypeName = node.targetType.name.name;
    } else {
      throw new Error(`Interpreter Error: Methods target type evaluation not fully implemented for ${node.targetType.type}.`);
    }
    if (!targetTypeName) {
      throw new Error(`Interpreter Error: Could not determine target type name for methods block.`);
    }

    // 2. Create runtime method closures
    const methodsMap = new Map<string, AbleFunction>();
    for (const funcDef of node.definitions) {
      const method: AbleFunction = {
        kind: "function",
        node: funcDef,
        closureEnv: environment, // Capture methods block's environment
      };
      methodsMap.set(funcDef.id.name, method);
    }

    // 3. Create and store the methods collection
    const methodsCollection: AbleMethodsCollection = {
      kind: "methods_collection",
      methodsNode: node,
      methods: methodsMap,
      closureEnv: environment,
    };

    // TODO: Handle merging if multiple methods blocks for the same type?
    if (this.inherentMethods.has(targetTypeName)) {
      console.warn(`Warning: Overwriting existing inherent methods for type '${targetTypeName}'.`);
    }
    this.inherentMethods.set(targetTypeName, methodsCollection);
  }

  private evaluateUnionDefinition(node: AST.UnionDefinition, environment: Environment): void {
    // TODO: Handle generics, privacy? Currently AST doesn't support privacy marker here.
    if (this.unions.has(node.id.name)) {
        // Allow redefinition for now? Or error?
        console.warn(`Interpreter Warning: Redefining union '${node.id.name}'.`);
    }
    const unionDef: AbleUnionDefinition = {
        kind: "union_definition",
        name: node.id.name,
        definitionNode: node,
    };
    // Store globally for now, similar to interfaces
    this.unions.set(node.id.name, unionDef);
    // Also define in the current environment so it can be referenced (e.g., potentially later for type checks?)
    environment.define(node.id.name, unionDef);
  }

  // --- Method Lookup & Binding ---

  // Finds a method (inherent or interface) for a given object and name
  private findMethod = findMethod;

  private bindMethod = bindMethod;

  // --- Helpers ---
  // Determine truthiness according to Able rules (Spec TBD, basic version here)
  private isTruthy = isTruthy;

  private valueToString = valueToString;

  private matchPattern = matchPattern;

  private evaluateMatchExpression = evaluateMatchExpression;

  private evaluateRaiseStatement(node: AST.RaiseStatement, environment: Environment): never {
    const errorValue = this.evaluate(node.expression, environment);
    // TODO: Spec discussion - should we enforce wrapping in AbleError?
    // For now, raise the evaluated value directly.
    throw new RaiseSignal(errorValue);
  }

  private evaluateRescueExpression(node: AST.RescueExpression, environment: Environment): AbleValue {
    try {
        // Evaluate the expression that might raise an error
        return this.evaluate(node.monitoredExpression, environment);
    } catch (signal) {
        if (signal instanceof RaiseSignal) {
            // A raise occurred, try to rescue it
            const raisedValue = signal.value;

            for (const clause of node.clauses) {
                 // 1. Attempt to match the pattern against the *raised value*
                 // Pass the current environment for pattern literal evaluation, but the bindings
                 // will go into a new environment derived from the current one.
                const matchEnv = this.matchPattern(clause.pattern, raisedValue, environment);

                if (matchEnv) {
                    // Pattern matched!
                    let guardResult = true;
                    // 2. Evaluate the guard condition (if present) in the match environment
                    if (clause.guard) {
                    const guardValue = this.evaluate(clause.guard, matchEnv); // Use matchEnv with bindings
                    guardResult = this.isTruthy(guardValue);
                    }

                    // 3. If guard passed (or no guard), evaluate the body
                    if (guardResult) {
                    // Evaluate the body in the environment created by the match (with bindings)
                    return this.evaluate(clause.body, matchEnv);
                    }
                }
                // If pattern didn't match or guard failed, continue to the next clause
            }
            // If no rescue clause matched, re-throw the original signal
            throw signal;
        } else {
             // If it wasn't a RaiseSignal (e.g., ReturnSignal, BreakSignal, JS Error),
             // propagate it up. Rescue only catches 'raise'.
            throw signal;
        }
    }
  }
} // <-- End of Interpreter class

// --- Entry Point ---

// Example function to interpret a module AST
export function interpret(moduleNode: AST.Module) {
  const interpreter = new Interpreter();
  interpreter.interpretModule(moduleNode);
}

// Example Usage (in another file or here):
// import sampleModule from './sample1'; // Assuming sample1.ts exports the AST
// interpret(sampleModule);
