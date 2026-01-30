#include "tree_sitter/parser.h"
#include <stdbool.h>
#include <ctype.h>
#include <string.h>

enum TokenType {
  NEWLINE,
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

  while (lexer->lookahead == ' ' || lexer->lookahead == '\t' ||
         lexer->lookahead == '\v' || lexer->lookahead == '\f') {
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
