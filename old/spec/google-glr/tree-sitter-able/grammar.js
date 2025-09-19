/**
 * @file Able grammar for tree-sitter
 * @author David <david@conquerthelawn.com>
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

// grammar.js for Able (Draft)

// Helper function for comma-separated lists
const commaSep1 = rule => seq(rule, repeat(seq(',', rule)));
const commaSep = rule => optional(commaSep1(rule));

module.exports = grammar({
  name: 'able',

  // Whitespace and comments ignored by the main parser structure
  extras: $ => [
    /\s/,      // Whitespace
    $.comment,
  ],

  // Operator precedence
  precedences: $ => [
    ['left', 'assign'], // Lowest precedence for assignment-like?
    ['left', 'or'],
    ['left', 'and'],
    ['left', 'comparison'],
    ['left', 'term'],
    ['left', 'factor'],
    ['right', 'unary'],
    ['left', 'call', 'member_access', 'ufcs_access'], // Function calls, dot access
    ['left', 'pipe'], // Pipe operator |>
    ['nonassoc', 'range'], // Range operators .. ...
    ['left', 'type_apply_space'], // Space in type application
    ['left', 'type_apply_dot'], // Dot in type application
  ],

  // Potential ambiguities the grammar should try to resolve
  conflicts: $ => [
    // Conflicts often arise with block-like structures
    [$._expression, $.block],
    [$.struct_literal, $.block],
    [$.lambda_expression, $.block],
    // Type parameter lists vs. binary operators if types can be complex expressions
    [$.generic_type_parameters, $._expression],
    [$._type_constructor_application, $._expression], // Disambiguate 'Map K V' (type) vs 'Map + K' (expr)
    [$._pattern, $._expression], // Patterns share syntax with expressions
  ],

  rules: {
    // Top-level rule: a sequence of declarations or statements
    source_file: $ => repeat($._declaration_or_statement),

    _declaration_or_statement: $ => choice(
      $.function_definition,
      $.struct_definition,
      $.union_definition,
      $.interface_definition,
      $.implementation_definition,
      $.package_declaration,
      $.import_declaration,
      $.assignment_statement, // Using '=' at top level
      $._statement // Other statements like expression statements
    ),

    // --- Declarations ---

    package_declaration: $ => seq(
        'package',
        field('name', $._package_name)
        // Assuming implicit newline/semicolon termination handled by sequence
    ),

    _package_name: $ => choice(
      $.identifier,
      $.qualified_identifier // foo.bar
    ),

    import_declaration: $ => seq(
      'import',
      field('path', $._import_path),
      optional(field('alias', seq('as', $.identifier))),
      optional(field('selective', $._import_selective))
      // Assuming implicit newline/semicolon
    ),
    _import_path: $ => choice(
      $.identifier,
      $.qualified_identifier,
      seq($.qualified_identifier, '.*') // Wildcard import
    ),
    _import_selective: $ => seq(
        '{',
        commaSep(choice(
            $.identifier,
            seq($.identifier, 'as', $.identifier) // Selective alias
        )),
        '}'
    ),

    function_definition: $ => seq(
      optional('private'),
      'fn',
      field('name', $.identifier),
      optional(field('generics', $.generic_parameters)), // <T>
      field('parameters', $.parameter_list),
      optional(seq('->', field('return_type', $._type_expression))),
      field('body', $.block)
    ),

    struct_definition: $ => seq(
      optional('private'),
      'struct',
      field('name', $.identifier),
      optional(field('generics', $.generic_type_parameters)), // Space-separated generics
      field('body', seq('{', repeat($.field_definition), '}'))
    ),

    union_definition: $ => seq(
      optional('private'),
      'union',
      field('name', $.identifier),
      optional(field('generics', $.generic_type_parameters)),
      '=',
      field('variants', seq($._type_expression, repeat(seq('|', $._type_expression))))
    ),

    interface_definition: $ => seq(
      optional('private'),
      'interface',
      field('name', $.identifier),
      choice(
        // Regular interface: interface MyInterface<T> { ... }
        seq(
          optional(field('generics', $.generic_parameters)), // <T> for interface itself? Spec unclear, using <> for consistency
          field('body', seq('{', repeat($._associated_item_signature), '}'))
        ),
        // HKT interface: interface MyInterface for M _ { ... }
        seq(
          'for',
          field('hkt_param', $.identifier), // M
          '_', // Placeholder
          field('body', seq('{', repeat($._associated_item_signature), '}'))
        )
      )
    ),

    implementation_definition: $ => seq(
        'impl',
        optional(field('impl_generics', $.generic_parameters)), // <T> for impl itself
        field('interface_name', $.identifier), // Can be qualified?
        // Handle different impl targets: regular type vs HKT constructor
        choice(
            // for TargetType
            seq(
                optional(field('interface_args', $.generic_type_arguments)), // Interface<Arg>
                'for',
                field('target_type', $._type_expression)
            ),
            // TypeParamName for TypeConstructor (HKT impl)
            seq(
                field('hkt_impl_param', $.identifier), // A in 'impl Mappable A for Array'
                'for',
                field('target_constructor', $.identifier) // Array
            )
        ),
        optional($.where_clause),
        field('body', seq('{', repeat($._associated_item_impl), '}'))
    ),


    field_definition: $ => seq(
      field('name', $.identifier),
      ':',
      field('type', $._type_expression)
      // Assuming implicit newline/semicolon termination
    ),

    parameter_list: $ => seq(
        '(',
         commaSep($.parameter),
        ')'
    ),
    parameter: $ => seq(
      field('name', $.identifier),
      optional(seq(':', field('type', $._type_expression)))
    ),

    // --- Statements ---
    _statement: $ => choice(
      $.expression_statement,
      $.if_expression, // 'if' is an expression, but can appear statement context
      $.while_statement,
      $.for_statement,
      $.match_expression, // 'match' is an expression
      $.breakpoint_expression, // 'breakpoint' is an expression
      $.return_statement, // If added
      $.block
    ),

    expression_statement: $ => seq(
      $._expression
      // Implicit newline/semicolon handled by block/sequence structure? Need careful testing.
      // optional(';')
    ),

    assignment_statement: $ => prec('assign', seq( // Use precedence helper
      field('left', $._pattern), // Allow patterns on LHS
      '=',
      field('right', $._expression)
      // optional(';')
    )),

     while_statement: $ => seq(
       'while',
       field('condition', $._expression),
       field('body', $.block)
     ),

     for_statement: $ => seq(
       'for',
       field('pattern', $._pattern),
       'in',
       field('iterable', $._expression),
       field('body', $.block)
     ),

     return_statement: $ => seq( // Example if 'return' was added
       'return',
       optional(field('value', $._expression))
       // optional(';')
     ),

    // --- Expressions ---
    _expression: $ => choice(
        $.literal,
        $.identifier,
        $.unary_expression,
        $.binary_expression,
        $.parenthesized_expression,
        $.call_expression,
        $.member_access, // instance.field
        $.ufcs_access,   // instance.function - parsed same as member_access initially
        $.block,
        $.lambda_expression,
        $.if_expression,
        $.match_expression,
        $.array_literal,
        $.struct_literal,
        $.struct_update_expression,
        $.range_expression,
        $.partial_application_placeholder,
        $.breakpoint_expression,
        $.break_expression,
        $.interpolated_string // Treat as literal for now
    ),

    parenthesized_expression: $ => seq('(', $._expression, ')'),

    unary_expression: $ => prec('unary', seq(
      field('operator', choice('-', '!')), // Add other unary ops
      field('operand', $._expression)
    )),

    binary_expression: $ => choice(
        prec.left('pipe', seq($._expression, '|>', $._expression)),
        prec.left('or', seq($._expression, '||', $._expression)),
        prec.left('and', seq($._expression, '&&', $._expression)),
        prec.left('comparison', seq($._expression, choice('==', '!=', '<', '<=', '>', '>='), $._expression)),
        prec.left('term', seq($._expression, choice('+', '-'), $._expression)),
        prec.left('factor', seq($._expression, choice('*', '/', '%'), $._expression))
        // Add other operators and precedence levels
    ),

    // Combined rule for function call, method call, or UFCS call (syntax is similar)
    call_expression: $ => prec('call', seq(
      field('function', $._expression), // Callee: identifier, member_access, grouping etc.
      field('arguments', alias($._argument_list, $.arguments)),
      optional(field('trailing_lambda', $.lambda_expression)) // Trailing lambda support
    )),

    _argument_list: $ => seq(
      '(',
      commaSep($._expression),
      ')'
    ),

    // Handles both instance.field and instance.function for UFCS
    // Semantic analysis differentiates field access, method call, UFCS call, UFCS partial app
    member_access: $ => prec('member_access', seq(
        field('object', $._expression),
        '.',
        field('field', $.identifier)
    )),
    // UFCS access parses identically to member_access initially
    ufcs_access: $ => alias($.member_access, $.ufcs_access),

    block: $ => seq(
      '{',
      repeat($._declaration_or_statement), // Allow declarations inside blocks
      optional($._expression), // Optional final expression for block value
      '}'
    ),

    lambda_expression: $ => alias($._lambda_expression_body, $.lambda),
    _lambda_expression_body: $ => seq(
        '{',
        optional(field('parameters', $.lambda_parameter_list)),
        optional(seq('->', field('return_type', $._type_expression))),
        '=>',
        field('body', $._expression),
        '}'
    ),
    lambda_parameter_list: $ => commaSep1($.lambda_parameter), // Must have at least one if present? No, zero args is no list.
    lambda_parameter: $ => seq(
        field('name', $.identifier),
        optional(seq(':', field('type', $._type_expression)))
    ),


    if_expression: $ => seq(
      'if',
      field('condition', $._expression),
      field('consequence', $.block),
      'else', // Mandatory for expression
      field('alternative', $.block) // Could also be another 'if' directly for 'else if'
    ),

    match_expression: $ => seq(
      field('value', $._expression),
      'match',
      '{',
       commaSep($.match_arm),
      '}'
    ),
    match_arm: $ => seq(
      field('pattern', $._pattern),
      optional(seq('if', field('guard', $._expression))),
      '=>',
      field('result', $._expression) // Or block? Spec had ResultExpression
    ),

    array_literal: $ => seq(
        '[',
         commaSep($._expression),
        ']'
    ),

    struct_literal: $ => seq(
        optional($.identifier), // Optional type name before literal
        '{',
         commaSep($.struct_field_initializer),
        '}'
    ),
    struct_field_initializer: $ => choice(
        $.identifier, // Shorthand
        seq(field('name', $.identifier), ':', field('value', $._expression))
    ),

    struct_update_expression: $ => seq(
        optional($.identifier), // Optional type name
        '{',
        commaSep($.struct_field_initializer),
        optional(','), // Allow trailing comma before ...
        '...',
        field('base', $._expression),
        '}'
    ),

    range_expression: $ => prec('range', seq(
      $._expression,
      field('operator', choice('..', '...')),
      $._expression
    )),

    partial_application_placeholder: $ => '_',

    breakpoint_expression: $ => seq(
        'breakpoint',
        field('label', $.label),
        field('body', $.block)
    ),

    break_expression: $ => seq(
        'break',
        field('label', $.label),
        field('value', $._expression)
    ),

    label: $ => seq("'", $.identifier), // e.g., 'myLoop

    // --- Patterns (Simplified) ---
    _pattern: $ => choice(
      $.literal,
      $.identifier,
      $.wildcard_pattern,
      $.struct_pattern, // Basic form
      $.array_pattern   // Basic form
      // Add tuple patterns, type patterns, etc.
    ),
    wildcard_pattern: $ => '_',
    // Need more detailed rules for struct/array destructuring patterns
    struct_pattern: $ => seq(
        optional($.identifier), // Type name
        '{',
        // Highly simplified - just identifiers for now
        commaSep(choice($.identifier, seq($.identifier, '@', $.identifier))),
        '}'),
    array_pattern: $ => seq(
        '[',
        // Highly simplified
        commaSep(choice($.identifier, $.wildcard_pattern, $.literal)), // Add ...rest pattern
        ']'),

    // --- Types ---
    _type_expression: $ => choice(
        $.type_identifier,
        $.nullable_type,
        $.function_type,
        $._type_constructor_application,
        $.parenthesized_type
    ),
    type_identifier: $ => $.identifier, // Could be qualified
    nullable_type: $ => seq('?', $._type_expression),
    parenthesized_type: $ => seq('(', $._type_expression, ')'),
    function_type: $ => seq(
        '(', commaSep($._type_expression), ')',
        '->',
        $._type_expression
    ),
    // Handles 'Map K V' and 'Pair.Float' via precedence
    _type_constructor_application: $ => choice(
        prec.left('type_apply_space', seq($._type_expression, $._type_expression)),
        prec.left('type_apply_dot', seq($._type_expression, '.', $.identifier)) // Use identifier for type param after dot? Or full type expr? Needs spec clarification. Assuming identifier.
    ),

    // --- Generics / Interface Helpers ---
    generic_parameters: $ => seq('<', commaSep1($.identifier), '>'), // e.g. <T, U>
    generic_type_parameters: $ => repeat1($.identifier), // Space separated e.g. T U
    generic_type_arguments: $ => repeat1($._type_expression), // Space separated e.g. i32 String
    where_clause: $ => seq('where', commaSep1($._where_predicate)),
    _where_predicate: $ => seq($.identifier, ':', $._interface_bound), // T: Display
    _interface_bound: $ => seq($.identifier, repeat(seq('+', $.identifier))), // Display + Clone

     _associated_item_signature: $ => choice(
         $.method_signature,
         $.associated_type_signature,
         $.associated_const_signature
     ),
     method_signature: $ => seq(
         'fn', $.identifier, optional($.generic_parameters), $.parameter_list, optional(seq('->', $._type_expression)), ';'
     ),
     associated_type_signature: $ => seq(
         'type', $.identifier, optional(seq(':', $._interface_bound)), ';'
     ),
     associated_const_signature: $ => seq(
         'const', $.identifier, ':', $._type_expression, ';'
     ),

      _associated_item_impl: $ => choice(
         $.function_definition, // Full function for methods
         $.associated_type_impl,
         $.associated_const_impl
     ),
     associated_type_impl: $ => seq(
         'type', $.identifier, '=', $._type_expression
     ),
      associated_const_impl: $ => seq(
         'const', $.identifier, /* optional type */ '=', $._expression
     ),

    // --- Terminals ---
    identifier: $ => /[a-zA-Z_][a-zA-Z0-9_]*/,
    // Handle Self keyword specifically if needed, overlaps identifier
    // Self: $ => 'Self',

    qualified_identifier: $ => seq($.identifier, repeat1(seq('.', $.identifier))),

    literal: $ => choice(
      $.number,
      $.string,
      $.interpolated_string, // Treat as string token for now
      $.char,
      $.boolean,
      $.nil
    ),

    number: $ => /\d+(\.\d+)?([eE][+-]?\d+)?(_\d+)*/, // Basic number, allow underscores anywhere after first digit? Simplified.
    string: $ => /"[^"]*"/, // Basic string, no escapes
    interpolated_string: $ => /`[^`]*`/, // Basic, no interpolation logic or escapes
    char:   $ => /'([^'\\]|\\.)'/, // Basic char literal with escape support
    boolean: $ => choice('true', 'false'),
    nil: $ => 'nil',

    comment: $ => token(seq('##', /.*/)),

  }
});
