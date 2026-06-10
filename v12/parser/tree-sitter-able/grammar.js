/**
 * @file Able language parser (v12 spec)
 * @author David Ellis <david@conquerthelawn.com>
 * @license epl-2.0
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check
const declarationRules = require("./grammar/declarations");
const expressionRules = require("./grammar/expressions");
const literalPatternTypeRules = require("./grammar/literals-patterns-types");

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
    ...declarationRules,
    ...expressionRules,
    ...literalPatternTypeRules,
  },
});
