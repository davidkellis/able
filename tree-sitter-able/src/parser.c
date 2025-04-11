#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 26
#define LARGE_STATE_COUNT 5
#define SYMBOL_COUNT 31
#define ALIAS_COUNT 0
#define TOKEN_COUNT 18
#define EXTERNAL_TOKEN_COUNT 0
#define FIELD_COUNT 0
#define MAX_ALIAS_SEQUENCE_LENGTH 3
#define PRODUCTION_ID_COUNT 1

enum ts_symbol_identifiers {
  sym__integer_literal_base = 1,
  sym__integer_type_suffix = 2,
  sym__float_literal_base = 3,
  sym__float_type_suffix = 4,
  anon_sym_true = 5,
  anon_sym_false = 6,
  anon_sym_SQUOTE = 7,
  sym__char_content = 8,
  anon_sym_DQUOTE = 9,
  sym__string_content = 10,
  anon_sym_BQUOTE = 11,
  sym__interpolated_content = 12,
  anon_sym_DOLLAR_LBRACE = 13,
  anon_sym_RBRACE = 14,
  sym_escape_sequence = 15,
  sym_nil_literal = 16,
  sym_line_comment = 17,
  sym_source_file = 18,
  sym__expression = 19,
  sym__literal = 20,
  sym_integer_literal = 21,
  sym_float_literal = 22,
  sym_boolean_literal = 23,
  sym_char_literal = 24,
  sym_string_literal = 25,
  sym_interpolated_string_literal = 26,
  sym_interpolation = 27,
  aux_sym_source_file_repeat1 = 28,
  aux_sym_string_literal_repeat1 = 29,
  aux_sym_interpolated_string_literal_repeat1 = 30,
};

static const char * const ts_symbol_names[] = {
  [ts_builtin_sym_end] = "end",
  [sym__integer_literal_base] = "_integer_literal_base",
  [sym__integer_type_suffix] = "_integer_type_suffix",
  [sym__float_literal_base] = "_float_literal_base",
  [sym__float_type_suffix] = "_float_type_suffix",
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
  [sym_integer_literal] = "integer_literal",
  [sym_float_literal] = "float_literal",
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
  [sym__integer_literal_base] = sym__integer_literal_base,
  [sym__integer_type_suffix] = sym__integer_type_suffix,
  [sym__float_literal_base] = sym__float_literal_base,
  [sym__float_type_suffix] = sym__float_type_suffix,
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
  [sym_integer_literal] = sym_integer_literal,
  [sym_float_literal] = sym_float_literal,
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
  [sym__integer_literal_base] = {
    .visible = false,
    .named = true,
  },
  [sym__integer_type_suffix] = {
    .visible = false,
    .named = true,
  },
  [sym__float_literal_base] = {
    .visible = false,
    .named = true,
  },
  [sym__float_type_suffix] = {
    .visible = false,
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
  [sym_integer_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_float_literal] = {
    .visible = true,
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
  [22] = 22,
  [23] = 23,
  [24] = 24,
  [25] = 25,
};

static bool ts_lex(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (eof) ADVANCE(42);
      ADVANCE_MAP(
        '"', 57,
        '#', 2,
        '$', 27,
        '\'', 54,
        '0', 5,
        '\\', 15,
        '_', 19,
        '`', 62,
        'f', 16,
        'n', 21,
        't', 24,
        '}', 70,
      );
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(41);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(6);
      END_STATE();
    case 1:
      if (lookahead == '"') ADVANCE(57);
      if (lookahead == '#') ADVANCE(60);
      if (lookahead == '\\') ADVANCE(15);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(59);
      if (lookahead != 0) ADVANCE(61);
      END_STATE();
    case 2:
      if (lookahead == '#') ADVANCE(74);
      END_STATE();
    case 3:
      if (lookahead == '#') ADVANCE(65);
      if (lookahead == '$') ADVANCE(66);
      if (lookahead == '\\') ADVANCE(15);
      if (lookahead == '`') ADVANCE(62);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(64);
      if (lookahead != 0) ADVANCE(67);
      END_STATE();
    case 4:
      if (lookahead == '#') ADVANCE(56);
      if (lookahead == '\\') ADVANCE(15);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(55);
      if (lookahead != 0 &&
          lookahead != '\'') ADVANCE(55);
      END_STATE();
    case 5:
      ADVANCE_MAP(
        '.', 48,
        '_', 6,
        'B', 36,
        'b', 36,
        'E', 35,
        'e', 35,
        'O', 37,
        'o', 37,
        'X', 39,
        'x', 39,
      );
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(43);
      END_STATE();
    case 6:
      if (lookahead == '.') ADVANCE(48);
      if (lookahead == '_') ADVANCE(6);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(35);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(43);
      END_STATE();
    case 7:
      if (lookahead == '1') ADVANCE(10);
      if (lookahead == '3') ADVANCE(8);
      if (lookahead == '6') ADVANCE(12);
      if (lookahead == '8') ADVANCE(47);
      END_STATE();
    case 8:
      if (lookahead == '2') ADVANCE(47);
      END_STATE();
    case 9:
      if (lookahead == '2') ADVANCE(51);
      END_STATE();
    case 10:
      if (lookahead == '2') ADVANCE(14);
      if (lookahead == '6') ADVANCE(47);
      END_STATE();
    case 11:
      if (lookahead == '3') ADVANCE(9);
      if (lookahead == '6') ADVANCE(13);
      END_STATE();
    case 12:
      if (lookahead == '4') ADVANCE(47);
      END_STATE();
    case 13:
      if (lookahead == '4') ADVANCE(51);
      END_STATE();
    case 14:
      if (lookahead == '8') ADVANCE(47);
      END_STATE();
    case 15:
      ADVANCE_MAP(
        '\\', 72,
        '"', 71,
        '$', 71,
        '\'', 71,
        '`', 71,
        'n', 71,
        'r', 71,
        't', 71,
      );
      END_STATE();
    case 16:
      if (lookahead == 'a') ADVANCE(22);
      END_STATE();
    case 17:
      if (lookahead == 'e') ADVANCE(52);
      END_STATE();
    case 18:
      if (lookahead == 'e') ADVANCE(53);
      END_STATE();
    case 19:
      if (lookahead == 'f') ADVANCE(11);
      if (lookahead == 'i') ADVANCE(7);
      if (lookahead == 'u') ADVANCE(7);
      END_STATE();
    case 20:
      if (lookahead == 'i') ADVANCE(7);
      if (lookahead == 'u') ADVANCE(7);
      END_STATE();
    case 21:
      if (lookahead == 'i') ADVANCE(23);
      END_STATE();
    case 22:
      if (lookahead == 'l') ADVANCE(25);
      END_STATE();
    case 23:
      if (lookahead == 'l') ADVANCE(73);
      END_STATE();
    case 24:
      if (lookahead == 'r') ADVANCE(26);
      END_STATE();
    case 25:
      if (lookahead == 's') ADVANCE(18);
      END_STATE();
    case 26:
      if (lookahead == 'u') ADVANCE(17);
      END_STATE();
    case 27:
      if (lookahead == '{') ADVANCE(68);
      END_STATE();
    case 28:
      if (lookahead == '{') ADVANCE(40);
      END_STATE();
    case 29:
      if (lookahead == '}') ADVANCE(71);
      END_STATE();
    case 30:
      if (lookahead == '}') ADVANCE(71);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(29);
      END_STATE();
    case 31:
      if (lookahead == '}') ADVANCE(71);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(30);
      END_STATE();
    case 32:
      if (lookahead == '}') ADVANCE(71);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(31);
      END_STATE();
    case 33:
      if (lookahead == '}') ADVANCE(71);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(32);
      END_STATE();
    case 34:
      if (lookahead == '}') ADVANCE(71);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(33);
      END_STATE();
    case 35:
      if (lookahead == '+' ||
          lookahead == '-') ADVANCE(38);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(50);
      END_STATE();
    case 36:
      if (lookahead == '0' ||
          lookahead == '1') ADVANCE(44);
      END_STATE();
    case 37:
      if (('0' <= lookahead && lookahead <= '7')) ADVANCE(45);
      END_STATE();
    case 38:
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(50);
      END_STATE();
    case 39:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(46);
      END_STATE();
    case 40:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(34);
      END_STATE();
    case 41:
      if (eof) ADVANCE(42);
      ADVANCE_MAP(
        '"', 57,
        '#', 2,
        '\'', 54,
        '0', 5,
        '_', 20,
        '`', 62,
        'f', 16,
        'n', 21,
        't', 24,
      );
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(41);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(6);
      END_STATE();
    case 42:
      ACCEPT_TOKEN(ts_builtin_sym_end);
      END_STATE();
    case 43:
      ACCEPT_TOKEN(sym__integer_literal_base);
      if (lookahead == '.') ADVANCE(48);
      if (lookahead == '_') ADVANCE(6);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(35);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(43);
      END_STATE();
    case 44:
      ACCEPT_TOKEN(sym__integer_literal_base);
      if (lookahead == '0' ||
          lookahead == '1' ||
          lookahead == '_') ADVANCE(44);
      END_STATE();
    case 45:
      ACCEPT_TOKEN(sym__integer_literal_base);
      if (('0' <= lookahead && lookahead <= '7') ||
          lookahead == '_') ADVANCE(45);
      END_STATE();
    case 46:
      ACCEPT_TOKEN(sym__integer_literal_base);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(46);
      END_STATE();
    case 47:
      ACCEPT_TOKEN(sym__integer_type_suffix);
      END_STATE();
    case 48:
      ACCEPT_TOKEN(sym__float_literal_base);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(35);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(49);
      END_STATE();
    case 49:
      ACCEPT_TOKEN(sym__float_literal_base);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(35);
      if (('0' <= lookahead && lookahead <= '9') ||
          lookahead == '_') ADVANCE(49);
      END_STATE();
    case 50:
      ACCEPT_TOKEN(sym__float_literal_base);
      if (('0' <= lookahead && lookahead <= '9') ||
          lookahead == '_') ADVANCE(50);
      END_STATE();
    case 51:
      ACCEPT_TOKEN(sym__float_type_suffix);
      END_STATE();
    case 52:
      ACCEPT_TOKEN(anon_sym_true);
      END_STATE();
    case 53:
      ACCEPT_TOKEN(anon_sym_false);
      END_STATE();
    case 54:
      ACCEPT_TOKEN(anon_sym_SQUOTE);
      END_STATE();
    case 55:
      ACCEPT_TOKEN(sym__char_content);
      END_STATE();
    case 56:
      ACCEPT_TOKEN(sym__char_content);
      if (lookahead == '#') ADVANCE(74);
      END_STATE();
    case 57:
      ACCEPT_TOKEN(anon_sym_DQUOTE);
      END_STATE();
    case 58:
      ACCEPT_TOKEN(sym__string_content);
      if (lookahead == '\n') ADVANCE(61);
      if (lookahead == '"' ||
          lookahead == '\\') ADVANCE(74);
      if (lookahead != 0) ADVANCE(58);
      END_STATE();
    case 59:
      ACCEPT_TOKEN(sym__string_content);
      if (lookahead == '#') ADVANCE(60);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(59);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '#' &&
          lookahead != '\\') ADVANCE(61);
      END_STATE();
    case 60:
      ACCEPT_TOKEN(sym__string_content);
      if (lookahead == '#') ADVANCE(58);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '#' &&
          lookahead != '\\') ADVANCE(61);
      END_STATE();
    case 61:
      ACCEPT_TOKEN(sym__string_content);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '\\') ADVANCE(61);
      END_STATE();
    case 62:
      ACCEPT_TOKEN(anon_sym_BQUOTE);
      END_STATE();
    case 63:
      ACCEPT_TOKEN(sym__interpolated_content);
      if (lookahead == '\n') ADVANCE(67);
      if (lookahead == '\\' ||
          lookahead == '`') ADVANCE(74);
      if (lookahead != 0) ADVANCE(63);
      END_STATE();
    case 64:
      ACCEPT_TOKEN(sym__interpolated_content);
      if (lookahead == '#') ADVANCE(65);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(64);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(67);
      END_STATE();
    case 65:
      ACCEPT_TOKEN(sym__interpolated_content);
      if (lookahead == '#') ADVANCE(63);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(67);
      END_STATE();
    case 66:
      ACCEPT_TOKEN(sym__interpolated_content);
      if (lookahead == '{') ADVANCE(69);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(67);
      END_STATE();
    case 67:
      ACCEPT_TOKEN(sym__interpolated_content);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(67);
      END_STATE();
    case 68:
      ACCEPT_TOKEN(anon_sym_DOLLAR_LBRACE);
      END_STATE();
    case 69:
      ACCEPT_TOKEN(anon_sym_DOLLAR_LBRACE);
      if (lookahead != 0 &&
          lookahead != '\\' &&
          lookahead != '`') ADVANCE(67);
      END_STATE();
    case 70:
      ACCEPT_TOKEN(anon_sym_RBRACE);
      END_STATE();
    case 71:
      ACCEPT_TOKEN(sym_escape_sequence);
      END_STATE();
    case 72:
      ACCEPT_TOKEN(sym_escape_sequence);
      if (lookahead == 'U' ||
          lookahead == 'u') ADVANCE(28);
      END_STATE();
    case 73:
      ACCEPT_TOKEN(sym_nil_literal);
      END_STATE();
    case 74:
      ACCEPT_TOKEN(sym_line_comment);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(74);
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
  [11] = {.lex_state = 0},
  [12] = {.lex_state = 0},
  [13] = {.lex_state = 0},
  [14] = {.lex_state = 0},
  [15] = {.lex_state = 3},
  [16] = {.lex_state = 3},
  [17] = {.lex_state = 3},
  [18] = {.lex_state = 1},
  [19] = {.lex_state = 1},
  [20] = {.lex_state = 1},
  [21] = {.lex_state = 3},
  [22] = {.lex_state = 4},
  [23] = {.lex_state = 0},
  [24] = {.lex_state = 0},
  [25] = {.lex_state = 0},
};

static const uint16_t ts_parse_table[LARGE_STATE_COUNT][SYMBOL_COUNT] = {
  [0] = {
    [ts_builtin_sym_end] = ACTIONS(1),
    [sym__integer_literal_base] = ACTIONS(1),
    [sym__integer_type_suffix] = ACTIONS(1),
    [sym__float_literal_base] = ACTIONS(1),
    [sym__float_type_suffix] = ACTIONS(1),
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
    [sym_source_file] = STATE(25),
    [sym__expression] = STATE(3),
    [sym__literal] = STATE(3),
    [sym_integer_literal] = STATE(3),
    [sym_float_literal] = STATE(3),
    [sym_boolean_literal] = STATE(3),
    [sym_char_literal] = STATE(3),
    [sym_string_literal] = STATE(3),
    [sym_interpolated_string_literal] = STATE(3),
    [aux_sym_source_file_repeat1] = STATE(3),
    [ts_builtin_sym_end] = ACTIONS(5),
    [sym__integer_literal_base] = ACTIONS(7),
    [sym__float_literal_base] = ACTIONS(9),
    [anon_sym_true] = ACTIONS(11),
    [anon_sym_false] = ACTIONS(11),
    [anon_sym_SQUOTE] = ACTIONS(13),
    [anon_sym_DQUOTE] = ACTIONS(15),
    [anon_sym_BQUOTE] = ACTIONS(17),
    [sym_nil_literal] = ACTIONS(19),
    [sym_line_comment] = ACTIONS(3),
  },
  [2] = {
    [sym__expression] = STATE(2),
    [sym__literal] = STATE(2),
    [sym_integer_literal] = STATE(2),
    [sym_float_literal] = STATE(2),
    [sym_boolean_literal] = STATE(2),
    [sym_char_literal] = STATE(2),
    [sym_string_literal] = STATE(2),
    [sym_interpolated_string_literal] = STATE(2),
    [aux_sym_source_file_repeat1] = STATE(2),
    [ts_builtin_sym_end] = ACTIONS(21),
    [sym__integer_literal_base] = ACTIONS(23),
    [sym__float_literal_base] = ACTIONS(26),
    [anon_sym_true] = ACTIONS(29),
    [anon_sym_false] = ACTIONS(29),
    [anon_sym_SQUOTE] = ACTIONS(32),
    [anon_sym_DQUOTE] = ACTIONS(35),
    [anon_sym_BQUOTE] = ACTIONS(38),
    [sym_nil_literal] = ACTIONS(41),
    [sym_line_comment] = ACTIONS(3),
  },
  [3] = {
    [sym__expression] = STATE(2),
    [sym__literal] = STATE(2),
    [sym_integer_literal] = STATE(2),
    [sym_float_literal] = STATE(2),
    [sym_boolean_literal] = STATE(2),
    [sym_char_literal] = STATE(2),
    [sym_string_literal] = STATE(2),
    [sym_interpolated_string_literal] = STATE(2),
    [aux_sym_source_file_repeat1] = STATE(2),
    [ts_builtin_sym_end] = ACTIONS(44),
    [sym__integer_literal_base] = ACTIONS(7),
    [sym__float_literal_base] = ACTIONS(9),
    [anon_sym_true] = ACTIONS(11),
    [anon_sym_false] = ACTIONS(11),
    [anon_sym_SQUOTE] = ACTIONS(13),
    [anon_sym_DQUOTE] = ACTIONS(15),
    [anon_sym_BQUOTE] = ACTIONS(17),
    [sym_nil_literal] = ACTIONS(46),
    [sym_line_comment] = ACTIONS(3),
  },
  [4] = {
    [sym__expression] = STATE(23),
    [sym__literal] = STATE(23),
    [sym_integer_literal] = STATE(23),
    [sym_float_literal] = STATE(23),
    [sym_boolean_literal] = STATE(23),
    [sym_char_literal] = STATE(23),
    [sym_string_literal] = STATE(23),
    [sym_interpolated_string_literal] = STATE(23),
    [sym__integer_literal_base] = ACTIONS(7),
    [sym__float_literal_base] = ACTIONS(9),
    [anon_sym_true] = ACTIONS(11),
    [anon_sym_false] = ACTIONS(11),
    [anon_sym_SQUOTE] = ACTIONS(13),
    [anon_sym_DQUOTE] = ACTIONS(15),
    [anon_sym_BQUOTE] = ACTIONS(17),
    [sym_nil_literal] = ACTIONS(48),
    [sym_line_comment] = ACTIONS(3),
  },
};

static const uint16_t ts_small_parse_table[] = {
  [0] = 4,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(52), 1,
      sym__integer_literal_base,
    ACTIONS(54), 1,
      sym__integer_type_suffix,
    ACTIONS(50), 9,
      ts_builtin_sym_end,
      sym__float_literal_base,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [21] = 4,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(58), 1,
      sym__integer_literal_base,
    ACTIONS(60), 1,
      sym__float_type_suffix,
    ACTIONS(56), 9,
      ts_builtin_sym_end,
      sym__float_literal_base,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [42] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(64), 1,
      sym__integer_literal_base,
    ACTIONS(62), 9,
      ts_builtin_sym_end,
      sym__float_literal_base,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [60] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(68), 1,
      sym__integer_literal_base,
    ACTIONS(66), 9,
      ts_builtin_sym_end,
      sym__float_literal_base,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [78] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(72), 1,
      sym__integer_literal_base,
    ACTIONS(70), 9,
      ts_builtin_sym_end,
      sym__float_literal_base,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [96] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(76), 1,
      sym__integer_literal_base,
    ACTIONS(74), 9,
      ts_builtin_sym_end,
      sym__float_literal_base,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [114] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(80), 1,
      sym__integer_literal_base,
    ACTIONS(78), 9,
      ts_builtin_sym_end,
      sym__float_literal_base,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [132] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(84), 1,
      sym__integer_literal_base,
    ACTIONS(82), 9,
      ts_builtin_sym_end,
      sym__float_literal_base,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [150] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(88), 1,
      sym__integer_literal_base,
    ACTIONS(86), 9,
      ts_builtin_sym_end,
      sym__float_literal_base,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [168] = 3,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(92), 1,
      sym__integer_literal_base,
    ACTIONS(90), 9,
      ts_builtin_sym_end,
      sym__float_literal_base,
      anon_sym_true,
      anon_sym_false,
      anon_sym_SQUOTE,
      anon_sym_DQUOTE,
      anon_sym_BQUOTE,
      anon_sym_RBRACE,
      sym_nil_literal,
  [186] = 6,
    ACTIONS(94), 1,
      anon_sym_BQUOTE,
    ACTIONS(96), 1,
      sym__interpolated_content,
    ACTIONS(99), 1,
      anon_sym_DOLLAR_LBRACE,
    ACTIONS(102), 1,
      sym_escape_sequence,
    ACTIONS(105), 1,
      sym_line_comment,
    STATE(15), 2,
      sym_interpolation,
      aux_sym_interpolated_string_literal_repeat1,
  [206] = 6,
    ACTIONS(105), 1,
      sym_line_comment,
    ACTIONS(107), 1,
      anon_sym_BQUOTE,
    ACTIONS(109), 1,
      sym__interpolated_content,
    ACTIONS(111), 1,
      anon_sym_DOLLAR_LBRACE,
    ACTIONS(113), 1,
      sym_escape_sequence,
    STATE(17), 2,
      sym_interpolation,
      aux_sym_interpolated_string_literal_repeat1,
  [226] = 6,
    ACTIONS(105), 1,
      sym_line_comment,
    ACTIONS(111), 1,
      anon_sym_DOLLAR_LBRACE,
    ACTIONS(115), 1,
      anon_sym_BQUOTE,
    ACTIONS(117), 1,
      sym__interpolated_content,
    ACTIONS(119), 1,
      sym_escape_sequence,
    STATE(15), 2,
      sym_interpolation,
      aux_sym_interpolated_string_literal_repeat1,
  [246] = 5,
    ACTIONS(105), 1,
      sym_line_comment,
    ACTIONS(121), 1,
      anon_sym_DQUOTE,
    ACTIONS(123), 1,
      sym__string_content,
    ACTIONS(125), 1,
      sym_escape_sequence,
    STATE(19), 1,
      aux_sym_string_literal_repeat1,
  [262] = 5,
    ACTIONS(105), 1,
      sym_line_comment,
    ACTIONS(127), 1,
      anon_sym_DQUOTE,
    ACTIONS(129), 1,
      sym__string_content,
    ACTIONS(131), 1,
      sym_escape_sequence,
    STATE(20), 1,
      aux_sym_string_literal_repeat1,
  [278] = 5,
    ACTIONS(105), 1,
      sym_line_comment,
    ACTIONS(133), 1,
      anon_sym_DQUOTE,
    ACTIONS(135), 1,
      sym__string_content,
    ACTIONS(138), 1,
      sym_escape_sequence,
    STATE(20), 1,
      aux_sym_string_literal_repeat1,
  [294] = 3,
    ACTIONS(105), 1,
      sym_line_comment,
    ACTIONS(143), 1,
      sym_escape_sequence,
    ACTIONS(141), 3,
      anon_sym_BQUOTE,
      sym__interpolated_content,
      anon_sym_DOLLAR_LBRACE,
  [306] = 3,
    ACTIONS(105), 1,
      sym_line_comment,
    ACTIONS(145), 1,
      sym__char_content,
    ACTIONS(147), 1,
      sym_escape_sequence,
  [316] = 2,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(149), 1,
      anon_sym_RBRACE,
  [323] = 2,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(151), 1,
      anon_sym_SQUOTE,
  [330] = 2,
    ACTIONS(3), 1,
      sym_line_comment,
    ACTIONS(153), 1,
      ts_builtin_sym_end,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(5)] = 0,
  [SMALL_STATE(6)] = 21,
  [SMALL_STATE(7)] = 42,
  [SMALL_STATE(8)] = 60,
  [SMALL_STATE(9)] = 78,
  [SMALL_STATE(10)] = 96,
  [SMALL_STATE(11)] = 114,
  [SMALL_STATE(12)] = 132,
  [SMALL_STATE(13)] = 150,
  [SMALL_STATE(14)] = 168,
  [SMALL_STATE(15)] = 186,
  [SMALL_STATE(16)] = 206,
  [SMALL_STATE(17)] = 226,
  [SMALL_STATE(18)] = 246,
  [SMALL_STATE(19)] = 262,
  [SMALL_STATE(20)] = 278,
  [SMALL_STATE(21)] = 294,
  [SMALL_STATE(22)] = 306,
  [SMALL_STATE(23)] = 316,
  [SMALL_STATE(24)] = 323,
  [SMALL_STATE(25)] = 330,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = true}}, SHIFT_EXTRA(),
  [5] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 0, 0, 0),
  [7] = {.entry = {.count = 1, .reusable = false}}, SHIFT(5),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT(6),
  [11] = {.entry = {.count = 1, .reusable = true}}, SHIFT(7),
  [13] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
  [15] = {.entry = {.count = 1, .reusable = true}}, SHIFT(18),
  [17] = {.entry = {.count = 1, .reusable = true}}, SHIFT(16),
  [19] = {.entry = {.count = 1, .reusable = true}}, SHIFT(3),
  [21] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [23] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(5),
  [26] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(6),
  [29] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(7),
  [32] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(22),
  [35] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(18),
  [38] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(16),
  [41] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(2),
  [44] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [46] = {.entry = {.count = 1, .reusable = true}}, SHIFT(2),
  [48] = {.entry = {.count = 1, .reusable = true}}, SHIFT(23),
  [50] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_integer_literal, 1, 0, 0),
  [52] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_integer_literal, 1, 0, 0),
  [54] = {.entry = {.count = 1, .reusable = true}}, SHIFT(9),
  [56] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_float_literal, 1, 0, 0),
  [58] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_float_literal, 1, 0, 0),
  [60] = {.entry = {.count = 1, .reusable = true}}, SHIFT(10),
  [62] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_boolean_literal, 1, 0, 0),
  [64] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_boolean_literal, 1, 0, 0),
  [66] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_interpolated_string_literal, 3, 0, 0),
  [68] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_interpolated_string_literal, 3, 0, 0),
  [70] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_integer_literal, 2, 0, 0),
  [72] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_integer_literal, 2, 0, 0),
  [74] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_float_literal, 2, 0, 0),
  [76] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_float_literal, 2, 0, 0),
  [78] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string_literal, 3, 0, 0),
  [80] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string_literal, 3, 0, 0),
  [82] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string_literal, 2, 0, 0),
  [84] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string_literal, 2, 0, 0),
  [86] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_char_literal, 3, 0, 0),
  [88] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_char_literal, 3, 0, 0),
  [90] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_interpolated_string_literal, 2, 0, 0),
  [92] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_interpolated_string_literal, 2, 0, 0),
  [94] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_interpolated_string_literal_repeat1, 2, 0, 0),
  [96] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_interpolated_string_literal_repeat1, 2, 0, 0), SHIFT_REPEAT(15),
  [99] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_interpolated_string_literal_repeat1, 2, 0, 0), SHIFT_REPEAT(4),
  [102] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_interpolated_string_literal_repeat1, 2, 0, 0), SHIFT_REPEAT(15),
  [105] = {.entry = {.count = 1, .reusable = false}}, SHIFT_EXTRA(),
  [107] = {.entry = {.count = 1, .reusable = false}}, SHIFT(14),
  [109] = {.entry = {.count = 1, .reusable = false}}, SHIFT(17),
  [111] = {.entry = {.count = 1, .reusable = false}}, SHIFT(4),
  [113] = {.entry = {.count = 1, .reusable = true}}, SHIFT(17),
  [115] = {.entry = {.count = 1, .reusable = false}}, SHIFT(8),
  [117] = {.entry = {.count = 1, .reusable = false}}, SHIFT(15),
  [119] = {.entry = {.count = 1, .reusable = true}}, SHIFT(15),
  [121] = {.entry = {.count = 1, .reusable = false}}, SHIFT(12),
  [123] = {.entry = {.count = 1, .reusable = false}}, SHIFT(19),
  [125] = {.entry = {.count = 1, .reusable = true}}, SHIFT(19),
  [127] = {.entry = {.count = 1, .reusable = false}}, SHIFT(11),
  [129] = {.entry = {.count = 1, .reusable = false}}, SHIFT(20),
  [131] = {.entry = {.count = 1, .reusable = true}}, SHIFT(20),
  [133] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_literal_repeat1, 2, 0, 0),
  [135] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_literal_repeat1, 2, 0, 0), SHIFT_REPEAT(20),
  [138] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_string_literal_repeat1, 2, 0, 0), SHIFT_REPEAT(20),
  [141] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_interpolation, 3, 0, 0),
  [143] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_interpolation, 3, 0, 0),
  [145] = {.entry = {.count = 1, .reusable = false}}, SHIFT(24),
  [147] = {.entry = {.count = 1, .reusable = true}}, SHIFT(24),
  [149] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [151] = {.entry = {.count = 1, .reusable = true}}, SHIFT(13),
  [153] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
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
