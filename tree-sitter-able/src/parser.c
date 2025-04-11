#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 21
#define LARGE_STATE_COUNT 4
#define SYMBOL_COUNT 37
#define ALIAS_COUNT 0
#define TOKEN_COUNT 29
#define EXTERNAL_TOKEN_COUNT 0
#define FIELD_COUNT 0
#define MAX_ALIAS_SEQUENCE_LENGTH 4
#define PRODUCTION_ID_COUNT 1

enum ts_symbol_identifiers {
  anon_sym_SEMI = 1,
  anon_sym_LF = 2,
  aux_sym_integer_literal_token1 = 3,
  aux_sym_integer_literal_token2 = 4,
  aux_sym_integer_literal_token3 = 5,
  aux_sym_integer_literal_token4 = 6,
  aux_sym_integer_literal_token5 = 7,
  anon_sym__ = 8,
  anon_sym_i8 = 9,
  anon_sym_i16 = 10,
  anon_sym_i32 = 11,
  anon_sym_i64 = 12,
  anon_sym_i128 = 13,
  anon_sym_u8 = 14,
  anon_sym_u16 = 15,
  anon_sym_u32 = 16,
  anon_sym_u64 = 17,
  anon_sym_u128 = 18,
  aux_sym_float_literal_token1 = 19,
  aux_sym_float_literal_token2 = 20,
  aux_sym_float_literal_token3 = 21,
  anon_sym_f32 = 22,
  anon_sym_f64 = 23,
  sym_string_literal = 24,
  anon_sym_true = 25,
  anon_sym_false = 26,
  sym_char_literal = 27,
  sym_nil_literal = 28,
  sym_source_file = 29,
  sym__statement = 30,
  sym_expression_statement = 31,
  sym__expression = 32,
  sym_integer_literal = 33,
  sym_float_literal = 34,
  sym_boolean_literal = 35,
  aux_sym_source_file_repeat1 = 36,
};

static const char * const ts_symbol_names[] = {
  [ts_builtin_sym_end] = "end",
  [anon_sym_SEMI] = ";",
  [anon_sym_LF] = "\n",
  [aux_sym_integer_literal_token1] = "integer_literal_token1",
  [aux_sym_integer_literal_token2] = "integer_literal_token2",
  [aux_sym_integer_literal_token3] = "integer_literal_token3",
  [aux_sym_integer_literal_token4] = "integer_literal_token4",
  [aux_sym_integer_literal_token5] = "integer_literal_token5",
  [anon_sym__] = "_",
  [anon_sym_i8] = "i8",
  [anon_sym_i16] = "i16",
  [anon_sym_i32] = "i32",
  [anon_sym_i64] = "i64",
  [anon_sym_i128] = "i128",
  [anon_sym_u8] = "u8",
  [anon_sym_u16] = "u16",
  [anon_sym_u32] = "u32",
  [anon_sym_u64] = "u64",
  [anon_sym_u128] = "u128",
  [aux_sym_float_literal_token1] = "float_literal_token1",
  [aux_sym_float_literal_token2] = "float_literal_token2",
  [aux_sym_float_literal_token3] = "float_literal_token3",
  [anon_sym_f32] = "f32",
  [anon_sym_f64] = "f64",
  [sym_string_literal] = "string_literal",
  [anon_sym_true] = "true",
  [anon_sym_false] = "false",
  [sym_char_literal] = "char_literal",
  [sym_nil_literal] = "nil_literal",
  [sym_source_file] = "source_file",
  [sym__statement] = "_statement",
  [sym_expression_statement] = "expression_statement",
  [sym__expression] = "_expression",
  [sym_integer_literal] = "integer_literal",
  [sym_float_literal] = "float_literal",
  [sym_boolean_literal] = "boolean_literal",
  [aux_sym_source_file_repeat1] = "source_file_repeat1",
};

static const TSSymbol ts_symbol_map[] = {
  [ts_builtin_sym_end] = ts_builtin_sym_end,
  [anon_sym_SEMI] = anon_sym_SEMI,
  [anon_sym_LF] = anon_sym_LF,
  [aux_sym_integer_literal_token1] = aux_sym_integer_literal_token1,
  [aux_sym_integer_literal_token2] = aux_sym_integer_literal_token2,
  [aux_sym_integer_literal_token3] = aux_sym_integer_literal_token3,
  [aux_sym_integer_literal_token4] = aux_sym_integer_literal_token4,
  [aux_sym_integer_literal_token5] = aux_sym_integer_literal_token5,
  [anon_sym__] = anon_sym__,
  [anon_sym_i8] = anon_sym_i8,
  [anon_sym_i16] = anon_sym_i16,
  [anon_sym_i32] = anon_sym_i32,
  [anon_sym_i64] = anon_sym_i64,
  [anon_sym_i128] = anon_sym_i128,
  [anon_sym_u8] = anon_sym_u8,
  [anon_sym_u16] = anon_sym_u16,
  [anon_sym_u32] = anon_sym_u32,
  [anon_sym_u64] = anon_sym_u64,
  [anon_sym_u128] = anon_sym_u128,
  [aux_sym_float_literal_token1] = aux_sym_float_literal_token1,
  [aux_sym_float_literal_token2] = aux_sym_float_literal_token2,
  [aux_sym_float_literal_token3] = aux_sym_float_literal_token3,
  [anon_sym_f32] = anon_sym_f32,
  [anon_sym_f64] = anon_sym_f64,
  [sym_string_literal] = sym_string_literal,
  [anon_sym_true] = anon_sym_true,
  [anon_sym_false] = anon_sym_false,
  [sym_char_literal] = sym_char_literal,
  [sym_nil_literal] = sym_nil_literal,
  [sym_source_file] = sym_source_file,
  [sym__statement] = sym__statement,
  [sym_expression_statement] = sym_expression_statement,
  [sym__expression] = sym__expression,
  [sym_integer_literal] = sym_integer_literal,
  [sym_float_literal] = sym_float_literal,
  [sym_boolean_literal] = sym_boolean_literal,
  [aux_sym_source_file_repeat1] = aux_sym_source_file_repeat1,
};

static const TSSymbolMetadata ts_symbol_metadata[] = {
  [ts_builtin_sym_end] = {
    .visible = false,
    .named = true,
  },
  [anon_sym_SEMI] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_LF] = {
    .visible = true,
    .named = false,
  },
  [aux_sym_integer_literal_token1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_integer_literal_token2] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_integer_literal_token3] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_integer_literal_token4] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_integer_literal_token5] = {
    .visible = false,
    .named = false,
  },
  [anon_sym__] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_i8] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_i16] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_i32] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_i64] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_i128] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_u8] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_u16] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_u32] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_u64] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_u128] = {
    .visible = true,
    .named = false,
  },
  [aux_sym_float_literal_token1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_float_literal_token2] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_float_literal_token3] = {
    .visible = false,
    .named = false,
  },
  [anon_sym_f32] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_f64] = {
    .visible = true,
    .named = false,
  },
  [sym_string_literal] = {
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
  [sym_char_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_nil_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_source_file] = {
    .visible = true,
    .named = true,
  },
  [sym__statement] = {
    .visible = false,
    .named = true,
  },
  [sym_expression_statement] = {
    .visible = true,
    .named = true,
  },
  [sym__expression] = {
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
  [aux_sym_source_file_repeat1] = {
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
};

static bool ts_lex(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (eof) ADVANCE(35);
      ADVANCE_MAP(
        '"', 1,
        '\'', 33,
        '0', 39,
        ';', 36,
        '_', 44,
        'f', 10,
        'i', 3,
        'n', 19,
        't', 22,
        'u', 4,
        '+', 38,
        '-', 38,
      );
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(0);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(40);
      END_STATE();
    case 1:
      if (lookahead == '"') ADVANCE(60);
      if (lookahead != 0) ADVANCE(1);
      END_STATE();
    case 2:
      if (lookahead == '\'') ADVANCE(63);
      END_STATE();
    case 3:
      if (lookahead == '1') ADVANCE(6);
      if (lookahead == '3') ADVANCE(7);
      if (lookahead == '6') ADVANCE(12);
      if (lookahead == '8') ADVANCE(45);
      END_STATE();
    case 4:
      if (lookahead == '1') ADVANCE(9);
      if (lookahead == '3') ADVANCE(8);
      if (lookahead == '6') ADVANCE(13);
      if (lookahead == '8') ADVANCE(50);
      END_STATE();
    case 5:
      if (lookahead == '2') ADVANCE(58);
      END_STATE();
    case 6:
      if (lookahead == '2') ADVANCE(14);
      if (lookahead == '6') ADVANCE(46);
      END_STATE();
    case 7:
      if (lookahead == '2') ADVANCE(47);
      END_STATE();
    case 8:
      if (lookahead == '2') ADVANCE(52);
      END_STATE();
    case 9:
      if (lookahead == '2') ADVANCE(15);
      if (lookahead == '6') ADVANCE(51);
      END_STATE();
    case 10:
      if (lookahead == '3') ADVANCE(5);
      if (lookahead == '6') ADVANCE(11);
      if (lookahead == 'a') ADVANCE(20);
      END_STATE();
    case 11:
      if (lookahead == '4') ADVANCE(59);
      END_STATE();
    case 12:
      if (lookahead == '4') ADVANCE(48);
      END_STATE();
    case 13:
      if (lookahead == '4') ADVANCE(53);
      END_STATE();
    case 14:
      if (lookahead == '8') ADVANCE(49);
      END_STATE();
    case 15:
      if (lookahead == '8') ADVANCE(54);
      END_STATE();
    case 16:
      if (lookahead == 'a') ADVANCE(20);
      END_STATE();
    case 17:
      if (lookahead == 'e') ADVANCE(61);
      END_STATE();
    case 18:
      if (lookahead == 'e') ADVANCE(62);
      END_STATE();
    case 19:
      if (lookahead == 'i') ADVANCE(21);
      END_STATE();
    case 20:
      if (lookahead == 'l') ADVANCE(23);
      END_STATE();
    case 21:
      if (lookahead == 'l') ADVANCE(64);
      END_STATE();
    case 22:
      if (lookahead == 'r') ADVANCE(24);
      END_STATE();
    case 23:
      if (lookahead == 's') ADVANCE(18);
      END_STATE();
    case 24:
      if (lookahead == 'u') ADVANCE(17);
      END_STATE();
    case 25:
      if (lookahead == '+' ||
          lookahead == '-') ADVANCE(30);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(57);
      END_STATE();
    case 26:
      if (lookahead == '+' ||
          lookahead == '-') ADVANCE(31);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(56);
      END_STATE();
    case 27:
      if (lookahead == '0' ||
          lookahead == '1') ADVANCE(43);
      END_STATE();
    case 28:
      if (('0' <= lookahead && lookahead <= '7')) ADVANCE(42);
      END_STATE();
    case 29:
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(55);
      END_STATE();
    case 30:
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(57);
      END_STATE();
    case 31:
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(56);
      END_STATE();
    case 32:
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(41);
      END_STATE();
    case 33:
      if (lookahead != 0 &&
          lookahead != '\'') ADVANCE(2);
      END_STATE();
    case 34:
      if (eof) ADVANCE(35);
      ADVANCE_MAP(
        '\n', 37,
        '"', 1,
        '\'', 33,
        '0', 39,
        ';', 36,
        '_', 44,
        'f', 16,
        'n', 19,
        't', 22,
        '+', 38,
        '-', 38,
      );
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(34);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(40);
      END_STATE();
    case 35:
      ACCEPT_TOKEN(ts_builtin_sym_end);
      END_STATE();
    case 36:
      ACCEPT_TOKEN(anon_sym_SEMI);
      END_STATE();
    case 37:
      ACCEPT_TOKEN(anon_sym_LF);
      if (lookahead == '\n') ADVANCE(37);
      END_STATE();
    case 38:
      ACCEPT_TOKEN(aux_sym_integer_literal_token1);
      END_STATE();
    case 39:
      ACCEPT_TOKEN(aux_sym_integer_literal_token2);
      ADVANCE_MAP(
        '.', 29,
        'B', 27,
        'b', 27,
        'E', 25,
        'e', 25,
        'O', 28,
        'o', 28,
        'X', 32,
        'x', 32,
      );
      if (('0' <= lookahead && lookahead <= '9') ||
          lookahead == '_') ADVANCE(40);
      END_STATE();
    case 40:
      ACCEPT_TOKEN(aux_sym_integer_literal_token2);
      if (lookahead == '.') ADVANCE(29);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(25);
      if (('0' <= lookahead && lookahead <= '9') ||
          lookahead == '_') ADVANCE(40);
      END_STATE();
    case 41:
      ACCEPT_TOKEN(aux_sym_integer_literal_token3);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'F') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'f')) ADVANCE(41);
      END_STATE();
    case 42:
      ACCEPT_TOKEN(aux_sym_integer_literal_token4);
      if (('0' <= lookahead && lookahead <= '7') ||
          lookahead == '_') ADVANCE(42);
      END_STATE();
    case 43:
      ACCEPT_TOKEN(aux_sym_integer_literal_token5);
      if (lookahead == '0' ||
          lookahead == '1' ||
          lookahead == '_') ADVANCE(43);
      END_STATE();
    case 44:
      ACCEPT_TOKEN(anon_sym__);
      END_STATE();
    case 45:
      ACCEPT_TOKEN(anon_sym_i8);
      END_STATE();
    case 46:
      ACCEPT_TOKEN(anon_sym_i16);
      END_STATE();
    case 47:
      ACCEPT_TOKEN(anon_sym_i32);
      END_STATE();
    case 48:
      ACCEPT_TOKEN(anon_sym_i64);
      END_STATE();
    case 49:
      ACCEPT_TOKEN(anon_sym_i128);
      END_STATE();
    case 50:
      ACCEPT_TOKEN(anon_sym_u8);
      END_STATE();
    case 51:
      ACCEPT_TOKEN(anon_sym_u16);
      END_STATE();
    case 52:
      ACCEPT_TOKEN(anon_sym_u32);
      END_STATE();
    case 53:
      ACCEPT_TOKEN(anon_sym_u64);
      END_STATE();
    case 54:
      ACCEPT_TOKEN(anon_sym_u128);
      END_STATE();
    case 55:
      ACCEPT_TOKEN(aux_sym_float_literal_token1);
      if (lookahead == 'E' ||
          lookahead == 'e') ADVANCE(26);
      if (('0' <= lookahead && lookahead <= '9') ||
          lookahead == '_') ADVANCE(55);
      END_STATE();
    case 56:
      ACCEPT_TOKEN(aux_sym_float_literal_token2);
      if (('0' <= lookahead && lookahead <= '9') ||
          lookahead == '_') ADVANCE(56);
      END_STATE();
    case 57:
      ACCEPT_TOKEN(aux_sym_float_literal_token3);
      if (('0' <= lookahead && lookahead <= '9') ||
          lookahead == '_') ADVANCE(57);
      END_STATE();
    case 58:
      ACCEPT_TOKEN(anon_sym_f32);
      END_STATE();
    case 59:
      ACCEPT_TOKEN(anon_sym_f64);
      END_STATE();
    case 60:
      ACCEPT_TOKEN(sym_string_literal);
      END_STATE();
    case 61:
      ACCEPT_TOKEN(anon_sym_true);
      END_STATE();
    case 62:
      ACCEPT_TOKEN(anon_sym_false);
      END_STATE();
    case 63:
      ACCEPT_TOKEN(sym_char_literal);
      END_STATE();
    case 64:
      ACCEPT_TOKEN(sym_nil_literal);
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
  [4] = {.lex_state = 34},
  [5] = {.lex_state = 34},
  [6] = {.lex_state = 34},
  [7] = {.lex_state = 34},
  [8] = {.lex_state = 34},
  [9] = {.lex_state = 34},
  [10] = {.lex_state = 34},
  [11] = {.lex_state = 34},
  [12] = {.lex_state = 34},
  [13] = {.lex_state = 34},
  [14] = {.lex_state = 0},
  [15] = {.lex_state = 0},
  [16] = {.lex_state = 0},
  [17] = {.lex_state = 0},
  [18] = {.lex_state = 0},
  [19] = {.lex_state = 0},
  [20] = {.lex_state = 0},
};

static const uint16_t ts_parse_table[LARGE_STATE_COUNT][SYMBOL_COUNT] = {
  [0] = {
    [ts_builtin_sym_end] = ACTIONS(1),
    [anon_sym_SEMI] = ACTIONS(1),
    [aux_sym_integer_literal_token1] = ACTIONS(1),
    [aux_sym_integer_literal_token2] = ACTIONS(1),
    [aux_sym_integer_literal_token3] = ACTIONS(1),
    [aux_sym_integer_literal_token4] = ACTIONS(1),
    [aux_sym_integer_literal_token5] = ACTIONS(1),
    [anon_sym__] = ACTIONS(1),
    [anon_sym_i8] = ACTIONS(1),
    [anon_sym_i16] = ACTIONS(1),
    [anon_sym_i32] = ACTIONS(1),
    [anon_sym_i64] = ACTIONS(1),
    [anon_sym_i128] = ACTIONS(1),
    [anon_sym_u8] = ACTIONS(1),
    [anon_sym_u16] = ACTIONS(1),
    [anon_sym_u32] = ACTIONS(1),
    [anon_sym_u64] = ACTIONS(1),
    [anon_sym_u128] = ACTIONS(1),
    [aux_sym_float_literal_token1] = ACTIONS(1),
    [aux_sym_float_literal_token2] = ACTIONS(1),
    [aux_sym_float_literal_token3] = ACTIONS(1),
    [anon_sym_f32] = ACTIONS(1),
    [anon_sym_f64] = ACTIONS(1),
    [sym_string_literal] = ACTIONS(1),
    [anon_sym_true] = ACTIONS(1),
    [anon_sym_false] = ACTIONS(1),
    [sym_char_literal] = ACTIONS(1),
    [sym_nil_literal] = ACTIONS(1),
  },
  [1] = {
    [sym_source_file] = STATE(20),
    [sym__statement] = STATE(2),
    [sym_expression_statement] = STATE(2),
    [sym__expression] = STATE(9),
    [sym_integer_literal] = STATE(9),
    [sym_float_literal] = STATE(9),
    [sym_boolean_literal] = STATE(9),
    [aux_sym_source_file_repeat1] = STATE(2),
    [ts_builtin_sym_end] = ACTIONS(3),
    [aux_sym_integer_literal_token1] = ACTIONS(5),
    [aux_sym_integer_literal_token2] = ACTIONS(7),
    [aux_sym_integer_literal_token3] = ACTIONS(9),
    [aux_sym_integer_literal_token4] = ACTIONS(9),
    [aux_sym_integer_literal_token5] = ACTIONS(9),
    [aux_sym_float_literal_token1] = ACTIONS(11),
    [aux_sym_float_literal_token2] = ACTIONS(13),
    [aux_sym_float_literal_token3] = ACTIONS(13),
    [sym_string_literal] = ACTIONS(15),
    [anon_sym_true] = ACTIONS(17),
    [anon_sym_false] = ACTIONS(17),
    [sym_char_literal] = ACTIONS(15),
    [sym_nil_literal] = ACTIONS(15),
  },
  [2] = {
    [sym__statement] = STATE(3),
    [sym_expression_statement] = STATE(3),
    [sym__expression] = STATE(9),
    [sym_integer_literal] = STATE(9),
    [sym_float_literal] = STATE(9),
    [sym_boolean_literal] = STATE(9),
    [aux_sym_source_file_repeat1] = STATE(3),
    [ts_builtin_sym_end] = ACTIONS(19),
    [aux_sym_integer_literal_token1] = ACTIONS(5),
    [aux_sym_integer_literal_token2] = ACTIONS(7),
    [aux_sym_integer_literal_token3] = ACTIONS(9),
    [aux_sym_integer_literal_token4] = ACTIONS(9),
    [aux_sym_integer_literal_token5] = ACTIONS(9),
    [aux_sym_float_literal_token1] = ACTIONS(11),
    [aux_sym_float_literal_token2] = ACTIONS(13),
    [aux_sym_float_literal_token3] = ACTIONS(13),
    [sym_string_literal] = ACTIONS(15),
    [anon_sym_true] = ACTIONS(17),
    [anon_sym_false] = ACTIONS(17),
    [sym_char_literal] = ACTIONS(15),
    [sym_nil_literal] = ACTIONS(15),
  },
  [3] = {
    [sym__statement] = STATE(3),
    [sym_expression_statement] = STATE(3),
    [sym__expression] = STATE(9),
    [sym_integer_literal] = STATE(9),
    [sym_float_literal] = STATE(9),
    [sym_boolean_literal] = STATE(9),
    [aux_sym_source_file_repeat1] = STATE(3),
    [ts_builtin_sym_end] = ACTIONS(21),
    [aux_sym_integer_literal_token1] = ACTIONS(23),
    [aux_sym_integer_literal_token2] = ACTIONS(26),
    [aux_sym_integer_literal_token3] = ACTIONS(29),
    [aux_sym_integer_literal_token4] = ACTIONS(29),
    [aux_sym_integer_literal_token5] = ACTIONS(29),
    [aux_sym_float_literal_token1] = ACTIONS(32),
    [aux_sym_float_literal_token2] = ACTIONS(35),
    [aux_sym_float_literal_token3] = ACTIONS(35),
    [sym_string_literal] = ACTIONS(38),
    [anon_sym_true] = ACTIONS(41),
    [anon_sym_false] = ACTIONS(41),
    [sym_char_literal] = ACTIONS(38),
    [sym_nil_literal] = ACTIONS(38),
  },
};

static const uint16_t ts_small_parse_table[] = {
  [0] = 3,
    ACTIONS(48), 1,
      anon_sym__,
    ACTIONS(44), 2,
      ts_builtin_sym_end,
      anon_sym_LF,
    ACTIONS(46), 14,
      anon_sym_SEMI,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token2,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token1,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [24] = 3,
    ACTIONS(54), 1,
      anon_sym__,
    ACTIONS(50), 2,
      ts_builtin_sym_end,
      anon_sym_LF,
    ACTIONS(52), 14,
      anon_sym_SEMI,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token2,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token1,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [48] = 3,
    ACTIONS(60), 1,
      anon_sym__,
    ACTIONS(56), 2,
      ts_builtin_sym_end,
      anon_sym_LF,
    ACTIONS(58), 14,
      anon_sym_SEMI,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token2,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token1,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [72] = 3,
    ACTIONS(66), 1,
      anon_sym__,
    ACTIONS(62), 2,
      ts_builtin_sym_end,
      anon_sym_LF,
    ACTIONS(64), 14,
      anon_sym_SEMI,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token2,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token1,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [96] = 2,
    ACTIONS(68), 2,
      ts_builtin_sym_end,
      anon_sym_LF,
    ACTIONS(70), 14,
      anon_sym_SEMI,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token2,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token1,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [117] = 4,
    ACTIONS(72), 1,
      ts_builtin_sym_end,
    ACTIONS(74), 1,
      anon_sym_SEMI,
    ACTIONS(76), 1,
      anon_sym_LF,
    ACTIONS(78), 13,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token2,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token1,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [142] = 2,
    ACTIONS(80), 2,
      ts_builtin_sym_end,
      anon_sym_LF,
    ACTIONS(82), 14,
      anon_sym_SEMI,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token2,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token1,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [163] = 2,
    ACTIONS(84), 2,
      ts_builtin_sym_end,
      anon_sym_LF,
    ACTIONS(86), 14,
      anon_sym_SEMI,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token2,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token1,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [184] = 2,
    ACTIONS(88), 2,
      ts_builtin_sym_end,
      anon_sym_LF,
    ACTIONS(90), 14,
      anon_sym_SEMI,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token2,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token1,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [205] = 2,
    ACTIONS(92), 2,
      ts_builtin_sym_end,
      anon_sym_LF,
    ACTIONS(94), 14,
      anon_sym_SEMI,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token2,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token1,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [226] = 2,
    ACTIONS(98), 2,
      aux_sym_integer_literal_token2,
      aux_sym_float_literal_token1,
    ACTIONS(96), 12,
      ts_builtin_sym_end,
      aux_sym_integer_literal_token1,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
      sym_string_literal,
      anon_sym_true,
      anon_sym_false,
      sym_char_literal,
      sym_nil_literal,
  [245] = 1,
    ACTIONS(100), 10,
      anon_sym_i8,
      anon_sym_i16,
      anon_sym_i32,
      anon_sym_i64,
      anon_sym_i128,
      anon_sym_u8,
      anon_sym_u16,
      anon_sym_u32,
      anon_sym_u64,
      anon_sym_u128,
  [258] = 1,
    ACTIONS(102), 10,
      anon_sym_i8,
      anon_sym_i16,
      anon_sym_i32,
      anon_sym_i64,
      anon_sym_i128,
      anon_sym_u8,
      anon_sym_u16,
      anon_sym_u32,
      anon_sym_u64,
      anon_sym_u128,
  [271] = 4,
    ACTIONS(104), 1,
      aux_sym_integer_literal_token2,
    ACTIONS(108), 1,
      aux_sym_float_literal_token1,
    ACTIONS(110), 2,
      aux_sym_float_literal_token2,
      aux_sym_float_literal_token3,
    ACTIONS(106), 3,
      aux_sym_integer_literal_token3,
      aux_sym_integer_literal_token4,
      aux_sym_integer_literal_token5,
  [287] = 1,
    ACTIONS(112), 2,
      anon_sym_f32,
      anon_sym_f64,
  [292] = 1,
    ACTIONS(114), 2,
      anon_sym_f32,
      anon_sym_f64,
  [297] = 1,
    ACTIONS(116), 1,
      ts_builtin_sym_end,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(4)] = 0,
  [SMALL_STATE(5)] = 24,
  [SMALL_STATE(6)] = 48,
  [SMALL_STATE(7)] = 72,
  [SMALL_STATE(8)] = 96,
  [SMALL_STATE(9)] = 117,
  [SMALL_STATE(10)] = 142,
  [SMALL_STATE(11)] = 163,
  [SMALL_STATE(12)] = 184,
  [SMALL_STATE(13)] = 205,
  [SMALL_STATE(14)] = 226,
  [SMALL_STATE(15)] = 245,
  [SMALL_STATE(16)] = 258,
  [SMALL_STATE(17)] = 271,
  [SMALL_STATE(18)] = 287,
  [SMALL_STATE(19)] = 292,
  [SMALL_STATE(20)] = 297,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 0, 0, 0),
  [5] = {.entry = {.count = 1, .reusable = true}}, SHIFT(17),
  [7] = {.entry = {.count = 1, .reusable = false}}, SHIFT(4),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT(4),
  [11] = {.entry = {.count = 1, .reusable = false}}, SHIFT(5),
  [13] = {.entry = {.count = 1, .reusable = true}}, SHIFT(5),
  [15] = {.entry = {.count = 1, .reusable = true}}, SHIFT(9),
  [17] = {.entry = {.count = 1, .reusable = true}}, SHIFT(8),
  [19] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [21] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [23] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(17),
  [26] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(4),
  [29] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(4),
  [32] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(5),
  [35] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(5),
  [38] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(9),
  [41] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(8),
  [44] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_integer_literal, 1, 0, 0),
  [46] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_integer_literal, 1, 0, 0),
  [48] = {.entry = {.count = 1, .reusable = false}}, SHIFT(15),
  [50] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_float_literal, 1, 0, 0),
  [52] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_float_literal, 1, 0, 0),
  [54] = {.entry = {.count = 1, .reusable = false}}, SHIFT(18),
  [56] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_float_literal, 2, 0, 0),
  [58] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_float_literal, 2, 0, 0),
  [60] = {.entry = {.count = 1, .reusable = false}}, SHIFT(19),
  [62] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_integer_literal, 2, 0, 0),
  [64] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_integer_literal, 2, 0, 0),
  [66] = {.entry = {.count = 1, .reusable = false}}, SHIFT(16),
  [68] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_boolean_literal, 1, 0, 0),
  [70] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_boolean_literal, 1, 0, 0),
  [72] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_expression_statement, 1, 0, 0),
  [74] = {.entry = {.count = 1, .reusable = false}}, SHIFT(14),
  [76] = {.entry = {.count = 1, .reusable = true}}, SHIFT(14),
  [78] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_expression_statement, 1, 0, 0),
  [80] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_integer_literal, 3, 0, 0),
  [82] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_integer_literal, 3, 0, 0),
  [84] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_float_literal, 3, 0, 0),
  [86] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_float_literal, 3, 0, 0),
  [88] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_integer_literal, 4, 0, 0),
  [90] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_integer_literal, 4, 0, 0),
  [92] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_float_literal, 4, 0, 0),
  [94] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_float_literal, 4, 0, 0),
  [96] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_expression_statement, 2, 0, 0),
  [98] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_expression_statement, 2, 0, 0),
  [100] = {.entry = {.count = 1, .reusable = true}}, SHIFT(10),
  [102] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [104] = {.entry = {.count = 1, .reusable = false}}, SHIFT(7),
  [106] = {.entry = {.count = 1, .reusable = true}}, SHIFT(7),
  [108] = {.entry = {.count = 1, .reusable = false}}, SHIFT(6),
  [110] = {.entry = {.count = 1, .reusable = true}}, SHIFT(6),
  [112] = {.entry = {.count = 1, .reusable = true}}, SHIFT(11),
  [114] = {.entry = {.count = 1, .reusable = true}}, SHIFT(13),
  [116] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
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
