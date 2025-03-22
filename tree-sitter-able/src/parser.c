#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 30
#define LARGE_STATE_COUNT 2
#define SYMBOL_COUNT 38
#define ALIAS_COUNT 0
#define TOKEN_COUNT 21
#define EXTERNAL_TOKEN_COUNT 0
#define FIELD_COUNT 0
#define MAX_ALIAS_SEQUENCE_LENGTH 5
#define PRODUCTION_ID_COUNT 1

enum ts_symbol_identifiers {
  anon_sym_EQ = 1,
  anon_sym_fn = 2,
  anon_sym_LPAREN = 3,
  anon_sym_RPAREN = 4,
  anon_sym_bool = 5,
  anon_sym_string = 6,
  anon_sym_int = 7,
  anon_sym_i32 = 8,
  anon_sym_i64 = 9,
  anon_sym_i128 = 10,
  anon_sym_float = 11,
  anon_sym_f32 = 12,
  anon_sym_f64 = 13,
  anon_sym_LBRACE = 14,
  anon_sym_RBRACE = 15,
  anon_sym_do = 16,
  anon_sym_end = 17,
  anon_sym_return = 18,
  sym_identifier = 19,
  sym_number = 20,
  sym_source_file = 21,
  sym__definition = 22,
  sym_assignment = 23,
  sym_simple_assignment = 24,
  sym_function_definition = 25,
  sym_parameter_list = 26,
  sym__type = 27,
  sym_block = 28,
  sym_brace_block = 29,
  sym_do_block = 30,
  sym__statement = 31,
  sym_return_statement = 32,
  sym__expression = 33,
  sym_literal = 34,
  sym_string = 35,
  aux_sym_source_file_repeat1 = 36,
  aux_sym_brace_block_repeat1 = 37,
};

static const char * const ts_symbol_names[] = {
  [ts_builtin_sym_end] = "end",
  [anon_sym_EQ] = "=",
  [anon_sym_fn] = "fn",
  [anon_sym_LPAREN] = "(",
  [anon_sym_RPAREN] = ")",
  [anon_sym_bool] = "bool",
  [anon_sym_string] = "string",
  [anon_sym_int] = "int",
  [anon_sym_i32] = "i32",
  [anon_sym_i64] = "i64",
  [anon_sym_i128] = "i128",
  [anon_sym_float] = "float",
  [anon_sym_f32] = "f32",
  [anon_sym_f64] = "f64",
  [anon_sym_LBRACE] = "{",
  [anon_sym_RBRACE] = "}",
  [anon_sym_do] = "do",
  [anon_sym_end] = "end",
  [anon_sym_return] = "return",
  [sym_identifier] = "identifier",
  [sym_number] = "number",
  [sym_source_file] = "source_file",
  [sym__definition] = "_definition",
  [sym_assignment] = "assignment",
  [sym_simple_assignment] = "simple_assignment",
  [sym_function_definition] = "function_definition",
  [sym_parameter_list] = "parameter_list",
  [sym__type] = "_type",
  [sym_block] = "block",
  [sym_brace_block] = "brace_block",
  [sym_do_block] = "do_block",
  [sym__statement] = "_statement",
  [sym_return_statement] = "return_statement",
  [sym__expression] = "_expression",
  [sym_literal] = "literal",
  [sym_string] = "string",
  [aux_sym_source_file_repeat1] = "source_file_repeat1",
  [aux_sym_brace_block_repeat1] = "brace_block_repeat1",
};

static const TSSymbol ts_symbol_map[] = {
  [ts_builtin_sym_end] = ts_builtin_sym_end,
  [anon_sym_EQ] = anon_sym_EQ,
  [anon_sym_fn] = anon_sym_fn,
  [anon_sym_LPAREN] = anon_sym_LPAREN,
  [anon_sym_RPAREN] = anon_sym_RPAREN,
  [anon_sym_bool] = anon_sym_bool,
  [anon_sym_string] = anon_sym_string,
  [anon_sym_int] = anon_sym_int,
  [anon_sym_i32] = anon_sym_i32,
  [anon_sym_i64] = anon_sym_i64,
  [anon_sym_i128] = anon_sym_i128,
  [anon_sym_float] = anon_sym_float,
  [anon_sym_f32] = anon_sym_f32,
  [anon_sym_f64] = anon_sym_f64,
  [anon_sym_LBRACE] = anon_sym_LBRACE,
  [anon_sym_RBRACE] = anon_sym_RBRACE,
  [anon_sym_do] = anon_sym_do,
  [anon_sym_end] = anon_sym_end,
  [anon_sym_return] = anon_sym_return,
  [sym_identifier] = sym_identifier,
  [sym_number] = sym_number,
  [sym_source_file] = sym_source_file,
  [sym__definition] = sym__definition,
  [sym_assignment] = sym_assignment,
  [sym_simple_assignment] = sym_simple_assignment,
  [sym_function_definition] = sym_function_definition,
  [sym_parameter_list] = sym_parameter_list,
  [sym__type] = sym__type,
  [sym_block] = sym_block,
  [sym_brace_block] = sym_brace_block,
  [sym_do_block] = sym_do_block,
  [sym__statement] = sym__statement,
  [sym_return_statement] = sym_return_statement,
  [sym__expression] = sym__expression,
  [sym_literal] = sym_literal,
  [sym_string] = sym_string,
  [aux_sym_source_file_repeat1] = aux_sym_source_file_repeat1,
  [aux_sym_brace_block_repeat1] = aux_sym_brace_block_repeat1,
};

static const TSSymbolMetadata ts_symbol_metadata[] = {
  [ts_builtin_sym_end] = {
    .visible = false,
    .named = true,
  },
  [anon_sym_EQ] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_fn] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_LPAREN] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_RPAREN] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_bool] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_string] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_int] = {
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
  [anon_sym_float] = {
    .visible = true,
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
  [anon_sym_LBRACE] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_RBRACE] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_do] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_end] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_return] = {
    .visible = true,
    .named = false,
  },
  [sym_identifier] = {
    .visible = true,
    .named = true,
  },
  [sym_number] = {
    .visible = true,
    .named = true,
  },
  [sym_source_file] = {
    .visible = true,
    .named = true,
  },
  [sym__definition] = {
    .visible = false,
    .named = true,
  },
  [sym_assignment] = {
    .visible = true,
    .named = true,
  },
  [sym_simple_assignment] = {
    .visible = true,
    .named = true,
  },
  [sym_function_definition] = {
    .visible = true,
    .named = true,
  },
  [sym_parameter_list] = {
    .visible = true,
    .named = true,
  },
  [sym__type] = {
    .visible = false,
    .named = true,
  },
  [sym_block] = {
    .visible = true,
    .named = true,
  },
  [sym_brace_block] = {
    .visible = true,
    .named = true,
  },
  [sym_do_block] = {
    .visible = true,
    .named = true,
  },
  [sym__statement] = {
    .visible = false,
    .named = true,
  },
  [sym_return_statement] = {
    .visible = true,
    .named = true,
  },
  [sym__expression] = {
    .visible = false,
    .named = true,
  },
  [sym_literal] = {
    .visible = true,
    .named = true,
  },
  [sym_string] = {
    .visible = true,
    .named = true,
  },
  [aux_sym_source_file_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_brace_block_repeat1] = {
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
  [24] = 15,
  [25] = 25,
  [26] = 26,
  [27] = 27,
  [28] = 28,
  [29] = 29,
};

static bool ts_lex(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (eof) ADVANCE(32);
      ADVANCE_MAP(
        '(', 36,
        ')', 37,
        '=', 33,
        'b', 22,
        'd', 19,
        'e', 16,
        'f', 6,
        'i', 2,
        'r', 12,
        's', 25,
        '{', 47,
        '}', 48,
      );
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(0);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(55);
      END_STATE();
    case 1:
      if (lookahead == '0') ADVANCE(55);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(1);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(54);
      if (('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(53);
      END_STATE();
    case 2:
      if (lookahead == '1') ADVANCE(4);
      if (lookahead == '3') ADVANCE(5);
      if (lookahead == '6') ADVANCE(8);
      if (lookahead == 'n') ADVANCE(26);
      END_STATE();
    case 3:
      if (lookahead == '2') ADVANCE(45);
      END_STATE();
    case 4:
      if (lookahead == '2') ADVANCE(9);
      END_STATE();
    case 5:
      if (lookahead == '2') ADVANCE(41);
      END_STATE();
    case 6:
      if (lookahead == '3') ADVANCE(3);
      if (lookahead == '6') ADVANCE(7);
      if (lookahead == 'l') ADVANCE(21);
      if (lookahead == 'n') ADVANCE(34);
      END_STATE();
    case 7:
      if (lookahead == '4') ADVANCE(46);
      END_STATE();
    case 8:
      if (lookahead == '4') ADVANCE(42);
      END_STATE();
    case 9:
      if (lookahead == '8') ADVANCE(43);
      END_STATE();
    case 10:
      if (lookahead == 'a') ADVANCE(28);
      END_STATE();
    case 11:
      if (lookahead == 'd') ADVANCE(50);
      END_STATE();
    case 12:
      if (lookahead == 'e') ADVANCE(27);
      END_STATE();
    case 13:
      if (lookahead == 'g') ADVANCE(39);
      END_STATE();
    case 14:
      if (lookahead == 'i') ADVANCE(17);
      END_STATE();
    case 15:
      if (lookahead == 'l') ADVANCE(38);
      END_STATE();
    case 16:
      if (lookahead == 'n') ADVANCE(11);
      END_STATE();
    case 17:
      if (lookahead == 'n') ADVANCE(13);
      END_STATE();
    case 18:
      if (lookahead == 'n') ADVANCE(51);
      END_STATE();
    case 19:
      if (lookahead == 'o') ADVANCE(49);
      END_STATE();
    case 20:
      if (lookahead == 'o') ADVANCE(15);
      END_STATE();
    case 21:
      if (lookahead == 'o') ADVANCE(10);
      END_STATE();
    case 22:
      if (lookahead == 'o') ADVANCE(20);
      END_STATE();
    case 23:
      if (lookahead == 'r') ADVANCE(14);
      END_STATE();
    case 24:
      if (lookahead == 'r') ADVANCE(18);
      END_STATE();
    case 25:
      if (lookahead == 't') ADVANCE(23);
      END_STATE();
    case 26:
      if (lookahead == 't') ADVANCE(40);
      END_STATE();
    case 27:
      if (lookahead == 't') ADVANCE(29);
      END_STATE();
    case 28:
      if (lookahead == 't') ADVANCE(44);
      END_STATE();
    case 29:
      if (lookahead == 'u') ADVANCE(24);
      END_STATE();
    case 30:
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(30);
      if (('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(53);
      END_STATE();
    case 31:
      if (eof) ADVANCE(32);
      if (lookahead == 'f') ADVANCE(52);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(31);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(30);
      if (('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(53);
      END_STATE();
    case 32:
      ACCEPT_TOKEN(ts_builtin_sym_end);
      END_STATE();
    case 33:
      ACCEPT_TOKEN(anon_sym_EQ);
      END_STATE();
    case 34:
      ACCEPT_TOKEN(anon_sym_fn);
      END_STATE();
    case 35:
      ACCEPT_TOKEN(anon_sym_fn);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(53);
      if (('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(53);
      END_STATE();
    case 36:
      ACCEPT_TOKEN(anon_sym_LPAREN);
      END_STATE();
    case 37:
      ACCEPT_TOKEN(anon_sym_RPAREN);
      END_STATE();
    case 38:
      ACCEPT_TOKEN(anon_sym_bool);
      END_STATE();
    case 39:
      ACCEPT_TOKEN(anon_sym_string);
      END_STATE();
    case 40:
      ACCEPT_TOKEN(anon_sym_int);
      END_STATE();
    case 41:
      ACCEPT_TOKEN(anon_sym_i32);
      END_STATE();
    case 42:
      ACCEPT_TOKEN(anon_sym_i64);
      END_STATE();
    case 43:
      ACCEPT_TOKEN(anon_sym_i128);
      END_STATE();
    case 44:
      ACCEPT_TOKEN(anon_sym_float);
      END_STATE();
    case 45:
      ACCEPT_TOKEN(anon_sym_f32);
      END_STATE();
    case 46:
      ACCEPT_TOKEN(anon_sym_f64);
      END_STATE();
    case 47:
      ACCEPT_TOKEN(anon_sym_LBRACE);
      END_STATE();
    case 48:
      ACCEPT_TOKEN(anon_sym_RBRACE);
      END_STATE();
    case 49:
      ACCEPT_TOKEN(anon_sym_do);
      END_STATE();
    case 50:
      ACCEPT_TOKEN(anon_sym_end);
      END_STATE();
    case 51:
      ACCEPT_TOKEN(anon_sym_return);
      END_STATE();
    case 52:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'n') ADVANCE(35);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(53);
      if (('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(53);
      END_STATE();
    case 53:
      ACCEPT_TOKEN(sym_identifier);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(53);
      if (('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(53);
      END_STATE();
    case 54:
      ACCEPT_TOKEN(sym_number);
      if (lookahead == '0') ADVANCE(55);
      if (('1' <= lookahead && lookahead <= '9')) ADVANCE(54);
      if (('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(53);
      END_STATE();
    case 55:
      ACCEPT_TOKEN(sym_number);
      if (('0' <= lookahead && lookahead <= '9')) ADVANCE(55);
      END_STATE();
    default:
      return false;
  }
}

static const TSLexMode ts_lex_modes[STATE_COUNT] = {
  [0] = {.lex_state = 0},
  [1] = {.lex_state = 31},
  [2] = {.lex_state = 0},
  [3] = {.lex_state = 0},
  [4] = {.lex_state = 31},
  [5] = {.lex_state = 31},
  [6] = {.lex_state = 0},
  [7] = {.lex_state = 1},
  [8] = {.lex_state = 0},
  [9] = {.lex_state = 0},
  [10] = {.lex_state = 0},
  [11] = {.lex_state = 1},
  [12] = {.lex_state = 0},
  [13] = {.lex_state = 0},
  [14] = {.lex_state = 31},
  [15] = {.lex_state = 31},
  [16] = {.lex_state = 31},
  [17] = {.lex_state = 31},
  [18] = {.lex_state = 31},
  [19] = {.lex_state = 31},
  [20] = {.lex_state = 31},
  [21] = {.lex_state = 0},
  [22] = {.lex_state = 31},
  [23] = {.lex_state = 31},
  [24] = {.lex_state = 0},
  [25] = {.lex_state = 0},
  [26] = {.lex_state = 0},
  [27] = {.lex_state = 1},
  [28] = {.lex_state = 0},
  [29] = {.lex_state = 0},
};

static const uint16_t ts_parse_table[LARGE_STATE_COUNT][SYMBOL_COUNT] = {
  [0] = {
    [ts_builtin_sym_end] = ACTIONS(1),
    [anon_sym_EQ] = ACTIONS(1),
    [anon_sym_fn] = ACTIONS(1),
    [anon_sym_LPAREN] = ACTIONS(1),
    [anon_sym_RPAREN] = ACTIONS(1),
    [anon_sym_bool] = ACTIONS(1),
    [anon_sym_string] = ACTIONS(1),
    [anon_sym_int] = ACTIONS(1),
    [anon_sym_i32] = ACTIONS(1),
    [anon_sym_i64] = ACTIONS(1),
    [anon_sym_i128] = ACTIONS(1),
    [anon_sym_float] = ACTIONS(1),
    [anon_sym_f32] = ACTIONS(1),
    [anon_sym_f64] = ACTIONS(1),
    [anon_sym_LBRACE] = ACTIONS(1),
    [anon_sym_RBRACE] = ACTIONS(1),
    [anon_sym_do] = ACTIONS(1),
    [anon_sym_end] = ACTIONS(1),
    [anon_sym_return] = ACTIONS(1),
    [sym_number] = ACTIONS(1),
  },
  [1] = {
    [sym_source_file] = STATE(26),
    [sym__definition] = STATE(5),
    [sym_assignment] = STATE(5),
    [sym_simple_assignment] = STATE(17),
    [sym_function_definition] = STATE(5),
    [aux_sym_source_file_repeat1] = STATE(5),
    [ts_builtin_sym_end] = ACTIONS(3),
    [anon_sym_fn] = ACTIONS(5),
    [sym_identifier] = ACTIONS(7),
  },
};

static const uint16_t ts_small_parse_table[] = {
  [0] = 2,
    STATE(13), 1,
      sym__type,
    ACTIONS(9), 9,
      anon_sym_bool,
      anon_sym_string,
      anon_sym_int,
      anon_sym_i32,
      anon_sym_i64,
      anon_sym_i128,
      anon_sym_float,
      anon_sym_f32,
      anon_sym_f64,
  [15] = 1,
    ACTIONS(11), 9,
      anon_sym_bool,
      anon_sym_string,
      anon_sym_int,
      anon_sym_i32,
      anon_sym_i64,
      anon_sym_i128,
      anon_sym_float,
      anon_sym_f32,
      anon_sym_f64,
  [27] = 5,
    ACTIONS(13), 1,
      ts_builtin_sym_end,
    ACTIONS(15), 1,
      anon_sym_fn,
    ACTIONS(18), 1,
      sym_identifier,
    STATE(17), 1,
      sym_simple_assignment,
    STATE(4), 4,
      sym__definition,
      sym_assignment,
      sym_function_definition,
      aux_sym_source_file_repeat1,
  [46] = 5,
    ACTIONS(5), 1,
      anon_sym_fn,
    ACTIONS(7), 1,
      sym_identifier,
    ACTIONS(21), 1,
      ts_builtin_sym_end,
    STATE(17), 1,
      sym_simple_assignment,
    STATE(4), 4,
      sym__definition,
      sym_assignment,
      sym_function_definition,
      aux_sym_source_file_repeat1,
  [65] = 3,
    ACTIONS(25), 1,
      anon_sym_return,
    ACTIONS(23), 2,
      anon_sym_RBRACE,
      anon_sym_end,
    STATE(6), 3,
      sym__statement,
      sym_return_statement,
      aux_sym_brace_block_repeat1,
  [78] = 4,
    ACTIONS(28), 1,
      sym_identifier,
    ACTIONS(30), 1,
      sym_number,
    STATE(24), 1,
      sym_string,
    STATE(21), 2,
      sym__expression,
      sym_literal,
  [92] = 3,
    ACTIONS(32), 1,
      anon_sym_RBRACE,
    ACTIONS(34), 1,
      anon_sym_return,
    STATE(10), 3,
      sym__statement,
      sym_return_statement,
      aux_sym_brace_block_repeat1,
  [104] = 3,
    ACTIONS(34), 1,
      anon_sym_return,
    ACTIONS(36), 1,
      anon_sym_end,
    STATE(6), 3,
      sym__statement,
      sym_return_statement,
      aux_sym_brace_block_repeat1,
  [116] = 3,
    ACTIONS(34), 1,
      anon_sym_return,
    ACTIONS(38), 1,
      anon_sym_RBRACE,
    STATE(6), 3,
      sym__statement,
      sym_return_statement,
      aux_sym_brace_block_repeat1,
  [128] = 4,
    ACTIONS(40), 1,
      sym_identifier,
    ACTIONS(42), 1,
      sym_number,
    STATE(15), 1,
      sym_string,
    STATE(16), 2,
      sym__expression,
      sym_literal,
  [142] = 3,
    ACTIONS(34), 1,
      anon_sym_return,
    ACTIONS(44), 1,
      anon_sym_end,
    STATE(9), 3,
      sym__statement,
      sym_return_statement,
      aux_sym_brace_block_repeat1,
  [154] = 4,
    ACTIONS(46), 1,
      anon_sym_LBRACE,
    ACTIONS(48), 1,
      anon_sym_do,
    STATE(14), 1,
      sym_block,
    STATE(18), 2,
      sym_brace_block,
      sym_do_block,
  [168] = 2,
    ACTIONS(50), 1,
      ts_builtin_sym_end,
    ACTIONS(52), 2,
      anon_sym_fn,
      sym_identifier,
  [176] = 2,
    ACTIONS(54), 1,
      ts_builtin_sym_end,
    ACTIONS(56), 2,
      anon_sym_fn,
      sym_identifier,
  [184] = 2,
    ACTIONS(58), 1,
      ts_builtin_sym_end,
    ACTIONS(60), 2,
      anon_sym_fn,
      sym_identifier,
  [192] = 2,
    ACTIONS(62), 1,
      ts_builtin_sym_end,
    ACTIONS(64), 2,
      anon_sym_fn,
      sym_identifier,
  [200] = 2,
    ACTIONS(66), 1,
      ts_builtin_sym_end,
    ACTIONS(68), 2,
      anon_sym_fn,
      sym_identifier,
  [208] = 2,
    ACTIONS(70), 1,
      ts_builtin_sym_end,
    ACTIONS(72), 2,
      anon_sym_fn,
      sym_identifier,
  [216] = 2,
    ACTIONS(74), 1,
      ts_builtin_sym_end,
    ACTIONS(76), 2,
      anon_sym_fn,
      sym_identifier,
  [224] = 1,
    ACTIONS(78), 3,
      anon_sym_RBRACE,
      anon_sym_end,
      anon_sym_return,
  [230] = 2,
    ACTIONS(80), 1,
      ts_builtin_sym_end,
    ACTIONS(82), 2,
      anon_sym_fn,
      sym_identifier,
  [238] = 2,
    ACTIONS(84), 1,
      ts_builtin_sym_end,
    ACTIONS(86), 2,
      anon_sym_fn,
      sym_identifier,
  [246] = 1,
    ACTIONS(54), 3,
      anon_sym_RBRACE,
      anon_sym_end,
      anon_sym_return,
  [252] = 2,
    ACTIONS(88), 1,
      anon_sym_LPAREN,
    STATE(2), 1,
      sym_parameter_list,
  [259] = 1,
    ACTIONS(90), 1,
      ts_builtin_sym_end,
  [263] = 1,
    ACTIONS(92), 1,
      sym_identifier,
  [267] = 1,
    ACTIONS(94), 1,
      anon_sym_RPAREN,
  [271] = 1,
    ACTIONS(96), 1,
      anon_sym_EQ,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(2)] = 0,
  [SMALL_STATE(3)] = 15,
  [SMALL_STATE(4)] = 27,
  [SMALL_STATE(5)] = 46,
  [SMALL_STATE(6)] = 65,
  [SMALL_STATE(7)] = 78,
  [SMALL_STATE(8)] = 92,
  [SMALL_STATE(9)] = 104,
  [SMALL_STATE(10)] = 116,
  [SMALL_STATE(11)] = 128,
  [SMALL_STATE(12)] = 142,
  [SMALL_STATE(13)] = 154,
  [SMALL_STATE(14)] = 168,
  [SMALL_STATE(15)] = 176,
  [SMALL_STATE(16)] = 184,
  [SMALL_STATE(17)] = 192,
  [SMALL_STATE(18)] = 200,
  [SMALL_STATE(19)] = 208,
  [SMALL_STATE(20)] = 216,
  [SMALL_STATE(21)] = 224,
  [SMALL_STATE(22)] = 230,
  [SMALL_STATE(23)] = 238,
  [SMALL_STATE(24)] = 246,
  [SMALL_STATE(25)] = 252,
  [SMALL_STATE(26)] = 259,
  [SMALL_STATE(27)] = 263,
  [SMALL_STATE(28)] = 267,
  [SMALL_STATE(29)] = 271,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 0, 0, 0),
  [5] = {.entry = {.count = 1, .reusable = false}}, SHIFT(27),
  [7] = {.entry = {.count = 1, .reusable = false}}, SHIFT(29),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT(13),
  [11] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_parameter_list, 2, 0, 0),
  [13] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [15] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(27),
  [18] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(29),
  [21] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [23] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_brace_block_repeat1, 2, 0, 0),
  [25] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_brace_block_repeat1, 2, 0, 0), SHIFT_REPEAT(7),
  [28] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [30] = {.entry = {.count = 1, .reusable = false}}, SHIFT(24),
  [32] = {.entry = {.count = 1, .reusable = true}}, SHIFT(19),
  [34] = {.entry = {.count = 1, .reusable = true}}, SHIFT(7),
  [36] = {.entry = {.count = 1, .reusable = true}}, SHIFT(23),
  [38] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
  [40] = {.entry = {.count = 1, .reusable = true}}, SHIFT(16),
  [42] = {.entry = {.count = 1, .reusable = false}}, SHIFT(15),
  [44] = {.entry = {.count = 1, .reusable = true}}, SHIFT(20),
  [46] = {.entry = {.count = 1, .reusable = true}}, SHIFT(8),
  [48] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [50] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_function_definition, 5, 0, 0),
  [52] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_function_definition, 5, 0, 0),
  [54] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_literal, 1, 0, 0),
  [56] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_literal, 1, 0, 0),
  [58] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_simple_assignment, 3, 0, 0),
  [60] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_simple_assignment, 3, 0, 0),
  [62] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_assignment, 1, 0, 0),
  [64] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_assignment, 1, 0, 0),
  [66] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_block, 1, 0, 0),
  [68] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_block, 1, 0, 0),
  [70] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_brace_block, 2, 0, 0),
  [72] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_brace_block, 2, 0, 0),
  [74] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_do_block, 2, 0, 0),
  [76] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_do_block, 2, 0, 0),
  [78] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_return_statement, 2, 0, 0),
  [80] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_brace_block, 3, 0, 0),
  [82] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_brace_block, 3, 0, 0),
  [84] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_do_block, 3, 0, 0),
  [86] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_do_block, 3, 0, 0),
  [88] = {.entry = {.count = 1, .reusable = true}}, SHIFT(28),
  [90] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [92] = {.entry = {.count = 1, .reusable = true}}, SHIFT(25),
  [94] = {.entry = {.count = 1, .reusable = true}}, SHIFT(3),
  [96] = {.entry = {.count = 1, .reusable = true}}, SHIFT(11),
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
