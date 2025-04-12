import * as AST from './ast'; // Import our AST definitions

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
    node: AST.FunctionDefinition;
    closureEnv: Environment; // Environment captured at definition time
}

// Represents a runtime struct definition
interface AbleStructDefinition {
    kind: 'struct_definition';
    name: string;
    definitionNode: AST.StructDefinition;
    // Store generic info if needed later
}

// Represents a runtime struct instance
interface AbleStructInstance {
    kind: 'struct_instance';
    definition: AbleStructDefinition;
    // Use array for positional, map for named
    values: AbleValue[] | Map<string, AbleValue>;
}

// Represents a runtime union definition (placeholder)
interface AbleUnionDefinition {
    kind: 'union_definition';
    name: string;
    definitionNode: AST.UnionDefinition;
}

// Represents a runtime interface definition (placeholder)
interface AbleInterfaceDefinition {
    kind: 'interface_definition';
    name: string;
    definitionNode: AST.InterfaceDefinition;
}

// Represents a runtime implementation definition (placeholder)
interface AbleImplementationDefinition {
    kind: 'implementation_definition';
    implNode: AST.ImplementationDefinition;
    // Link to interface and target type info
}

// Represents a runtime methods definition (placeholder)
interface AbleMethodsDefinition {
    kind: 'methods_definition';
    methodsNode: AST.MethodsDefinition;
    // Link to target type info
}

// Represents a runtime error value (for !, rescue)
interface AbleError {
    kind: 'error';
    // Based on spec's Error interface concept
    message: string;
    // Add cause, stack trace etc. later
    originalValue?: any; // The value raised
}

// Represents a runtime Proc handle (placeholder)
interface AbleProcHandle {
    kind: 'proc_handle';
    id: number; // Example ID
    // Add status, result promise/callback etc.
}

// Represents a runtime Thunk (placeholder)
interface AbleThunk {
    kind: 'thunk';
    id: number; // Example ID
    // Add logic for lazy evaluation and blocking
}

// Represents a runtime Array (placeholder)
interface AbleArray {
    kind: 'array';
    elements: AbleValue[];
}

// Represents a runtime Range (placeholder)
interface AbleRange {
    kind: 'range';
    start: number | bigint;
    end: number | bigint;
    inclusive: boolean;
}


type AbleValue =
  | AblePrimitive
  | { kind: "function"; value: AbleFunction }
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
  ;

// Special object to signal a `return` occurred
class ReturnSignal {
    constructor(public value: AbleValue) {}
}
// Special object to signal a `raise` occurred
class RaiseSignal {
    constructor(public value: AbleValue) {}
}

