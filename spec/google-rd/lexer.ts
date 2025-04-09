// --- lexer.ts ---

export enum TokenType {
  // Single-character tokens
  LEFT_PAREN, RIGHT_PAREN, LEFT_BRACE, RIGHT_BRACE, LEFT_BRACKET, RIGHT_BRACKET,
  COMMA, DOT, MINUS, PLUS, SLASH, STAR, SEMICOLON, COLON,
  BANG, EQUAL, LESS, GREATER, PIPE, UNDERSCORE, BACKTICK, DOLLAR, AT,

  // One or two character tokens
  BANG_EQUAL, EQUAL_EQUAL, LESS_EQUAL, GREATER_EQUAL,
  ARROW, FAT_ARROW, // -> =>
  DOT_DOT, DOT_DOT_DOT, // .. ...
  PIPE_GREATER, // |>
  QUESTION_MARK, // ?

  // Literals
  IDENTIFIER, STRING, NUMBER,

  // Keywords
  FN, STRUCT, UNION, INTERFACE, IMPL, FOR,
  IF, ELSE, WHILE, MATCH, BREAKPOINT, BREAK, RETURN, // Note: return not in spec yet, added for demo
  LET, VAR, // Assuming we might add these
  TRUE, FALSE, NIL,
  PACKAGE, IMPORT, PRIVATE, AS,
  TYPE, CONST, SELF,

  // Special
  COMMENT, // ## ...
  EOF
}

export interface Token {
  type: TokenType;
  lexeme: string; // The actual text
  value: any;    // Literal value (e.g., number, string content)
  line: number;
}

export class Lexer {
  private source: string;
  private tokens: Token[] = [];
  private start: number = 0;
  private current: number = 0;
  private line: number = 1;

  private static keywords: Map<string, TokenType> = new Map([
    ['fn', TokenType.FN],
    ['struct', TokenType.STRUCT],
    ['union', TokenType.UNION],
    ['interface', TokenType.INTERFACE],
    ['impl', TokenType.IMPL],
    ['for', TokenType.FOR],
    ['if', TokenType.IF],
    ['else', TokenType.ELSE],
    ['while', TokenType.WHILE],
    ['match', TokenType.MATCH],
    ['breakpoint', TokenType.BREAKPOINT],
    ['break', TokenType.BREAK],
    ['return', TokenType.RETURN], // Example addition
    ['let', TokenType.LET],       // Example addition
    ['var', TokenType.VAR],       // Example addition
    ['true', TokenType.TRUE],
    ['false', TokenType.FALSE],
    ['nil', TokenType.NIL],
    ['package', TokenType.PACKAGE],
    ['import', TokenType.IMPORT],
    ['private', TokenType.PRIVATE],
    ['as', TokenType.AS],
    ['type', TokenType.TYPE],
    ['const', TokenType.CONST],
    ['Self', TokenType.SELF], // Capital S
  ]);


  constructor(source: string) {
    this.source = source;
  }

  scanTokens(): Token[] {
    while (!this.isAtEnd()) {
      this.start = this.current;
      this.scanToken();
    }
    this.tokens.push({ type: TokenType.EOF, lexeme: '', value: null, line: this.line });
    return this.tokens;
  }

  private isAtEnd(): boolean {
    return this.current >= this.source.length;
  }

  private scanToken(): void {
    const c = this.advance();
    switch (c) {
      // Single char
      case '(': this.addToken(TokenType.LEFT_PAREN); break;
      case ')': this.addToken(TokenType.RIGHT_PAREN); break;
      case '{': this.addToken(TokenType.LEFT_BRACE); break;
      case '}': this.addToken(TokenType.RIGHT_BRACE); break;
      case '[': this.addToken(TokenType.LEFT_BRACKET); break;
      case ']': this.addToken(TokenType.RIGHT_BRACKET); break;
      case ',': this.addToken(TokenType.COMMA); break;
      case '-': this.addToken(this.match('>') ? TokenType.ARROW : TokenType.MINUS); break;
      case '+': this.addToken(TokenType.PLUS); break;
      case '*': this.addToken(TokenType.STAR); break;
      case '/': this.addToken(TokenType.SLASH); break;
      case ';': this.addToken(TokenType.SEMICOLON); break;
      case ':': this.addToken(TokenType.COLON); break;
      case '!': this.addToken(this.match('=') ? TokenType.BANG_EQUAL : TokenType.BANG); break;
      case '=': this.addToken(this.match('>') ? TokenType.FAT_ARROW : this.match('=') ? TokenType.EQUAL_EQUAL : TokenType.EQUAL); break;
      case '<': this.addToken(this.match('=') ? TokenType.LESS_EQUAL : TokenType.LESS); break;
      case '>': this.addToken(this.match('=') ? TokenType.GREATER_EQUAL : TokenType.GREATER); break;
      case '|': this.addToken(this.match('>') ? TokenType.PIPE_GREATER : TokenType.PIPE); break;
      case '_': this.addToken(TokenType.UNDERSCORE); break;
      case '.': this.addToken(this.match('.') ? (this.match('.') ? TokenType.DOT_DOT_DOT : TokenType.DOT_DOT) : TokenType.DOT); break;
      case '?': this.addToken(TokenType.QUESTION_MARK); break;
      case '@': this.addToken(TokenType.AT); break;
      case '$': this.addToken(TokenType.DOLLAR); break; // For interpolation start
      case '`': this.string('`'); break; // Interpolated string


      // Comments (##)
      case '#':
        if (this.match('#')) {
          // A comment goes until the end of the line.
          while (this.peek() !== '\n' && !this.isAtEnd()) this.advance();
          // Could optionally add COMMENT token, or just skip
        } else {
           // Handle potential future single '#' operators or error
           this.error("Unexpected character '#'. Did you mean '##' for a comment?");
        }
        break;

      // Whitespace
      case ' ':
      case '\r':
      case '\t':
        // Ignore whitespace.
        break;
      case '\n':
        this.line++;
        break;

      // Strings
      case '"': this.string('"'); break;
      case '\'': this.character(); break; // Assuming single quote for char

      default:
        if (this.isDigit(c)) {
          this.number();
        } else if (this.isAlpha(c)) {
          this.identifier();
        } else {
          this.error(`Unexpected character: ${c}`);
        }
        break;
    }
  }

  private advance(): string {
    return this.source.charAt(this.current++);
  }

  private addToken(type: TokenType, value: any = null): void {
    const text = this.source.substring(this.start, this.current);
    this.tokens.push({ type, lexeme: text, value, line: this.line });
  }

  private match(expected: string): boolean {
    if (this.isAtEnd()) return false;
    if (this.source.charAt(this.current) !== expected) return false;
    this.current++;
    return true;
  }

  private peek(): string {
    if (this.isAtEnd()) return '\0';
    return this.source.charAt(this.current);
  }

  private peekNext(): string {
    if (this.current + 1 >= this.source.length) return '\0';
    return this.source.charAt(this.current + 1);
  }

  private string(delimiter: '"' | '`'): void {
     // Basic string handling, no escape sequences or interpolation logic yet
    while (this.peek() !== delimiter && !this.isAtEnd()) {
      if (this.peek() === '\n') this.line++;
      // TODO: Handle escape sequences like \" or \`
      // TODO: Handle interpolation ${...} for backticks
      this.advance();
    }

    if (this.isAtEnd()) {
      this.error("Unterminated string.");
      return;
    }

    // The closing delimiter.
    this.advance();

    // Trim the surrounding quotes/backticks.
    const value = this.source.substring(this.start + 1, this.current - 1);
     // TODO: Process escapes and interpolation here
    this.addToken(TokenType.STRING, value);
  }

  private character(): void {
      // Basic char handling, allows one char
      if (this.peek() !== '\'' && !this.isAtEnd()) {
          this.advance(); // consume the character
      } else {
           this.error("Empty or invalid character literal.");
           return;
      }

      if (this.peek() !== '\'') {
          // Handle multi-char error or escapes later
           this.error("Unterminated or multi-character literal.");
           // Consume until quote potentially
           while (this.peek() !== '\'' && !this.isAtEnd()) this.advance();
      }

       if (this.isAtEnd()) {
           this.error("Unterminated character literal.");
           return;
       }

       this.advance(); // Consume the closing '

       // Simple: Get char between quotes
       const value = this.source.substring(this.start + 1, this.current - 1);
       if (value.length !== 1) {
           // Error recovery might allow adding token, or not
           this.error("Character literal must contain exactly one character.");
           // this.addToken(TokenType.CHAR, null); // Or add with error marker
       } else {
            this.addToken(TokenType.CHAR, value);
       }
  }

  private isDigit(c: string): boolean {
    return c >= '0' && c <= '9';
  }

  private number(): void {
    while (this.isDigit(this.peek())) this.advance();

    // Look for a fractional part.
    if (this.peek() === '.' && this.isDigit(this.peekNext())) {
      // Consume the "."
      this.advance();
      while (this.isDigit(this.peek())) this.advance();
    }
     // TODO: Handle exponents (e.g., 1e-10)

    this.addToken(TokenType.NUMBER,
        parseFloat(this.source.substring(this.start, this.current)));
  }

   private identifier(): void {
    while (this.isAlphaNumeric(this.peek())) this.advance();

    const text = this.source.substring(this.start, this.current);
    let type = Lexer.keywords.get(text);
    if (type === undefined) type = TokenType.IDENTIFIER;
    this.addToken(type);
  }

  private isAlpha(c: string): boolean {
    return (c >= 'a' && c <= 'z') ||
           (c >= 'A' && c <= 'Z') ||
            c === '_';
  }

  private isAlphaNumeric(c: string): boolean {
    return this.isAlpha(c) || this.isDigit(c);
  }

   private error(message: string): void {
       // Basic error reporting
       console.error(`[Line ${this.line}] Lexer Error: ${message}`);
       // In a real implementation, collect errors rather than logging directly
   }
}
