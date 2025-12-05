// =============================================================================
// Able v10 AST (minimal, complete coverage of concepts; TS-only data model)
// =============================================================================

// Base node with discriminant
export interface Position {
  line: number;
  column: number;
}

export interface Span {
  start: Position;
  end: Position;
}

export interface AstNode {
  type: string;
  span?: Span;
  origin?: string;
}

// -----------------------------------------------------------------------------
// Identifiers and common utility nodes
// -----------------------------------------------------------------------------

export interface Identifier extends AstNode {
  type: 'Identifier';
  name: string;
}
export function identifier(name: string): Identifier {
  return { type: 'Identifier', name };
}

// -----------------------------------------------------------------------------
// Literals
// -----------------------------------------------------------------------------

export interface StringLiteral extends AstNode { type: 'StringLiteral'; value: string; }
export interface IntegerLiteral extends AstNode {
  type: 'IntegerLiteral';
  value: bigint | number;
  integerType?: 'i8' | 'i16' | 'i32' | 'i64' | 'i128' | 'u8' | 'u16' | 'u32' | 'u64' | 'u128';
}
export interface FloatLiteral extends AstNode { type: 'FloatLiteral'; value: number; floatType?: 'f32' | 'f64'; }
export interface BooleanLiteral extends AstNode { type: 'BooleanLiteral'; value: boolean; }
export interface NilLiteral extends AstNode { type: 'NilLiteral'; value: null; }
export interface CharLiteral extends AstNode { type: 'CharLiteral'; value: string; }

export interface ArrayLiteral extends AstNode { type: 'ArrayLiteral'; elements: Expression[]; }
export interface MapLiteralEntry extends AstNode { type: 'MapLiteralEntry'; key: Expression; value: Expression; }
export interface MapLiteralSpread extends AstNode { type: 'MapLiteralSpread'; expression: Expression; }
export interface MapLiteral extends AstNode { type: 'MapLiteral'; entries: (MapLiteralEntry | MapLiteralSpread)[]; }

export type Literal =
  | StringLiteral
  | IntegerLiteral
  | FloatLiteral
  | BooleanLiteral
  | NilLiteral
  | CharLiteral
  | ArrayLiteral
  | MapLiteral;

export function stringLiteral(value: string): StringLiteral { return { type: 'StringLiteral', value }; }
export function integerLiteral(value: bigint | number, integerType?: IntegerLiteral['integerType']): IntegerLiteral {
  return { type: 'IntegerLiteral', value, integerType };
}
export function floatLiteral(value: number, floatType?: FloatLiteral['floatType']): FloatLiteral {
  return { type: 'FloatLiteral', value, floatType };
}
export function booleanLiteral(value: boolean): BooleanLiteral { return { type: 'BooleanLiteral', value }; }
export function nilLiteral(): NilLiteral { return { type: 'NilLiteral', value: null }; }
export function charLiteral(value: string): CharLiteral { return { type: 'CharLiteral', value }; }
export function arrayLiteral(elements: Expression[]): ArrayLiteral { return { type: 'ArrayLiteral', elements }; }
export function mapLiteralEntry(key: Expression, value: Expression): MapLiteralEntry {
  return { type: 'MapLiteralEntry', key, value };
}
export function mapLiteralSpread(expression: Expression): MapLiteralSpread {
  return { type: 'MapLiteralSpread', expression };
}
export function mapLiteral(entries: (MapLiteralEntry | MapLiteralSpread)[]): MapLiteral {
  return { type: 'MapLiteral', entries };
}

// -----------------------------------------------------------------------------
// Types (surface type expressions sufficient to represent v10 authoring)
// -----------------------------------------------------------------------------

export interface SimpleTypeExpression extends AstNode { type: 'SimpleTypeExpression'; name: Identifier; }
export interface GenericTypeExpression extends AstNode { type: 'GenericTypeExpression'; base: TypeExpression; arguments: TypeExpression[]; }
export interface FunctionTypeExpression extends AstNode { type: 'FunctionTypeExpression'; paramTypes: TypeExpression[]; returnType: TypeExpression; }
export interface NullableTypeExpression extends AstNode { type: 'NullableTypeExpression'; innerType: TypeExpression; }
export interface ResultTypeExpression extends AstNode { type: 'ResultTypeExpression'; innerType: TypeExpression; }
export interface UnionTypeExpression extends AstNode { type: 'UnionTypeExpression'; members: TypeExpression[]; }
export interface WildcardTypeExpression extends AstNode { type: 'WildcardTypeExpression'; }

export type TypeExpression =
  | SimpleTypeExpression
  | GenericTypeExpression
  | FunctionTypeExpression
  | NullableTypeExpression
  | ResultTypeExpression
  | UnionTypeExpression
  | WildcardTypeExpression;

export function simpleTypeExpression(name: Identifier | string): SimpleTypeExpression {
  return { type: 'SimpleTypeExpression', name: typeof name === 'string' ? identifier(name) : name };
}
export function genericTypeExpression(base: TypeExpression, args: TypeExpression[]): GenericTypeExpression {
  return { type: 'GenericTypeExpression', base, arguments: args };
}
export function functionTypeExpression(paramTypes: TypeExpression[], returnType: TypeExpression): FunctionTypeExpression {
  return { type: 'FunctionTypeExpression', paramTypes, returnType };
}
export function nullableTypeExpression(innerType: TypeExpression): NullableTypeExpression { return { type: 'NullableTypeExpression', innerType }; }
export function resultTypeExpression(innerType: TypeExpression): ResultTypeExpression { return { type: 'ResultTypeExpression', innerType }; }
export function unionTypeExpression(members: TypeExpression[]): UnionTypeExpression { return { type: 'UnionTypeExpression', members }; }
export function wildcardTypeExpression(): WildcardTypeExpression { return { type: 'WildcardTypeExpression' }; }

// Generics and constraints
export interface InterfaceConstraint extends AstNode { type: 'InterfaceConstraint'; interfaceType: TypeExpression; }
export interface GenericParameter extends AstNode {
  type: 'GenericParameter';
  name: Identifier;
  constraints?: InterfaceConstraint[];
  isInferred?: boolean;
}
export interface WhereClauseConstraint extends AstNode { type: 'WhereClauseConstraint'; typeParam: Identifier; constraints: InterfaceConstraint[]; }

export function interfaceConstraint(interfaceType: TypeExpression): InterfaceConstraint { return { type: 'InterfaceConstraint', interfaceType }; }
export function genericParameter(
  name: Identifier | string,
  constraints?: InterfaceConstraint[],
  options?: { isInferred?: boolean },
): GenericParameter {
  return {
    type: 'GenericParameter',
    name: typeof name === 'string' ? identifier(name) : name,
    constraints,
    isInferred: options?.isInferred,
  };
}
export function whereClauseConstraint(typeParam: Identifier | string, constraints: InterfaceConstraint[]): WhereClauseConstraint {
  return { type: 'WhereClauseConstraint', typeParam: typeof typeParam === 'string' ? identifier(typeParam) : typeParam, constraints };
}

// -----------------------------------------------------------------------------
// Patterns (assignment, match, for-loop binding)
// -----------------------------------------------------------------------------

export interface WildcardPattern extends AstNode { type: 'WildcardPattern'; }
export interface LiteralPattern extends AstNode { type: 'LiteralPattern'; literal: Literal; }
export interface StructPatternField extends AstNode {
  type: 'StructPatternField';
  fieldName?: Identifier;
  pattern: Pattern;
  binding?: Identifier;
  typeAnnotation?: TypeExpression;
}
export interface StructPattern extends AstNode { type: 'StructPattern'; structType?: Identifier; fields: StructPatternField[]; isPositional: boolean; }
export interface ArrayPattern extends AstNode { type: 'ArrayPattern'; elements: Pattern[]; restPattern?: Identifier | WildcardPattern; }

export interface TypedPattern extends AstNode { type: 'TypedPattern'; pattern: Pattern; typeAnnotation: TypeExpression; }

export type Pattern = Identifier | WildcardPattern | LiteralPattern | StructPattern | ArrayPattern | TypedPattern;

export function typedPattern(pattern: Pattern, typeAnnotation: TypeExpression): TypedPattern {
  return { type: 'TypedPattern', pattern, typeAnnotation };
}

export function wildcardPattern(): WildcardPattern { return { type: 'WildcardPattern' }; }
export function literalPattern(literal: Literal): LiteralPattern { return { type: 'LiteralPattern', literal }; }
export function structPatternField(
  pattern: Pattern,
  fieldName?: Identifier | string,
  binding?: Identifier | string,
  typeAnnotation?: TypeExpression,
): StructPatternField {
  return {
    type: 'StructPatternField',
    fieldName: typeof fieldName === 'string' ? identifier(fieldName) : fieldName,
    pattern,
    binding: typeof binding === 'string' ? identifier(binding) : binding,
    typeAnnotation,
  };
}
export function structPattern(fields: StructPatternField[], isPositional: boolean, structType?: Identifier | string): StructPattern {
  return { type: 'StructPattern', structType: typeof structType === 'string' ? identifier(structType) : structType, fields, isPositional };
}
export function arrayPattern(elements: Pattern[], restPattern?: Identifier | WildcardPattern | string): ArrayPattern {
  let rest: Identifier | WildcardPattern | undefined = undefined;
  if (typeof restPattern === 'string') rest = identifier(restPattern);
  else if (restPattern) rest = restPattern;
  return { type: 'ArrayPattern', elements, restPattern: rest };
}

// -----------------------------------------------------------------------------
// Expressions
// -----------------------------------------------------------------------------

export interface UnaryExpression extends AstNode { type: 'UnaryExpression'; operator: '-' | '!' | '~'; operand: Expression; }
export interface BinaryExpression extends AstNode { type: 'BinaryExpression'; operator: string; left: Expression; right: Expression; }
export interface FunctionCall extends AstNode { type: 'FunctionCall'; callee: Expression; arguments: Expression[]; typeArguments?: TypeExpression[]; isTrailingLambda: boolean; }
export interface BlockExpression extends AstNode { type: 'BlockExpression'; body: Statement[]; }
export interface AssignmentExpression extends AstNode { type: 'AssignmentExpression'; operator: ':=' | '=' | '+=' | '-=' | '*=' | '/=' | '%=' | '&=' | '|=' | '\\xor=' | '<<=' | '>>='; left: Pattern | MemberAccessExpression | IndexExpression; right: Expression; }
export interface RangeExpression extends AstNode { type: 'RangeExpression'; start: Expression; end: Expression; inclusive: boolean; }
export interface StringInterpolation extends AstNode { type: 'StringInterpolation'; parts: (StringLiteral | Expression)[]; }
export interface MemberAccessExpression extends AstNode {
  type: "MemberAccessExpression";
  object: Expression;
  member: Identifier | IntegerLiteral;
  isSafe?: boolean;
}
export interface IndexExpression extends AstNode { type: 'IndexExpression'; object: Expression; index: Expression; }
export interface LambdaExpression extends AstNode { type: 'LambdaExpression'; genericParams?: GenericParameter[]; params: FunctionParameter[]; returnType?: TypeExpression; body: Expression | BlockExpression; whereClause?: WhereClauseConstraint[]; isVerboseSyntax: boolean; }
export interface ProcExpression extends AstNode { type: 'ProcExpression'; expression: FunctionCall | BlockExpression; }
export interface SpawnExpression extends AstNode { type: 'SpawnExpression'; expression: FunctionCall | BlockExpression; }
export interface AwaitExpression extends AstNode { type: 'AwaitExpression'; expression: Expression; }
export interface PropagationExpression extends AstNode { type: 'PropagationExpression'; expression: Expression; }
export interface OrElseExpression extends AstNode { type: 'OrElseExpression'; expression: Expression; handler: BlockExpression; errorBinding?: Identifier; }
export interface BreakpointExpression extends AstNode { type: 'BreakpointExpression'; label: Identifier; body: BlockExpression; }
export interface IteratorLiteral extends AstNode {
  type: 'IteratorLiteral';
  body: Statement[];
  binding?: Identifier;
  elementType?: TypeExpression;
}
export interface ImplicitMemberExpression extends AstNode { type: 'ImplicitMemberExpression'; member: Identifier; }
export interface PlaceholderExpression extends AstNode { type: 'PlaceholderExpression'; index?: number; }
export interface TopicReferenceExpression extends AstNode { type: 'TopicReferenceExpression'; }

export type Expression =
  | Identifier
  | Literal
  | UnaryExpression
  | BinaryExpression
  | FunctionCall
  | BlockExpression
  | AssignmentExpression
  | RangeExpression
  | StringInterpolation
  | MemberAccessExpression
  | IndexExpression
  | LambdaExpression
  | ProcExpression
  | SpawnExpression
  | AwaitExpression
  | PropagationExpression
  | OrElseExpression
  | BreakpointExpression
  | IteratorLiteral
  | ImplicitMemberExpression
  | PlaceholderExpression
  | TopicReferenceExpression
  | IfExpression
  | MatchExpression
  | StructLiteral
  | MapLiteral
  | RescueExpression
  | EnsureExpression;

export function unaryExpression(operator: UnaryExpression['operator'], operand: Expression): UnaryExpression { return { type: 'UnaryExpression', operator, operand }; }
export function binaryExpression(operator: string, left: Expression, right: Expression): BinaryExpression { return { type: 'BinaryExpression', operator, left, right }; }
export function functionCall(callee: Expression, args: Expression[], typeArgs?: TypeExpression[], isTrailingLambda = false): FunctionCall {
  return { type: 'FunctionCall', callee, arguments: args, typeArguments: typeArgs, isTrailingLambda };
}
export function blockExpression(body: Statement[]): BlockExpression { return { type: 'BlockExpression', body }; }
export function assignmentExpression(operator: AssignmentExpression['operator'], left: Pattern | MemberAccessExpression | IndexExpression, right: Expression): AssignmentExpression {
  return { type: 'AssignmentExpression', operator, left, right };
}
export function rangeExpression(start: Expression, end: Expression, inclusive: boolean): RangeExpression { return { type: 'RangeExpression', start, end, inclusive }; }
export function stringInterpolation(parts: (StringLiteral | Expression)[]): StringInterpolation { return { type: 'StringInterpolation', parts }; }
export function memberAccessExpression(
  object: Expression,
  member: Identifier | IntegerLiteral | string,
  options?: { isSafe?: boolean },
): MemberAccessExpression {
  const memberNode = typeof member === "string" ? identifier(member) : member;
  const expr: MemberAccessExpression = { type: "MemberAccessExpression", object, member: memberNode };
  if (options?.isSafe) expr.isSafe = true;
  return expr;
}
export function indexExpression(object: Expression, index: Expression): IndexExpression { return { type: 'IndexExpression', object, index }; }

export function lambdaExpression(
  params: FunctionParameter[],
  body: Expression | BlockExpression,
  returnType?: TypeExpression,
  genericParams?: GenericParameter[],
  whereClause?: WhereClauseConstraint[],
  isVerboseSyntax = false,
): LambdaExpression {
  return { type: 'LambdaExpression', params, body, returnType, genericParams, whereClause, isVerboseSyntax };
}

export function procExpression(expression: FunctionCall | BlockExpression): ProcExpression {
  return { type: 'ProcExpression', expression };
}

export function spawnExpression(expression: FunctionCall | BlockExpression): SpawnExpression {
  return { type: 'SpawnExpression', expression };
}

export function awaitExpression(expression: Expression): AwaitExpression {
  return { type: 'AwaitExpression', expression };
}

export function propagationExpression(expression: Expression): PropagationExpression {
  return { type: 'PropagationExpression', expression };
}

export function orElseExpression(expression: Expression, handler: BlockExpression, errorBinding?: Identifier | string): OrElseExpression {
  return { type: 'OrElseExpression', expression, handler, errorBinding: typeof errorBinding === 'string' ? identifier(errorBinding) : errorBinding };
}

export function breakpointExpression(label: Identifier | string, body: BlockExpression): BreakpointExpression {
  return { type: 'BreakpointExpression', label: typeof label === 'string' ? identifier(label) : label, body };
}

export function iteratorLiteral(
  body: Statement[],
  binding?: Identifier | string,
  elementType?: TypeExpression,
): IteratorLiteral {
  return {
    type: 'IteratorLiteral',
    body,
    binding: binding ? (typeof binding === 'string' ? identifier(binding) : binding) : undefined,
    elementType,
  };
}

export function implicitMemberExpression(member: Identifier | string): ImplicitMemberExpression {
  return { type: 'ImplicitMemberExpression', member: typeof member === 'string' ? identifier(member) : member };
}

export function placeholderExpression(index?: number): PlaceholderExpression {
  return index === undefined ? { type: 'PlaceholderExpression' } : { type: 'PlaceholderExpression', index };
}

export function topicReferenceExpression(): TopicReferenceExpression {
  return { type: 'TopicReferenceExpression' };
}

// -----------------------------------------------------------------------------
// Control Flow
// -----------------------------------------------------------------------------

export interface OrClause extends AstNode { type: 'OrClause'; condition?: Expression; body: BlockExpression; }
export interface IfExpression extends AstNode { type: 'IfExpression'; ifCondition: Expression; ifBody: BlockExpression; orClauses: OrClause[]; }
export interface MatchClause extends AstNode { type: 'MatchClause'; pattern: Pattern; guard?: Expression; body: Expression; }
export interface MatchExpression extends AstNode { type: 'MatchExpression'; subject: Expression; clauses: MatchClause[]; }
export interface WhileLoop extends AstNode { type: 'WhileLoop'; condition: Expression; body: BlockExpression; }
export interface ForLoop extends AstNode { type: 'ForLoop'; pattern: Pattern; iterable: Expression; body: BlockExpression; }
export interface LoopExpression extends AstNode { type: 'LoopExpression'; body: BlockExpression; }
export interface BreakStatement extends AstNode { type: 'BreakStatement'; label?: Identifier; value?: Expression; }
export interface ContinueStatement extends AstNode { type: 'ContinueStatement'; label?: Identifier; }
export interface YieldStatement extends AstNode { type: 'YieldStatement'; expression?: Expression; }

export function orClause(body: BlockExpression, condition?: Expression): OrClause { return { type: 'OrClause', condition, body }; }
export function ifExpression(ifCondition: Expression, ifBody: BlockExpression, orClauses: OrClause[] = []): IfExpression { return { type: 'IfExpression', ifCondition, ifBody, orClauses }; }
export function matchClause(pattern: Pattern, body: Expression, guard?: Expression): MatchClause { return { type: 'MatchClause', pattern, guard, body }; }
export function matchExpression(subject: Expression, clauses: MatchClause[]): MatchExpression { return { type: 'MatchExpression', subject, clauses }; }
export function whileLoop(condition: Expression, body: BlockExpression): WhileLoop { return { type: 'WhileLoop', condition, body }; }
export function forLoop(pattern: Pattern, iterable: Expression, body: BlockExpression): ForLoop { return { type: 'ForLoop', pattern, iterable, body }; }
export function loopExpression(body: BlockExpression): LoopExpression { return { type: 'LoopExpression', body }; }
export function breakStatement(label?: Identifier | string, value?: Expression): BreakStatement {
  const stmt: BreakStatement = { type: 'BreakStatement' };
  if (label !== undefined) stmt.label = typeof label === 'string' ? identifier(label) : label;
  if (value !== undefined) stmt.value = value;
  return stmt;
}
export function continueStatement(label?: Identifier | string): ContinueStatement {
  return { type: 'ContinueStatement', label: label !== undefined ? (typeof label === 'string' ? identifier(label) : label) : undefined };
}

export function yieldStatement(expression?: Expression): YieldStatement {
  return { type: 'YieldStatement', expression };
}

// -----------------------------------------------------------------------------
// Error Handling
// -----------------------------------------------------------------------------

export interface RaiseStatement extends AstNode { type: 'RaiseStatement'; expression: Expression; }
export interface RescueExpression extends AstNode { type: 'RescueExpression'; monitoredExpression: Expression; clauses: MatchClause[]; }
export interface EnsureExpression extends AstNode { type: 'EnsureExpression'; tryExpression: Expression; ensureBlock: BlockExpression; }
export interface RethrowStatement extends AstNode { type: 'RethrowStatement'; }

export function raiseStatement(expression: Expression): RaiseStatement { return { type: 'RaiseStatement', expression }; }
export function rescueExpression(monitoredExpression: Expression, clauses: MatchClause[]): RescueExpression { return { type: 'RescueExpression', monitoredExpression, clauses }; }
export function ensureExpression(tryExpression: Expression, ensureBlock: BlockExpression): EnsureExpression { return { type: 'EnsureExpression', tryExpression, ensureBlock }; }
export function rethrowStatement(): RethrowStatement { return { type: 'RethrowStatement' }; }

// -----------------------------------------------------------------------------
// Definitions (structs, unions, interfaces, impls, methods, functions)
// -----------------------------------------------------------------------------

export interface StructFieldDefinition extends AstNode { type: 'StructFieldDefinition'; name?: Identifier; fieldType: TypeExpression; }
export function structFieldDefinition(fieldType: TypeExpression, name?: Identifier | string): StructFieldDefinition {
  return { type: 'StructFieldDefinition', name: typeof name === 'string' ? identifier(name) : name, fieldType };
}

export interface StructDefinition extends AstNode {
  type: 'StructDefinition';
  id: Identifier;
  genericParams?: GenericParameter[];
  fields: StructFieldDefinition[];
  whereClause?: WhereClauseConstraint[];
  kind: 'singleton' | 'named' | 'positional';
  isPrivate?: boolean;
}
export function structDefinition(id: Identifier | string, fields: StructFieldDefinition[], kind: StructDefinition['kind'], genericParams?: GenericParameter[], whereClause?: WhereClauseConstraint[], isPrivate?: boolean): StructDefinition {
  return { type: 'StructDefinition', id: typeof id === 'string' ? identifier(id) : id, fields, kind, genericParams, whereClause, isPrivate };
}

export interface StructFieldInitializer extends AstNode { type: 'StructFieldInitializer'; name?: Identifier; value: Expression; isShorthand: boolean; }
export function structFieldInitializer(value: Expression, name?: Identifier | string, isShorthand = false): StructFieldInitializer {
  return { type: 'StructFieldInitializer', name: typeof name === 'string' ? identifier(name) : name, value, isShorthand };
}

export interface StructLiteral extends AstNode {
  type: 'StructLiteral';
  structType?: Identifier;
  fields: StructFieldInitializer[];
  isPositional: boolean;
  functionalUpdateSources?: Expression[];
  typeArguments?: TypeExpression[];
}
export function structLiteral(
  fields: StructFieldInitializer[],
  isPositional: boolean,
  structType?: Identifier | string,
  functionalUpdateSources?: Expression[],
  typeArguments?: TypeExpression[],
): StructLiteral {
  return {
    type: 'StructLiteral',
    structType: typeof structType === 'string' ? identifier(structType) : structType,
    fields,
    isPositional,
    functionalUpdateSources,
    typeArguments,
  };
}

export interface UnionDefinition extends AstNode { type: 'UnionDefinition'; id: Identifier; genericParams?: GenericParameter[]; variants: TypeExpression[]; whereClause?: WhereClauseConstraint[]; isPrivate?: boolean; }
export function unionDefinition(id: Identifier | string, variants: TypeExpression[], genericParams?: GenericParameter[], whereClause?: WhereClauseConstraint[], isPrivate?: boolean): UnionDefinition {
  return { type: 'UnionDefinition', id: typeof id === 'string' ? identifier(id) : id, genericParams, variants, whereClause, isPrivate };
}

export interface TypeAliasDefinition extends AstNode {
  type: 'TypeAliasDefinition';
  id: Identifier;
  genericParams?: GenericParameter[];
  targetType: TypeExpression;
  whereClause?: WhereClauseConstraint[];
  isPrivate?: boolean;
}
export function typeAliasDefinition(
  id: Identifier | string,
  targetType: TypeExpression,
  genericParams?: GenericParameter[],
  whereClause?: WhereClauseConstraint[],
  isPrivate?: boolean,
): TypeAliasDefinition {
  return {
    type: 'TypeAliasDefinition',
    id: typeof id === 'string' ? identifier(id) : id,
    targetType,
    genericParams,
    whereClause,
    isPrivate,
  };
}

export interface FunctionParameter extends AstNode { type: 'FunctionParameter'; name: Pattern; paramType?: TypeExpression; }
export function functionParameter(name: Pattern | Identifier | string, paramType?: TypeExpression): FunctionParameter {
  let pattern: Pattern;
  if (typeof name === 'string') pattern = identifier(name);
  else if ((name as Identifier).type === 'Identifier') pattern = name as Identifier;
  else pattern = name as Pattern;
  return { type: 'FunctionParameter', name: pattern, paramType };
}

export interface FunctionDefinition extends AstNode {
  type: 'FunctionDefinition';
  id: Identifier;
  genericParams?: GenericParameter[];
  inferredGenericParams?: GenericParameter[];
  params: FunctionParameter[];
  returnType?: TypeExpression;
  body: BlockExpression;
  whereClause?: WhereClauseConstraint[];
  isMethodShorthand: boolean;
  isPrivate: boolean;
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
  options?: { inferredGenericParams?: GenericParameter[] },
): FunctionDefinition {
  return {
    type: 'FunctionDefinition',
    id: typeof id === 'string' ? identifier(id) : id,
    params,
    body,
    returnType,
    genericParams,
    inferredGenericParams: options?.inferredGenericParams,
    whereClause,
    isMethodShorthand,
    isPrivate,
  };
}

export interface FunctionSignature extends AstNode {
  type: 'FunctionSignature';
  name: Identifier;
  genericParams?: GenericParameter[];
  inferredGenericParams?: GenericParameter[];
  params: FunctionParameter[];
  returnType?: TypeExpression;
  whereClause?: WhereClauseConstraint[];
  defaultImpl?: BlockExpression;
}
export function functionSignature(
  name: Identifier | string,
  params: FunctionParameter[],
  returnType?: TypeExpression,
  genericParams?: GenericParameter[],
  whereClause?: WhereClauseConstraint[],
  defaultImpl?: BlockExpression,
  options?: { inferredGenericParams?: GenericParameter[] },
): FunctionSignature {
  return {
    type: 'FunctionSignature',
    name: typeof name === 'string' ? identifier(name) : name,
    params,
    returnType,
    genericParams,
    inferredGenericParams: options?.inferredGenericParams,
    whereClause,
    defaultImpl,
  };
}

export interface InterfaceDefinition extends AstNode {
  type: 'InterfaceDefinition';
  id: Identifier;
  genericParams?: GenericParameter[];
  selfTypePattern?: TypeExpression; // for T / Array T / M _
  signatures: FunctionSignature[];
  whereClause?: WhereClauseConstraint[];
  baseInterfaces?: TypeExpression[];
  isPrivate: boolean;
}
export function interfaceDefinition(id: Identifier | string, signatures: FunctionSignature[], genericParams?: GenericParameter[], selfTypePattern?: TypeExpression, whereClause?: WhereClauseConstraint[], baseInterfaces?: TypeExpression[], isPrivate = false): InterfaceDefinition {
  return { type: 'InterfaceDefinition', id: typeof id === 'string' ? identifier(id) : id, genericParams, selfTypePattern, signatures, whereClause, baseInterfaces, isPrivate };
}

export interface ImplementationDefinition extends AstNode {
  type: 'ImplementationDefinition';
  implName?: Identifier; // named impl
  genericParams?: GenericParameter[];
  interfaceName: Identifier;
  interfaceArgs?: TypeExpression[]; // supports HKT/higher-kinded args as needed
  targetType: TypeExpression;
  definitions: FunctionDefinition[];
  whereClause?: WhereClauseConstraint[];
  isPrivate?: boolean;
}
export function implementationDefinition(interfaceName: Identifier | string, targetType: TypeExpression, definitions: FunctionDefinition[], implName?: Identifier | string, genericParams?: GenericParameter[], interfaceArgs?: TypeExpression[], whereClause?: WhereClauseConstraint[]): ImplementationDefinition {
  return { type: 'ImplementationDefinition', implName: typeof implName === 'string' ? identifier(implName) : implName, genericParams, interfaceName: typeof interfaceName === 'string' ? identifier(interfaceName) : interfaceName, interfaceArgs, targetType, definitions, whereClause };
}

export interface MethodsDefinition extends AstNode { type: 'MethodsDefinition'; targetType: TypeExpression; genericParams?: GenericParameter[]; definitions: FunctionDefinition[]; whereClause?: WhereClauseConstraint[]; }
export function methodsDefinition(targetType: TypeExpression, definitions: FunctionDefinition[], genericParams?: GenericParameter[], whereClause?: WhereClauseConstraint[]): MethodsDefinition {
  return { type: 'MethodsDefinition', targetType, genericParams, definitions, whereClause };
}

// -----------------------------------------------------------------------------
// Packages & Imports
// -----------------------------------------------------------------------------

export interface PackageStatement extends AstNode { type: 'PackageStatement'; namePath: Identifier[]; isPrivate?: boolean; }
export function packageStatement(namePath: (Identifier | string)[], isPrivate?: boolean): PackageStatement {
  return { type: 'PackageStatement', namePath: namePath.map(p => typeof p === 'string' ? identifier(p) : p), isPrivate };
}

export interface ImportSelector extends AstNode { type: 'ImportSelector'; name: Identifier; alias?: Identifier; }
export function importSelector(name: Identifier | string, alias?: Identifier | string): ImportSelector {
  return { type: 'ImportSelector', name: typeof name === 'string' ? identifier(name) : name, alias: typeof alias === 'string' ? identifier(alias) : alias };
}

export interface ImportStatement extends AstNode { type: 'ImportStatement'; packagePath: Identifier[]; isWildcard: boolean; selectors?: ImportSelector[]; alias?: Identifier; }
export function importStatement(packagePath: (Identifier | string)[], isWildcard = false, selectors?: ImportSelector[], alias?: Identifier | string): ImportStatement {
  return { type: 'ImportStatement', packagePath: packagePath.map(p => typeof p === 'string' ? identifier(p) : p), isWildcard, selectors, alias: typeof alias === 'string' ? identifier(alias) : alias };
}

// -----------------------------------------------------------------------------
// Module root
// -----------------------------------------------------------------------------

export interface Module extends AstNode { type: 'Module'; package?: PackageStatement; imports: ImportStatement[]; body: Statement[]; }
export function module(body: Statement[], imports: ImportStatement[] = [], pkg?: PackageStatement): Module {
  return { type: 'Module', package: pkg, imports, body };
}

// -----------------------------------------------------------------------------
// Statements
// -----------------------------------------------------------------------------

export interface ReturnStatement extends AstNode { type: 'ReturnStatement'; argument?: Expression; }
export function returnStatement(argument?: Expression): ReturnStatement { return { type: 'ReturnStatement', argument }; }

export type Statement =
  | Expression
  | FunctionDefinition
  | StructDefinition
  | UnionDefinition
  | TypeAliasDefinition
  | InterfaceDefinition
  | ImplementationDefinition
  | MethodsDefinition
  | ImportStatement
  | PackageStatement
  | ReturnStatement
  | RaiseStatement
  | RethrowStatement
  | BreakStatement
  | ContinueStatement
  | WhileLoop
  | ForLoop
  | LoopExpression
  | YieldStatement
  | PreludeStatement
  | ExternFunctionBody
  | DynImportStatement;

// -----------------------------------------------------------------------------
// Dynamic imports (runtime-bound names)
// -----------------------------------------------------------------------------

export interface DynImportStatement extends AstNode {
  type: 'DynImportStatement';
  packagePath: Identifier[];
  isWildcard: boolean;
  selectors?: ImportSelector[];
  alias?: Identifier;
}
export function dynImportStatement(packagePath: (Identifier | string)[], isWildcard = false, selectors?: ImportSelector[], alias?: Identifier | string): DynImportStatement {
  return { type: 'DynImportStatement', packagePath: packagePath.map(p => typeof p === 'string' ? identifier(p) : p), isWildcard, selectors, alias: typeof alias === 'string' ? identifier(alias) : alias };
}

// -----------------------------------------------------------------------------
// Host interop: preludes and extern function bodies
// -----------------------------------------------------------------------------

export type HostTarget = 'go' | 'crystal' | 'typescript' | 'python' | 'ruby';

export interface PreludeStatement extends AstNode {
  type: 'PreludeStatement';
  target: HostTarget;
  code: string; // raw host code
}
export function preludeStatement(target: HostTarget, code: string): PreludeStatement {
  return { type: 'PreludeStatement', target, code };
}

export interface ExternFunctionBody extends AstNode {
  type: 'ExternFunctionBody';
  target: HostTarget;
  signature: FunctionDefinition; // function header (id/generics/params/returnType)
  body: string; // raw host code body
}
export function externFunctionBody(target: HostTarget, signature: FunctionDefinition, body: string): ExternFunctionBody {
  return { type: 'ExternFunctionBody', target, signature, body };
}


// -----------------------------------------------------------------------------
// DSL helpers (aliases and convenience builders)
// -----------------------------------------------------------------------------

// Literals & identifiers
export const id = identifier;
export const str = stringLiteral;
export const int = integerLiteral;
export const flt = floatLiteral;
export const bool = booleanLiteral;
export const nil = nilLiteral;
export const chr = charLiteral;
export function arr(...elements: Expression[]): ArrayLiteral { return arrayLiteral(elements); }
export const mapEntry = mapLiteralEntry;
export const mapSpread = mapLiteralSpread;
export function mapLit(entries: (MapLiteralEntry | MapLiteralSpread)[]): MapLiteral { return mapLiteral(entries); }

// Types
export const ty = simpleTypeExpression;
export const gen = genericTypeExpression;
export const fnType = functionTypeExpression;
export const nullable = nullableTypeExpression;
export const result = resultTypeExpression;
export const unionT = unionTypeExpression;
export const wildT = wildcardTypeExpression;

// Patterns
export const wc = wildcardPattern;
export const litP = literalPattern;
export const typedP = typedPattern;
export function fieldP(
  pattern: Pattern,
  fieldName?: Identifier | string,
  binding?: Identifier | string,
  typeAnnotation?: TypeExpression,
): StructPatternField {
  return structPatternField(pattern, fieldName, binding, typeAnnotation);
}
export function structP(fields: StructPatternField[], isPositional: boolean, structType?: Identifier | string): StructPattern {
  return structPattern(fields, isPositional, structType);
}
export function arrP(elements: Pattern[], restPattern?: Identifier | WildcardPattern | string): ArrayPattern {
  return arrayPattern(elements, restPattern);
}

// Expressions
export const un = unaryExpression;
export const bin = binaryExpression;
export function call(callee: string | Expression, ...args: Expression[]): FunctionCall {
  const calleeExpr = typeof callee === 'string' ? identifier(callee) : callee;
  return functionCall(calleeExpr, args);
}
export function callT(callee: string | Expression, typeArgs: TypeExpression[], ...args: Expression[]): FunctionCall {
  const calleeExpr = typeof callee === 'string' ? identifier(callee) : callee;
  return functionCall(calleeExpr, args, typeArgs);
}
export function block(...statements: Statement[]): BlockExpression { return blockExpression(statements); }
export function assign(left: string | Pattern | MemberAccessExpression | IndexExpression, value: Expression, op: AssignmentExpression['operator'] = ':='): AssignmentExpression {
  const lhs = typeof left === 'string' ? identifier(left) : left;
  return assignmentExpression(op, lhs as Pattern | MemberAccessExpression | IndexExpression, value);
}
export function assignMember(object: string | Expression, member: string | number, value: Expression): AssignmentExpression {
  const objExpr = typeof object === 'string' ? identifier(object) : object;
  const memberNode = typeof member === 'string' ? identifier(member) : integerLiteral(member as number);
  return assignmentExpression('=', memberAccessExpression(objExpr, memberNode), value);
}
export function assignIndex(object: string | Expression, indexExpr: Expression, value: Expression): AssignmentExpression {
  const objExpr = typeof object === 'string' ? identifier(object) : object;
  return assignmentExpression('=', indexExpression(objExpr, indexExpr), value);
}
export const range = rangeExpression;
export const interp = stringInterpolation;
export const member = memberAccessExpression;
export const index = indexExpression;
export const lam = lambdaExpression;
export const proc = procExpression;
export const spawn = spawnExpression;
export const prop = propagationExpression;
export function orelse(expression: Expression, handlerStmts: Statement[], errorBinding?: Identifier | string): OrElseExpression {
  return orElseExpression(expression, blockExpression(handlerStmts), errorBinding);
}
export function bp(label: Identifier | string, ...stmts: Statement[]): BreakpointExpression {
  return breakpointExpression(label, blockExpression(stmts));
}
export const iter = iteratorLiteral;

// Control flow
export function iff(condition: Expression, ...stmts: Statement[]): IfExpression {
  return ifExpression(condition, blockExpression(stmts));
}
export const orC = orClause;
export const wloop = whileLoop;
export const loopExpr = loopExpression;
export function forIn(pattern: Pattern | string, iterable: Expression, ...stmts: Statement[]): ForLoop {
  const p = typeof pattern === 'string' ? identifier(pattern) : pattern;
  return forLoop(p, iterable, blockExpression(stmts));
}
export const brk = breakStatement;
export const cont = continueStatement;

// Error handling
export const raise = raiseStatement;
export const rescue = rescueExpression;
export function ensure(expr: Expression, ...stmts: Statement[]): EnsureExpression {
  return ensureExpression(expr, blockExpression(stmts));
}
export const rethrow = rethrowStatement;

// Match
export const mc = matchClause;
export const match = matchExpression;

// Definitions
export const fieldDef = structFieldDefinition;
export const structDef = structDefinition;
export function fieldInit(value: Expression, name?: Identifier | string): StructFieldInitializer { return structFieldInitializer(value, name); }
export function shorthandField(name: Identifier | string): StructFieldInitializer {
  const idNode = typeof name === 'string' ? identifier(name) : name;
  return structFieldInitializer(idNode, idNode, true);
}
export const unionDef = unionDefinition;
export const param = functionParameter;
export function fn(name: Identifier | string, params: FunctionParameter[], bodyStmts: Statement[], returnType?: TypeExpression, genericParams?: GenericParameter[], whereClause?: WhereClauseConstraint[], isMethodShorthand = false, isPrivate = false): FunctionDefinition {
  return functionDefinition(name, params, blockExpression(bodyStmts), returnType, genericParams, whereClause, isMethodShorthand, isPrivate);
}
export const fnSig = functionSignature;
export const iface = interfaceDefinition;
export const impl = implementationDefinition;
export const methods = methodsDefinition;

// Packages & imports
export const pkg = packageStatement;
export const impSel = importSelector;
export const imp = importStatement;
export const dynImp = dynImportStatement;

// Module
export const mod = module;

// Statements
export const ret = returnStatement;
export const yld = yieldStatement;

// Host interop
export const prelude = preludeStatement;
export const extern = externFunctionBody;
