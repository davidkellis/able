// follow the guide at https://tomassetti.me/incremental-parsing-using-tree-sitter/ for how to build a custom grammar and use it

/**
 * @file Able programming langauge
 * @author David Ellis <david@conquerthelawn.com>
 * @license EPL-2.0
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

///////////////////////////////////////////////////////////////////////////////////////////////////

// grammar.js for Able Language (Corrected Interpolated String Content)

module.exports = grammar({
  name: "able",

  extras: ($) => [
    /\s/,
    $.line_comment,
    // $.block_comment, // TBD
  ],

  precedences: ($) => [],

  conflicts: ($) => [],

  rules: {
    source_file: ($) => repeat($._expression),

    _expression: ($) => $._literal,

    _literal: ($) =>
      choice(
        $.integer_literal,
        $.float_literal,
        $.boolean_literal,
        $.char_literal,
        $.string_literal,
        $.interpolated_string_literal, // Keep as syntactic rule
        $.nil_literal
      ),

    // --- Other Literals (Unchanged from previous correct version) ---
    integer_literal: ($) =>
      seq(
        $._integer_literal_base,
        optional($._integer_type_suffix)
      ),

    _integer_literal_base: ($) =>
      token(
        choice(
          /\d[\d_]*\d/, // Decimal
          seq(/0[xX]/, /[a-fA-F0-9][a-fA-F0-9_]*/), // Hexadecimal
          seq(/0[oO]/, /[0-7][0-7_]*/), // Octal
          seq(/0[bB]/, /[01][01_]*/) // Binary
        )
      ),

    _integer_type_suffix: ($) =>
      token(
        choice(
          /_i8/, /_u8/, /_i16/, /_u16/, /_i32/, /_u32/, /_i64/, /_u64/, /_i128/, /_u128/
        )
      ),

    float_literal: ($) =>
      seq(
        $._float_literal_base,
        optional($._float_type_suffix)
      ),

    _float_literal_base: ($) =>
      token(
        choice(
          seq(/\d[\d_]*\.(\d[\d_]*)?/, optional(seq(/[eE]/, optional(/[+-]/), /\d[\d_]*/))), // Decimal with optional exponent
          seq(/\d[\d_]*/, /[eE]/, optional(/[+-]/), /\d[\d_]*/) // Exponent-only form
        )
      ),

    _float_type_suffix: ($) =>
      token.immediate(
        choice(
          /_f32/, /_f64/
        )
      ),

    boolean_literal: ($) => choice("true", "false"),
    char_literal: ($) =>
      seq(/* ... */ "'", choice($._char_content, $.escape_sequence), "'"),
    _char_content: ($) => token.immediate(/[^\\']/),
    string_literal: ($) =>
      seq(
        /* ... */ '"',
        repeat(choice($._string_content, $.escape_sequence)),
        '"'
      ),
    _string_content: ($) => token.immediate(/[^\\"]+/),
    // --- End Unchanged Literals ---

    // Interpolated String Literal (Syntactic Rule)
    interpolated_string_literal: ($) =>
      seq(
        "`",
        repeat(
          choice(
            $.interpolation, // Handles ${expression}
            $.escape_sequence, // Handles \`, \$, \\, etc. (using the common rule)
            $._interpolated_content // Handles any other character sequence
          )
        ),
        "`"
      ),

    // **Corrected Token** for content within interpolated string
    // Matches any sequence of characters that are not backslash or backtick.
    // The parser will try $.interpolation first if it sees '$',
    // and $.escape_sequence first if it sees '\'.
    _interpolated_content: ($) => token.immediate(/[^\\`]+/),

    // Interpolation block `${expression}` (Syntactic Rule)
    interpolation: ($) =>
      seq(
        token.immediate("${"), // Use token.immediate
        $._expression,
        token.immediate("}") // Use token.immediate
      ),

    // Common Escape Sequence (Lexical/Token Rule)
    // Handles \n, \r, \t, \\, \', \", \`, \$ and \u{...}
    // Note: Added \` and \$ explicitly here for clarity, though \\. covers them.
    escape_sequence: ($) =>
      token.immediate(/\\(?:[nrt\\'"`$]|\\[uU]\{[0-9a-fA-F]{1,6}\})/), // Match \ followed by specific chars or unicode sequence

    // Alternative escape_sequence using simpler `\\.` if the above is too restrictive
    // escape_sequence: $ => token.immediate(/\\./), // Matches \ followed by ANY single character

    // Nil Literal
    nil_literal: ($) => "nil",

    // Line Comment
    line_comment: ($) => token(seq("##", /.*/)),

    // block_comment: $ => token(seq( ... TBD ... )),
  },
});
