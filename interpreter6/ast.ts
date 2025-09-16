// =============================================================================
// Base AST Node
// =============================================================================

export interface AstNode {
  type: string;
  // Optional: Add common properties like location/span if needed later
  // loc?: { start: Position; end: Position };
}

// =============================================================================
// Identifiers
// =============================================================================

export interface Identifier extends AstNode {
  type: 'Identifier';
  name: string;
}
export function identifier(name: string): Identifier {
  return { type: 'Identifier', name };
}

// =============================================================================
// Literals
// =============================================================================

export type Literal =
  | StringLiteral
  | IntegerLiteral
  | FloatLiteral
  | BooleanLiteral
  | NilLiteral
  | CharLiteral
  | ArrayLiteral // Added ArrayLiteral here as it's a literal form
  ;

export interface StringLiteral extends AstNode {
  type: 'StringLiteral';
  value: string;
  // isInterpolated: boolean; // Removed, handled by StringInterpolation node
}
export function stringLiteral(value: string): StringLiteral {
  return { type: 'StringLiteral', value };
}

export interface IntegerLiteral extends AstNode {
  type: 'IntegerLiteral';
  value: bigint | number; // Use bigint for larger types if needed by target
  integerType?: 'i8' | 'i16' | 'i32' | 'i64' | 'i128' | 'u8' | 'u16' | 'u32' | 'u64' | 'u128'; // Suffix determines this
}
export function integerLiteral(value: bigint | number, type?: IntegerLiteral['integerType']): IntegerLiteral {
  return { type: 'IntegerLiteral', value, integerType: type };
}

export interface FloatLiteral extends AstNode {
    type: 'FloatLiteral';
    value: number;
    floatType?: 'f32' | 'f64'; // Suffix determines this
}
export function floatLiteral(value: number, type?: FloatLiteral['floatType']): FloatLiteral {
    return { type: 'FloatLiteral', value, floatType: type };
}

export interface BooleanLiteral extends AstNode {
  type: 'BooleanLiteral';
  value: boolean;
}
export function booleanLiteral(value: boolean): BooleanLiteral {
  return { type: 'BooleanLiteral', value };
}

export interface NilLiteral extends AstNode {
  type: 'NilLiteral';
  value: null; // Explicitly null
}
export function nilLiteral(): NilLiteral {
  return { type: 'NilLiteral', value: null };
}

export interface CharLiteral extends AstNode {
    type: 'CharLiteral';
    value: string; // Store the single character as a string
}
export function charLiteral(value: string): CharLiteral {
    // TODO: Add validation? Ensure it's a single char?
    return { type: 'CharLiteral', value };
}

export interface ArrayLiteral extends AstNode {
    type: 'ArrayLiteral';
    elements: Expression[];
}
export function arrayLiteral(elements: Expression[]): ArrayLiteral {
    return { type: 'ArrayLiteral', elements };
}

// =============================================================================
// Type Expressions
// =============================================================================

export type TypeExpression =
  | SimpleTypeExpression // e.g., i32, string, MyType
  | GenericTypeExpression // e.g., Array i32, Map string User
  | FunctionTypeExpression // e.g., (i32, string) -> bool
  | NullableTypeExpression // e.g., ?string
  | ResultTypeExpression // e.g., !string
  | WildcardTypeExpression // e.g., _
  ;

export interface SimpleTypeExpression extends AstNode {
    type: 'SimpleTypeExpression';
    name: Identifier; // e.g., i32, string, MyType, Self
}
export function simpleTypeExpression(name: Identifier | string): SimpleTypeExpression {
    return { type: 'SimpleTypeExpression', name: typeof name === 'string' ? identifier(name) : name };
}

export interface GenericTypeExpression extends AstNode {
    type: 'GenericTypeExpression';
    base: TypeExpression; // The base type constructor (e.g., Array, Map string)
    arguments: TypeExpression[]; // The type arguments (e.g., [i32] for Array i32)
}
export function genericTypeExpression(base: TypeExpression, args: TypeExpression[]): GenericTypeExpression {
    return { type: 'GenericTypeExpression', base, arguments: args };
}

export interface FunctionTypeExpression extends AstNode {
    type: 'FunctionTypeExpression';
    paramTypes: TypeExpression[];
    returnType: TypeExpression;
}
export function functionTypeExpression(paramTypes: TypeExpression[], returnType: TypeExpression): FunctionTypeExpression {
    return { type: 'FunctionTypeExpression', paramTypes, returnType };
}

export interface NullableTypeExpression extends AstNode {
    type: 'NullableTypeExpression'; // Represents ?Type
    innerType: TypeExpression;
}
export function nullableTypeExpression(innerType: TypeExpression): NullableTypeExpression {
    return { type: 'NullableTypeExpression', innerType };
}

export interface ResultTypeExpression extends AstNode {
    type: 'ResultTypeExpression'; // Represents !Type
    innerType: TypeExpression;
}
export function resultTypeExpression(innerType: TypeExpression): ResultTypeExpression {
    return { type: 'ResultTypeExpression', innerType };
}

export interface WildcardTypeExpression extends AstNode {
    type: 'WildcardTypeExpression'; // Represents _ in type context
}
export function wildcardTypeExpression(): WildcardTypeExpression {
    return { type: 'WildcardTypeExpression' };
}

// =============================================================================
// Generics & Constraints
// =============================================================================

export interface InterfaceConstraint extends AstNode {
    type: 'InterfaceConstraint';
    interfaceType: TypeExpression; // e.g., Display, Mappable K V
}
export function interfaceConstraint(interfaceType: TypeExpression): InterfaceConstraint {
    return { type: 'InterfaceConstraint', interfaceType };
}

export interface GenericParameter extends AstNode {
    type: 'GenericParameter';
    name: Identifier;
    constraints?: InterfaceConstraint[]; // T: Display + Clone
}
export function genericParameter(name: Identifier | string, constraints?: InterfaceConstraint[]): GenericParameter {
    return { type: 'GenericParameter', name: typeof name === 'string' ? identifier(name) : name, constraints };
}

export interface WhereClauseConstraint extends AstNode { // For where clauses
    type: 'WhereClauseConstraint';
    typeParam: Identifier; // The generic param being constrained (e.g., T)
    constraints: InterfaceConstraint[]; // The interfaces it must implement
}
export function whereClauseConstraint(typeParam: Identifier | string, constraints: InterfaceConstraint[]): WhereClauseConstraint {
    return { type: 'WhereClauseConstraint', typeParam: typeof typeParam === 'string' ? identifier(typeParam) : typeParam, constraints };
}

// =============================================================================
// Patterns (for Assignment/Destructuring/Match)
// =============================================================================

export type Pattern =
  | Identifier // Simple binding
  | WildcardPattern
  | LiteralPattern
  | StructPattern
  | ArrayPattern
  ;

export interface WildcardPattern extends AstNode {
    type: 'WildcardPattern'; // Represents _
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
    fieldName?: Identifier; // Null for positional
    pattern: Pattern; // Nested pattern for the field
    binding?: Identifier; // For `field @ binding : Pattern` (only with := ?) - TBD if needed in AST
}
export function structPatternField(pattern: Pattern, fieldName?: Identifier | string): StructPatternField {
    const name = typeof fieldName === 'string' ? identifier(fieldName) : fieldName;
    return { type: 'StructPatternField', fieldName: name, pattern };
}

export interface StructPattern extends AstNode {
    type: 'StructPattern';
    structType?: Identifier; // Optional type check (e.g., Point { ... })
    fields: StructPatternField[]; // Named or positional fields
    isPositional: boolean; // Distinguishes { a, b } from { x: a, y: b }
}
export function structPattern(fields: StructPatternField[], isPositional: boolean, structType?: Identifier | string): StructPattern {
    const typeId = typeof structType === 'string' ? identifier(structType) : structType;
    return { type: 'StructPattern', structType: typeId, fields, isPositional };
}

export interface ArrayPattern extends AstNode {
    type: 'ArrayPattern';
    elements: Pattern[];
    restPattern?: Identifier | WildcardPattern; // Optional identifier or wildcard for ...rest or ...
}
export function arrayPattern(elements: Pattern[], restPattern?: Identifier | WildcardPattern | string): ArrayPattern {
    let rest: Identifier | WildcardPattern | undefined = undefined;
    if (typeof restPattern === 'string') {
        rest = identifier(restPattern);
    } else if (restPattern) {
        rest = restPattern;
    }
    return { type: 'ArrayPattern', elements, restPattern: rest };
}

// =============================================================================
// Expressions
// =============================================================================

export type Expression =
  | Identifier
  | Literal // Includes ArrayLiteral
  | BinaryExpression
  | UnaryExpression
  | FunctionCall
  | BlockExpression
  | AssignmentExpression
  | IfExpression
  | MatchExpression
  | StructLiteral
  | MemberAccessExpression
  | PropagationExpression // For `!`
  | OrElseExpression // For `else {}`
  | RangeExpression
  | StringInterpolation
  | LambdaExpression // Added for { => }
  | ProcExpression // Added
  | SpawnExpression // Added
  | BreakpointExpression // Added
  | RescueExpression // Added
  ;

export interface UnaryExpression extends AstNode {
    type: 'UnaryExpression';
    operator: '-' | '!' | '~';
    operand: Expression;
}
export function unaryExpression(operator: UnaryExpression['operator'], operand: Expression): UnaryExpression {
    return { type: 'UnaryExpression', operator, operand };
}

export interface BinaryExpression extends AstNode {
  type: 'BinaryExpression';
  operator: string; // +, -, *, /, %, ==, !=, <, >, <=, >=, &&, ||, &, |, ^, <<, >>, |>
  left: Expression;
  right: Expression;
}
export function binaryExpression(operator: string, left: Expression, right: Expression): BinaryExpression {
  return { type: 'BinaryExpression', operator, left, right };
}

export interface FunctionCall extends AstNode {
  type: 'FunctionCall';
  callee: Expression; // Identifier, MemberAccess, or other expression evaluating to a function
  arguments: Expression[];
  typeArguments?: TypeExpression[]; // For explicit generic args like func<T>()
  isTrailingLambda: boolean; // Indicates if the last arg was provided using trailing lambda syntax
}
export function functionCall(callee: Expression, args: Expression[], typeArgs?: TypeExpression[], isTrailingLambda = false): FunctionCall {
  return { type: 'FunctionCall', callee, arguments: args, typeArguments: typeArgs, isTrailingLambda };
}

export interface BlockExpression extends AstNode {
    type: 'BlockExpression'; // Represents `do { ... }` or implicit blocks in control flow
    body: Statement[]; // Sequence of statements/expressions
}
export function blockExpression(body: Statement[]): BlockExpression {
    return { type: 'BlockExpression', body };
}

export interface AssignmentExpression extends AstNode {
    type: 'AssignmentExpression';
    operator: ':=' | '=' | '+=' | '-=' | '*=' | '/=' | '%=' | '&=' | '|=' | '^=' | '<<=' | '>>='; // Declaration or Assignment/Mutation
    left: Pattern | MemberAccessExpression; // LHS: Pattern for := or destructuring =, MemberAccess for mutation =
    right: Expression;
}
export function assignmentExpression(operator: AssignmentExpression['operator'], left: Pattern | MemberAccessExpression, right: Expression): AssignmentExpression {
    return { type: 'AssignmentExpression', operator, left, right };
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

export interface StringInterpolation extends AstNode {
    type: 'StringInterpolation'; // Represents `... ${expr} ...`
    parts: (StringLiteral | Expression)[]; // Alternating string parts and expressions
}
export function stringInterpolation(parts: (StringLiteral | Expression)[]): StringInterpolation {
    return { type: 'StringInterpolation', parts };
}

export interface MemberAccessExpression extends AstNode {
    type: 'MemberAccessExpression';
    object: Expression; // The struct instance or other object
    member: Identifier | IntegerLiteral; // Field name or positional index (e.g., .0, .1)
    // isImplicitSelf: boolean; // Maybe track `#member` usage? Or handle in interpreter. Let's omit for now.
}
export function memberAccessExpression(object: Expression, member: Identifier | IntegerLiteral | string): MemberAccessExpression {
    let memberNode: Identifier | IntegerLiteral;
    if (typeof member === 'string') {
        memberNode = identifier(member);
    } else {
        memberNode = member;
    }
    return { type: 'MemberAccessExpression', object, member: memberNode };
}

// =============================================================================
// Statements & Definitions
// =============================================================================

export type Statement =
  | Expression // Expressions can be statements
  | FunctionDefinition
  | StructDefinition
  | UnionDefinition
  | InterfaceDefinition
  | ImplementationDefinition
  | MethodsDefinition
  | ImportStatement
  | PackageStatement
  | ReturnStatement
  | RaiseStatement
  | BreakStatement // Added
  | WhileLoop
  | ForLoop
  ;

// Note: VariableDeclaration is handled by AssignmentExpression with operator ':='

// =============================================================================
// Structs
// =============================================================================

export interface StructFieldDefinition extends AstNode {
    type: 'StructFieldDefinition';
    name?: Identifier; // Null for positional fields
    fieldType: TypeExpression; // Type annotation
}
export function structFieldDefinition(fieldType: TypeExpression, name?: Identifier | string): StructFieldDefinition {
    const id = typeof name === 'string' ? identifier(name) : name;
    return { type: 'StructFieldDefinition', name: id, fieldType };
}

export interface StructDefinition extends AstNode {
    type: 'StructDefinition';
    id: Identifier;
    genericParams?: GenericParameter[];
    fields: StructFieldDefinition[];
    whereClause?: WhereClauseConstraint[];
    kind: 'singleton' | 'named' | 'positional';
}
export function structDefinition(
    id: Identifier | string,
    fields: StructFieldDefinition[],
    kind: StructDefinition['kind'],
    genericParams?: GenericParameter[],
    whereClause?: WhereClauseConstraint[]
): StructDefinition {
    const typeId = typeof id === 'string' ? identifier(id) : id;
    return { type: 'StructDefinition', id: typeId, fields, kind, genericParams, whereClause };
}

export interface StructFieldInitializer extends AstNode {
    type: 'StructFieldInitializer';
    name?: Identifier; // Null for positional or shorthand `{ name }`
    value: Expression;
    isShorthand: boolean; // Indicates if `{ name }` shorthand was used
}
export function structFieldInitializer(value: Expression, name?: Identifier | string, isShorthand = false): StructFieldInitializer {
    const id = typeof name === 'string' ? identifier(name) : name;
    return { type: 'StructFieldInitializer', name: id, value, isShorthand };
}

export interface StructLiteral extends AstNode {
    type: 'StructLiteral';
    structType?: Identifier; // Optional: Name of the struct being instantiated (often inferred)
    fields: StructFieldInitializer[]; // Can be named or positional
    isPositional: boolean; // Distinguishes { a, b } from { x: a, y: b }
    functionalUpdateSource?: Expression; // For `{ ...source, field: val }`
    // Add generic arguments? typeArguments?: TypeExpression[];
}
export function structLiteral(
    fields: StructFieldInitializer[],
    isPositional: boolean,
    structType?: Identifier | string,
    functionalUpdateSource?: Expression
): StructLiteral {
    const typeId = typeof structType === 'string' ? identifier(structType) : structType;
    return { type: 'StructLiteral', structType: typeId, fields, isPositional, functionalUpdateSource };
}

// =============================================================================
// Unions
// =============================================================================

export interface UnionDefinition extends AstNode {
    type: 'UnionDefinition';
    id: Identifier;
    genericParams?: GenericParameter[];
    variants: TypeExpression[]; // List of types in the union
    whereClause?: WhereClauseConstraint[];
}
export function unionDefinition(
    id: Identifier | string,
    variants: TypeExpression[],
    genericParams?: GenericParameter[],
    whereClause?: WhereClauseConstraint[]
): UnionDefinition {
    const typeId = typeof id === 'string' ? identifier(id) : id;
    return { type: 'UnionDefinition', id: typeId, genericParams, variants, whereClause };
}

// =============================================================================
// Functions & Lambdas
// =============================================================================

export interface FunctionParameter extends AstNode {
    type: 'FunctionParameter';
    name: Pattern; // CHANGED: was Identifier, now Pattern to allow destructuring
    paramType?: TypeExpression; // Type annotation (optional for lambdas?)
}
export function functionParameter(name: Pattern | Identifier | string, paramType?: TypeExpression): FunctionParameter {
    // Accept Pattern, Identifier, or string for convenience
    let pattern: Pattern;
    if (typeof name === 'string') {
        pattern = identifier(name);
    } else if ((name as Identifier).type === 'Identifier') {
        pattern = name as Identifier;
    } else {
        pattern = name as Pattern;
    }
    return { type: 'FunctionParameter', name: pattern, paramType };
}

// Represents `fn name(...) { ... }`
export interface FunctionDefinition extends AstNode {
  type: 'FunctionDefinition';
  id: Identifier; // Must have a name
  genericParams?: GenericParameter[];
  params: FunctionParameter[];
  returnType?: TypeExpression;
  body: BlockExpression; // Always a block for named functions
  whereClause?: WhereClauseConstraint[];
  isMethodShorthand: boolean; // True if defined using `fn #method`
  isPrivate: boolean; // Added for `private fn ...`
}
export function functionDefinition(
    id: Identifier | string,
    params: FunctionParameter[],
    body: BlockExpression,
    returnType?: TypeExpression,
    genericParams?: GenericParameter[],
    whereClause?: WhereClauseConstraint[],
    isMethodShorthand = false,
    isPrivate = false,
): FunctionDefinition {
  const funcId = typeof id === 'string' ? identifier(id) : id;
  return { type: 'FunctionDefinition', id: funcId, params, body, returnType, genericParams, whereClause, isMethodShorthand, isPrivate };
}

// Represents `{ [params] [-> Type] => expression }` or `fn([params]) [-> Type] { body }`
export interface LambdaExpression extends AstNode {
    type: 'LambdaExpression';
    genericParams?: GenericParameter[]; // Only for verbose `fn(...) { ... }` form
    params: FunctionParameter[];
    returnType?: TypeExpression;
    body: Expression | BlockExpression; // Expression for `=>`, Block for `fn{}`
    whereClause?: WhereClauseConstraint[]; // Only for verbose `fn(...) { ... }` form
    isVerboseSyntax: boolean; // true for `fn(...) { ... }`, false for `{ => }`
}
export function lambdaExpression(
    params: FunctionParameter[],
    body: Expression | BlockExpression,
    returnType?: TypeExpression,
    genericParams?: GenericParameter[],
    whereClause?: WhereClauseConstraint[],
    isVerboseSyntax = false
): LambdaExpression {
    return { type: 'LambdaExpression', params, body, returnType, genericParams, whereClause, isVerboseSyntax };
}


export interface ReturnStatement extends AstNode {
    type: 'ReturnStatement';
    argument?: Expression; // Optional argument (for `return;` vs `return val;`)
}
export function returnStatement(argument?: Expression): ReturnStatement {
    return { type: 'ReturnStatement', argument };
}

// =============================================================================
// Control Flow
// =============================================================================

export interface OrClause extends AstNode {
    type: 'OrClause';
    condition?: Expression; // Condition is optional for the final 'else' case `or { ... }`
    body: BlockExpression;
}
export function orClause(body: BlockExpression, condition?: Expression): OrClause {
    return { type: 'OrClause', condition, body };
}

export interface IfExpression extends AstNode {
    type: 'IfExpression'; // Represents `if cond { } or cond2 { } or { }`
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
    guard?: Expression; // `if guard`
    body: Expression; // Result expression (can be BlockExpression)
}
export function matchClause(pattern: Pattern, body: Expression, guard?: Expression): MatchClause {
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
    pattern: Pattern; // Pattern to bind loop variable(s)
    iterable: Expression; // Expression evaluating to an Iterable
    body: BlockExpression;
}
export function forLoop(pattern: Pattern, iterable: Expression, body: BlockExpression): ForLoop {
    return { type: 'ForLoop', pattern, iterable, body };
}

export interface BreakpointExpression extends AstNode {
    type: 'BreakpointExpression'; // Represents `breakpoint 'label { ... }`
    label: Identifier;
    body: BlockExpression;
}
export function breakpointExpression(label: Identifier | string, body: BlockExpression): BreakpointExpression {
    const id = typeof label === 'string' ? identifier(label) : label;
    return { type: 'BreakpointExpression', label: id, body };
}

export interface BreakStatement extends AstNode {
    type: 'BreakStatement'; // Represents `break 'label value`
    label: Identifier;
    value: Expression;
}
export function breakStatement(label: Identifier | string, value: Expression): BreakStatement {
    const id = typeof label === 'string' ? identifier(label) : label;
    return { type: 'BreakStatement', label: id, value };
}

// =============================================================================
// Error Handling
// =============================================================================

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
export function orElseExpression(expression: Expression, handler: BlockExpression, errorBinding?: Identifier | string): OrElseExpression {
    const id = typeof errorBinding === 'string' ? identifier(errorBinding) : errorBinding;
    return { type: 'OrElseExpression', expression, handler, errorBinding: id };
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

// =============================================================================
// Interfaces, Implementations, Methods
// =============================================================================

export interface FunctionSignature extends AstNode {
    type: 'FunctionSignature';
    name: Identifier;
    genericParams?: GenericParameter[];
    params: FunctionParameter[];
    returnType?: TypeExpression;
    whereClause?: WhereClauseConstraint[];
    // hasSelfParam: boolean; // Infer from params[0] name/type? Let's check params[0].name === 'self'
    defaultImpl?: BlockExpression; // Added for default method implementations
}
export function functionSignature(
    name: Identifier | string,
    params: FunctionParameter[],
    returnType?: TypeExpression,
    genericParams?: GenericParameter[],
    whereClause?: WhereClauseConstraint[],
    defaultImpl?: BlockExpression,
): FunctionSignature {
    const id = typeof name === 'string' ? identifier(name) : name;
    return { type: 'FunctionSignature', name: id, params, returnType, genericParams, whereClause, defaultImpl };
}

export interface InterfaceDefinition extends AstNode {
    type: 'InterfaceDefinition';
    id: Identifier;
    genericParams?: GenericParameter[]; // Generics for the interface itself (e.g., `interface Mappable K V`)
    selfTypePattern?: TypeExpression; // For `interface X for SelfTypePattern` (e.g., `T`, `M _`, `Array T`)
    signatures: FunctionSignature[];
    whereClause?: WhereClauseConstraint[];
    baseInterfaces?: TypeExpression[]; // For composite interfaces `interface X = A + B`
    isPrivate: boolean;
}
export function interfaceDefinition(
    id: Identifier | string,
    signatures: FunctionSignature[],
    genericParams?: GenericParameter[],
    selfTypePattern?: TypeExpression,
    whereClause?: WhereClauseConstraint[],
    baseInterfaces?: TypeExpression[],
    isPrivate = false,
): InterfaceDefinition {
    const typeId = typeof id === 'string' ? identifier(id) : id;
    return { type: 'InterfaceDefinition', id: typeId, genericParams, selfTypePattern, signatures, whereClause, baseInterfaces, isPrivate };
}

export interface ImplementationDefinition extends AstNode {
    type: 'ImplementationDefinition';
    implName?: Identifier; // For named impls: `MyImpl = impl ...`
    genericParams?: GenericParameter[]; // Generics for the impl block `<T: Clone>`
    interfaceName: Identifier;
    interfaceArgs?: TypeExpression[]; // Args for the interface `impl Display for Array T` -> interfaceName=Display, interfaceArgs=[] ? No, this seems wrong. Should be part of targetType?
                                      // Let's rethink: `impl<T> Mappable A for Array` -> interfaceName=Mappable, interfaceArgs=[A]? Yes.
    targetType: TypeExpression; // The type implementing the interface (e.g., `Point`, `Array T`, `Array`)
    definitions: FunctionDefinition[]; // Concrete method implementations
    whereClause?: WhereClauseConstraint[];
}
export function implementationDefinition(
    interfaceName: Identifier | string,
    targetType: TypeExpression,
    definitions: FunctionDefinition[],
    implName?: Identifier | string,
    genericParams?: GenericParameter[],
    interfaceArgs?: TypeExpression[],
    whereClause?: WhereClauseConstraint[],
): ImplementationDefinition {
    const ifaceId = typeof interfaceName === 'string' ? identifier(interfaceName) : interfaceName;
    const nameId = typeof implName === 'string' ? identifier(implName) : implName;
    return { type: 'ImplementationDefinition', implName: nameId, genericParams, interfaceName: ifaceId, interfaceArgs, targetType, definitions, whereClause };
}

export interface MethodsDefinition extends AstNode {
    type: 'MethodsDefinition';
    targetType: TypeExpression; // The type the methods are for (e.g., `Point`, `Array T`)
    genericParams?: GenericParameter[]; // Generics for the methods block itself (rare)
    definitions: FunctionDefinition[]; // Inherent methods (instance or static)
    whereClause?: WhereClauseConstraint[];
}
export function methodsDefinition(
    targetType: TypeExpression,
    definitions: FunctionDefinition[],
    genericParams?: GenericParameter[],
    whereClause?: WhereClauseConstraint[],
): MethodsDefinition {
    return { type: 'MethodsDefinition', targetType, genericParams, definitions, whereClause };
}

// =============================================================================
// Packages & Modules
// =============================================================================

export interface PackageStatement extends AstNode {
    type: 'PackageStatement';
    namePath: Identifier[]; // Full path declared, e.g., ['utils', 'fmt'] for `package utils.fmt;`
}
export function packageStatement(namePath: (Identifier | string)[]): PackageStatement {
    const path = namePath.map(p => typeof p === 'string' ? identifier(p) : p);
    return { type: 'PackageStatement', namePath: path };
}

export interface ImportSelector extends AstNode {
    type: 'ImportSelector';
    name: Identifier;
    alias?: Identifier;
}
export function importSelector(name: Identifier | string, alias?: Identifier | string): ImportSelector {
    const nameId = typeof name === 'string' ? identifier(name) : name;
    const aliasId = typeof alias === 'string' ? identifier(alias) : alias;
    return { type: 'ImportSelector', name: nameId, alias: aliasId };
}

export interface ImportStatement extends AstNode {
    type: 'ImportStatement';
    packagePath: Identifier[]; // e.g., ['io', 'network'] for import io.network
    isWildcard: boolean; // For `import io.*`
    selectors?: ImportSelector[]; // For `import io.{puts, gets}`
    alias?: Identifier; // For `import io as myio`
}
export function importStatement(
    packagePath: (Identifier | string)[],
    isWildcard = false,
    selectors?: ImportSelector[],
    alias?: Identifier | string
): ImportStatement {
    const path = packagePath.map(p => typeof p === 'string' ? identifier(p) : p);
    const aliasId = typeof alias === 'string' ? identifier(alias) : alias;
    return { type: 'ImportStatement', packagePath: path, isWildcard, selectors, alias: aliasId };
}

// =============================================================================
// Concurrency
// =============================================================================

export interface ProcExpression extends AstNode {
    type: 'ProcExpression'; // Represents `proc expr`
    expression: FunctionCall | BlockExpression;
}
export function procExpression(expression: FunctionCall | BlockExpression): ProcExpression {
    return { type: 'ProcExpression', expression };
}

export interface SpawnExpression extends AstNode {
    type: 'SpawnExpression'; // Represents `spawn expr`
    expression: FunctionCall | BlockExpression;
}
export function spawnExpression(expression: FunctionCall | BlockExpression): SpawnExpression {
    return { type: 'SpawnExpression', expression };
}

// =============================================================================
// Program / Module Root
// =============================================================================

export interface Module extends AstNode {
    type: 'Module';
    package?: PackageStatement;
    imports: ImportStatement[];
    body: Statement[]; // Top-level definitions and statements
}
export function module(body: Statement[], imports: ImportStatement[] = [], pkg?: PackageStatement): Module {
    return { type: 'Module', package: pkg, imports, body };
}
