/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

module.exports = grammar({
  name: "able",

  rules: {
    source_file: ($) => repeat($._definition),

    _definition: ($) =>
      choice(
        $.function_definition,
        $.assignment
        // TODO: other kinds of definitions
      ),

    variable_declaration: ($) => seq($.identifier, "=", $._expression),

    assignment: ($) => choice($.simple_assignment),

    simple_assignment: ($) => seq($.identifier, "=", $._expression),

    function_definition: ($) =>
      seq("fn", $.identifier, $.parameter_list, $._type, $.block),

    parameter_list: ($) =>
      seq(
        "(",
        // TODO: parameters
        ")"
      ),

    _type: ($) =>
      choice(
        "bool",
        "string",
        "int", // i64
        "i32",
        "i64",
        "i128",
        "float", // f64
        "f32",
        "f64"
      ),

    block: ($) => choice($.brace_block, $.do_block),

    brace_block: ($) => seq("{", repeat($._statement), "}"),

    do_block: ($) => seq("do", repeat($._statement), "end"),

    _statement: ($) =>
      choice(
        $.return_statement
        // TODO: other kinds of statements
      ),

    return_statement: ($) => seq("return", $._expression),

    _expression: ($) =>
      choice(
        $.identifier,
        $.literal
        // TODO: other kinds of expressions
      ),

    literal: ($) => choice($.number, $.string),

    string: ($) => choice(),

    identifier: ($) => /[a-zA-Z1-9_]*[a-zA-Z_]+[a-zA-Z1-9_]*/,

    number: ($) => /\d+/,
  },
});
