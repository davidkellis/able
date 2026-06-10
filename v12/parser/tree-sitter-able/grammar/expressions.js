/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

const { ASSIGN_OPERATORS, PREC, sep1, lineOp, lineKeyword } = require("./helpers");

module.exports = {
  expression: $ => $.low_precedence_pipe_expression,

  low_precedence_pipe_expression: $ => prec.left(
    PREC.low_pipe,
    seq(
      $.assignment_expression,
      repeat(seq(
        choice("|>>", lineOp("|>>")),
        optional($._line_breaks),
        $.assignment_expression,
      )),
    ),
  ),

  pipe_expression: $ => prec.left(
    PREC.pipe,
    seq(
      $.matchable_expression,
      repeat(seq(
        choice("|>", lineOp("|>")),
        optional($._line_breaks),
        $.matchable_expression,
      )),
    ),
  ),

  matchable_expression: $ => choice(
    $.ensure_expression,
    $.rescue_expression,
    $.handling_expression,
    $.pipe_operand_base,
  ),

  pipe_operand_base: $ => choice(
    $.spawn_expression,
    $.await_expression,
    $.breakpoint_expression,
    $.match_expression,
    $.range_expression,
  ),

  ensure_expression: $ => seq(
    choice($.rescue_expression, $.handling_expression, $.pipe_operand_base),
    choice("ensure", lineKeyword("ensure")),
    optional($._line_breaks),
    field("ensure", $.block),
  ),

  rescue_expression: $ => seq(
    choice($.handling_expression, $.pipe_operand_base),
    choice("rescue", lineKeyword("rescue")),
    optional($._line_breaks),
    field("rescue", $.rescue_block),
  ),

  handling_expression: $ => prec.left(seq(
    choice($.rescue_postfix_expression, $.pipe_operand_base),
    repeat1($.or_handler_clause),
  )),

  or_handler_clause: $ => seq(
    choice("or", lineKeyword("or")),
    optional($._line_breaks),
    field("handler", $.handling_block),
  ),

  handling_block: $ => seq(
    "{",
    repeat($._statement_sep),
    optional(seq(
      field("binding", $.identifier),
      "=>",
      repeat($._statement_sep),
    )),
    repeat(seq($.statement, repeat1($._statement_sep))),
    optional($.statement),
    "}",
  ),

  match_expression: $ => seq(
    field("subject", alias($._postfix_expression_without_match, $.postfix_expression)),
    "match",
    "{",
    repeat($._statement_sep),
    repeat1(seq($.match_clause, repeat($._statement_sep))),
    "}",
  ),

  match_clause: $ => seq(
    "case",
    field("pattern", $.pattern),
    optional(field("guard", $.match_guard)),
    "=>",
    field("body", choice($.block, $.expression)),
    optional(","),
  ),

  match_guard: $ => seq("if", $.expression),

  rescue_block: $ => seq(
    "{",
    repeat($._statement_sep),
    repeat1(seq($.match_clause, repeat($._statement_sep))),
    "}",
  ),

  spawn_expression: $ => seq(
    "spawn",
    choice(
      $.block,
      $.do_expression,
      $.call_target,
    ),
  ),

  await_expression: $ => seq(
    "await",
    $.postfix_expression,
  ),

  breakpoint_expression: $ => seq(
    "breakpoint",
    optional(field("label", $.label)),
    $.block,
  ),

  label: _ => token(prec(-1, seq("'", /[A-Za-z_][A-Za-z0-9_]*/))),

  assignment_expression: $ => choice(
    prec.right(
      PREC.assignment,
      seq(
        field("left", $.assignment_target),
        field("operator", $.assignment_operator),
        optional($._line_breaks),
        field("right", $.assignment_expression),
      ),
    ),
    $.pipe_expression,
  ),

  assignment_target: $ => choice(
    $.pattern,
    $.postfix_expression,
  ),

  assignment_operator: _ => choice(...ASSIGN_OPERATORS),

  range_expression: $ => choice(
    prec.left(
      PREC.range,
      seq(
        $.logical_or_expression,
        field("operator", choice("..", "...", lineOp(".."), lineOp("..."))),
        optional($._line_breaks),
        $.logical_or_expression,
      ),
    ),
    $.logical_or_expression,
  ),

  logical_or_expression: $ => prec.left(
    PREC.logical_or,
    seq(
      $.logical_and_expression,
      repeat(seq(
        choice("||", lineOp("||")),
        optional($._line_breaks),
        $.logical_and_expression,
      )),
    ),
  ),

  logical_and_expression: $ => prec.left(
    PREC.logical_and,
    seq(
      $.bitwise_or_expression,
      repeat(seq(
        choice("&&", lineOp("&&")),
        optional($._line_breaks),
        $.bitwise_or_expression,
      )),
    ),
  ),

  bitwise_or_expression: $ => prec.left(
    PREC.bit_or,
    seq(
      $.bitwise_xor_expression,
      repeat(seq(
        choice(".|", lineOp(".|")),
        optional($._line_breaks),
        $.bitwise_xor_expression,
      )),
    ),
  ),

  bitwise_xor_expression: $ => prec.left(
    PREC.bit_xor,
    seq(
      $.bitwise_and_expression,
      repeat(seq(
        choice(".^", lineOp(".^")),
        optional($._line_breaks),
        $.bitwise_and_expression,
      )),
    ),
  ),

  bitwise_and_expression: $ => prec.left(
    PREC.bit_and,
    seq(
      $.equality_expression,
      repeat(seq(
        choice(".&", lineOp(".&")),
        optional($._line_breaks),
        $.equality_expression,
      )),
    ),
  ),

  equality_expression: $ => prec.left(
    PREC.equality,
    seq(
      $.comparison_expression,
      repeat(seq(
        choice(
          "==",
          "!=",
          lineOp("=="),
          lineOp("!="),
        ),
        optional($._line_breaks),
        $.comparison_expression,
      )),
    ),
  ),

  comparison_expression: $ => prec.left(
    PREC.comparison,
    seq(
      $.shift_expression,
      repeat(seq(
        choice(
          ">",
          "<",
          ">=",
          "<=",
          lineOp(">"),
          lineOp("<"),
          lineOp(">="),
          lineOp("<="),
        ),
        optional($._line_breaks),
        $.shift_expression,
      )),
    ),
  ),

  shift_expression: $ => prec.left(
    PREC.shift,
    seq(
      $.additive_expression,
      repeat(seq(
        choice(
          ".<<",
          ".>>",
          lineOp(".<<"),
          lineOp(".>>"),
        ),
        optional($._line_breaks),
        $.additive_expression,
      )),
    ),
  ),

  additive_expression: $ => prec.left(
    PREC.additive,
    seq(
      $.multiplicative_expression,
      repeat(seq(
        choice(
          "+",
          "-",
          lineOp("+"),
          lineOp("-"),
        ),
        optional($._line_breaks),
        $.multiplicative_expression,
      )),
    ),
  ),

  multiplicative_expression: $ => prec.left(
    PREC.multiplicative,
    seq(
      choice($.cast_expression, $.unary_expression),
      repeat(seq(
        choice(
          "//",
          "%",
          "/%",
          "*",
          "/",
          lineOp("//"),
          lineOp("%"),
          lineOp("/%"),
          lineOp("*"),
          lineOp("/"),
        ),
        optional($._line_breaks),
        choice($.cast_expression, $.unary_expression),
      )),
    ),
  ),

  unary_expression: $ => choice(
    prec.right(
      PREC.unary,
      seq(
        choice("-", "!", ".~"),
        optional($._line_breaks),
        $.unary_expression,
      ),
    ),
    $.exponent_expression,
  ),

  cast_expression: $ => prec.left(
    PREC.cast,
    seq(
      $.unary_expression,
      repeat1(seq("as", optional($._line_breaks), $.type_expression)),
    ),
  ),

  exponent_expression: $ => choice(
    prec.right(
      PREC.exponent,
      seq(
        $.postfix_expression,
        choice("^", lineOp("^")),
        optional($._line_breaks),
        $.exponent_expression,
      ),
    ),
    $.postfix_expression,
  ),

  postfix_expression: $ => prec.left(
      PREC.call,
      seq(
        choice(
          $.primary_expression,
          $.spawn_expression,
          $.await_expression,
          $.breakpoint_expression,
          $.match_expression,
        ),
      repeat(choice(
        $.type_arguments,
        $.call_suffix,
        $.index_suffix,
        $.propagate_suffix,
        $.member_access,
      )),
      optional($.lambda_expression),
    ),
  ),

  _postfix_expression_without_match: $ => prec.left(
    PREC.call,
    seq(
      choice(
        $.primary_expression,
        $.spawn_expression,
        $.breakpoint_expression,
      ),
      repeat(choice(
        $.type_arguments,
        $.call_suffix,
        $.index_suffix,
        $.propagate_suffix,
        $.member_access,
      )),
      optional($.lambda_expression),
    ),
  ),

  rescue_postfix_expression: $ => prec.left(
    PREC.call,
    seq(
      $.rescue_expression,
      repeat(choice(
        $.type_arguments,
        $.call_suffix,
        $.index_suffix,
        $.propagate_suffix,
        $.member_access,
      )),
      optional($.lambda_expression),
    ),
  ),

  call_target: $ => prec.left(
    PREC.call,
    seq(
      choice(
        $.primary_expression,
        $.spawn_expression,
        $.breakpoint_expression,
      ),
      repeat(choice(
        $.type_arguments,
        $.member_access,
        $.index_suffix,
        $.propagate_suffix,
      )),
      $.call_suffix,
      repeat(choice(
        $.type_arguments,
        $.call_suffix,
        $.index_suffix,
        $.propagate_suffix,
        $.member_access,
      )),
      optional($.lambda_expression),
    ),
  ),

  call_suffix: $ => prec.dynamic(-1, seq(
    token.immediate("("),
    choice(
      seq(
        optional($._line_breaks),
        ")",
      ),
      seq(
        optional($._line_breaks),
        sep1($, $.expression, $._comma_sep),
        optional($._comma_sep),
        optional($._line_breaks),
        ")",
      ),
    ),
  )),

  type_arguments: $ => prec.dynamic(1, seq(
    token.immediate("<"),
    optional($._line_breaks),
    sep1($, $.type_expression, $._comma_sep),
    optional($._line_breaks),
    ">",
  )),

  index_suffix: $ => seq(
    token.immediate("["),
    optional($._line_breaks),
    $.expression,
    optional($._line_breaks),
    optional(seq(
      ":",
      optional($._line_breaks),
      $.expression,
      optional($._line_breaks),
    )),
    "]",
  ),

  propagate_suffix: $ => token.immediate("!"),

  member_access: $ => prec.left(
    PREC.member,
    seq(
      field("operator", choice(".", "?.", lineOp("."), lineOp("?."))),
      field("member", choice($.identifier, $.keyword_identifier, $.numeric_member)),
    ),
  ),

  primary_expression: $ => choice(
    $.literal,
    $.identifier,
    $.placeholder_expression,
    $.implicit_member_expression,
    $.if_expression,
    $.loop_expression,
    $.do_expression,
    $.iterator_literal,
    $.verbose_lambda_expression,
    $.lambda_expression,
    $.parenthesized_expression,
  ),

  verbose_lambda_expression: $ => seq(
    "fn",
    field("type_parameters", optional($.type_parameter_list)),
    field("parameters", $.parameter_list),
    field("return_type", optional($.return_type)),
    field("where_clause", optional($.where_clause)),
    field("body", $.block),
  ),

  lambda_expression: $ => prec.right(
    PREC.lambda,
    choice(
      seq(
        "{",
        optional(field("parameters", $.lambda_parameter_list)),
        optional(seq(
          "->",
          field("return_type", $.type_expression),
        )),
        "=>",
        repeat1($._line_breaks),
        field("body", $.expression_list),
        repeat($._statement_sep),
        "}",
      ),
      seq(
        "{",
        optional(field("parameters", $.lambda_parameter_list)),
        optional(seq(
          "->",
          field("return_type", $.type_expression),
        )),
        "=>",
        repeat1($._line_breaks),
        field("body", $.block),
        optional($._line_breaks),
        "}",
      ),
      seq(
        "{",
        optional(field("parameters", $.lambda_parameter_list)),
        optional(seq(
          "->",
          field("return_type", $.type_expression),
        )),
        "=>",
        field("body", choice($.block, $.expression_list)),
        optional($._line_breaks),
        "}",
      ),
    ),
  ),

  expression_list: $ => prec.right(seq(
    $.expression_statement,
    repeat(seq(
      repeat1($._statement_sep),
      $.expression_statement,
    )),
    optional(repeat1($._statement_sep)),
  )),

  lambda_parameter_list: $ => seq(
    sep1($, $.lambda_parameter, $._comma_sep),
  ),

  lambda_parameter: $ => field("name", $.identifier),

  if_expression: $ => choice(
    $.if_expression_with_else,
    $.if_expression_without_else,
  ),

  if_expression_with_else: $ => prec.right(
    PREC.logical_or,
    seq(
      "if",
      field("condition", $.expression),
      field("consequence", $.block),
      repeat(field("elsif_clause", $.elsif_clause)),
      field("else_clause", $.else_clause),
    ),
  ),

  if_expression_without_else: $ => prec.right(
    PREC.logical_or - 1,
    seq(
      "if",
      field("condition", $.expression),
      field("consequence", $.block),
      repeat(field("elsif_clause", $.elsif_clause)),
    ),
  ),

  elsif_clause: $ => seq(
    "elsif",
    field("condition", $.expression),
    field("consequence", $.block),
  ),

  else_clause: $ => seq(
    "else",
    field("alternative", $.block),
  ),

  elsif_clause_statement: $ => seq(
    "elsif",
    field("condition", $.expression),
    field("consequence", $.block),
    optional(field("else_clause", $.else_clause)),
  ),

  else_clause_statement: $ => seq(
    "else",
    field("alternative", $.block),
  ),


  do_expression: $ => seq(
    "do",
    $.block,
  ),

  parenthesized_expression: $ => seq(
    "(",
    optional($._line_breaks),
    $.expression,
    optional($._line_breaks),
    ")",
  ),

  iterator_literal: $ => seq(
    "Iterator",
    optional(field("element_type", $.type_expression)),
    field("body", $.iterator_block),
  ),

  iterator_block: $ => seq(
    "{",
    repeat($._statement_sep),
    optional(seq(
      field("binding", $.identifier),
      "=>",
      repeat($._statement_sep),
    )),
    repeat(seq($.statement, repeat1($._statement_sep))),
    optional($.statement),
    "}",
  ),

};
