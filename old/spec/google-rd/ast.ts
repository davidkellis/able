// --- ast.ts ---

import { Token } from './lexer';

export interface Visitor<R> {
  visitAssignmentExpr(expr: Assignment): R;
  visitBinaryExpr(expr: Binary): R;
  visitCallExpr(expr: Call): R;
  visitGroupingExpr(expr: Grouping): R;
  visitLiteralExpr(expr: Literal): R;
  visitVariableExpr(expr: Variable): R;
  visitBlockStmt(stmt: Block): R;
  visitExpressionStmt(stmt: ExpressionStatement): R;
  visitFunctionStmt(stmt: FunctionDeclaration): R;
  visitIfStmt(stmt: IfStatement): R; // Treat if as statement for simplicity here
  visitReturnStmt(stmt: ReturnStatement): R;
  visitWhileStmt(stmt: WhileStatement): R;
}

export interface Expr {
  accept<R>(visitor: Visitor<R>): R;
}

export interface Stmt {
 accept<R>(visitor: Visitor<R>): R;
}

// --- Expressions ---
export class Assignment implements Expr {
  constructor(public name: Token, public value: Expr) {}
  accept<R>(visitor: Visitor<R>): R { return visitor.visitAssignmentExpr(this); }
}

export class Binary implements Expr {
  constructor(public left: Expr, public operator: Token, public right: Expr) {}
  accept<R>(visitor: Visitor<R>): R { return visitor.visitBinaryExpr(this); }
}

export class Call implements Expr {
 constructor(public callee: Expr, public paren: Token, public args: Expr[]) {}
 accept<R>(visitor: Visitor<R>): R { return visitor.visitCallExpr(this); }
}

export class Grouping implements Expr {
  constructor(public expression: Expr) {}
  accept<R>(visitor: Visitor<R>): R { return visitor.visitGroupingExpr(this); }
}

export class Literal implements Expr {
  constructor(public value: any) {}
  accept<R>(visitor: Visitor<R>): R { return visitor.visitLiteralExpr(this); }
}

export class Variable implements Expr {
  constructor(public name: Token) {}
  accept<R>(visitor: Visitor<R>): R { return visitor.visitVariableExpr(this); }
}

// --- Statements ---
export class Block implements Stmt {
 constructor(public statements: Stmt[]) {}
 accept<R>(visitor: Visitor<R>): R { return visitor.visitBlockStmt(this); }
}

export class ExpressionStatement implements Stmt {
 constructor(public expression: Expr) {}
 accept<R>(visitor: Visitor<R>): R { return visitor.visitExpressionStmt(this); }
}

export class FunctionParameter {
    constructor(public name: Token, public typeAnnotation?: Token /* Simple Type Name */) {}
}

export class FunctionDeclaration implements Stmt {
 constructor(
     public name: Token,
     public params: FunctionParameter[],
     public returnType?: Token, /* Simple Type Name */
     public body?: Block // Body is optional only if it's an interface decl
 ) {}
 accept<R>(visitor: Visitor<R>): R { return visitor.visitFunctionStmt(this); }
}

export class IfStatement implements Stmt {
 constructor(public condition: Expr, public thenBranch: Block, public elseBranch?: Block) {}
 accept<R>(visitor: Visitor<R>): R { return visitor.visitIfStmt(this); }
}

export class ReturnStatement implements Stmt {
 constructor(public keyword: Token, public value?: Expr) {}
 accept<R>(visitor: Visitor<R>): R { return visitor.visitReturnStmt(this); }
}

export class WhileStatement implements Stmt {
 constructor(public condition: Expr, public body: Block) {}
 accept<R>(visitor: Visitor<R>): R { return visitor.visitWhileStmt(this); }
}

// Add other node types as needed (StructDef, UnionDef, MatchExpr, etc.)
