import * as AST from "./ast"; // Import our AST definitions

// --- Runtime Values ---
// More closely aligned with spec types
type AblePrimitive =
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
  | { kind: "char"; value: string } // Single Unicode char as string
  | { kind: "nil"; value: null }
  | { kind: "void"; value: undefined }; // Represent void

// Represents a runtime function (closure)
interface AbleFunction {
  kind: "function"; // Added kind for type narrowing
  node: AST.FunctionDefinition | AST.LambdaExpression; // Allow both definition types
  closureEnv: Environment; // Environment captured at definition time
}

// Represents a runtime struct definition
interface AbleStructDefinition {
  kind: "struct_definition";
  name: string;
  definitionNode: AST.StructDefinition;
  // Store generic info if needed later
}

// Represents a runtime struct instance
interface AbleStructInstance {
  kind: "struct_instance";
  definition: AbleStructDefinition;
  // Use array for positional, map for named
  values: AbleValue[] | Map<string, AbleValue>;
}

// Represents a runtime union definition (placeholder)
interface AbleUnionDefinition {
  kind: "union_definition";
  name: string;
  definitionNode: AST.UnionDefinition;
}

// Represents a runtime interface definition (placeholder)
interface AbleInterfaceDefinition {
  kind: "interface_definition";
  name: string;
  definitionNode: AST.InterfaceDefinition;
}

// Represents a runtime implementation definition (placeholder)
interface AbleImplementationDefinition {
  kind: "implementation_definition";
  implNode: AST.ImplementationDefinition;
  // Link to interface and target type info
}

// Represents a runtime methods definition (placeholder)
interface AbleMethodsDefinition {
  kind: "methods_definition";
  methodsNode: AST.MethodsDefinition;
  // Link to target type info
}

// Represents a runtime error value (for !, rescue)
interface AbleError {
  kind: "error";
  // Based on spec's Error interface concept
  message: string;
  // Add cause, stack trace etc. later
  originalValue?: any; // The value raised
}
function createError(message: string, originalValue?: any): AbleError {
  return { kind: "error", message, originalValue };
}

// Represents a runtime Proc handle (placeholder)
interface AbleProcHandle {
  kind: "proc_handle";
  id: number; // Example ID
  // Add status, result promise/callback etc.
}

// Represents a runtime Thunk (placeholder)
interface AbleThunk {
  kind: "thunk";
  id: number; // Example ID
  // Add logic for lazy evaluation and blocking
}

// Represents a runtime Array (placeholder)
interface AbleArray {
  kind: "array";
  elements: AbleValue[];
}

// Represents a runtime Range (placeholder)
interface AbleRange {
  kind: "range";
  start: number | bigint;
  end: number | bigint;
  inclusive: boolean;
}

// Represents a runtime Iterator (implements spec's Iterator T concept)
interface AbleIterator {
  kind: "AbleIterator"; // Distinguish from other values
  // The core method to get the next value
  next: () => AbleValue | typeof IteratorEnd;
  // Internal state would be captured here (e.g., index, source array/range)
}

type AbleValue =
  | AblePrimitive
  | AbleFunction // Updated kind
  | AbleStructDefinition
  | AbleStructInstance
  | AbleUnionDefinition // Added
  | AbleInterfaceDefinition // Added
  | AbleImplementationDefinition // Added
  | AbleMethodsDefinition // Added
  | AbleError // Added
  | AbleProcHandle // Added
  | AbleThunk // Added
  | AbleArray // Added
  | AbleRange // Added
  | AbleIterator; // Added for for-loops

// Special object to signal a `return` occurred
class ReturnSignal extends Error {
  // Inherit from Error for stack trace
  constructor(public value: AbleValue) {
    super(`ReturnSignal: ${JSON.stringify(value)}`);
    this.name = "ReturnSignal";
  }
}
// Special object to signal a `raise` occurred
class RaiseSignal extends Error {
  // Inherit from Error for better stack traces?
  constructor(public value: AbleValue) {
    // Should typically be AbleError
    super(`RaiseSignal: ${JSON.stringify(value)}`); // Add message for debugging
    this.name = "RaiseSignal";
  }
}
// Special object to signal a `break` occurred
class BreakSignal extends Error {
  constructor(public label: string, public value: AbleValue) {
    super(`BreakSignal: '${label}' with ${JSON.stringify(value)}`);
    this.name = "BreakSignal";
  }
}
// Add IteratorEnd signal/value
const IteratorEnd: AblePrimitive = { kind: "nil", value: null }; // Use nil for now, maybe dedicated type later

// --- Environment ---

class Environment {
  private values: Map<string, AbleValue> = new Map();
  constructor(private enclosing: Environment | null = null) {}

  // Define a new variable in the current scope
  define(name: string, value: AbleValue): void {
    // TODO: Handle := error if already defined in *current* scope per spec
    if (this.values.has(name)) {
      // For now, allow redefinition for simplicity, spec says := errors
      console.warn(`Warning: Redefining variable "${name}" in the same scope.`);
    }
    this.values.set(name, value);
  }

  // Assign a value to an existing variable (searches up scopes)
  assign(name: string, value: AbleValue): void {
    if (this.values.has(name)) {
      this.values.set(name, value);
      return;
    }
    if (this.enclosing !== null) {
      this.enclosing.assign(name, value);
      return;
    }
    throw new Error(`Interpreter Error: Undefined variable '${name}' for assignment.`);
  }

  // Get a variable's value (searches up scopes)
  get(name: string): AbleValue {
    if (this.values.has(name)) {
      return this.values.get(name)!;
    }
    if (this.enclosing !== null) {
      return this.enclosing.get(name);
    }
    throw new Error(`Interpreter Error: Undefined variable '${name}'.`);
  }

  // Check if variable exists in current scope only
  hasInCurrentScope(name: string): boolean {
    return this.values.has(name);
  }
}

// --- Interpreter ---

class Interpreter {
  private globalEnv: Environment = new Environment();

  // Store builtins separately for potential import resolution
  private builtins: Map<string, AbleValue> = new Map();

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
      console.log(`Interpreting package: ${moduleNode.package.path.map((id) => id.name).join(".")}`);
    }

    // 2. Handle Imports
    const moduleEnv = new Environment(this.globalEnv);
    this.processImports(moduleNode.imports, moduleEnv);

    // 3. Evaluate Definitions in Module Scope
    try {
      for (const stmt of moduleNode.body) {
        if (
          stmt.type === "FunctionDefinition" ||
          stmt.type === "StructDefinition" ||
          stmt.type === "UnionDefinition" ||
          stmt.type === "InterfaceDefinition" ||
          stmt.type === "ImplementationDefinition" ||
          stmt.type === "MethodsDefinition"
        ) {
          this.evaluate(stmt, moduleEnv); // Use moduleEnv
        }
      }

      // 4. Find and call the 'main' function if it exists in the module scope
      let mainFunc: AbleValue | undefined;
      try {
        mainFunc = moduleEnv.get("main"); // Look in moduleEnv
      } catch (e) {
        // 'main' not defined is okay
      }

      if (mainFunc && mainFunc.kind === "function") {
        this.executeFunction(mainFunc, [], moduleEnv); // Call main with moduleEnv as call site env
      } else {
        if (mainFunc) {
          console.warn("Warning: 'main' was found but is not a function.");
        }
        // If no main, evaluate remaining top-level expressions? (Spec TBD)
        // For now, evaluate all statements sequentially if no main? Or require main?
        // Let's assume 'main' is the entry point if present.
      }
    } catch (error) {
      // Catch runtime errors from the interpreter itself (including uncaught signals)
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
      const modulePath = imp.path.map((id) => id.name).join(".");
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
    let lastValue: AbleValue = { kind: "nil", value: null }; // Default result of a block is nil unless specified

    for (const stmt of statements) {
      // Evaluate each statement. Signals are thrown and caught by callers.
      lastValue = this.evaluate(stmt, environment);
    }
    return lastValue; // Return the normal value of the last statement/expression
  }

  // The core evaluation dispatcher - returns AbleValue, throws signals
  evaluate(node: AST.AstNode | null, environment: Environment): AbleValue {
    try {
      // Add try block to help pinpoint errors
      if (!node) return { kind: "nil", value: null }; // Handle null nodes if they appear

      // Use type assertions within each case block for clarity and safety
      switch (node.type) {
        // --- Literals ---
        case "StringLiteral": {
          const typedNode = node as AST.StringLiteral;
          return { kind: "string", value: typedNode.value };
        }
        case "IntegerLiteral": {
          const typedNode = node as AST.IntegerLiteral;
          const intType = typedNode.integerType || "i32";
          // Ensure value is treated correctly based on target type
          if (["i64", "i128", "u64", "u128"].includes(intType)) {
            // Use BigInt for 64-bit+ integers
            return { kind: intType as any, value: BigInt(typedNode.value.toString()) };
          } else {
            // Use Number for smaller integers
            return { kind: intType as any, value: Number(typedNode.value) };
          }
        }
        case "FloatLiteral": {
          const typedNode = node as AST.FloatLiteral;
          return { kind: typedNode.floatType || "f64", value: typedNode.value };
        }
        case "BooleanLiteral": {
          const typedNode = node as AST.BooleanLiteral;
          return { kind: "bool", value: typedNode.value };
        }
        case "NilLiteral":
          return { kind: "nil", value: null };
        case "CharLiteral": {
          const typedNode = node as AST.CharLiteral;
          return { kind: "char", value: typedNode.value };
        }
        case "ArrayLiteral": {
          const typedNode = node as AST.ArrayLiteral;
          const elements = typedNode.elements.map((el) => this.evaluate(el, environment));
          return { kind: "array", elements };
        }

        // --- Expressions ---
        case "Identifier": {
          const typedNode = node as AST.Identifier;
          return environment.get(typedNode.name);
        }
        case "BlockExpression":
          return this.evaluateBlockExpression(node as AST.BlockExpression, environment);
        case "UnaryExpression":
          return this.evaluateUnaryExpression(node as AST.UnaryExpression, environment);
        case "BinaryExpression":
          return this.evaluateBinaryExpression(node as AST.BinaryExpression, environment);
        case "AssignmentExpression":
          return this.evaluateAssignmentExpression(node as AST.AssignmentExpression, environment);
        case "FunctionCall":
          return this.evaluateFunctionCall(node as AST.FunctionCall, environment);
        case "IfExpression":
          return this.evaluateIfExpression(node as AST.IfExpression, environment);
        case "StructLiteral":
          return this.evaluateStructLiteral(node as AST.StructLiteral, environment);
        case "MemberAccessExpression":
          return this.evaluateMemberAccess(node as AST.MemberAccessExpression, environment);
        case "StringInterpolation":
          return this.evaluateStringInterpolation(node as AST.StringInterpolation, environment);
        case "LambdaExpression":
          return this.evaluateLambdaExpression(node as AST.LambdaExpression, environment);
        case "RangeExpression":
          return this.evaluateRangeExpression(node as AST.RangeExpression, environment);

        // --- Statements/Definitions ---
        case "FunctionDefinition":
          this.evaluateFunctionDefinition(node as AST.FunctionDefinition, environment);
          return { kind: "nil", value: null }; // Definitions don't produce a value themselves
        case "StructDefinition":
          this.evaluateStructDefinition(node as AST.StructDefinition, environment);
          return { kind: "nil", value: null };
        case "ReturnStatement": {
          const typedNode = node as AST.ReturnStatement;
          const returnValue = typedNode.argument ? this.evaluate(typedNode.argument, environment) : ({ kind: "void", value: undefined } as AblePrimitive);
          // Throw signal to unwind stack
          throw new ReturnSignal(returnValue);
        }

        // --- Placeholders for other nodes ---
        case "UnionDefinition":
        case "InterfaceDefinition":
        case "ImplementationDefinition":
        case "MethodsDefinition":
        case "ImportStatement":
        case "PackageStatement": // Handled in interpretModule
        case "ImportStatement": // Handled in processImports
          return { kind: "nil", value: null };

        // --- Control Flow & Others (Placeholders) ---
        // Assuming RaiseStatement/BreakStatement are the correct AST types for now
        case "RaiseStatement":
        case "BreakStatement":
        case "WhileLoop":
        case "ForLoop":
        case "MatchExpression":
        case "ProcExpression":
        case "SpawnExpression":
        case "BreakpointExpression":
        case "RescueExpression":
        case "PropagationExpression":
        case "OrElseExpression":
        // --- Types & Patterns (Not directly evaluated) ---
        case "SimpleTypeExpression":
        case "FunctionTypeExpression":
        case "GenericParameter":
        case "FunctionParameter":
        case "StructFieldDefinition":
        case "StructFieldInitializer": // Handled within StructLiteral eval
        case "ImportSelector": // Handled within ImportStatement eval
        case "OrClause": // Handled within IfExpression eval
        case "MatchClause": // Handled within MatchExpression eval
        case "WildcardPattern": // Handled within assignment/match
        case "LiteralPattern": // Handled within assignment/match
        case "StructPattern": // Handled within assignment/match
        case "ArrayPattern": // Handled within assignment/match
        case "FunctionSignature": // Not evaluated directly
          // These nodes are part of other structures or declarations and aren't evaluated standalone.
          // console.warn(`Interpreter Warning: Evaluation not applicable for node type ${node.type}`);
          return { kind: "nil", value: null };

        default:
          // Use type assertion for exhaustive check
          const _exhaustiveCheck: never = node;
          throw new Error(`Interpreter Error: Unknown or unhandled AST node type: ${(_exhaustiveCheck as any).type}`);
      }
    } catch (e) {
      // Add context to errors
      if (e instanceof Error && !(e instanceof ReturnSignal || e instanceof RaiseSignal || e instanceof BreakSignal)) {
        console.error(`Error during evaluation of ${node?.type} node:`, node);
      }
      throw e; // Re-throw signal or error
    }
  }

  // --- Specific Evaluators ---

  private evaluateBlockExpression(node: AST.BlockExpression, environment: Environment): AbleValue {
    const blockEnv = new Environment(environment); // Create new scope for the block
    // evaluateStatements now throws signals directly
    return this.evaluateStatements(node.body, blockEnv);
  }

  private evaluateUnaryExpression(node: AST.UnaryExpression, environment: Environment): AbleValue {
    const operand = this.evaluate(node.operand, environment);

    switch (node.operator) {
      case "-":
        // Add checks for other numeric kinds (i8, i16, etc.)
        if (operand.kind === "i32" && typeof operand.value === "number") return { kind: "i32", value: -operand.value };
        if (operand.kind === "f64" && typeof operand.value === "number") return { kind: "f64", value: -operand.value };
        if (operand.kind === "i64" && typeof operand.value === "bigint") return { kind: "i64", value: -operand.value };
        // Add other bigint types (i128, u64, u128 - though negation might change type for unsigned)
        throw new Error(`Interpreter Error: Unary '-' not supported for type ${operand.kind}`);
      case "!":
        if (operand.kind === "bool") {
          return { kind: "bool", value: !operand.value };
        }
        throw new Error(`Interpreter Error: Unary '!' not supported for type ${operand.kind}`);
      case "~":
        // Add checks for all integer kinds
        if (operand.kind === "i32" && typeof operand.value === "number") return { kind: "i32", value: ~operand.value };
        if (operand.kind === "i64" && typeof operand.value === "bigint") return { kind: "i64", value: ~operand.value };
        // Add other integer types
        throw new Error(`Interpreter Error: Unary '~' not supported for type ${operand.kind}`);
    }
    throw new Error(`Interpreter Error: Unknown unary operator ${node.operator}`);
  }

  private evaluateBinaryExpression(node: AST.BinaryExpression, environment: Environment): AbleValue {
    const left = this.evaluate(node.left, environment);
    // Handle short-circuiting operators first
    if (node.operator === "&&") {
      if (left.kind !== "bool") throw new Error("Interpreter Error: Left operand of && must be boolean");
      if (!left.value) return { kind: "bool", value: false }; // Short-circuit
      const right = this.evaluate(node.right, environment);
      if (right.kind !== "bool") throw new Error("Interpreter Error: Right operand of && must be boolean");
      return right;
    }
    if (node.operator === "||") {
      if (left.kind !== "bool") throw new Error("Interpreter Error: Left operand of || must be boolean");
      if (left.value) return { kind: "bool", value: true }; // Short-circuit
      const right = this.evaluate(node.right, environment);
      if (right.kind !== "bool") throw new Error("Interpreter Error: Right operand of || must be boolean");
      return right;
    }

    const right = this.evaluate(node.right, environment);

    // --- Arithmetic --- (Needs expansion for all type combinations and bigint)
    if (["+", "-", "*", "/", "%"].includes(node.operator)) {
      // Ensure both operands are primitives before accessing .value
      if ("value" in left && "value" in right) {
        // Check for number operations (assuming matching kinds for now)
        if (
          typeof left.value === "number" &&
          typeof right.value === "number" &&
          typeof left.kind === "string" &&
          left.kind.match(/^(i(8|16|32)|u(8|16|32)|f(32|64))$/) &&
          left.kind === right.kind
        ) {
          const kind = left.kind; // Use the specific kind
          switch (node.operator) {
            case "+":
              return { kind, value: left.value + right.value };
            case "-":
              return { kind, value: left.value - right.value };
            case "*":
              return { kind, value: left.value * right.value };
            case "/":
              if (right.value === 0) throw createError("Division by zero");
              const result = kind.startsWith("f") ? left.value / right.value : Math.trunc(left.value / right.value);
              return { kind, value: result };
            case "%":
              if (right.value === 0) throw createError("Division by zero");
              return { kind, value: left.value % right.value };
          }
        }
        // Check for bigint operations (assuming matching kinds)
        if (
          typeof left.value === "bigint" &&
          typeof right.value === "bigint" &&
          typeof left.kind === "string" &&
          left.kind.match(/^(i(64|128)|u(64|128))$/) &&
          left.kind === right.kind
        ) {
          const kind = left.kind; // Use the specific kind
          switch (node.operator) {
            case "+":
              return { kind, value: left.value + right.value };
            case "-":
              return { kind, value: left.value - right.value };
            case "*":
              return { kind, value: left.value * right.value };
            case "/":
              if (right.value === 0n) throw createError("Division by zero");
              return { kind, value: left.value / right.value }; // BigInt division truncates
            case "%":
              if (right.value === 0n) throw createError("Division by zero");
              return { kind, value: left.value % right.value };
          }
        }
        // String concatenation
        if (node.operator === "+" && left.kind === "string" && right.kind === "string") {
          return { kind: "string", value: left.value + right.value };
        }
      }
      // TODO: Add type promotion rules (e.g., i32 + f64 -> f64)
      throw new Error(`Interpreter Error: Operator '${node.operator}' not supported for types ${left.kind} and ${right.kind}`);
    }

    // --- Comparison --- (Needs refinement for different types)
    if ([">", "<", ">=", "<=", "==", "!="].includes(node.operator)) {
      // Ensure both operands are primitives before accessing .value
      if ("value" in left && "value" in right) {
        // Handle nil comparisons first
        if (left.kind === "nil" || right.kind === "nil") {
          if (node.operator === "==") return { kind: "bool", value: left.kind === right.kind };
          if (node.operator === "!=") return { kind: "bool", value: left.kind !== right.kind };
          // Other comparisons with nil are likely errors or false
          throw new Error(`Interpreter Error: Operator '${node.operator}' not supported for nil.`);
        }

        // Basic comparison for non-nil primitives
        const lVal = left.value; // Now guaranteed not null
        const rVal = right.value; // Now guaranteed not null

        // Check for > < >= <= (only valid for numbers/bigints/strings/chars?)
        if ([">", "<", ">=", "<="].includes(node.operator)) {
          if (
            !(
              (typeof lVal === "number" || typeof lVal === "bigint" || typeof lVal === "string" || typeof lVal === "string") && // char is string
              (typeof rVal === "number" || typeof rVal === "bigint" || typeof rVal === "string" || typeof rVal === "string")
            )
          ) {
            throw new Error(`Interpreter Error: Operator '${node.operator}' requires comparable types (numbers, bigints, strings, chars), got ${left.kind} and ${right.kind}.`);
          }
          // Allow JS comparison for these types (might need refinement for specific Able rules)
          try {
            switch (node.operator) {
              case ">":
                return { kind: "bool", value: lVal > rVal };
              case "<":
                return { kind: "bool", value: lVal < rVal };
              case ">=":
                return { kind: "bool", value: lVal >= rVal };
              case "<=":
                return { kind: "bool", value: lVal <= rVal };
            }
          } catch (e) {
            // Catch potential errors comparing incompatible types (e.g., bigint > string)
            throw new Error(`Interpreter Error: Cannot compare ${left.kind} and ${right.kind} with ${node.operator}`);
          }
        }

        // Handle == !=
        if (node.operator === "==") return { kind: "bool", value: left.kind === right.kind && lVal === rVal }; // Strict equality requires same kind
        if (node.operator === "!=") return { kind: "bool", value: left.kind !== right.kind || lVal !== rVal };
      } else {
        //      // Handle comparison involving non-primitives (structs, arrays etc.) - likely false for ==/!= unless Eq implemented
        // TODO: Implement Eq/PartialEq interface checks
        if (node.operator === "==") return { kind: "bool", value: false }; // Default non-primitive comparison
        if (node.operator === "!=") return { kind: "bool", value: true };
        throw new Error(`Interpreter Error: Comparison operator '${node.operator}' not supported for non-primitive types ${left.kind} and ${right.kind}`);
      }
      //  } else {
      //      // Handle comparison involving non-primitives (structs, arrays etc.) - likely false for ==/!= unless Eq implemented
      //      if (node.operator === '==') return { kind: 'bool', value: false }; // Default non-primitive comparison
      //      if (node.operator === '!=') return { kind: 'bool', value: true };
      //      throw new Error(`Interpreter Error: Comparison operator '${node.operator}' not supported for non-primitive types ${left.kind} and ${right.kind}`);
      //  }
    }

    // --- Bitwise --- (Needs expansion for all integer types)
    if (["&", "|", "^", "<<", ">>"].includes(node.operator)) {
      // Ensure both operands are primitives with 'value' property
      if ("value" in left && "value" in right) {
        // Check for number operations (assuming matching integer kinds)
        if (typeof left.value === "number" && typeof right.value === "number" && left.kind.match(/^i|^u/) && left.kind === right.kind) {
          const kind = left.kind; // Use specific kind
          switch (node.operator) {
            case "&":
              return { kind, value: left.value & right.value };
            case "|":
              return { kind, value: left.value | right.value };
            case "^":
              return { kind, value: left.value ^ right.value };
            case "<<":
              return { kind, value: left.value << right.value };
            case ">>":
              return { kind, value: left.value >> right.value }; // Sign-propagating right shift
          }
        }
        // Check for bigint operations (assuming matching integer kinds)
        if (typeof left.value === "bigint" && typeof right.value === "bigint" && left.kind.match(/^i|^u/) && left.kind === right.kind) {
          const kind = left.kind; // Use specific kind
          // Note: Right operand for shift in BigInt must not be negative
          const rightShiftVal = node.operator === "<<" || node.operator === ">>" ? BigInt(Number(right.value)) : right.value; // Convert shift count carefully
          if ((node.operator === "<<" || node.operator === ">>") && rightShiftVal < 0n) {
            throw new Error("Interpreter Error: Shift amount cannot be negative.");
          }
          switch (node.operator) {
            case "&":
              return { kind, value: left.value & right.value };
            case "|":
              return { kind, value: left.value | right.value };
            case "^":
              return { kind, value: left.value ^ right.value };
            case "<<":
              return { kind, value: left.value << rightShiftVal };
            case ">>":
              return { kind, value: left.value >> rightShiftVal };
          }
        }
      }
      throw new Error(`Interpreter Error: Bitwise operator '${node.operator}' not supported for types ${left.kind} and ${right.kind}`);
    }

    throw new Error(`Interpreter Error: Unknown binary operator ${node.operator}`);
  }

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
        throw new Error(`Interpreter Error: Compound operator ${node.operator} not fully implemented for ${leftVal.kind} and ${rightVal.kind}`);
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

      if (obj.kind === "struct_instance") {
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
      } else if (obj.kind === "array") {
        // Handle array mutation obj.elements[index] = valueToAssign;
        if (member.type !== "IntegerLiteral") throw new Error("Interpreter Error: Array index must be an integer literal.");
        const index = Number(member.value);
        if (index < 0 || index >= obj.elements.length) {
          throw new Error(`Interpreter Error: Array index ${index} out of bounds (length ${obj.elements.length}).`);
        }
        obj.elements[index] = valueToAssign;
      } else {
        throw new Error(`Interpreter Error: Cannot assign to member of type ${obj.kind}.`);
      }
    } else {
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
        } else if (value.kind === "nil" && patternVal.kind === "nil") {
          // Specifically allow nil to match nil literal
        } else {
          throw new Error(`Interpreter Error: Cannot match literal pattern ${this.valueToString(patternVal)} against non-primitive value ${this.valueToString(value)}.`);
        }
        break;
      }
      case "StructPattern": {
        if (value.kind !== "struct_instance") {
          throw new Error(`Interpreter Error: Cannot destructure non-struct value (got ${value.kind}) with a struct pattern.`);
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
            // Positional patterns in AST don't have fieldName, use index
            this.evaluatePatternAssignment(pattern.fields[i].pattern, value.values[i], environment, isDeclaration);
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
        if (value.kind !== "array") {
          throw new Error(`Interpreter Error: Cannot destructure non-array value (got ${value.kind}) with an array pattern.`);
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
          this.evaluatePatternAssignment(pattern.elements[i], value.elements[i], environment, isDeclaration);
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
    if (!structDefVal || structDefVal.kind !== "struct_definition") {
      throw new Error(`Interpreter Error: Cannot instantiate unknown or non-struct type '${node.structType?.name}'.`);
    }
    const structDef = structDefVal as AbleStructDefinition;

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
        if (sourceVal.kind !== "struct_instance" || sourceVal.definition !== structDef) {
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
    const object = this.evaluate(node.object, environment);

    if (object.kind === "struct_instance") {
      const member = node.member;
      if (member.type === "Identifier") {
        // Named field access
        if (!(object.values instanceof Map)) throw new Error(`Interpreter Error: Expected named fields map for struct instance '${object.definition.name}'.`);
        const fieldName = member.name;
        if (object.values.has(fieldName)) {
          return object.values.get(fieldName)!;
        } else {
          // TODO: Check for inherent/interface methods here
          throw new Error(`Interpreter Error: Struct '${object.definition.name}' has no field or method named '${fieldName}'.`);
        }
      } else {
        // Positional field access (IntegerLiteral)
        if (!Array.isArray(object.values)) throw new Error(`Interpreter Error: Expected positional fields array for struct instance '${object.definition.name}'.`);
        const index = Number(member.value); // Assuming integer literal for index
        if (index < 0 || index >= object.values.length) {
          throw new Error(`Interpreter Error: Index ${index} out of bounds for struct '${object.definition.name}'.`);
        }
        return object.values[index];
      }
    } else if (object.kind === "array") {
      // Handle array indexing
      if (node.member.type !== "IntegerLiteral") throw new Error("Interpreter Error: Array index must be an integer literal.");
      const index = Number(node.member.value);
      if (index < 0 || index >= object.elements.length) {
        throw new Error(`Interpreter Error: Array index ${index} out of bounds (length ${object.elements.length}).`);
      }
      return object.elements[index];
    }
    // TODO: Handle method calls via member access (UFCS, inherent, interface)
    // TODO: Handle static method access (e.g., Point.origin()) - might need different AST node or check object type

    throw new Error(`Interpreter Error: Cannot access member on type ${object.kind}.`);
  }

  private evaluateFunctionCall(node: AST.FunctionCall, environment: Environment): AbleValue {
    const callee = this.evaluate(node.callee, environment);

    if (callee.kind !== "function") {
      // TODO: Check for callable objects implementing Apply interface
      throw new Error(`Interpreter Error: Cannot call non-function type ${callee.kind}.`);
    }

    const func = callee as AbleFunction; // Runtime function value
    const args = node.arguments.map((arg) => this.evaluate(arg, environment));

    // Handle native functions (identified by null node and having 'apply')
    if (func.node === null && typeof (func as any).apply === "function") {
      // Native functions might throw errors directly or return AbleError values
      try {
        return (func as any).apply(args);
      } catch (e: any) {
        // Wrap native errors if necessary
        throw createError(e.message || "Native function error", e);
      }
    }

    // Handle user-defined functions
    const funcDef = func.node; // AST.FunctionDefinition or AST.LambdaExpression
    if (!funcDef) throw new Error("Interpreter Error: Function definition node is missing."); // Should not happen

    if (args.length !== funcDef.params.length) {
      const funcName = funcDef.type === "FunctionDefinition" && funcDef.id ? funcDef.id.name : "(anonymous)";
      throw new Error(`Interpreter Error: Expected ${funcDef.params.length} arguments but got ${args.length} for function '${funcName}'.`);
    }

    // Create new environment for the function call
    // Enclosing scope is the environment where the function was DEFINED (closure)
    const funcEnv = new Environment(func.closureEnv);

    // Bind arguments to parameters
    for (let i = 0; i < funcDef.params.length; i++) {
      funcEnv.define(funcDef.params[i].name.name, args[i]);
    }

    // Execute the function body
    try {
      let lastValue: AbleValue;
      if (funcDef.body.type === "BlockExpression") {
        lastValue = this.evaluateStatements(funcDef.body.body, funcEnv);
      } else {
        // Single expression lambda body
        lastValue = this.evaluate(funcDef.body, funcEnv);
      }
      // Implicit return of the last expression's value
      return lastValue;
    } catch (signal) {
      // Catch ReturnSignal specifically
      if (signal instanceof ReturnSignal) {
        return signal.value; // Return the value from the signal
      }
      // Propagate other signals (Raise, Break) or errors
      throw signal;
    }
  }

  // Helper to execute a function value directly (used for 'main')
  private executeFunction(func: AbleFunction, args: AbleValue[], callSiteEnv: Environment): AbleValue {
    // Similar logic to evaluateFunctionCall, but takes AbleFunction directly
    const funcDef = func.node;
    if (!funcDef) throw new Error("Interpreter Error: Function definition node is missing.");

    if (args.length !== funcDef.params.length) {
      throw new Error(`Interpreter Error: Argument count mismatch during direct function execution.`);
    }
    const funcEnv = new Environment(func.closureEnv);
    for (let i = 0; i < funcDef.params.length; i++) {
      funcEnv.define(funcDef.params[i].name.name, args[i]);
    }
    // Execute and handle signals similar to evaluateFunctionCall
    try {
      let lastValue: AbleValue;
      if (funcDef.body.type === "BlockExpression") {
        lastValue = this.evaluateStatements(funcDef.body.body, funcEnv);
      } else {
        lastValue = this.evaluate(funcDef.body, funcEnv);
      }
      // Implicit return
      return lastValue;
    } catch (signal) {
      if (signal instanceof ReturnSignal) return signal.value; // Catch return from within
      // Propagate other signals (Raise, Break) or errors
      throw signal;
    }
  }

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

  private evaluateRangeExpression(node: AST.RangeExpression, environment: Environment): AbleRange {
    const startVal = this.evaluate(node.start, environment);
    const endVal = this.evaluate(node.end, environment);

    // Add type guards before accessing .value
    if ("value" in startVal && "value" in endVal) {
      // Basic validation - ensure both are numbers or both are bigints
      if (!((typeof startVal.value === "number" && typeof endVal.value === "number") || (typeof startVal.value === "bigint" && typeof endVal.value === "bigint"))) {
        throw new Error(`Interpreter Error: Range boundaries must be both numbers or both bigints. Got ${startVal.kind} and ${endVal.kind}.`);
      }

      return {
        kind: "range",
        start: startVal.value as number | bigint, // Cast is safe due to check above
        end: endVal.value as number | bigint,
        inclusive: node.inclusive,
      };
    } else {
      throw new Error(`Interpreter Error: Range boundaries must be primitive types. Got ${startVal.kind} and ${endVal.kind}.`);
    }
  }

  // --- Helpers ---

  // Determine truthiness according to Able rules (Spec TBD, basic version here)
  private isTruthy(value: AbleValue): boolean {
    if (value.kind === "bool") {
      return value.value;
    }
    if (value.kind === "nil" || value.kind === "void") {
      return false;
    }
    // TODO: Define truthiness for other types (numbers != 0, strings not empty, collections not empty etc.)
    // Per discussion, let's make 0, 0n, "" false for now.
    if (
      (value.kind === "i32" ||
        value.kind === "f64" ||
        value.kind === "i8" ||
        value.kind === "i16" ||
        value.kind === "u8" ||
        value.kind === "u16" ||
        value.kind === "u32" ||
        value.kind === "f32") &&
      value.value === 0
    )
      return false;
    if ((value.kind === "i64" || value.kind === "i128" || value.kind === "u64" || value.kind === "u128") && value.value === 0n) return false;
    if (value.kind === "string" && value.value === "") return false;
    if (value.kind === "array" && value.elements.length === 0) return false;

    // Most other things are truthy by default
    return true;
  }

  // Convert runtime value to string for printing (basic)
  private valueToString(value: AbleValue): string {
    // Removed EvaluationResult as signals are errors now
    // if (value instanceof ReturnSignal) return `<return ${this.valueToString(value.value)}>`; // Handled by catching Error
    // if (value instanceof RaiseSignal) return `<raise ${this.valueToString(value.value)}>`; // Handled by catching Error
    // if (value instanceof BreakSignal) return `<break '${value.label}' ${this.valueToString(value.value)}>`; // Handled by catching Error

    if (value === null || value === undefined) return "<?>"; // Should not happen with typed values

    switch (value.kind) {
      case "i8":
      case "i16":
      case "i32":
      case "u8":
      case "u16":
      case "u32":
      case "f32":
      case "f64":
        return value.value.toString();
      case "i64":
      case "i128":
      case "u64":
      case "u128":
        return value.value.toString(); // BigInt already has toString
      case "string":
        return value.value; // Consider adding quotes for clarity? No, raw string.
      case "bool":
        return value.value.toString();
      case "char":
        return `'${value.value}'`; // Add quotes around char
      case "nil":
        return "nil";
      case "void":
        return "void";
      case "function":
        const funcName = value.node?.type === "FunctionDefinition" && value.node.id ? value.node.id.name : "(anonymous)";
        return `<function ${funcName}>`;
      case "struct_definition":
        return `<struct ${value.name}>`;
      case "struct_instance":
        // TODO: Call Display interface if implemented, otherwise default repr
        if (Array.isArray(value.values)) {
          return `${value.definition.name} { ${value.values.map((v) => this.valueToString(v)).join(", ")} }`;
        } else {
          // Correctly handle map iteration for named fields
          const fields = Array.from(value.values.entries())
            .map(([key, val]: [string, AbleValue]) => `${key}: ${this.valueToString(val)}`)
            .join(", ");
          return `${value.definition.name} { ${fields} }`;
        }
      case "array":
        return `[${value.elements.map((v) => this.valueToString(v)).join(", ")}]`;
      case "error":
        return `<error: ${value.message}>`;
      // Add other types
      case "union_definition":
        return `<union ${value.name}>`;
      case "interface_definition":
        return `<interface ${value.name}>`;
      // Safely access properties for ImplementationDefinition and MethodsDefinition
      case "implementation_definition":
        const ifaceName = value.implNode.interfaceName?.name ?? "?";
        // Need a way to represent targetType better
        const targetTypeName = (value.implNode.targetType as any)?.name?.name ?? "?"; // Basic attempt
        return `<impl ${ifaceName} for ${targetTypeName}>`;
      case "methods_definition":
        // Need a way to represent targetType better
        const methodsTargetTypeName = (value.methodsNode.targetType as any)?.name?.name ?? "?"; // Basic attempt
        return `<methods for ${methodsTargetTypeName}>`;
      case "proc_handle":
        return `<proc ${value.id}>`;
      case "thunk":
        return `<thunk ${value.id}>`;
      case "range":
        return `${value.start}${value.inclusive ? ".." : "..."}${value.end}`;
      case "AbleIterator":
        return `<iterator>`; // Placeholder
      default:
        // Use type assertion for exhaustive check
        const _exhaustiveCheck: never = value;
        return `<${(_exhaustiveCheck as any).kind}>`; // Use kind property for unknown types
    }
  }
}

// --- Entry Point ---

// Example function to interpret a module AST
export function interpret(moduleNode: AST.Module) {
  const interpreter = new Interpreter();
  interpreter.interpretModule(moduleNode);
}

// Example Usage (in another file or here):
// import sampleModule from './sample1'; // Assuming sample1.ts exports the AST
// interpret(sampleModule);
