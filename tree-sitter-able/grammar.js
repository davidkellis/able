// follow the guide at https://tomassetti.me/incremental-parsing-using-tree-sitter/ for how to build a custom grammar and use it

/**
 * @file Able programming langauge
 * @author David Ellis <david@conquerthelawn.com>
 * @license EPL-2.0
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

module.exports = grammar({
  name: "able",

  rules: {
    source_file: ($) => repeat($._statement),

    _statement: ($) => choice($.expression_statement),

    expression_statement: ($) =>
      seq($._expression, optional(choice(";", "\n"))),

    _expression: ($) =>
      choice(
        $.integer_literal,
        $.float_literal,
        $.string_literal,
        $.boolean_literal,
        $.char_literal,
        $.nil_literal
      ),

    // Integer literals (no decimal allowed)
    integer_literal: ($) =>
      prec(
        1,
        seq(
          optional(/[+-]/),
          choice(
            /[0-9][0-9_]*?/, // Decimal
            /0[xX][0-9a-fA-F][0-9a-fA-F_]*?/, // Hex
            /0[oO][0-7][0-7_]*?/, // Octal
            /0[bB][0-1][0-1_]*?/ // Binary
          ),
          optional(
            seq(
              "_",
              choice(
                "i8",
                "i16",
                "i32",
                "i64",
                "i128",
                "u8",
                "u16",
                "u32",
                "u64",
                "u128"
              )
            )
          )
        )
      ),

    // Float literals (must include decimal or exponent)
    // float_literal: ($) =>
    //   prec.left(
    //     2,
    //     seq(
    //       optional(/[+-]/),
    //       choice(
    //         // Digits with optional decimal and/or exponent
    //         seq(
    //           /[0-9][0-9_]*/,
    //           choice(
    //             seq(
    //               /\./,
    //               optional(/[0-9][0-9_]*/),
    //               optional(/[eE][+-]?[0-9][0-9_]*/)
    //             ), // 123., 123.456, 123.456e10
    //             /[eE][+-]?[0-9][0-9_]*/ // 123e10
    //           )
    //         ),
    //         // Decimal starting with a point
    //         seq(/\./, /[0-9][0-9_]*/, optional(/[eE][+-]?[0-9][0-9_]*/)) // .5, .5e10
    //       ),
    //       optional(choice("_f32", "_f64"))
    //     )
    //   ),
    float_literal: ($) =>
      prec(
        2,
        seq(
          optional(/[+-]/),
          choice(
            /[0-9][0-9_]*\.[0-9][0-9_]*/, // Decimal
            /[0-9][0-9_]*\.[0-9][0-9_]*[eE][+-]?[0-9][0-9_]*/, // Decimal with exponent
            /[0-9][0-9_]*[eE][+-]?[0-9][0-9_]*/ // Exponent only
          ),
          optional(seq("_", choice("f32", "f64")))
        )
      ),

    string_literal: ($) => /"[^"]*"/,

    boolean_literal: ($) => choice("true", "false"),

    char_literal: ($) => /'[^']'/,

    nil_literal: ($) => "nil",
  },
});
