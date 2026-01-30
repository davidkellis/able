#include "tree_sitter/parser.h"
#include <stdbool.h>
#include <ctype.h>
#include <string.h>

enum TokenType {
  NEWLINE,
  TYPE_APPLICATION_SEP,
};

static bool is_ident_char(int32_t c) {
  return isalnum(c) || c == '_';
}

static size_t scan_identifier(TSLexer *lexer, char *buffer, size_t max_len) {
  size_t len = 0;
  while (is_ident_char(lexer->lookahead) && len + 1 < max_len) {
    buffer[len++] = (char)lexer->lookahead;
    lexer->advance(lexer, false);
  }
  buffer[len] = '\0';
  return len;
}

static bool is_line_continuation(TSLexer *lexer) {
  switch (lexer->lookahead) {
    case '.':
      return true;
    case '?':
      lexer->advance(lexer, false);
      return lexer->lookahead == '.';
    case '|':
      lexer->advance(lexer, false);
      return lexer->lookahead == '|' || lexer->lookahead == '>';
    case '&':
      lexer->advance(lexer, false);
      return lexer->lookahead == '&';
    case '=':
      lexer->advance(lexer, false);
      return lexer->lookahead == '=';
    case '!':
      lexer->advance(lexer, false);
      return lexer->lookahead == '=';
    case '>':
    case '<':
    case '*':
    case '%':
    case '^':
    case '/':
      return true;
    case '+':
    case '-':
      lexer->advance(lexer, false);
      return lexer->lookahead == ' ' || lexer->lookahead == '\t' ||
        lexer->lookahead == '\r' || lexer->lookahead == '\n';
    default:
      if (isalpha(lexer->lookahead)) {
        char ident[8];
        scan_identifier(lexer, ident, sizeof(ident));
        return (
          strcmp(ident, "or") == 0 ||
          strcmp(ident, "ensure") == 0 ||
          strcmp(ident, "rescue") == 0 ||
          strcmp(ident, "where") == 0
        );
      }
      return false;
  }
}

static bool is_horizontal_space(int32_t c) {
  return c == ' ' || c == '\t' || c == '\v' || c == '\f';
}

static bool is_type_prefix_start(int32_t c) {
  return isalpha(c) || c == '_' || c == '?' || c == '!' || c == '(';
}

static bool is_disallowed_type_keyword(const char *ident) {
  return (
    strcmp(ident, "fn") == 0 ||
    strcmp(ident, "struct") == 0 ||
    strcmp(ident, "union") == 0 ||
    strcmp(ident, "interface") == 0 ||
    strcmp(ident, "impl") == 0 ||
    strcmp(ident, "methods") == 0 ||
    strcmp(ident, "type") == 0 ||
    strcmp(ident, "package") == 0 ||
    strcmp(ident, "import") == 0 ||
    strcmp(ident, "dynimport") == 0 ||
    strcmp(ident, "extern") == 0 ||
    strcmp(ident, "prelude") == 0 ||
    strcmp(ident, "private") == 0 ||
    strcmp(ident, "do") == 0 ||
    strcmp(ident, "return") == 0 ||
    strcmp(ident, "if") == 0 ||
    strcmp(ident, "elsif") == 0 ||
    strcmp(ident, "or") == 0 ||
    strcmp(ident, "else") == 0 ||
    strcmp(ident, "while") == 0 ||
    strcmp(ident, "loop") == 0 ||
    strcmp(ident, "for") == 0 ||
    strcmp(ident, "in") == 0 ||
    strcmp(ident, "match") == 0 ||
    strcmp(ident, "case") == 0 ||
    strcmp(ident, "breakpoint") == 0 ||
    strcmp(ident, "break") == 0 ||
    strcmp(ident, "continue") == 0 ||
    strcmp(ident, "raise") == 0 ||
    strcmp(ident, "rescue") == 0 ||
    strcmp(ident, "ensure") == 0 ||
    strcmp(ident, "rethrow") == 0 ||
    strcmp(ident, "spawn") == 0 ||
    strcmp(ident, "await") == 0 ||
    strcmp(ident, "as") == 0 ||
    strcmp(ident, "true") == 0 ||
    strcmp(ident, "false") == 0 ||
    strcmp(ident, "where") == 0
  );
}

static bool scan_type_application_sep(TSLexer *lexer, const bool *valid_symbols) {
  if (!valid_symbols[TYPE_APPLICATION_SEP]) {
    return false;
  }

  if (!is_horizontal_space(lexer->lookahead)) {
    return false;
  }

  while (is_horizontal_space(lexer->lookahead)) {
    lexer->advance(lexer, true);
  }

  lexer->mark_end(lexer);

  if (!is_type_prefix_start(lexer->lookahead)) {
    return false;
  }

  if (isalpha(lexer->lookahead) || lexer->lookahead == '_') {
    char ident[16];
    scan_identifier(lexer, ident, sizeof(ident));
    if (strcmp(ident, "Self") != 0 && strcmp(ident, "nil") != 0 &&
        strcmp(ident, "void") != 0 && strcmp(ident, "Iterator") != 0 &&
        is_disallowed_type_keyword(ident)) {
      return false;
    }
  }

  lexer->result_symbol = TYPE_APPLICATION_SEP;
  return true;
}

void *tree_sitter_able_external_scanner_create(void) {
  return NULL;
}

void tree_sitter_able_external_scanner_destroy(void *payload) {
  (void)payload;
}

void tree_sitter_able_external_scanner_reset(void *payload) {
  (void)payload;
}

unsigned tree_sitter_able_external_scanner_serialize(void *payload, char *buffer) {
  (void)payload;
  (void)buffer;
  return 0;
}

void tree_sitter_able_external_scanner_deserialize(void *payload, const char *buffer, unsigned length) {
  (void)payload;
  (void)buffer;
  (void)length;
}

bool tree_sitter_able_external_scanner_scan(void *payload, TSLexer *lexer, const bool *valid_symbols) {
  (void)payload;

  if (scan_type_application_sep(lexer, valid_symbols)) {
    return true;
  }

  while (is_horizontal_space(lexer->lookahead)) {
    lexer->advance(lexer, true);
  }

  if (lexer->lookahead != '\n' && lexer->lookahead != '\r') {
    return false;
  }

  lexer->advance(lexer, false);
  if (lexer->lookahead == '\n') {
    lexer->advance(lexer, false);
  }
  lexer->mark_end(lexer);

  while (lexer->lookahead == ' ' || lexer->lookahead == '\t' ||
         lexer->lookahead == '\v' || lexer->lookahead == '\f') {
    lexer->advance(lexer, true);
  }

  if (is_line_continuation(lexer)) {
    return false;
  }

  if (!valid_symbols[NEWLINE]) {
    return false;
  }

  lexer->result_symbol = NEWLINE;
  return true;
}
