// --- parser.ts ---

import { Token, TokenType } from './lexer';
import * as Ast from './ast';

class ParseError extends Error {}

export class Parser {
  private tokens: Token[];
  private current: number = 0;
  private hadError: boolean = false;

  constructor(tokens: Token[]) {
    this.tokens = tokens;
  }

  parse(): Ast.Stmt[] {
    const statements: Ast.Stmt[] = [];
    while (!this.isAtEnd()) {
        try {
            // Allow null if synchronize skips tokens to a point where declaration returns null
            const decl = this.declaration();
            if (decl) {
                statements.push(decl);
            }
        } catch (error) {
            if (error instanceof ParseError) {
                 this.synchronize();
            } else {
                // Rethrow unexpected errors
                throw error;
            }
        }
    }
    return statements;
  }

  // --- Grammar Rules ---

  // declaration -> funcDecl | varDecl | statement ;
  private declaration(): Ast.Stmt | null {
      if (this.match(TokenType.FN)) return this.functionDeclaration("function");
      // Add structDecl, unionDecl, interfaceDecl, implDecl etc. here
      // Simplified: Treat assignments like variable declarations for now
      if (this.match(TokenType.LET, TokenType.VAR) || this.peek().type === TokenType.IDENTIFIER && this.peekNext().type === TokenType.EQUAL) {
         return this.assignmentStatement(); // Simplified
      }

      return this.statement();
  }

  // funcDecl -> "fn" IDENTIFIER "(" parameters? ")" ("->" type)? block ;
  private functionDeclaration(kind: string): Ast.FunctionDeclaration {
      const name = this.consume(TokenType.IDENTIFIER, `Expect ${kind} name.`);
      this.consume(TokenType.LEFT_PAREN, `Expect '(' after ${kind} name.`);
      const parameters: Ast.FunctionParameter[] = [];
      if (!this.check(TokenType.RIGHT_PAREN)) {
          do {
              if (parameters.length >= 255) { // Arity limit example
                  this.error(this.peek(), "Can't have more than 255 parameters.");
              }
              const paramName = this.consume(TokenType.IDENTIFIER, "Expect parameter name.");
              let paramType;
              if (this.match(TokenType.COLON)) {
                  // Super simplified type parsing - only allows single identifier
                  paramType = this.consume(TokenType.IDENTIFIER, "Expect type name after ':'.");
              }
              parameters.push(new Ast.FunctionParameter(paramName, paramType));
          } while (this.match(TokenType.COMMA));
      }
      this.consume(TokenType.RIGHT_PAREN, "Expect ')' after parameters.");

      let returnType;
      if (this.match(TokenType.ARROW)) {
          // Super simplified type parsing
          returnType = this.consume(TokenType.IDENTIFIER, "Expect return type name after '->'.")
      }

      this.consume(TokenType.LEFT_BRACE, `Expect '{' before ${kind} body.`);
      const body = this.block(); // Block expects Stmt[] but returns Block node
      return new Ast.FunctionDeclaration(name, parameters, returnType, body);
  }


  // Simplified assignment statement (instead of full varDecl)
  // assignmentStmt -> IDENTIFIER "=" expression (";" | NEWLINE);
   private assignmentStatement(): Ast.Stmt {
        // This is highly simplified, doesn't handle full patterns 'Pattern = Expr'
        // Assumes 'let'/'var' was matched or lookahead saw 'ident ='.
        // If we used let/var, consume IDENTIFIER here. If not, parse identifier first.

        // Let's assume simple `identifier = expression`
        const name = this.previous().type === TokenType.IDENTIFIER ? this.previous() : this.consume(TokenType.IDENTIFIER, "Expect variable name.");

        if (this.match(TokenType.EQUAL)) {
             const value = this.expression();
             // We need expression statements for the AST
             const assignExpr = new Ast.Assignment(name, value); // Assignment is an expression
             // Consume optional ; or handle newline implicitly
             this.consumeSemicolonOrNewline();
             return new Ast.ExpressionStatement(assignExpr); // Wrap in statement
        } else {
            // If no '=', maybe it's just an expression statement starting with identifier?
            this.current--; // Backtrack if we consumed IDENTIFIER but didn't find '='
            return this.statement();
        }

    }

  // statement -> exprStmt | ifStmt | whileStmt | returnStmt | block ;
  private statement(): Ast.Stmt {
    if (this.match(TokenType.IF)) return this.ifStatement();
    if (this.match(TokenType.WHILE)) return this.whileStatement();
    if (this.match(TokenType.RETURN)) return this.returnStatement(); // Example addition
    if (this.match(TokenType.LEFT_BRACE)) return this.block();
    // Add matchStmt, breakpointStmt etc.

    return this.expressionStatement();
  }

  // ifStmt -> "if" expression block ( "else" block )? ;
  private ifStatement(): Ast.Stmt {
      // Assuming condition isn't parenthesized based on spec examples
      const condition = this.expression();
      this.consume(TokenType.LEFT_BRACE, "Expect '{' after if condition.");
      const thenBranch = this.block();
      let elseBranch: Ast.Block | undefined = undefined;
      if (this.match(TokenType.ELSE)) {
          // Handle 'else if' potentially here, or just simple 'else'
          if(this.match(TokenType.IF)) {
              // Simplification: treat 'else if' as nested if within else block
               const elseIfStmt = this.ifStatement();
               elseBranch = new Ast.Block([elseIfStmt]);
          } else {
               this.consume(TokenType.LEFT_BRACE, "Expect '{' after else.");
               elseBranch = this.block();
          }
      }
       // Note: Spec requires mandatory else for 'if' *expressions*.
       // If 'if' can be a statement without else, this is okay.
       // If 'if' must be expression, require else and type compatibility check later.
      return new Ast.IfStatement(condition, thenBranch, elseBranch);
  }

  // whileStmt -> "while" expression block ;
  private whileStatement(): Ast.Stmt {
       const condition = this.expression();
       this.consume(TokenType.LEFT_BRACE, "Expect '{' after while condition.");
       const body = this.block();
       return new Ast.WhileStatement(condition, body);
  }

  // returnStmt -> "return" expression? (";" | NEWLINE) ;
  private returnStatement(): Ast.Stmt {
       const keyword = this.previous();
       let value: Ast.Expr | undefined = undefined;
       // Check if return is followed by something other than ; or newline implicitly
       // This check is basic, doesn't handle all block endings correctly
       if (!this.check(TokenType.SEMICOLON) && !this.check(TokenType.RIGHT_BRACE) && !this.check(TokenType.EOF) ) {
            // Approximation: If next isn't ; or }, assume there's a value
           value = this.expression();
       }
       this.consumeSemicolonOrNewline();
       return new Ast.ReturnStatement(keyword, value);
   }


  // block -> "{" declaration* "}" ;
  private block(): Ast.Block {
    const statements: Ast.Stmt[] = [];
    while (!this.check(TokenType.RIGHT_BRACE) && !this.isAtEnd()) {
      const decl = this.declaration();
      if (decl) {
         statements.push(decl);
      }
    }
    this.consume(TokenType.RIGHT_BRACE, "Expect '}' after block.");
    return new Ast.Block(statements);
  }

  // exprStmt -> expression (";" | NEWLINE) ;
  private expressionStatement(): Ast.Stmt {
    const expr = this.expression();
    this.consumeSemicolonOrNewline();
    return new Ast.ExpressionStatement(expr);
  }

  // --- Expression Parsing (Highly Simplified - No Pratt Parsing) ---

  // expression -> assignment | addition ; // Lowest precedence
  private expression(): Ast.Expr {
      // Simple version: just parse binary ops like '+'
      return this.addition();
  }

   // assignment -> IDENTIFIER "=" assignment | addition ; // Right-associative
   // This is complex to fit into simple precedence, handle assignment separately?
   // Let's skip assignment *expressions* for now and handle it only as a statement.

   // addition -> term ( "+" term )* ;
   private addition(): Ast.Expr {
       let expr = this.term(); // Parse higher precedence first

       while (this.match(TokenType.PLUS)) { // Only handling '+' here
           const operator = this.previous();
           const right = this.term();
           expr = new Ast.Binary(expr, operator, right);
       }
       return expr;
   }

   // term -> factor ( "*" factor )* ; // Example higher precedence
   private term(): Ast.Expr {
       let expr = this.primary(); // Parse highest precedence

       while (this.match(TokenType.STAR)) { // Only handling '*' here
           const operator = this.previous();
           const right = this.primary();
           expr = new Ast.Binary(expr, operator, right);
       }
       return expr;
   }


   // primary -> NUMBER | STRING | TRUE | FALSE | NIL | IDENTIFIER
   //          | "(" expression ")" | callExpr ;
   private primary(): Ast.Expr {
       if (this.match(TokenType.FALSE)) return new Ast.Literal(false);
       if (this.match(TokenType.TRUE)) return new Ast.Literal(true);
       if (this.match(TokenType.NIL)) return new Ast.Literal(null);

       if (this.match(TokenType.NUMBER, TokenType.STRING)) {
           return new Ast.Literal(this.previous().value);
       }

       if (this.match(TokenType.IDENTIFIER)) {
           const identifierExpr = new Ast.Variable(this.previous());
            // Check for function call right after identifier
           if (this.match(TokenType.LEFT_PAREN)) {
               return this.finishCall(identifierExpr);
           }
           return identifierExpr;
       }

       if (this.match(TokenType.LEFT_PAREN)) {
           const expr = this.expression();
           this.consume(TokenType.RIGHT_PAREN, "Expect ')' after expression.");
            // Check for function call after grouping
           if (this.match(TokenType.LEFT_PAREN)) {
                return this.finishCall(new Ast.Grouping(expr));
           }
           return new Ast.Grouping(expr);
       }

       // Add lambda parsing, struct literal parsing, array literal etc here

       throw this.error(this.peek(), "Expect expression.");
   }

   // Helper to finish parsing a call expression once the callee and '(' are found
   private finishCall(callee: Ast.Expr): Ast.Expr {
       const args: Ast.Expr[] = [];
       if (!this.check(TokenType.RIGHT_PAREN)) {
           do {
                if (args.length >= 255) {
                    this.error(this.peek(), "Can't have more than 255 arguments.");
                }
               args.push(this.expression());
           } while (this.match(TokenType.COMMA));
       }
       const paren = this.consume(TokenType.RIGHT_PAREN, "Expect ')' after arguments.");

       // Handle trailing lambda here
       if (this.check(TokenType.LEFT_BRACE) && this.peekNext().type !== TokenType.RIGHT_BRACE) { // Basic check for lambda
            // TODO: Parse lambda expression here and add to args
            // this.error(this.peek(), "Trailing lambda parsing not implemented yet.");
       }


       return new Ast.Call(callee, paren, args);
   }


  // --- Utility Methods ---

  private match(...types: TokenType[]): boolean {
    for (const type of types) {
      if (this.check(type)) {
        this.advance();
        return true;
      }
    }
    return false;
  }

  private consume(type: TokenType, message: string): Token {
    if (this.check(type)) return this.advance();
    throw this.error(this.peek(), message);
  }

   private consumeSemicolonOrNewline(): void {
        // Able allows newline or semicolon as statement terminator in blocks
        if (this.match(TokenType.SEMICOLON)) return;
        // Check if the previous token was followed by a newline implicitly
        // This requires lexer to track newline significance or parser to infer
        // Simple version: Assume newline handling is implicit if no semicolon
        // Or require semicolon for now? Let's assume optional semicolon
        if (this.check(TokenType.RIGHT_BRACE) || this.isAtEnd()) {
            // Likely end of block or file, no terminator needed
            return;
        }
        // Maybe check previous token's line vs current token's line? Complex.
        // Simplification: Just allow omitting semicolon.
   }

  private check(type: TokenType): boolean {
    if (this.isAtEnd()) return false;
    return this.peek().type === type;
  }

  private advance(): Token {
    if (!this.isAtEnd()) this.current++;
    return this.previous();
  }

  private isAtEnd(): boolean {
    return this.peek().type === TokenType.EOF;
  }

  private peek(): Token {
    return this.tokens[this.current];
  }

   private peekNext(): Token {
    if (this.isAtEnd()) return this.peek(); // EOF
    return this.tokens[this.current + 1];
   }


  private previous(): Token {
    return this.tokens[this.current - 1];
  }

  private error(token: Token, message: string): ParseError {
      this.hadError = true; // Mark that an error occurred
      if (token.type === TokenType.EOF) {
          console.error(`[Line ${token.line}] Parser Error at end: ${message}`);
      } else {
           console.error(`[Line ${token.line}] Parser Error at '${token.lexeme}': ${message}`);
      }
       // In a real parser, collect errors instead of logging
      return new ParseError(); // Throw custom error to enable catching
  }

  // Basic error recovery: discard tokens until a likely statement boundary
  private synchronize(): void {
      this.advance(); // Consume the token that caused the error

      while (!this.isAtEnd()) {
           // If previous was semicolon, likely end of statement
           if (this.previous().type === TokenType.SEMICOLON) return;
            // Or if next token is a keyword that starts a declaration/statement
           switch (this.peek().type) {
               case TokenType.FN:
               case TokenType.STRUCT:
               case TokenType.UNION:
               case TokenType.INTERFACE:
               case TokenType.IMPL:
               case TokenType.LET: // Example
               case TokenType.VAR: // Example
               case TokenType.IF:
               case TokenType.WHILE:
               case TokenType.FOR:
               case TokenType.RETURN:
               case TokenType.MATCH:
               case TokenType.BREAKPOINT:
                   return; // Start parsing from this keyword
           }
           this.advance(); // Discard token and continue synchronizing
      }
  }
}
