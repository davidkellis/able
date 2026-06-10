/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

const { PREC, sep1, lineKeyword, renameOperator } = require("./helpers");

module.exports = {
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

  export_statement: $ => choice(
    seq("export", field("name", $.identifier)),
    seq("export", "*", "from", field("path", $.import_path)),
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
      $.export_statement,
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

  // Prefer the declaration when an empty implementation body could otherwise
  // be parsed as the right side of an assignment expression with a struct
  // literal. `impl` is declaration syntax in this position.
  named_implementation_definition: $ => prec.dynamic(1, seq(
    field("name", $.identifier),
    "=",
    field("implementation", $.implementation_definition),
  )),

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

};
