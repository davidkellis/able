// Base interface for all AST nodes
export interface AstNode {
  type: string;
  // Optional: Add common properties like location/span if needed later
  // loc?: { start: Position; end: Position };
}

// Represents an identifier
export interface Identifier extends AstNode {
  type: 'Identifier';
  name: string;
}
export function identifier(name: string): Identifier {
  return { type: 'Identifier', name };
}

// --- Literals ---
// Base for all literals
export interface Literal extends AstNode {
    type: 'Literal';
    // value: any; // Keep value specific to subtypes
}

export interface StringLiteral extends Literal {
  type: 'StringLiteral'; // More specific type
  value: string;
  isInterpolated: boolean; // Track if it's a backtick string
  // If interpolated, parts might be needed: parts: (string | Expression)[];
}
export function stringLiteral(value: string, isInterpolated: boolean = false): StringLiteral {
  return { type: 'StringLiteral', value, isInterpolated };
}

// Integer Literals - Add specific kinds based on spec
export interface IntegerLiteral extends Literal {
  type: 'IntegerLiteral';
  value: bigint | number; // Use bigint for larger types
  integerType: 'i8' | 'i16' | 'i32' | 'i64' | 'i128' | 'u8' | 'u16' | 'u32' | 'u64' | 'u128';
}
// DSL needs refinement based on how parser provides type info
export function integerLiteral(value: bigint | number, type: IntegerLiteral['integerType'] = 'i32'): IntegerLiteral {
  // Basic type inference or use provided type
  return { type: 'IntegerLiteral', value, integerType: type };
}


export interface FloatLiteral extends Literal {
    type: 'FloatLiteral'; // More specific type
    value: number;
    floatType: 'f32' | 'f64';
}
export function floatLiteral(value: number, type: FloatLiteral['floatType'] = 'f64'): FloatLiteral {
    return { type: 'FloatLiteral', value, floatType: type };
}

export interface BooleanLiteral extends Literal {
  type: 'BooleanLiteral'; // More specific type
  value: boolean;
}
export function booleanLiteral(value: boolean): BooleanLiteral {
  return { type: 'BooleanLiteral', value };
}

export interface NilLiteral extends Literal {
  type: 'NilLiteral'; // More specific type
  value: null; // Explicitly null
}
export function nilLiteral(): NilLiteral {
  return { type: 'NilLiteral', value: null };
}

export interface CharLiteral extends Literal {
    type: 'CharLiteral';
    value: string; // Store the single character as a string
}
export function charLiteral(value: string): CharLiteral {
    // Add validation? Ensure it's a single char?
    return { type: 'CharLiteral', value };
}


// --- Expressions ---

export type Expression =
  | Identifier
  | Literal // Keep subtypes separate where possible
  | StringLiteral
  | IntegerLiteral
  | FloatLiteral
  | BooleanLiteral
  | NilLiteral
  | CharLiteral
  | BinaryExpression
  | UnaryExpression // Added
  | FunctionCall
  | BlockExpression
  | AssignmentExpression
  | IfExpression
  | MatchExpression // Added
  | StructLiteral
  | MemberAccessExpression
  | PropagationExpression // Added for `!`
  | OrElseExpression // Added for `else {}`
  | ArrayLiteral // Added
  | RangeExpression // Added
  | StringInterpolation // Added (if needed separately from StringLiteral)
  ;
  // Add more expression types here...

export interface UnaryExpression extends AstNode {
    type: 'UnaryExpression';
    operator: '-' | '!' | '~'; // Add other unary ops if needed
    operand: Expression;
}
export function unaryExpression(operator: UnaryExpression['operator'], operand: Expression): UnaryExpression {
    return { type: 'UnaryExpression', operator, operand };
}


export interface BinaryExpression extends AstNode {
  type: 'BinaryExpression';
  operator: string; // Keep as string for flexibility, refine in interpreter
  left: Expression;
  right: Expression;
}
export function binaryExpression(operator: string, left: Expression, right: Expression): BinaryExpression {
  return { type: 'BinaryExpression', operator, left, right };
}

export interface FunctionCall extends AstNode {
  type: 'FunctionCall';
  callee: Expression; // Usually an Identifier, but could be other expressions
  arguments: Expression[];
  // Optional: Add type arguments if needed later
  // typeArguments?: TypeExpression[];
}
export function functionCall(callee: Expression, args: Expression[]): FunctionCall {
  return { type: 'FunctionCall', callee, arguments: args };
}

export interface BlockExpression extends AstNode {
    type: 'BlockExpression';
    body: Statement[]; // Sequence of statements/expressions
}
export function blockExpression(body: Statement[]): BlockExpression {
    return { type: 'BlockExpression', body };
}

export interface AssignmentExpression extends AstNode {
    type: 'AssignmentExpression';
    operator: ':=' | '='; // Declaration or Assignment/Mutation
    left: Pattern | MemberAccessExpression; // LHS: Pattern for := or destructuring =, MemberAccess for mutation =
    right: Expression;
}
// DSL function remains the same for now
export function assignmentExpression(operator: ':=' | '=', left: Pattern | MemberAccessExpression, right: Expression): AssignmentExpression {
    return { type: 'AssignmentExpression', operator, left, right };
}

export interface ArrayLiteral extends AstNode {
    type: 'ArrayLiteral';
    elements: Expression[];
}
export function arrayLiteral(elements: Expression[]): ArrayLiteral {
    return { type: 'ArrayLiteral', elements };
}

export interface RangeExpression extends AstNode {
    type: 'RangeExpression';
    start: Expression;
    end: Expression;
    inclusive: boolean; // true for '..', false for '...'
}
export function rangeExpression(start: Expression, end: Expression, inclusive: boolean): RangeExpression {
    return { type: 'RangeExpression', start, end, inclusive };
}

// If string interpolation needs complex structure:
export interface StringInterpolation extends AstNode {
    type: 'StringInterpolation';
    parts: (StringLiteral | Expression)[]; // Alternating string parts and expressions
}
export function stringInterpolation(parts: (StringLiteral | Expression)[]): StringInterpolation {
    return { type: 'StringInterpolation', parts };
}


// --- Statements/Definitions ---

export type Statement =
  | Expression // Expressions can be statements
  // | VariableDeclaration // Use AssignmentExpression with :=
  | FunctionDefinition
  | StructDefinition
  | UnionDefinition // Added
  | InterfaceDefinition // Added
  | ImplementationDefinition // Added
  | MethodsDefinition // Added
  | ImportStatement // Added
  | PackageStatement // Added
  | ReturnStatement // Added
  | RaiseStatement // Added
  | WhileLoop // Added
  | ForLoop // Added
  ;
  // Add more statement/definition types here...

// Using AssignmentExpression for := for now, might refine later
export type VariableDeclaration = AssignmentExpression & { operator: ':=' };
export function variableDeclaration(left: Pattern, right: Expression): VariableDeclaration {
    // Type assertion for clarity, structure matches AssignmentExpression
    return assignmentExpression(':=', left, right) as VariableDeclaration;
}

// --- Structs ---

export interface StructFieldDefinition extends AstNode {
    type: 'StructFieldDefinition';
    name: Identifier;
    fieldType: TypeExpression; // Type annotation
    isPositional: boolean; // Added to distinguish positional vs named
}
// DSL needs update
export function structFieldDefinition(name: Identifier, fieldType: TypeExpression, isPositional: boolean = false): StructFieldDefinition {
    return { type: 'StructFieldDefinition', name, fieldType, isPositional };
}

export interface StructDefinition extends AstNode {
    type: 'StructDefinition';
    id: Identifier;
    genericParams?: GenericParameter[]; // Added
    fields: StructFieldDefinition[];
    whereClause?: Constraint[]; // Added
    // Add flag for singleton, named, positional based on fields?
    kind: 'singleton' | 'named' | 'positional'; // Added
}
// DSL needs update
export function structDefinition(
    id: Identifier,
    fields: StructFieldDefinition[],
    kind: StructDefinition['kind'],
    genericParams?: GenericParameter[],
    whereClause?: Constraint[]
): StructDefinition {
    return { type: 'StructDefinition', id, fields, kind, genericParams, whereClause };
}


export interface StructFieldInitializer extends AstNode {
    type: 'StructFieldInitializer';
    name?: Identifier; // Optional for positional
    value: Expression;
    isPositional: boolean; // Added
}
// DSL needs update
export function structFieldInitializer(value: Expression, name?: Identifier): StructFieldInitializer {
    return { type: 'StructFieldInitializer', name, value, isPositional: !name };
}


export interface StructLiteral extends AstNode {
    type: 'StructLiteral';
    structType: Identifier; // Name of the struct being instantiated
    fields: StructFieldInitializer[]; // Can be named or positional
    // Add generic arguments? typeArguments?: TypeExpression[];
}
// DSL needs update
export function structLiteral(structType: Identifier, fields: StructFieldInitializer[]): StructLiteral {
    return { type: 'StructLiteral', structType, fields };
}

export interface MemberAccessExpression extends AstNode {
    type: 'MemberAccessExpression';
    object: Expression; // The struct instance or other object
    member: Identifier | IntegerLiteral; // Field name or positional index
}
// DSL needs update
export function memberAccessExpression(object: Expression, member: Identifier | IntegerLiteral): MemberAccessExpression {
    return { type: 'MemberAccessExpression', object, member };
}

// --- Unions ---
export interface UnionDefinition extends AstNode {
    type: 'UnionDefinition';
    id: Identifier;
    genericParams?: GenericParameter[];
    variants: TypeExpression[]; // List of types in the union
    whereClause?: Constraint[];
}
export function unionDefinition(
    id: Identifier,
    variants: TypeExpression[],
    genericParams?: GenericParameter[],
    whereClause?: Constraint[]
): UnionDefinition {
    return { type: 'UnionDefinition', id, genericParams, variants, whereClause };
}


// --- Functions ---

export interface FunctionParameter extends AstNode {
    type: 'FunctionParameter';
    name: Identifier;
    paramType?: TypeExpression; // Type annotation
}
export function functionParameter(name: Identifier, paramType?: TypeExpression): FunctionParameter {
    return { type: 'FunctionParameter', name, paramType };
}

export interface FunctionDefinition extends AstNode {
  type: 'FunctionDefinition';
  id: Identifier | null; // Null for anonymous functions/lambdas
  genericParams?: GenericParameter[]; // Added
  params: FunctionParameter[];
  returnType?: TypeExpression; // Optional return type annotation
  body: BlockExpression | Expression; // Allow single expression for lambdas
  whereClause?: Constraint[]; // Added
  isLambda: boolean; // Added to distinguish fn {} from { => }
  isMethodShorthand?: boolean; // Added for `fn #method`
}
// DSL needs update
export function functionDefinition(
    id: Identifier | null,
    params: FunctionParameter[],
    body: BlockExpression | Expression,
    returnType?: TypeExpression,
    genericParams?: GenericParameter[],
    whereClause?: Constraint[],
    isLambda: boolean = false,
    isMethodShorthand: boolean = false
): FunctionDefinition {
  return { type: 'FunctionDefinition', id, params, body, returnType, genericParams, whereClause, isLambda, isMethodShorthand };
}

export interface ReturnStatement extends AstNode {
    type: 'ReturnStatement';
    argument?: Expression; // Optional argument (for `return;` vs `return val;`)
}
export function returnStatement(argument?: Expression): ReturnStatement {
    return { type: 'ReturnStatement', argument };
}


// --- Patterns (for Assignment/Destructuring/Match) ---
export type Pattern =
  | Identifier // Simple binding
  | WildcardPattern
  | LiteralPattern // Added
  | StructPattern // Added
  | ArrayPattern // Added
  ;
  // Add StructPattern, ArrayPattern, LiteralPattern later

export interface WildcardPattern extends AstNode {
    type: 'WildcardPattern';
}
export function wildcardPattern(): WildcardPattern {
    return { type: 'WildcardPattern' };
}

export interface LiteralPattern extends AstNode {
    type: 'LiteralPattern';
    literal: Literal; // Match against a specific literal value
}
export function literalPattern(literal: Literal): LiteralPattern {
    return { type: 'LiteralPattern', literal };
}

export interface StructPatternField extends AstNode {
    type: 'StructPatternField';
    fieldName?: Identifier; // Optional for positional
    pattern: Pattern; // Nested pattern for the field
    isPositional: boolean;
}
export function structPatternField(pattern: Pattern, fieldName?: Identifier): StructPatternField {
    return { type: 'StructPatternField', fieldName, pattern, isPositional: !fieldName };
}

export interface StructPattern extends AstNode {
    type: 'StructPattern';
    structType?: Identifier; // Optional type check
    fields: StructPatternField[]; // Named or positional fields
}
export function structPattern(fields: StructPatternField[], structType?: Identifier): StructPattern {
    return { type: 'StructPattern', structType, fields };
}

export interface ArrayPattern extends AstNode {
    type: 'ArrayPattern';
    elements: Pattern[];
    restPattern?: Identifier; // Optional identifier for ...rest
}
export function arrayPattern(elements: Pattern[], restPattern?: Identifier): ArrayPattern {
    return { type: 'ArrayPattern', elements, restPattern };
}


// --- Type Expressions ---
// Needs significant expansion based on spec details
export interface TypeExpression extends AstNode {
    type: 'TypeExpression';
    // Details TBD: name, arguments, function type structure, etc.
    // Simplistic representation for now:
    baseName: Identifier | null; // e.g., 'Array', 'Map', 'i32', null for function types
    arguments?: TypeExpression[]; // e.g., [i32] for Array i32
    isNullable?: boolean; // For ?Type
    isResult?: boolean; // For !Type
    // For function types:
    paramTypes?: TypeExpression[];
    returnType?: TypeExpression;
}
// Basic DSL, needs refinement
export function typeExpression(
    baseName: Identifier | string | null,
    args?: TypeExpression[],
    isNullable = false,
    isResult = false
): TypeExpression {
    const baseId = typeof baseName === 'string' ? identifier(baseName) : baseName;
    return { type: 'TypeExpression', baseName: baseId, arguments: args, isNullable, isResult };
}

// --- Generics & Constraints (Placeholders) ---
export interface GenericParameter extends AstNode {
    type: 'GenericParameter';
    name: Identifier;
    constraints?: InterfaceConstraint[]; // T: Display + Clone
}
export function genericParameter(name: Identifier, constraints?: InterfaceConstraint[]): GenericParameter {
    return { type: 'GenericParameter', name, constraints };
}

export interface InterfaceConstraint extends AstNode {
    type: 'InterfaceConstraint';
    interfaceType: TypeExpression; // e.g., Display, Mappable K V
}
export function interfaceConstraint(interfaceType: TypeExpression): InterfaceConstraint {
    return { type: 'InterfaceConstraint', interfaceType };
}

export interface Constraint extends AstNode { // For where clauses
    type: 'Constraint';
    typeParam: Identifier; // The generic param being constrained (e.g., T)
    constraints: InterfaceConstraint[]; // The interfaces it must implement
}
export function constraint(typeParam: Identifier, constraints: InterfaceConstraint[]): Constraint {
    return { type: 'Constraint', typeParam, constraints };
}


// --- Control Flow ---

export interface OrClause extends AstNode {
    type: 'OrClause';
    condition?: Expression; // Condition is optional for the final 'else' case
    body: BlockExpression;
}
export function orClause(body: BlockExpression, condition?: Expression): OrClause {
    return { type: 'OrClause', condition, body };
}


export interface IfExpression extends AstNode {
    type: 'IfExpression';
    ifCondition: Expression;
    ifBody: BlockExpression;
    orClauses: OrClause[]; // Includes 'or condition {}' and the final 'or {}'
}
export function ifExpression(
    ifCondition: Expression,
    ifBody: BlockExpression,
    orClauses: OrClause[] = []
): IfExpression {
    return { type: 'IfExpression', ifCondition, ifBody, orClauses };
}

export interface MatchClause extends AstNode {
    type: 'MatchClause';
    pattern: Pattern;
    guard?: Expression;
    body: BlockExpression | Expression; // Allow single expression result
}
export function matchClause(pattern: Pattern, body: BlockExpression | Expression, guard?: Expression): MatchClause {
    return { type: 'MatchClause', pattern, guard, body };
}

export interface MatchExpression extends AstNode {
    type: 'MatchExpression';
    subject: Expression;
    clauses: MatchClause[];
}
export function matchExpression(subject: Expression, clauses: MatchClause[]): MatchExpression {
    return { type: 'MatchExpression', subject, clauses };
}

export interface WhileLoop extends AstNode {
    type: 'WhileLoop';
    condition: Expression;
    body: BlockExpression;
}
export function whileLoop(condition: Expression, body: BlockExpression): WhileLoop {
    return { type: 'WhileLoop', condition, body };
}

export interface ForLoop extends AstNode {
    type: 'ForLoop';
    pattern: Pattern;
    iterable: Expression;
    body: BlockExpression;
}
export function forLoop(pattern: Pattern, iterable: Expression, body: BlockExpression): ForLoop {
    return { type: 'ForLoop', pattern, iterable, body };
}

// --- Error Handling ---

export interface PropagationExpression extends AstNode {
    type: 'PropagationExpression'; // Represents `expr!`
    expression: Expression;
}
export function propagationExpression(expression: Expression): PropagationExpression {
    return { type: 'PropagationExpression', expression };
}

export interface OrElseExpression extends AstNode {
    type: 'OrElseExpression'; // Represents `expr else { handler }` or `expr else |err| { handler }`
    expression: Expression;
    handler: BlockExpression;
    errorBinding?: Identifier; // Optional binding for the error in `|err|`
}
export function orElseExpression(expression: Expression, handler: BlockExpression, errorBinding?: Identifier): OrElseExpression {
    return { type: 'OrElseExpression', expression, handler, errorBinding };
}

export interface RaiseStatement extends AstNode {
    type: 'RaiseStatement';
    expression: Expression; // The error value being raised
}
export function raiseStatement(expression: Expression): RaiseStatement {
    return { type: 'RaiseStatement', expression };
}

export interface RescueExpression extends AstNode {
    type: 'RescueExpression'; // Represents `expr rescue { cases }`
    monitoredExpression: Expression;
    clauses: MatchClause[]; // Reuse MatchClause for handling exceptions
}
export function rescueExpression(monitoredExpression: Expression, clauses: MatchClause[]): RescueExpression {
    return { type: 'RescueExpression', monitoredExpression, clauses };
}


// --- Interfaces, Impls, Methods (Placeholders/Basic Structure) ---

export interface FunctionSignature extends AstNode {
    type: 'FunctionSignature';
    name: Identifier;
    genericParams?: GenericParameter[];
    params: FunctionParameter[];
    returnType?: TypeExpression;
    whereClause?: Constraint[];
    hasSelfParam: boolean; // Distinguish static vs instance methods in interface
}
// DSL needed

export interface InterfaceDefinition extends AstNode {
    type: 'InterfaceDefinition';
    id: Identifier;
    genericParams?: GenericParameter[];
    selfTypePattern?: TypeExpression; // For `interface X for SelfTypePattern`
    signatures: FunctionSignature[];
    whereClause?: Constraint[];
    // Add composite interface info? baseInterfaces?: TypeExpression[]
}
// DSL needed

export interface ImplementationDefinition extends AstNode {
    type: 'ImplementationDefinition';
    implName?: Identifier; // For named impls: `MyImpl = impl ...`
    genericParams?: GenericParameter[];
    interfaceName: Identifier;
    interfaceArgs?: TypeExpression[];
    targetType: TypeExpression;
    definitions: FunctionDefinition[]; // Concrete method implementations
    whereClause?: Constraint[];
}
// DSL needed

export interface MethodsDefinition extends AstNode {
    type: 'MethodsDefinition';
    targetType: TypeExpression;
    definitions: FunctionDefinition[]; // Inherent methods
    genericParams?: GenericParameter[]; // Generics for the methods block itself (rare)
    whereClause?: Constraint[];
}
// DSL needed


// --- Packages & Modules (Placeholders) ---

export interface PackageStatement extends AstNode {
    type: 'PackageStatement';
    name: Identifier; // Unqualified name
}
export function packageStatement(name: Identifier): PackageStatement {
    return { type: 'PackageStatement', name };
}

export interface ImportSelector extends AstNode {
    type: 'ImportSelector';
    name: Identifier;
    alias?: Identifier;
}
export function importSelector(name: Identifier, alias?: Identifier): ImportSelector {
    return { type: 'ImportSelector', name, alias };
}

export interface ImportStatement extends AstNode {
    type: 'ImportStatement';
    packagePath: Identifier[]; // e.g., ['io', 'network'] for import io.network
    isWildcard: boolean; // For `import io.*`
    selectors?: ImportSelector[]; // For `import io.{puts, gets}`
    alias?: Identifier; // For `import io as myio`
}
// DSL needed


// --- Concurrency (Placeholders) ---
export interface ProcExpression extends AstNode {
    type: 'ProcExpression';
    expression: FunctionCall | BlockExpression;
}
export function procExpression(expression: FunctionCall | BlockExpression): ProcExpression {
    return { type: 'ProcExpression', expression };
}

export interface SpawnExpression extends AstNode {
    type: 'SpawnExpression';
    expression: FunctionCall | BlockExpression;
}
export function spawnExpression(expression: FunctionCall | BlockExpression): SpawnExpression {
    return { type: 'SpawnExpression', expression };
}


// Add more node types and DSL functions as needed...
