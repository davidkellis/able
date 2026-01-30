/**
 * @file Able language parser (v12 spec)
 * @author David Ellis <david@conquerthelawn.com>
 * @license epl-2.0
 */

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
  unary: 15,
  exponent: 16,
  call: 17,
  member: 18,
  type_application: 19,
  return_stmt: 20,
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

module.exports = grammar({
  name: "able",

  extras: $ => [
    /[ \t\f\v\uFEFF\u2060\u200B]/,
    $.comment,
  ],

  externals: $ => [
    $._newline,
    $._type_application_sep,
  ],

  word: $ => $.identifier,

  supertypes: $ => [
    $.statement,
    $.expression,
    $.type_expression,
    $.pattern,
    $.literal,
  ],

  conflicts: $ => [
    [$.primary_expression, $.qualified_identifier],
    [$.primary_expression, $.pattern_base],
    [$.named_implementation_definition, $.primary_expression, $.pattern_base],
    [$.literal, $.literal_pattern],
    [$.pattern_base, $.struct_pattern_field],
    [$.primary_expression, $.pattern_base, $.struct_pattern_field],
    [$.primary_expression, $.pattern_base, $.qualified_identifier],
    [$.lambda_parameter, $.pattern_base, $.struct_pattern_field],
    [$.spawn_expression, $.primary_expression],
    [$.array_literal, $.array_pattern],
    [$.expression_statement, $.lambda_expression],
    [$.assignment_target, $.array_pattern],
    [$.assignment_target, $.struct_pattern_element],
    [$.matchable_expression, $.ensure_expression, $.rescue_expression, $.handling_expression],
    [$.matchable_expression, $.ensure_expression, $.rescue_expression],
    [$.pattern, $.typed_pattern],
    [$.struct_literal, $.struct_pattern],
    [$.assignment_target, $.exponent_expression],
    [$.primary_expression, $.struct_literal_shorthand_field, $.pattern_base, $.struct_pattern_field],
    [$.primary_expression, $.struct_literal_shorthand_field, $.pattern_base],
    [$.struct_literal_field, $.pattern_base, $.struct_pattern_field],
    [$.primary_expression, $.struct_literal_shorthand_field],
    [$.struct_literal_field, $.pattern_base],
    [$.block, $.struct_pattern],
    [$.struct_record, $.struct_tuple],
    [$.type_identifier, $.nil_literal],
    [$.pattern_base, $.wildcard_type],
    [$.struct_pattern, $.type_identifier],
    [$.type_suffix, $.type_prefix],
    [$.struct_type_suffix, $.type_prefix],
    [$.type_suffix, $.struct_type_suffix],
    [$._line_breaks, $._statement_sep],
    [$._line_breaks, $._comma_or_newline_sep],
    [$.struct_definition],
    [$.function_signature],
    [$.interface_definition],
    [$.source_file],
    [$.block],
    [$.iterator_block],
    [$.handling_block],
    [$.if_expression],
    [$.if_expression_with_else, $.if_expression_without_else],
  ],

  rules: {
    source_file: $ => seq(
      optional(seq(
        repeat($._statement_sep),
        field("package", $.package_statement),
      )),
      repeat($._statement_sep),
      repeat(seq($.statement, repeat1($._statement_sep))),
      optional($.statement),
    ),

    comment: _ => token(seq("##", /[^\n]*/)),

    _line_breaks: $ => prec(1, repeat1($._newline)),
    _comma_sep: _ => token(/,(\r?\n)*/),
    _comma_or_newline_sep: $ => choice($._comma_sep, $._newline),
    _statement_sep: $ => choice($._newline, ";"),

    package_statement: $ => seq(
      "package",
      $.identifier,
    ),

    import_statement: $ => seq(
      field("kind", choice("import", "dynimport")),
      field("path", $.import_path),
      optional(choice(
        seq(renameOperator, field("alias", $.identifier)),
        field("clause", $.import_clause),
      )),
    ),

    import_clause: $ => choice(
      alias(token(".*"), $.import_wildcard_clause),
      seq(
        alias(token(seq(".", "{")), "{"),
        optional($._line_breaks),
        sep1($, $.import_selector, $._comma_sep),
        optional($._comma_sep),
        optional($._line_breaks),
        "}"
      ),
    ),

    import_path: $ => seq(
      $.identifier,
      repeat(seq(".", $.identifier)),
    ),

    import_selector: $ => seq(
      $.identifier,
      optional(seq(renameOperator, $.identifier)),
    ),

    statement: $ => seq(
      choice(
        $.import_statement,
        $.prelude_statement,
        $.extern_function,
        $.function_definition,
        $.struct_definition,
        $.union_definition,
        $.type_alias_definition,
        $.interface_definition,
        $.named_implementation_definition,
        $.implementation_definition,
        $.methods_definition,
        $.elsif_clause_statement,
        $.else_clause_statement,
        $.continue_statement,
        $.return_statement,
        $.raise_statement,
        $.break_statement,
        $.while_statement,
        $.for_statement,
        $.rethrow_statement,
        $.ellipsis_statement,
        $.expression_statement,
      ),
    ),

    function_definition: $ => seq(
      optional("private"),
      "fn",
      field("method_shorthand", optional("#")),
      field("name", $.identifier),
      field("type_parameters", optional($.type_parameter_list)),
      field("parameters", $.parameter_list),
      field("return_type", optional($.return_type)),
      field("where_clause", optional($.where_clause)),
      field("body", $.block),
    ),

    type_parameter_list: $ => seq(
      "<",
      repeat($._statement_sep),
      sep1($, $.type_parameter, $._comma_sep),
      repeat($._statement_sep),
      ">",
    ),

    type_parameter: $ => seq(
      $.identifier,
      optional(seq(":", $.type_bound_list)),
      optional(seq(token.immediate("="), $.type_expression)),
    ),

    generic_parameter_list: $ => repeat1($.generic_parameter),

    generic_parameter: $ => seq(
      $.identifier,
      optional(seq(":", $.type_bound_list)),
      optional(seq(token.immediate("="), $.type_expression)),
    ),

    type_bound_list: $ => seq(
      $.type_expression,
      repeat(seq(
        "+",
        optional($._line_breaks),
        $.type_expression,
      )),
    ),

    declaration_type_parameters: $ => choice(
      $.type_parameter_list,
      $.generic_parameter_list,
    ),

    interface_argument_clause: $ => repeat1($.interface_type_expression),

    parameter_list: $ => seq(
      "(",
      choice(
        seq(
          optional($._line_breaks),
          ")",
        ),
        seq(
          optional($._line_breaks),
          sep1($, $.parameter, $._comma_sep),
          optional($._line_breaks),
          ")",
        ),
      ),
    ),

    parameter: $ => seq(
      optional("mut"),
      field("pattern", $.pattern),
      optional(seq(":", field("type", $.type_expression))),
    ),

    return_type: $ => seq("->", $.type_expression),

    where_clause: $ => seq(
      choice("where", lineKeyword("where")),
      repeat($._statement_sep),
      sep1($, $.where_constraint, $._comma_sep),
    ),

    where_constraint: $ => seq(
      field("subject", $.type_expression),
      ":",
      $.type_bound_list,
    ),

    block: $ => seq(
      "{",
      repeat($._statement_sep),
      repeat(seq($.statement, repeat1($._statement_sep))),
      optional($.statement),
      "}",
    ),

    expression_statement: $ => seq(
      field("expression", $.expression),
    ),

    struct_definition: $ => seq(
      optional("private"),
      "struct",
      field("name", $.identifier),
      field("type_parameters", optional($.declaration_type_parameters)),
      optional(choice(
        field("record", $.struct_record),
        field("tuple", $.struct_tuple),
      )),
      optional($.where_clause),
    ),

    struct_record: $ => seq(
      "{",
      optional($._line_breaks),
      choice(
        "}",
        seq(
          sep1($, $.struct_field, $._comma_or_newline_sep),
          optional($._comma_or_newline_sep),
          optional($._line_breaks),
          "}",
        ),
      ),
    ),

    struct_field: $ => seq(
      optional("private"),
      $.identifier,
      ":",
      $.type_expression,
    ),

    struct_tuple: $ => seq(
      "{",
      optional($._line_breaks),
      choice(
        "}",
        seq(
          sep1($, $.type_expression, $._comma_or_newline_sep),
          optional($._comma_or_newline_sep),
          optional($._line_breaks),
          "}",
        ),
      ),
    ),

    union_definition: $ => seq(
      optional("private"),
      "union",
      field("name", $.identifier),
      field("type_parameters", optional($.declaration_type_parameters)),
      "=",
      optional($._line_breaks),
      sep1($, $.type_expression, $._comma_sep),
    ),

    type_alias_definition: $ => seq(
      optional("private"),
      "type",
      field("name", $.identifier),
      field("type_parameters", optional($.generic_parameter_list)),
      field("where_clause", optional($.where_clause)),
      "=",
      optional($._line_breaks),
      field("target", $.type_expression),
    ),

    interface_definition: $ => seq(
      optional("private"),
      "interface",
      field("name", $.identifier),
      field("type_parameters", optional($.declaration_type_parameters)),
      optional(seq(
        "for",
        field("self_type", $.type_expression),
        optional(seq(choice(":", "="), field("base_interfaces", $.interface_composition))),
      )),
      field("where_clause", optional($.where_clause)),
      repeat($._statement_sep),
      choice(
        seq(
          "{",
          repeat($._statement_sep),
          repeat(seq($.interface_member, repeat($._statement_sep))),
          "}",
        ),
        seq(
          "=",
          optional($._line_breaks),
          field("composite", $.interface_composition),
        ),
      ),
    ),

    interface_member: $ => choice(
      seq(
        field("signature", $.function_signature),
        field("default_body", $.block),
      ),
      seq(
        field("signature", $.function_signature),
      ),
    ),

    interface_composition: $ => $.type_bound_list,

    function_signature: $ => seq(
      "fn",
      field("method_shorthand", optional("#")),
      field("name", $.identifier),
      field("type_parameters", optional($.type_parameter_list)),
      field("parameters", $.parameter_list),
      field("return_type", optional($.return_type)),
      field("where_clause", optional($.where_clause)),
    ),

    methods_definition: $ => seq(
      "methods",
      field("type_parameters", optional($.type_parameter_list)),
      field("target", $.type_expression),
      field("where_clause", optional($.where_clause)),
      repeat($._statement_sep),
      "{",
      repeat($._statement_sep),
      repeat(seq($.method_member, repeat($._statement_sep))),
      "}",
    ),

    implementation_definition: $ => seq(
      optional("private"),
      "impl",
      field("type_parameters", optional($.type_parameter_list)),
      field("interface", $.qualified_identifier),
      field("interface_args", optional($.interface_argument_clause)),
      "for",
      field("target", $.type_expression),
      field("where_clause", optional($.where_clause)),
      repeat($._statement_sep),
      "{",
      repeat($._statement_sep),
      repeat(seq($.method_member, repeat($._statement_sep))),
      "}",
    ),

    method_member: $ => choice(
      $.function_definition,
    ),

    named_implementation_definition: $ => seq(
      field("name", $.identifier),
      "=",
      field("implementation", $.implementation_definition),
    ),

    prelude_statement: $ => seq(
      "prelude",
      field("target", $.host_language),
      field("body", $.host_code_block),
    ),

    extern_function: $ => seq(
      "extern",
      field("target", $.host_language),
      field("signature", $.function_signature),
      field("body", $.host_code_block),
    ),

    host_language: _ => choice("go", "crystal", "typescript", "python", "ruby"),

    host_code_block: $ => seq(
      "{",
      repeat(choice($.host_code_block, $.host_code_chunk)),
      "}",
    ),

    host_code_chunk: _ => token(prec(-1, /[^{}]+/)),

    return_statement: $ => choice(
      prec.left(PREC.return_stmt, seq(
        "return",
        field("argument", $.expression),
      )),
      seq("return"),
    ),

    raise_statement: $ => seq(
      "raise",
      $.expression,
    ),

    break_statement: $ => prec.left(seq(
      "break",
      optional(choice(
        seq(field("label", $.label), optional(field("value", $.expression))),
        field("value", $.expression),
      )),
    )),

    continue_statement: $ => seq(
      "continue",
    ),

    rethrow_statement: $ => seq("rethrow"),

    ellipsis_statement: _ => "...",

    while_statement: $ => seq(
      "while",
      $.expression,
      $.block,
    ),

    loop_expression: $ => prec.left(1, seq(
      "loop",
      $.block,
    )),

    for_statement: $ => seq(
      "for",
      $.pattern,
      "in",
      $.expression,
      $.block,
    ),

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
      seq(
        $.unary_expression,
        repeat1(seq("as", $.type_expression)),
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
          field("body", choice($.block, $.expression)),
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
  },
});
