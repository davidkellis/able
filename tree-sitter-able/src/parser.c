#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 22
#define LARGE_STATE_COUNT 5
#define SYMBOL_COUNT 27
#define ALIAS_COUNT 0
#define TOKEN_COUNT 16
#define EXTERNAL_TOKEN_COUNT 0
#define FIELD_COUNT 0
#define MAX_ALIAS_SEQUENCE_LENGTH 3
#define PRODUCTION_ID_COUNT 1

enum ts_symbol_identifiers {
  sym_integer_literal = 1,
  sym_float_literal = 2,
  anon_sym_true = 3,
  anon_sym_false = 4,
  anon_sym_SQUOTE = 5,
  sym__char_content = 6,
  anon_sym_DQUOTE = 7,
  sym__string_content = 8,
  anon_sym_BQUOTE = 9,
  sym__interpolated_content = 10,
  anon_sym_DOLLAR_LBRACE = 11,
  anon_sym_RBRACE = 12,
  sym_escape_sequence = 13,
  sym_nil_literal = 14,
  sym_line_comment = 15,
  sym_source_file = 16,
  sym__expression = 17,
  sym__literal = 18,
  sym_boolean_literal = 19,
  sym_char_literal = 20,
  sym_string_literal = 21,
  sym_interpolated_string_literal = 22,
  sym_interpolation = 23,
  aux_sym_source_file_repeat1 = 24,
  aux_sym_string_literal_repeat1 = 25,
  aux_sym_interpolated_string_literal_repeat1 = 26,
};

static const char * const ts_symbol_names[] = {
  [ts_builtin_sym_end] = "end",
  [sym_integer_literal] = "integer_literal",
  [sym_float_literal] = "float_literal",
  [anon_sym_true] = "true",
  [anon_sym_false] = "false",
  [anon_sym_SQUOTE] = "'",
  [sym__char_content] = "_char_content",
  [anon_sym_DQUOTE] = "\"",
  [sym__string_content] = "_string_content",
  [anon_sym_BQUOTE] = "`",
  [sym__interpolated_content] = "_interpolated_content",
  [anon_sym_DOLLAR_LBRACE] = "${",
  [anon_sym_RBRACE] = "}",
  [sym_escape_sequence] = "escape_sequence",
  [sym_nil_literal] = "nil_literal",
  [sym_line_comment] = "line_comment",
  [sym_source_file] = "source_file",
  [sym__expression] = "_expression",
  [sym__literal] = "_literal",
  [sym_boolean_literal] = "boolean_literal",
  [sym_char_literal] = "char_literal",
  [sym_string_literal] = "string_literal",
  [sym_interpolated_string_literal] = "interpolated_string_literal",
  [sym_interpolation] = "interpolation",
  [aux_sym_source_file_repeat1] = "source_file_repeat1",
  [aux_sym_string_literal_repeat1] = "string_literal_repeat1",
  [aux_sym_interpolated_string_literal_repeat1] = "interpolated_string_literal_repeat1",
};

static const TSSymbol ts_symbol_map[] = {
  [ts_builtin_sym_end] = ts_builtin_sym_end,
  [sym_integer_literal] = sym_integer_literal,
  [sym_float_literal] = sym_float_literal,
  [anon_sym_true] = anon_sym_true,
  [anon_sym_false] = anon_sym_false,
  [anon_sym_SQUOTE] = anon_sym_SQUOTE,
  [sym__char_content] = sym__char_content,
  [anon_sym_DQUOTE] = anon_sym_DQUOTE,
  [sym__string_content] = sym__string_content,
  [anon_sym_BQUOTE] = anon_sym_BQUOTE,
  [sym__interpolated_content] = sym__interpolated_content,
  [anon_sym_DOLLAR_LBRACE] = anon_sym_DOLLAR_LBRACE,
  [anon_sym_RBRACE] = anon_sym_RBRACE,
  [sym_escape_sequence] = sym_escape_sequence,
  [sym_nil_literal] = sym_nil_literal,
  [sym_line_comment] = sym_line_comment,
  [sym_source_file] = sym_source_file,
  [sym__expression] = sym__expression,
  [sym__literal] = sym__literal,
  [sym_boolean_literal] = sym_boolean_literal,
  [sym_char_literal] = sym_char_literal,
  [sym_string_literal] = sym_string_literal,
  [sym_interpolated_string_literal] = sym_interpolated_string_literal,
  [sym_interpolation] = sym_interpolation,
  [aux_sym_source_file_repeat1] = aux_sym_source_file_repeat1,
  [aux_sym_string_literal_repeat1] = aux_sym_string_literal_repeat1,
  [aux_sym_interpolated_string_literal_repeat1] = aux_sym_interpolated_string_literal_repeat1,
};

static const TSSymbolMetadata ts_symbol_metadata[] = {
  [ts_builtin_sym_end] = {
    .visible = false,
    .named = true,
  },
  [sym_integer_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_float_literal] = {
    .visible = true,
    .named = true,
  },
  [anon_sym_true] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_false] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_SQUOTE] = {
    .visible = true,
    .named = false,
  },
  [sym__char_content] = {
    .visible = false,
    .named = true,
  },
  [anon_sym_DQUOTE] = {
    .visible = true,
    .named = false,
  },
  [sym__string_content] = {
    .visible = false,
    .named = true,
  },
  [anon_sym_BQUOTE] = {
    .visible = true,
    .named = false,
  },
  [sym__interpolated_content] = {
    .visible = false,
    .named = true,
  },
  [anon_sym_DOLLAR_LBRACE] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_RBRACE] = {
    .visible = true,
    .named = false,
  },
  [sym_escape_sequence] = {
    .visible = true,
    .named = true,
  },
  [sym_nil_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_line_comment] = {
    .visible = true,
    .named = true,
  },
  [sym_source_file] = {
    .visible = true,
    .named = true,
  },
  [sym__expression] = {
    .visible = false,
    .named = true,
  },
  [sym__literal] = {
    .visible = false,
    .named = true,
  },
  [sym_boolean_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_char_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_string_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_interpolated_string_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_interpolation] = {
    .visible = true,
    .named = true,
  },
  [aux_sym_source_file_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_string_literal_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_interpolated_string_literal_repeat1] = {
    .visible = false,
    .named = false,
  },
};

static const TSSymbol ts_alias_sequences[PRODUCTION_ID_COUNT][MAX_ALIAS_SEQUENCE_LENGTH] = {
  [0] = {0},
};

static const uint16_t ts_non_terminal_alias_map[] = {
  0,
};

static const TSStateId ts_primary_state_ids[STATE_COUNT] = {
  [0] = 0,
  [1] = 1,
  [2] = 2,
  [3] = 3,
  [4] = 4,
  [5] = 5,
  [6] = 6,
  [7] = 7,
  [8] = 8,
  [9] = 9,
  [10] = 10,
  [11] = 11,
  [12] = 12,
  [13] = 13,
  [14] = 14,
  [15] = 15,
  [16] = 16,
  [17] = 17,
  [18] = 18,
  [19] = 19,
  [20] = 20,
  [21] = 21,
};

static bool ts_lex(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (eof) ADVANCE(39);
      ADVANCE_MAP(
        '"', 61,
        '#', 2,
        '$', 24,
        '\'', 58,
        '0', 42,
        '\\', 13,
        '`', 66,
        'f', 14,
        'n', 18,
        't', 21,
        '}', 74,
      );
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(38);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(43);
      END_STATE();
    case 1:
      if (lookahead == '"') ADVANCE(61);
      if (lookahead == '#') ADVANCE(64);
      if (lookahead == '\\') ADVANCE(13);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(63);
      if (lookahead != 0) ADVANCE(65);
      END_STATE();
    case 2:
      if (lookahead == '#') ADVANCE(78);
      END_STATE();
    case 3:
      if (lookahead == '#') ADVANCE(69);
      if (lookahead == '$') ADVANCE(70);
      if (lookahead == '\\') ADVANCE(13);
      if (lookahead == '`') ADVANCE(66);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(68);
      if (lookahead != 0) ADVANCE(71);
      END_STATE();
    case 4:
      if (lookahead == '#') ADVANCE(60);
      if (lookahead == '\\') ADVANCE(13);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(59);
      if (lookahead != 0 &&
          lookahead != '\'') ADVANCE(59);
      END_STATE();
    case 5:
      if (lookahead == '1') ADVANCE(8);
      if (lookahead == '3') ADVANCE(6);
      if (lookahead == '6') ADVANCE(10);
      if (lookahead == '8') ADVANCE(40);
      END_STATE();
    case 6:
      if (lookahead == '2') ADVANCE(40);
      END_STATE();
    case 7:
      if (lookahead == '2') ADVANCE(50);
      END_STATE();
    case 8:
      if (lookahead == '2') ADVANCE(12);
      if (lookahead == '6') ADVANCE(40);
      END_STATE();
    case 9:
      if (lookahead == '3') ADVANCE(7);
      if (lookahead == '6') ADVANCE(11);
      END_STATE();
    case 10:
      if (lookahead == '4') ADVANCE(40);
      END_STATE();
    case 11:
      if (lookahead == '4') ADVANCE(50);
      END_STATE();
    case 12:
      if (lookahead == '8') ADVANCE(40);
      END_STATE();
    case 13:
      ADVANCE_MAP(
        '\\', 76,
        '"', 75,
        '$', 75,
        '\'', 75,
        '`', 75,
        'n', 75,
        'r', 75,
        't', 75,
      );
      END_STATE();
    case 14:
      if (lookahead == 'a') ADVANCE(19);
      END_STATE();
    case 15:
      if (lookahead == 'e') ADVANCE(56);
      END_STATE();
    case 16:
      if (lookahead == 'e') ADVANCE(57);
      END_STATE();
    case 17:
      if (lookahead == 'f') ADVANCE(9);
      END_STATE();
    case 18:
      if (lookahead == 'i') ADVANCE(20);
      END_STATE();
    case 19:
      if (lookahead == 'l') ADVANCE(22);
      END_STATE();
    case 20:
      if (lookahead == 'l') ADVANCE(77);
      END_STATE();
    case 21:
      if (lookahead == 'r') ADVANCE(23);
      END_STATE();
    case 22:
      if (lookahead == 's') ADVANCE(16);
      END_STATE();
    case 23:
      if (lookahead == 'u') ADVANCE(15);
      END_STATE();
    case 24:
      if (lookahead == '{') ADVANCE(72);
      END_STATE();
    case 25:
      if (lookahead == '{') ADVANCE(37);
      END_STATE();
    case 26:
      if (lookahead == '}') ADVANCE(75);
      END_STATE();
    case 27:
      if (lookahead == '}') ADVANCE(75);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(26);
      END_STATE();
    case 28:
      if (lookahead == '}') ADVANCE(75);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(27);
      END_STATE();
    case 29:
      if (lookahead == '}') ADVANCE(75);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(28);
      END_STATE();
    case 30:
      if (lookahead == '}') ADVANCE(75);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(29);
      END_STATE();
    case 31:
      if (lookahead == '}') ADVANCE(75);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(30);
      END_STATE();
    case 32:
      if (lookahead == '+' ||
          lookahead == '-') ADVANCE(35);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(55);
      END_STATE();
    case 33:
      if (lookahead == '0' ||
          lookahead == '1') ADVANCE(45);
      END_STATE();
    case 34:
      if (('0' <= lookahead && lookahead <= '7')) ADVANCE(47);
      END_STATE();
    case 35:
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(55);
      END_STATE();
    case 36:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(49);
      END_STATE();
    case 37:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(31);
      END_STATE();
    case 38:
      if (eof) ADVANCE(39);
      ADVANCE_MAP(
        '"', 61,
        '#', 2,
        '\'', 58,
        '0', 42,
        '`', 66,
        'f', 14,
        'n', 18,
        't', 21,
      );
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(38);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(43);
      END_STATE();
    case 39:
      ACCEPT_TOKEN(ts_builtin_sym_end);
      END_STATE();
    case 40:
      ACCEPT_TOKEN(sym_integer_literal);
      END_STATE();
    case 41:
      ACCEPT_TOKEN(sym_integer_literal);
      if (lookahead == '.') ADVANCE(51);
      if (lookahead == '_') ADVANCE(41);
      if (lookahead == 'f') ADVANCE(9);
      if (lookahead == 'i') ADVANCE(5);
      if (lookahead == 'u') ADVANCE(5);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(32);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(43);
      END_STATE();
    case 42:
      ACCEPT_TOKEN(sym_integer_literal);
      ADVANCE_MAP(
        '.', 51,
        '_', 41,
        'B', 33,
        'b', 33,
        'E', 32,
        'e', 32,
        'O', 34,
        'o', 34,
        'X', 36,
        'x', 36,
      );
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(43);
      END_STATE();
    case 43:
      ACCEPT_TOKEN(sym_integer_literal);
      if (lookahead == '.') ADVANCE(51);
      if (lookahead == '_') ADVANCE(41);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(32);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(43);
      END_STATE();
    case 44:
      ACCEPT_TOKEN(sym_integer_literal);
      if (lookahead == '_') ADVANCE(44);
      if (lookahead == 'i') ADVANCE(5);
      if (lookahead == 'u') ADVANCE(5);
      if (lookahead == '0' ||
          lookahead == '1') ADVANCE(45);
      END_STATE();
    case 45:
      ACCEPT_TOKEN(sym_integer_literal);
      if (lookahead == '_') ADVANCE(44);
      if (lookahead == '0' ||
          lookahead == '1') ADVANCE(45);
      END_STATE();
    case 46:
      ACCEPT_TOKEN(sym_integer_literal);
      if (lookahead == '_') ADVANCE(46);
      if (lookahead == 'i') ADVANCE(5);
      if (lookahead == 'u') ADVANCE(5);
      if (('0' <= lookahead && lookahead <= '7')) ADVANCE(47);
      END_STATE();
    case 47:
      ACCEPT_TOKEN(sym_integer_literal);
      if (lookahead == '_') ADVANCE(46);
      if (('0' <= lookahead && lookahead <= '7')) ADVANCE(47);
      END_STATE();
    case 48:
      ACCEPT_TOKEN(sym_integer_literal);
      if (lookahead == '_') ADVANCE(48);
      if (lookahead == 'i') ADVANCE(5);
      if (lookahead == 'u') ADVANCE(5);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(49);
      END_STATE();
    case 49:
      ACCEPT_TOKEN(sym_integer_literal);
      if (lookahead == '_') ADVANCE(48);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(49);
      END_STATE();
    case 50:
      ACCEPT_TOKEN(sym_float_literal);
      END_STATE();
    case 51:
      ACCEPT_TOKEN(sym_float_literal);
      if (lookahead == '_') ADVANCE(17);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(32);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(53);
      END_STATE();
    case 52:
      ACCEPT_TOKEN(sym_float_literal);
      if (lookahead == '_') ADVANCE(52);
      if (lookahead == 'f') ADVANCE(9);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(32);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(53);
      END_STATE();
    case 53:
      ACCEPT_TOKEN(sym_float_literal);
      if (lookahead == '_') ADVANCE(52);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(32);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(53);
      END_STATE();
    case 54:
      ACCEPT_TOKEN(sym_float_literal);
      if (lookahead == '_') ADVANCE(54);
      if (lookahead == 'f') ADVANCE(9);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(55);
      END_STATE();
    case 55:
      ACCEPT_TOKEN(sym_float_literal);
      if (lookahead == '_') ADVANCE(54);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(55);
      END_STATE();
    case 56:
      ACCEPT_TOKEN(anon_sym_true);
      END_STATE();
    case 57:
      ACCEPT_TOKEN(anon_sym_false);
      END_STATE();
    case 58:
      ACCEPT_TOKEN(anon_sym_SQUOTE);
      END_STATE();
    case 59:
      ACCEPT_TOKEN(sym__char_content);
      END_STATE();
    case 60:
      ACCEPT_TOKEN(sym__char_content);
      if (lookahead == '#') ADVANCE(78);
      END_STATE();
    case 61:
      ACCEPT_TOKEN(anon_sym_DQUOTE);
      END_STATE();
    case 62:
      ACCEPT_TOKEN(sym__string_content);
      if (lookahead == '\n') ADVANCE(65);
      if (lookahead == '"' ||
          lookahead == '\\') ADVANCE(78);
      if (lookahead != 0) ADVANCE(62);
      END_STATE();
    case 63:
      ACCEPT_TOKEN(sym__string_content);
      if (lookahead == '#') ADVANCE(64);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(63);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '#' &&
          lookahead != '\\') ADVANCE(65);
      END_STATE();
    case 64:
      ACCEPT_TOKEN(sym__string_content);
      if (lookahead == '#') ADVANCE(62);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '#' &&
          lookahead != '\\') ADVANCE(65);
      END_STATE();
    case 65:
      ACCEPT_TOKEN(sym__string_content);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '\\') ADVANCE(65);
      END_STATE();
    case 66:
      ACCEPT_TOKEN(anon_sym_BQUOTE);
      END_STATE();
    case 67:
      ACCEPT_TOKEN(sym__interpolated_content);
      if (lookahead == '\n') ADVANCE(71);
      if (lookahead == '\\' ||
          lookahead == '`') ADVANCE(78);
      if (lookahead != 0) ADVANCE(67);
      END_STATE();
    case 68:
      ACCEPT_TOKEN(sym__interpolated_content);
      if (lookahead == '#') ADVANCE(69);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(68);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(71);
      END_STATE();
    case 69:
      ACCEPT_TOKEN(sym__interpolated_content);
      if (lookahead == '#') ADVANCE(67);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(71);
      END_STATE();
    case 70:
      ACCEPT_TOKEN(sym__interpolated_content);
      if (lookahead == '{') ADVANCE(73);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(71);
      END_STATE();
    case 71:
      ACCEPT_TOKEN(sym__interpolated_content);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(71);
      END_STATE();
    case 72:
      ACCEPT_TOKEN(anon_sym_DOLLAR_LBRACE);
      END_STATE();
    case 73:
      ACCEPT_TOKEN(anon_sym_DOLLAR_LBRACE);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(71);
      END_STATE();
    case 74:
      ACCEPT_TOKEN(anon_sym_RBRACE);
      END_STATE();
    case 75:
      ACCEPT_TOKEN(sym_escape_sequence);
      END_STATE();
    case 76:
      ACCEPT_TOKEN(sym_escape_sequence);
      if (lookahead == 'U' ||
          lookahead == 'u') ADVANCE(25);
      END_STATE();
    case 77:
      ACCEPT_TOKEN(sym_nil_literal);
      END_STATE();
    case 78:
      ACCEPT_TOKEN(sym_line_comment);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(78);
      END_STATE();
    default:
      return false;
  }
}

static const TSLexMode ts_lex_modes[STATE_COUNT] = {
  [0] = {.lex_state = 0},
  [1] = {.lex_state = 0},
  [2] = {.lex_state = 0},
  [3] = {.lex_state = 0},
  [4] = {.lex_state = 0},
  [5] = {.lex_state = 0},
  [6] = {.lex_state = 0},
  [7] = {.lex_state = 0},
  [8] = {.lex_state = 0},
  [9] = {.lex_state = 0},
  [10] = {.lex_state = 0},
  [11] = {.lex_state = 3},
  [12] = {.lex_state = 3},
  [13] = {.lex_state = 3},
  [14] = {.lex_state = 1},
  [15] = {.lex_state = 1},
  [16] = {.lex_state = 1},
  [17] = {.lex_state = 3},
  [18] = {.lex_state = 4},
  [19] = {.lex_state = 0},
  [20] = {.lex_state = 0},
  [21] = {.lex_state = 0},
};

static const uint16_t ts_parse_table[LARGE_STATE_COUNT][SYMBOL_COUNT] = {
  [0] = {
    [ts_builtin_sym_end] = ACTIONS(1),
    [sym_integer_literal] = ACTIONS(1),
    [sym_float_literal] = ACTIONS(1),
    [anon_sym_true] = ACTIONS(1),
    [anon_sym_false] = ACTIONS(1),
    [anon_sym_SQUOTE] = ACTIONS(1),
    [anon_sym_DQUOTE] = ACTIONS(1),
    [anon_sym_BQUOTE] = ACTIONS(1),
    [anon_sym_DOLLAR_LBRACE] = ACTIONS(1),
    [anon_sym_RBRACE] = ACTIONS(1),
    [sym_escape_sequence] = ACTIONS(1),
    [sym_nil_literal] = ACTIONS(1),
    [sym_line_comment] = ACTIONS(3),
  },
  [1] = {
    [sym_source_file] = STATE(20),
    [sym__expression] = STATE(2),
    [sym__literal] = STATE(2),
    [sym_boolean_literal] = STATE(2),
    [sym_char_literal] = STATE(2),
    [sym_string_literal] = STATE(2),
    [sym_interpolated_string_literal] = STATE(2),
    [aux_sym_source_file_repeat1] = STATE(2),
    [ts_builtin_sym_end] = ACTIONS(5),
    [sym_integer_literal] = ACTIONS(7),
    [sym_float_literal] = ACTIONS(9),
    [anon_sym_true] = ACTIONS(11),
    [anon_sym_false] = ACTIONS(11),
    [anon_sym_SQUOTE] = ACTIONS(13),
    [anon_sym_DQUOTE] = ACTIONS(15),
    [anon_sym_BQUOTE] = ACTIONS(17),
    [sym_nil_literal] = ACTIONS(9),
    [sym_line_comment] = ACTIONS(3),
  },
  [2] = {
    [sym__expression] = STATE(3),
    [sym__literal] = STATE(3),
    [sym_boolean_literal] = STATE(3),
    [sym_char_literal] = STATE(3),
    [sym_string_literal] = STATE(3),
    [sym_interpolated_string_literal] = STATE(3),
    [aux_sym_source_file_repeat1] = STATE(3),
    [ts_builtin_sym_end] = ACTIONS(19),
    [sym_integer_literal] = ACTIONS(21),
    [sym_float_literal] = ACTIONS(23),
    [anon_sym_true] = ACTIONS(11),
    [anon_sym_false] = ACTIONS(11),
    [anon_sym_SQUOTE] = ACTIONS(13),
    [anon_sym_DQUOTE] = ACTIONS(15),
    [anon_sym_BQUOTE] = ACTIONS(17),
    [sym_nil_literal] = ACTIONS(23),
    [sym_line_comment] = ACTIONS(3),
  },
  [3] = {
    [sym__expression] = STATE(3),
    [sym__literal] = STATE(3),
    [sym_boolean_literal] = STATE(3),
    [sym_char_literal] = STATE(3),
    [sym_string_literal] = STATE(3),
    [sym_interpolated_string_literal] = STATE(3),
    [aux_sym_source_file_repeat1] = STATE(3),
    [ts_builtin_sym_end] = ACTIONS(25),
    [sym_integer_literal] = ACTIONS(27),
    [sym_float_literal] = ACTIONS(30),
    [anon_sym_true] = ACTIONS(33),
    [anon_sym_false] = ACTIONS(33),
    [anon_sym_SQUOTE] = ACTIONS(36),
    [anon_sym_DQUOTE] = ACTIONS(39),
    [anon_sym_BQUOTE] = ACTIONS(42),
    [sym_nil_literal] = ACTIONS(30),
    [sym_line_comment] = ACTIONS(3),
  },
  [4] = {
    [sym__expression] = STATE(21),
    [sym__literal] = STATE(21),
    [sym_boolean_literal] = STATE(21),
    [sym_char_literal] = STATE(21),
    [sym_string_literal] = STATE(21),
    [sym_interpolated_string_literal] = STATE(21),
    [sym_integer_literal] = ACTIONS(45),
    [sym_float_literal] = ACTIONS(47),
    [anon_sym_true] = ACTIONS(11),
    [anon_sym_false] = ACTIONS(11),
    [anon_sym_SQUOTE] = ACTIONS(13),
    [anon_sym_DQUOTE] = ACTIONS(15),
    [anon_sym_BQUOTE] = ACTIONS(17),
    [sym_nil_literal] = ACTIONS(47),
    [sym_line_comment] = ACTIONS(3),
  },
};

static const uint16_t ts_small_parse_table[] = {
  [0] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(51), 1,
      sym_integer_literal,
    ACTIONS(49), 9,
      ts_builtin_sym_end,
      sym_float_literal,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [18] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(55), 1,
      sym_integer_literal,
    ACTIONS(53), 9,
      ts_builtin_sym_end,
      sym_float_literal,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [36] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(59), 1,
      sym_integer_literal,
    ACTIONS(57), 9,
      ts_builtin_sym_end,
      sym_float_literal,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [54] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(63), 1,
      sym_integer_literal,
    ACTIONS(61), 9,
      ts_builtin_sym_end,
      sym_float_literal,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [72] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(67), 1,
      sym_integer_literal,
    ACTIONS(65), 9,
      ts_builtin_sym_end,
      sym_float_literal,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [90] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(71), 1,
      sym_integer_literal,
    ACTIONS(69), 9,
      ts_builtin_sym_end,
      sym_float_literal,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [108] = 6,
    ACTIONS(73), 1,
      anon_sym_BQUOTE,
    ACTIONS(75), 1,
      sym__interpolated_content,
    ACTIONS(77), 1,
      anon_sym_DOLLAR_LBRACE,
    ACTIONS(79), 1,
      sym_escape_sequence,
    ACTIONS(81), 1,
      sym_line_comment,
    STATE(12), 2,
      sym_interpolation,
      aux_sym_interpolated_string_literal_repeat1,
  [128] = 6,
    ACTIONS(77), 1,
      anon_sym_DOLLAR_LBRACE,
    ACTIONS(81), 1,
      sym_line_comment,
    ACTIONS(83), 1,
      anon_sym_BQUOTE,
    ACTIONS(85), 1,
      sym__interpolated_content,
    ACTIONS(87), 1,
      sym_escape_sequence,
    STATE(13), 2,
      sym_interpolation,
      aux_sym_interpolated_string_literal_repeat1,
  [148] = 6,
    ACTIONS(81), 1,
      sym_line_comment,
    ACTIONS(89), 1,
      anon_sym_BQUOTE,
    ACTIONS(91), 1,
      sym__interpolated_content,
    ACTIONS(94), 1,
      anon_sym_DOLLAR_LBRACE,
    ACTIONS(97), 1,
      sym_escape_sequence,
    STATE(13), 2,
      sym_interpolation,
      aux_sym_interpolated_string_literal_repeat1,
  [168] = 5,
    ACTIONS(81), 1,
      sym_line_comment,
    ACTIONS(100), 1,
      anon_sym_DQUOTE,
    ACTIONS(102), 1,
      sym__string_content,
    ACTIONS(104), 1,
      sym_escape_sequence,
    STATE(15), 1,
      aux_sym_string_literal_repeat1,
  [184] = 5,
    ACTIONS(81), 1,
      sym_line_comment,
    ACTIONS(106), 1,
      anon_sym_DQUOTE,
    ACTIONS(108), 1,
      sym__string_content,
    ACTIONS(111), 1,
      sym_escape_sequence,
    STATE(15), 1,
      aux_sym_string_literal_repeat1,
  [200] = 5,
    ACTIONS(81), 1,
      sym_line_comment,
    ACTIONS(114), 1,
      anon_sym_DQUOTE,
    ACTIONS(116), 1,
      sym__string_content,
    ACTIONS(118), 1,
      sym_escape_sequence,
    STATE(14), 1,
      aux_sym_string_literal_repeat1,
  [216] = 3,
    ACTIONS(81), 1,
      sym_line_comment,
    ACTIONS(122), 1,
      sym_escape_sequence,
    ACTIONS(120), 3,
      anon_sym_BQUOTE,
      sym__interpolated_content,
      anon_sym_DOLLAR_LBRACE,
  [228] = 3,
    ACTIONS(81), 1,
      sym_line_comment,
    ACTIONS(124), 1,
      sym__char_content,
    ACTIONS(126), 1,
      sym_escape_sequence,
  [238] = 2,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(128), 1,
      anon_sym_SQUOTE,
  [245] = 2,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(130), 1,
      ts_builtin_sym_end,
  [252] = 2,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(132), 1,
      anon_sym_RBRACE,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(5)] = 0,
  [SMALL_STATE(6)] = 18,
  [SMALL_STATE(7)] = 36,
  [SMALL_STATE(8)] = 54,
  [SMALL_STATE(9)] = 72,
  [SMALL_STATE(10)] = 90,
  [SMALL_STATE(11)] = 108,
  [SMALL_STATE(12)] = 128,
  [SMALL_STATE(13)] = 148,
  [SMALL_STATE(14)] = 168,
  [SMALL_STATE(15)] = 184,
  [SMALL_STATE(16)] = 200,
  [SMALL_STATE(17)] = 216,
  [SMALL_STATE(18)] = 228,
  [SMALL_STATE(19)] = 238,
  [SMALL_STATE(20)] = 245,
  [SMALL_STATE(21)] = 252,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = true}}, SHIFT_EXTRA(),
  [5] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 0, 0, 0),
  [7] = {.entry = {.count = 1, .reusable = false}}, SHIFT(2),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT(2),
  [11] = {.entry = {.count = 1, .reusable = true}}, SHIFT(6),
  [13] = {.entry = {.count = 1, .reusable = true}}, SHIFT(18),
  [15] = {.entry = {.count = 1, .reusable = true}}, SHIFT(16),
  [17] = {.entry = {.count = 1, .reusable = true}}, SHIFT(11),
  [19] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [21] = {.entry = {.count = 1, .reusable = false}}, SHIFT(3),
  [23] = {.entry = {.count = 1, .reusable = true}}, SHIFT(3),
  [25] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [27] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(3),
  [30] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(3),
  [33] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(6),
  [36] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(18),
  [39] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(16),
  [42] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(11),
  [45] = {.entry = {.count = 1, .reusable = false}}, SHIFT(21),
  [47] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [49] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_interpolated_string_literal, 2, 0, 0),
  [51] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_interpolated_string_literal, 2, 0, 0),
  [53] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_boolean_literal, 1, 0, 0),
  [55] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_boolean_literal, 1, 0, 0),
  [57] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_interpolated_string_literal, 3, 0, 0),
  [59] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_interpolated_string_literal, 3, 0, 0),
  [61] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string_literal, 3, 0, 0),
  [63] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string_literal, 3, 0, 0),
  [65] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_char_literal, 3, 0, 0),
  [67] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_char_literal, 3, 0, 0),
  [69] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string_literal, 2, 0, 0),
  [71] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string_literal, 2, 0, 0),
  [73] = {.entry = {.count = 1, .reusable = false}}, SHIFT(5),
  [75] = {.entry = {.count = 1, .reusable = false}}, SHIFT(12),
  [77] = {.entry = {.count = 1, .reusable = false}}, SHIFT(4),
  [79] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [81] = {.entry = {.count = 1, .reusable = false}}, SHIFT_EXTRA(),
  [83] = {.entry = {.count = 1, .reusable = false}}, SHIFT(7),
  [85] = {.entry = {.count = 1, .reusable = false}}, SHIFT(13),
  [87] = {.entry = {.count = 1, .reusable = true}}, SHIFT(13),
  [89] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_interpolated_string_literal_repeat1, 2, 0, 0),
  [91] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_interpolated_string_literal_repeat1, 2, 0, 0), SHIFT_REPEAT(13),
  [94] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_interpolated_string_literal_repeat1, 2, 0, 0), SHIFT_REPEAT(4),
  [97] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_interpolated_string_literal_repeat1, 2, 0, 0), SHIFT_REPEAT(13),
  [100] = {.entry = {.count = 1, .reusable = false}}, SHIFT(8),
  [102] = {.entry = {.count = 1, .reusable = false}}, SHIFT(15),
  [104] = {.entry = {.count = 1, .reusable = true}}, SHIFT(15),
  [106] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_literal_repeat1, 2, 0, 0),
  [108] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_literal_repeat1, 2, 0, 0), SHIFT_REPEAT(15),
  [111] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_string_literal_repeat1, 2, 0, 0), SHIFT_REPEAT(15),
  [114] = {.entry = {.count = 1, .reusable = false}}, SHIFT(10),
  [116] = {.entry = {.count = 1, .reusable = false}}, SHIFT(14),
  [118] = {.entry = {.count = 1, .reusable = true}}, SHIFT(14),
  [120] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_interpolation, 3, 0, 0),
  [122] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_interpolation, 3, 0, 0),
  [124] = {.entry = {.count = 1, .reusable = false}}, SHIFT(19),
  [126] = {.entry = {.count = 1, .reusable = true}}, SHIFT(19),
  [128] = {.entry = {.count = 1, .reusable = true}}, SHIFT(9),
  [130] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [132] = {.entry = {.count = 1, .reusable = true}}, SHIFT(17),
};

#ifdef __cplusplus
extern "C" {
#endif
#ifdef TREE_SITTER_HIDE_SYMBOLS
#define TS_PUBLIC
#elif defined(_WIN32)
#define TS_PUBLIC __declspec(dllexport)
#else
#define TS_PUBLIC __attribute__((visibility("default")))
#endif

TS_PUBLIC const TSLanguage *tree_sitter_able(void) {
  static const TSLanguage language = {
    .version = LANGUAGE_VERSION,
    .symbol_count = SYMBOL_COUNT,
    .alias_count = ALIAS_COUNT,
    .token_count = TOKEN_COUNT,
    .external_token_count = EXTERNAL_TOKEN_COUNT,
    .state_count = STATE_COUNT,
    .large_state_count = LARGE_STATE_COUNT,
    .production_id_count = PRODUCTION_ID_COUNT,
    .field_count = FIELD_COUNT,
    .max_alias_sequence_length = MAX_ALIAS_SEQUENCE_LENGTH,
    .parse_table = &ts_parse_table[0][0],
    .small_parse_table = ts_small_parse_table,
    .small_parse_table_map = ts_small_parse_table_map,
    .parse_actions = ts_parse_actions,
    .symbol_names = ts_symbol_names,
    .symbol_metadata = ts_symbol_metadata,
    .public_symbol_map = ts_symbol_map,
    .alias_map = ts_non_terminal_alias_map,
    .alias_sequences = &ts_alias_sequences[0][0],
    .lex_modes = ts_lex_modes,
    .lex_fn = ts_lex,
    .primary_state_ids = ts_primary_state_ids,
  };
  return &language;
}
#ifdef __cplusplus
}
#endif
