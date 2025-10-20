/**
 * @file Able language parser (v10 spec)
 * @author David Ellis <david@conquerthelawn.com>
 * @license epl-2.0
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

const PREC = {
  assignment: 1,
  logical_or: 2,
  logical_and: 3,
  equality: 4,
  comparison: 5,
  additive: 6,
  multiplicative: 7,
  unary: 8,
  call: 9,
  member: 10,
};

const commaSep = rule => optional(commaSep1(rule));
const commaSep1 = rule => seq(rule, repeat(seq(',', rule)));

module.exports = grammar({
  name: "able",

  extras: $ => [
    /[\s\uFEFF\u2060\u200B]/,
    $.comment,
  ],

  word: $ => $.identifier,

  supertypes: $ => [
    $.statement,
    $.expression,
    $.type_expression,
    $.pattern,
    $.literal,
  ],

  rules: {
    source_file: $ => seq(
      field("package", optional($.package_statement)),
      repeat(field("imports", $.import_statement)),
      repeat(field("statements", $.statement)),
    ),

    comment: _ => token(seq("##", /[^\n]*/)),

    package_statement: $ => seq(
      "package",
      $.identifier,
      repeat(seq(".", $.identifier)),
      optional(";"),
    ),

    import_statement: $ => seq(
      field("kind", choice("import", "dynimport")),
      field("path", $.qualified_identifier),
      optional(field("clause", $.import_clause)),
      optional(";"),
    ),

    import_clause: $ => choice(
      seq(".", "*"),
      seq(".", "{", commaSep1($.import_selector), optional(","), "}"),
      seq("as", $.identifier),
    ),

    import_selector: $ => seq(
      $.identifier,
      optional(seq("as", $.identifier)),
    ),

    statement: $ => choice(
      $.function_definition,
      $.struct_definition,
      $.expression_statement,
    ),

    function_definition: $ => seq(
      optional("private"),
      "fn",
      field("name", $.identifier),
      field("type_parameters", optional($.type_parameter_list)),
      field("parameters", $.parameter_list),
      field("return_type", optional($.return_type)),
      field("where_clause", optional($.where_clause)),
      field("body", $.block),
    ),

    type_parameter_list: $ => seq(
      "<",
      commaSep1($.type_parameter),
      optional(","),
      ">",
    ),

    type_parameter: $ => seq(
      $.identifier,
      optional(seq(":", commaSep1($.type_expression))),
    ),

    parameter_list: $ => seq(
      "(",
      commaSep($.parameter),
      ")",
    ),

    parameter: $ => seq(
      field("pattern", $.pattern),
      optional(seq(":", field("type", $.type_expression))),
    ),

    return_type: $ => seq("->", $.type_expression),

    where_clause: $ => seq(
      "where",
      commaSep1($.where_constraint),
    ),

    where_constraint: $ => seq(
      $.identifier,
      ":",
      commaSep1($.type_expression),
    ),

    block: $ => seq(
      "{",
      repeat($.statement),
      "}",
    ),

    expression_statement: $ => seq(
      field("expression", $.expression),
      optional(";"),
    ),

    struct_definition: $ => seq(
      optional("private"),
      "struct",
      field("name", $.identifier),
      field("type_parameters", optional($.type_parameter_list)),
      choice(
        field("record", $.struct_record),
        field("tuple", $.struct_tuple),
      ),
      optional($.where_clause),
    ),

    struct_record: $ => seq(
      "{",
      commaSep($.struct_field),
      optional(","),
      "}",
    ),

    struct_field: $ => seq(
      optional("private"),
      $.identifier,
      ":",
      $.type_expression,
    ),

    struct_tuple: $ => seq(
      "(",
      commaSep($.type_expression),
      optional(","),
      ")",
    ),

    expression: $ => choice(
      $.literal,
      $.identifier,
    ),

    literal: $ => choice(
      $.number_literal,
      $.string_literal,
      $.boolean_literal,
      $.nil_literal,
    ),

    pattern: $ => choice(
      $.identifier,
      "_",
    ),

    type_expression: $ => $.qualified_identifier,

    qualified_identifier: $ => seq(
      $.identifier,
      repeat(seq(".", $.identifier)),
    ),

    identifier: _ => token(prec(-1, /[A-Za-z_][A-Za-z0-9_]*/)),

    number_literal: _ => token(/[0-9][0-9_]*/),

    string_literal: _ => token(seq(
      '"',
      repeat(choice(/[^"\\\n]+/, /\\./)),
      '"',
    )),

    boolean_literal: _ => choice("true", "false"),

    nil_literal: _ => "nil",
  },
});
