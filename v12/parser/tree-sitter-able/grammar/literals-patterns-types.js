/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

const {
  CHARACTER_LITERAL_PATTERN,
  FLOAT_LITERAL_PATTERN,
  FLOAT_SUFFIX_ONLY_PATTERN,
  INTEGER_LITERAL_PATTERN,
  INTERPOLATION_CHUNK_PATTERN,
  PREC,
  lineOp,
  renameOperator,
  sep1,
} = require("./helpers");

module.exports = {
  literal: $ => choice(
    $.number_literal,
    $.character_literal,
    $.string_literal,
    $.interpolated_string,
    $.struct_literal,
    $.boolean_literal,
    $.nil_literal,
    $.array_literal,
    $.map_literal,
  ),

  literal_pattern: $ => choice(
    $.number_literal,
    $.character_literal,
    $.string_literal,
    $.boolean_literal,
    $.nil_literal,
  ),

  array_literal: $ => seq(
    "[",
    choice(
      seq(
        optional($._line_breaks),
        "]",
      ),
      seq(
        optional($._line_breaks),
        sep1($, $.expression, $._comma_sep),
        optional($._comma_sep),
        optional($._line_breaks),
        "]",
      ),
    ),
  ),

  struct_literal: $ => prec.left(-1, seq(
    field("type", alias($.struct_type_suffix, $.type_suffix)),
    "{",
    choice(
      seq(
        optional($._line_breaks),
        "}",
      ),
      seq(
        optional($._line_breaks),
        sep1($, $.struct_literal_element, $._comma_sep),
        optional($._comma_sep),
        optional($._line_breaks),
        "}",
      ),
    ),
  )),

  struct_literal_element: $ => choice(
    $.struct_literal_spread,
    $.struct_literal_field,
    $.struct_literal_shorthand_field,
    $.expression,
  ),

  struct_literal_field: $ => seq(
    field("name", $.identifier),
    ":",
    field("value", $.expression),
  ),

  struct_literal_shorthand_field: $ => field("name", $.identifier),

  struct_literal_spread: $ => seq(
    choice("...", lineOp("...")),
    field("source", $.expression),
  ),

  map_literal: $ => seq(
    "#{",
    choice(
      seq(
        optional($._line_breaks),
        "}",
      ),
      seq(
        optional($._line_breaks),
        sep1($, $.map_literal_element, $._comma_sep),
        optional($._comma_sep),
        optional($._line_breaks),
        "}",
      ),
    ),
  ),

  map_literal_element: $ => choice(
    $.map_literal_entry,
    $.map_literal_spread,
  ),

  map_literal_entry: $ => seq(
    field("key", $.expression),
    ":",
    field("value", $.expression),
  ),

  map_literal_spread: $ => seq(
    choice("...", lineOp("...")),
    field("expression", $.expression),
  ),

  pattern: $ => choice(
    $.typed_pattern,
    $.pattern_base,
  ),

  typed_pattern: $ => prec.right(
    seq(
      $.pattern_base,
      ":",
      $.type_expression,
    ),
  ),

  pattern_base: $ => choice(
    $.identifier,
    "_",
    $.literal_pattern,
    $.struct_pattern,
    $.array_pattern,
    $.parenthesized_pattern,
  ),

  struct_pattern: $ => seq(
    optional(field("type", $.qualified_identifier)),
    "{",
    choice(
      seq(
        optional($._line_breaks),
        "}",
      ),
      seq(
        optional($._line_breaks),
        sep1($, $.struct_pattern_element, $._comma_sep),
        optional($._comma_sep),
        optional($._line_breaks),
        "}",
      ),
    ),
  ),

  struct_pattern_element: $ => choice(
    $.struct_pattern_field,
    $.pattern,
  ),

  struct_pattern_field: $ => seq(
    field("field", $.identifier),
    optional(seq(renameOperator, field("binding", $.identifier))),
    optional(seq(
      ":",
      field("type", $.type_expression),
      optional(field("value", $.pattern)),
    )),
  ),

  array_pattern: $ => seq(
    "[",
    choice(
      seq(
        optional($._line_breaks),
        "]",
      ),
      seq(
        optional($._line_breaks),
        $.array_pattern_rest,
        optional($._line_breaks),
        "]",
      ),
      seq(
        optional($._line_breaks),
        sep1($, $.pattern, $._comma_sep),
        optional(seq($._comma_sep, $.array_pattern_rest)),
        optional($._line_breaks),
        "]",
      ),
    ),
  ),

  array_pattern_rest: $ => seq(
    "...",
    optional($.identifier),
  ),

  parenthesized_pattern: $ => seq(
    "(",
    $.pattern,
    ")",
  ),

  interface_type_expression: $ => $.interface_type_union,

  interface_type_union: $ => choice(
    prec.left(seq(
      $.interface_type_arrow,
      repeat1(seq(
        "|",
        optional($._line_breaks),
        $.interface_type_arrow,
      )),
    )),
    $.interface_type_arrow,
  ),

  interface_type_arrow: $ => choice(
    prec.right(seq(
      $.interface_type_suffix,
      "->",
      optional($._line_breaks),
      $.interface_type_arrow,
    )),
    $.interface_type_suffix,
  ),

  interface_type_suffix: $ => prec.left(
    seq(
      $.interface_type_prefix,
      repeat($.type_arguments),
    ),
  ),

  interface_type_prefix: $ => choice(
    seq("?", $.interface_type_prefix),
    seq("!", $.interface_type_prefix),
    $.interface_type_atom,
  ),

  interface_type_atom: $ => choice(
    $.parenthesized_type,
    $.type_identifier,
    $.wildcard_type,
  ),

  type_expression: $ => $.type_union,

  type_union: $ => choice(
    prec.left(seq(
      $.type_arrow,
      repeat1(seq(
        "|",
        optional($._line_breaks),
        $.type_arrow,
      )),
    )),
    $.type_arrow,
  ),

  type_arrow: $ => choice(
    prec.right(seq(
      $.type_suffix,
      "->",
      optional($._line_breaks),
      $.type_arrow,
    )),
    $.type_suffix,
  ),

  type_suffix: $ => choice(
    prec.left(
      PREC.type_application,
      seq(
        $.type_prefix,
        repeat1(choice(
          $.type_arguments,
          seq($._type_application_sep, $.type_prefix),
          alias($.parenthesized_type_immediate, $.parenthesized_type),
        )),
      ),
    ),
    $.type_prefix,
  ),

  struct_type_suffix: $ => choice(
    prec.left(
      PREC.type_application,
      seq(
        $.type_prefix,
        repeat1(choice($.type_prefix, $.type_arguments)),
      ),
    ),
    $.type_prefix,
  ),

  type_prefix: $ => choice(
    seq("?", $.type_prefix),
    seq("!", $.type_prefix),
    $.type_atom,
  ),

  type_generic_application: $ => prec.left(
    PREC.type_application,
    seq(
      $.type_atom,
      repeat1($.type_atom),
    ),
  ),

  type_atom: $ => choice(
    $.parenthesized_type,
    $.type_identifier,
    $.wildcard_type,
  ),

  parenthesized_type: $ => seq(
    "(",
    choice(
      seq(
        optional($._line_breaks),
        ")",
      ),
      seq(
        optional($._line_breaks),
        sep1($, $.type_expression, $._comma_sep),
        optional($._line_breaks),
        ")",
      ),
    ),
  ),

  parenthesized_type_immediate: $ => seq(
    token.immediate("("),
    choice(
      seq(
        optional($._line_breaks),
        ")",
      ),
      seq(
        optional($._line_breaks),
        sep1($, $.type_expression, $._comma_sep),
        optional($._line_breaks),
        ")",
      ),
    ),
  ),

  wildcard_type: _ => "_",

  type_identifier: $ => choice(
    $.qualified_identifier,
    "Self",
    "nil",
    "void",
  ),

  qualified_identifier: $ => seq(
    $.identifier,
    repeat(seq(".", $.identifier)),
  ),

  placeholder_expression: _ => token(choice("@", /@[1-9][0-9]*/)),
  implicit_member_expression: $ => seq(
    "#",
    field("member", $.identifier),
  ),

  identifier: _ => token(prec(-1, /[A-KM-Za-km-zA-Z_][A-Za-z0-9_]*|l|lo|loo|l[A-Za-np-zA-Z0-9_][A-Za-z0-9_]*|lo[A-Za-np-zA-Z0-9_][A-Za-z0-9_]*|loo[A-Za-oq-zA-Z0-9_][A-Za-z0-9_]*|loop[A-Za-z0-9_]+/)),
  keyword_identifier: _ => "package",

  numeric_member: _ => token.immediate(/[0-9]+/),

  number_literal: _ => token(choice(
    new RegExp(FLOAT_LITERAL_PATTERN),
    new RegExp(FLOAT_SUFFIX_ONLY_PATTERN),
    new RegExp(INTEGER_LITERAL_PATTERN),
  )),

  character_literal: _ => token(prec(1, new RegExp(CHARACTER_LITERAL_PATTERN))),

  string_literal: _ => token(seq(
    '"',
    repeat(choice(/[^"\\]+/, /\\./)),
    '"',
  )),

  interpolated_string: $ => seq(
    "`",
    repeat(choice(
      $.interpolation_text,
      $.string_interpolation,
    )),
    "`",
  ),

  interpolation_text: _ => token(prec(1, new RegExp(`${INTERPOLATION_CHUNK_PATTERN}+`))),

  string_interpolation: $ => seq(
    token.immediate("${"),
    field("expression", $.expression),
    "}",
  ),

  boolean_literal: _ => choice("true", "false"),

  nil_literal: _ => "nil",
};
