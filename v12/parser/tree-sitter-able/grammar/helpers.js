/// <reference types="tree-sitter-cli/dsl" />
// @ts-check


const PREC = {
  lambda: 0,
  low_pipe: 1,
  pipe: 2,
  assignment: 3,
  range: 4,
  logical_or: 5,
  logical_and: 6,
  equality: 7,
  comparison: 8,
  bit_or: 9,
  bit_xor: 10,
  bit_and: 11,
  shift: 12,
  additive: 13,
  multiplicative: 14,
  cast: 15,
  unary: 16,
  exponent: 17,
  call: 18,
  member: 19,
  type_application: 20,
  return_stmt: 21,
};

const sep = ($, rule, separator) => optional(sep1($, rule, separator));
const sep1 = ($, rule, separator) => seq(
  rule,
  repeat(seq(separator, optional($._line_breaks), rule)),
);
const lineBreak = /(?:\r?\n[ \t]*)+/;
const lineOp = op => alias(token(prec(1, seq(lineBreak, op))), op);
const lineKeyword = kw => alias(token(prec(1, seq(lineBreak, kw, /[ \t\r\n]+/))), kw);
const renameOperator = token("::");

const KEYWORDS = [
  "fn",
  "struct",
  "union",
  "interface",
  "impl",
  "methods",
  "type",
  "package",
  "import",
  "dynimport",
  "extern",
  "prelude",
  "private",
  "Self",
  "do",
    "return",
    "if",
    "elsif",
    "or",
    "else",
  "while",
  "loop",
  "for",
  "in",
  "match",
  "case",
  "breakpoint",
  "break",
  "continue",
  "raise",
  "rescue",
  "ensure",
  "rethrow",
  "spawn",
  "await",
  "as",
  "nil",
  "void",
  "true",
  "false",
  "where",
  "Iterator",
];

const ASSIGN_OPERATORS = [
  ':=',
  '=',
  '+=',
  '-=',
  '*=',
  '/=',
  '%=',
  '.&=',
  '.|=',
  '.^=',
  '.<<=',
  '.>>=',
];

const DECIMAL_DIGITS = "[0-9](?:_?[0-9])*";
const BINARY_DIGITS = "[01](?:_?[01])*";
const OCTAL_DIGITS = "[0-7](?:_?[0-7])*";
const HEX_DIGITS = "[0-9a-fA-F](?:_?[0-9a-fA-F])*";
const EXPONENT_PART = "[eE][+-]?[0-9](?:_?[0-9])*";
const INTEGER_SUFFIX = "_(?:i|u)(?:8|16|32|64|128)";
const FLOAT_SUFFIX = "_(?:f32|f64)";
const INTEGER_BODY = `(?:${DECIMAL_DIGITS}|0[bB]${BINARY_DIGITS}|0[oO]${OCTAL_DIGITS}|0[xX]${HEX_DIGITS})`;
const INTEGER_LITERAL_PATTERN = `${INTEGER_BODY}(?:${INTEGER_SUFFIX})?`;
const DECIMAL_FLOAT = `${DECIMAL_DIGITS}\\.${DECIMAL_DIGITS}(?:${EXPONENT_PART})?`;
const DECIMAL_EXPONENT = `${DECIMAL_DIGITS}${EXPONENT_PART}`;
const FLOAT_CORE = `(?:${DECIMAL_FLOAT}|${DECIMAL_EXPONENT})`;
const FLOAT_LITERAL_PATTERN = `${FLOAT_CORE}(?:${FLOAT_SUFFIX})?`;
const FLOAT_SUFFIX_ONLY_PATTERN = `${DECIMAL_DIGITS}${FLOAT_SUFFIX}`;
const CHARACTER_LITERAL_PATTERN = "'(?:[^'\\\\\\n]|\\\\[nrt0'\"\\\\]|\\\\u\\{[0-9a-fA-F]{1,6}\\})'";
const INTERPOLATION_CHUNK_PATTERN = "(?:[^`\\\\$]+|\\\\[`$\\\\]|\\\\.)";
const ITERATOR_KEYWORD = token(seq("Iterator", /[A-Za-z_0-9]*/));

module.exports = {
  PREC,
  sep1,
  lineOp,
  lineKeyword,
  renameOperator,
  ASSIGN_OPERATORS,
  INTEGER_LITERAL_PATTERN,
  FLOAT_LITERAL_PATTERN,
  FLOAT_SUFFIX_ONLY_PATTERN,
  CHARACTER_LITERAL_PATTERN,
  INTERPOLATION_CHUNK_PATTERN,
};
