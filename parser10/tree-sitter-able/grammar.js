/**
 * @file Able language parser (v10 spec)
 * @author David Ellis <david@conquerthelawn.com>
 * @license epl-2.0
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

const PREC = {
  lambda: 0,
  pipe: 1,
  assignment: 2,
  range: 3,
  logical_or: 4,
  logical_and: 5,
  equality: 6,
  comparison: 7,
  bit_or: 8,
  bit_xor: 9,
  bit_and: 10,
  shift: 11,
  additive: 12,
  multiplicative: 13,
  unary: 14,
  exponent: 15,
  call: 16,
  member: 17,
};

const commaSep = rule => optional(commaSep1(rule));
const commaSep1 = rule => seq(rule, repeat(seq(',', rule)));

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
  "or",
  "else",
  "while",
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
  "proc",
  "spawn",
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
  '&=',
  '|=',
  '\\xor=',
  '<<=',
  '>>=',
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

  conflicts: $ => [
    [$.primary_expression, $.qualified_identifier],
    [$.primary_expression, $.pattern_base],
    [$.named_implementation_definition, $.primary_expression, $.pattern_base],
    [$.literal, $.literal_pattern],
    [$.pattern_base, $.struct_pattern_field],
    [$.primary_expression, $.pattern_base, $.struct_pattern_field],
    [$.lambda_parameter, $.pattern_base, $.struct_pattern_field],
    [$.proc_expression, $.primary_expression],
    [$.spawn_expression, $.primary_expression],
    [$.array_literal, $.array_pattern],
    [$.pattern, $.typed_pattern],
    [$.struct_literal, $.struct_pattern],
    [$.primary_expression, $.struct_literal_shorthand_field, $.pattern_base, $.struct_pattern_field],
    [$.struct_literal_field, $.pattern_base, $.struct_pattern_field],
    [$.primary_expression, $.struct_literal_shorthand_field],
    [$.struct_literal_field, $.pattern_base],
    [$.block, $.struct_pattern],
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
      alias(token(".*"), $.import_wildcard_clause),
      seq(
        alias(token(seq(".", "{")), "{"),
        commaSep1($.import_selector),
        optional(","),
        "}"
      ),
      seq("as", $.identifier),
    ),

    import_selector: $ => seq(
      $.identifier,
      optional(seq("as", $.identifier)),
    ),

    statement: $ => choice(
      $.prelude_statement,
      $.extern_function,
      $.function_definition,
      $.struct_definition,
      $.union_definition,
      $.interface_definition,
      $.named_implementation_definition,
      $.implementation_definition,
      $.methods_definition,
      $.continue_statement,
      $.return_statement,
      $.raise_statement,
      $.break_statement,
      $.while_statement,
      $.for_statement,
      $.rethrow_statement,
      $.expression_statement,
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
      commaSep1($.type_parameter),
      optional(","),
      ">",
    ),

    type_parameter: $ => seq(
      $.identifier,
      optional(seq(":", $.type_bound_list)),
    ),

    generic_parameter_list: $ => repeat1($.generic_parameter),

    generic_parameter: $ => seq(
      $.identifier,
      optional(seq(":", $.type_bound_list)),
    ),

    type_bound_list: $ => seq(
      $.type_expression,
      repeat(seq("+", $.type_expression)),
    ),

    declaration_type_parameters: $ => choice(
      $.type_parameter_list,
      $.generic_parameter_list,
    ),

    interface_argument_clause: $ => repeat1($.type_identifier),

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
      $.type_bound_list,
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
      field("type_parameters", optional($.declaration_type_parameters)),
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

    union_definition: $ => seq(
      optional("private"),
      "union",
      field("name", $.identifier),
      field("type_parameters", optional($.declaration_type_parameters)),
      "=",
      commaSep1($.type_expression),
    ),

    interface_definition: $ => seq(
      optional("private"),
      "interface",
      field("name", $.identifier),
      field("type_parameters", optional($.declaration_type_parameters)),
      optional(seq("for", field("self_type", $.type_expression))),
      field("where_clause", optional($.where_clause)),
      choice(
        seq(
          "{",
          repeat($.interface_member),
          "}",
        ),
        seq(
          "=",
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
        optional(";"),
      ),
    ),

    interface_composition: $ => $.type_bound_list,

    function_signature: $ => seq(
      "fn",
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
      "{",
      repeat($.method_member),
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
      "{",
      repeat($.method_member),
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

    return_statement: $ => prec.left(seq(
      "return",
      optional($.expression),
      optional(";"),
    )),

    raise_statement: $ => seq(
      "raise",
      $.expression,
      optional(";"),
    ),

    break_statement: $ => prec.left(seq(
      "break",
      optional(choice(
        seq(field("label", $.label), optional(field("value", $.expression))),
        field("value", $.expression),
      )),
      optional(";"),
    )),

    continue_statement: $ => seq(
      "continue",
      optional(";"),
    ),

    rethrow_statement: $ => seq("rethrow", optional(";")),

    while_statement: $ => seq(
      "while",
      $.expression,
      $.block,
    ),

    for_statement: $ => seq(
      "for",
      $.pattern,
      "in",
      $.expression,
      $.block,
    ),

    expression: $ => $.pipe_expression,

    pipe_expression: $ => prec.left(
      PREC.pipe,
      seq(
        $.matchable_expression,
        repeat(seq("|>", $.matchable_expression)),
      ),
    ),

    matchable_expression: $ => choice(
      $.ensure_expression,
      $.rescue_expression,
      $.handling_expression,
      $.proc_expression,
      $.spawn_expression,
    $.breakpoint_expression,
    $.match_expression,
    $.assignment_expression,
  ),

    ensure_expression: $ => seq(
      choice($.rescue_expression, $.handling_expression, $.assignment_expression),
      "ensure",
      field("ensure", $.block),
    ),

    rescue_expression: $ => seq(
      choice($.handling_expression, $.assignment_expression),
      "rescue",
      field("rescue", $.rescue_block),
    ),

    handling_expression: $ => prec.left(seq(
      $.assignment_expression,
      repeat1($.else_clause),
    )),

    else_clause: $ => seq(
      "else",
      field("handler", $.handling_block),
    ),

    handling_block: $ => seq(
      "{",
      optional(seq(
        "|",
        field("binding", $.identifier),
        "|",
      )),
      repeat($.statement),
      "}",
    ),

    match_expression: $ => seq(
      field("subject", $.assignment_expression),
      "match",
      "{",
      commaSep1($.match_clause),
      optional(","),
      "}",
    ),

    match_clause: $ => seq(
      "case",
      field("pattern", $.pattern),
      optional(field("guard", $.match_guard)),
      "=>",
      field("body", choice($.block, $.expression)),
    ),

    match_guard: $ => seq("if", $.expression),

    rescue_block: $ => seq(
      "{",
      commaSep1($.match_clause),
      optional(","),
      "}",
    ),

    proc_expression: $ => seq(
      "proc",
      choice(
        $.block,
        $.do_expression,
        $.call_target,
      ),
    ),

    spawn_expression: $ => seq(
      "spawn",
      choice(
        $.block,
        $.do_expression,
        $.call_target,
      ),
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
          field("right", $.assignment_expression),
        ),
      ),
      $.range_expression,
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
          field("operator", choice("..", "...")),
          $.logical_or_expression,
        ),
      ),
      $.logical_or_expression,
    ),

    logical_or_expression: $ => prec.left(
      PREC.logical_or,
      seq(
        $.logical_and_expression,
        repeat(seq("||", $.logical_and_expression)),
      ),
    ),

    logical_and_expression: $ => prec.left(
      PREC.logical_and,
      seq(
        $.bitwise_or_expression,
        repeat(seq("&&", $.bitwise_or_expression)),
      ),
    ),

    bitwise_or_expression: $ => prec.left(
      PREC.bit_or,
      seq(
        $.bitwise_xor_expression,
        repeat(seq("|", $.bitwise_xor_expression)),
      ),
    ),

    bitwise_xor_expression: $ => prec.left(
      PREC.bit_xor,
      seq(
        $.bitwise_and_expression,
        repeat(seq("\\xor", $.bitwise_and_expression)),
      ),
    ),

    bitwise_and_expression: $ => prec.left(
      PREC.bit_and,
      seq(
        $.equality_expression,
        repeat(seq("&", $.equality_expression)),
      ),
    ),

    equality_expression: $ => prec.left(
      PREC.equality,
      seq(
        $.comparison_expression,
        repeat(seq(choice("==", "!="), $.comparison_expression)),
      ),
    ),

    comparison_expression: $ => prec.left(
      PREC.comparison,
      seq(
        $.shift_expression,
        repeat(seq(choice(">", "<", ">=", "<="), $.shift_expression)),
      ),
    ),

    shift_expression: $ => prec.left(
      PREC.shift,
      seq(
        $.additive_expression,
        repeat(seq(choice("<<", ">>"), $.additive_expression)),
      ),
    ),

    additive_expression: $ => prec.left(
      PREC.additive,
      seq(
        $.multiplicative_expression,
        repeat(seq(choice("+", "-"), $.multiplicative_expression)),
      ),
    ),

    multiplicative_expression: $ => prec.left(
      PREC.multiplicative,
      seq(
        $.unary_expression,
        repeat(seq(choice("*", "/", "%"), $.unary_expression)),
      ),
    ),

    unary_expression: $ => choice(
      prec.right(
        PREC.unary,
        seq(
          choice("-", "!", "~"),
          $.unary_expression,
        ),
      ),
      $.exponent_expression,
    ),

    exponent_expression: $ => choice(
      prec.right(
        PREC.exponent,
        seq(
          $.postfix_expression,
          "^",
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
          $.proc_expression,
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

    call_target: $ => prec.left(
      PREC.call,
      seq(
        choice(
          $.primary_expression,
          $.proc_expression,
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
      commaSep($.expression),
      optional(","),
      ")",
    )),

    type_arguments: $ => prec.dynamic(1, seq(
      token.immediate("<"),
      commaSep1($.type_expression),
      optional(","),
      ">",
    )),

    index_suffix: $ => seq(
      "[",
      $.expression,
      optional(seq(":", $.expression)),
      "]",
    ),

    propagate_suffix: $ => "!",

    member_access: $ => prec.left(
      PREC.member,
      seq(
        ".",
        $.identifier,
      ),
    ),

    primary_expression: $ => choice(
      $.literal,
      $.identifier,
      $.placeholder_expression,
      $.implicit_member_expression,
      $.topic_reference,
      $.if_expression,
      $.do_expression,
      $.iterator_literal,
      $.lambda_expression,
      $.parenthesized_expression,
    ),

    lambda_expression: $ => prec.right(
      PREC.lambda,
      seq(
        "{",
        optional(field("parameters", $.lambda_parameter_list)),
        optional(seq("->", field("return_type", $.type_expression))),
        "=>",
        field("body", $.expression),
        "}",
      ),
    ),

    lambda_parameter_list: $ => seq(
      $.lambda_parameter,
      repeat(seq(",", $.lambda_parameter)),
      optional(","),
    ),

    lambda_parameter: $ => field("name", $.identifier),

    if_expression: $ => prec.right(
      PREC.logical_or,
      seq(
      "if",
      field("condition", $.expression),
      field("consequence", $.block),
      repeat(field("or_clause", $.or_clause)),
      optional(seq("else", $.block)),
    )),

    or_clause: $ => seq(
      "or",
      field("condition", optional($.expression)),
      field("consequence", $.block),
    ),

    do_expression: $ => seq(
      "do",
      $.block,
    ),

    parenthesized_expression: $ => seq(
      "(",
      $.expression,
      ")",
    ),

    iterator_literal: $ => seq(
      "Iterator",
      optional(field("element_type", $.type_expression)),
      "{",
      field("binding", $.identifier),
      "=>",
      field("body", $.iterator_body),
      "}",
    ),

    iterator_body: $ => seq($.statement, repeat($.statement)),

    literal: $ => choice(
      $.number_literal,
      $.character_literal,
      $.string_literal,
      $.interpolated_string,
      $.struct_literal,
      $.boolean_literal,
      $.nil_literal,
      $.array_literal,
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
      commaSep($.expression),
      optional(","),
      "]",
    ),

    struct_literal: $ => seq(
      field("type", $.qualified_identifier),
      field("type_arguments", optional($.type_arguments)),
      "{",
      optional(seq(
        commaSep1($.struct_literal_element),
        optional(","),
      )),
      "}",
    ),

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
      "...",
      field("source", $.expression),
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
      optional(commaSep1($.struct_pattern_element)),
      optional(","),
      "}",
    ),

    struct_pattern_element: $ => choice(
      $.struct_pattern_field,
      $.pattern,
    ),

    struct_pattern_field: $ => seq(
      field("field", $.identifier),
      optional(seq("as", field("binding", $.identifier))),
      optional(seq(":", field("value", $.pattern))),
    ),

    array_pattern: $ => seq(
      "[",
      optional(choice(
        $.array_pattern_rest,
        seq(
          commaSep1($.pattern),
          optional(seq(",", $.array_pattern_rest)),
        ),
      )),
      optional(","),
      "]",
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

    type_expression: $ => $.type_union,

    type_union: $ => choice(
      prec.left(seq($.type_arrow, repeat1(seq("|", $.type_arrow)))),
      $.type_arrow,
    ),

    type_arrow: $ => choice(
      prec.right(seq($.type_suffix, "->", $.type_arrow)),
      $.type_suffix,
    ),

    type_suffix: $ => prec.left(seq($.type_prefix, repeat($.type_prefix))),

    type_prefix: $ => choice(
      seq("?", $.type_prefix),
      seq("!", $.type_prefix),
      $.type_atom,
    ),

    type_atom: $ => choice(
      $.parenthesized_type,
      $.type_identifier,
      $.wildcard_type,
    ),

    parenthesized_type: $ => seq(
      "(",
      optional(seq(commaSep1($.type_expression), optional(","))),
      ")",
    ),

    wildcard_type: _ => "_",

    type_identifier: $ => choice(
      $.qualified_identifier,
      "Self",
      "nil",
      "void",
    ),

    qualified_identifier: $ => prec.right(seq(
      $.identifier,
      repeat(seq(".", $.identifier)),
    )),

    placeholder_expression: _ => token(choice("@", /@[1-9][0-9]*/)),
    implicit_member_expression: $ => seq(
      "#",
      field("member", $.identifier),
    ),

    topic_reference: _ => token("%"),

    identifier: _ => token(prec(-1, /[A-Za-z_][A-Za-z0-9_]*/)),

    number_literal: _ => token(choice(
      new RegExp(FLOAT_LITERAL_PATTERN),
      new RegExp(FLOAT_SUFFIX_ONLY_PATTERN),
      new RegExp(INTEGER_LITERAL_PATTERN),
    )),

    character_literal: _ => token(prec(1, new RegExp(CHARACTER_LITERAL_PATTERN))),

    string_literal: _ => token(seq(
      '"',
      repeat(choice(/[^"\\\n]+/, /\\./)),
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
