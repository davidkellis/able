grammar Able;

sourceFile
  : packageDecl eos ( packageLevelDecl eos)* EOF
  ;

// package foo
// package internationalization
packageDecl
  : 'package' LWS packageIdentifier
  ;

// import io
// import io.*
// import io.{puts, gets}
// import io.{puts as p}
// import internationalization as i18n
// import internationalization as i18n.{Unicode}
// import davidkellis.matrix as mat.{Matrix, Vector as Vec}
// import davidkellis.matrix as mat.{Matrix, Vector as Vec,}
// import davidkellis.matrix as mat.{
//   Matrix as Mat,
//   Vector as Vec,
// }
importDecl
  : 'import' LWS fqPackageIdentifier (LWS 'as' LWS packageIdentifier)?       # ImportDeclBasePackage
  | 'import' LWS fqPackageIdentifier (LWS 'as' LWS packageIdentifier)? '.*'  # ImportDeclAllMembers
  | 'import' LWS fqPackageIdentifier (LWS 'as' LWS packageIdentifier)? 
      '.{' (WS? identifier (LWS 'as' LWS identifier)? WS? ',')* 
           (WS? identifier (LWS 'as' LWS identifier)?)? WS? '}'    # ImportDeclSpecificMembers
  ;

// packageLevelDecl are any of the following:
// variable declarations
// assignment expressions
// struct/union/function/interface/impl/macro definitions
packageLevelDecl
  : importDecl
  | assignmentExpr
  | variableDecl
  | structDefn
  | unionDefn
  | functionDefn
  | interfaceDefn
  | implDefn
  | macroDefn
  ;

// v = expr
// v: type = expr
// v1, v2 = expr, expr              # parallel assignment
// v1: type, v2: type = expr, expr  # parallel assignment
// (v1, v2) = expr
// Array{v1, v2, v3*} = expr
// MyStruct{v1, v2} = expr
// v += expr
// v -= expr
// v *= expr
assignmentExpr
  : multipleAssignmentLhs LWS? '=' LWS? expressionList                      # AssignmentExprMultipleAssignment
  | optionallyTypedIdentifier LWS? infixAssignmentOperator LWS? expression  # AssignmentExprOperatorAssignment
  ;

// Only operators that produce a value in the same domain as their arguments (i.e. operators whose domain and range are the same) should be allowed here.
// It wouldn't make sense to allow =< or =>, because those would return bool, when the arguments would likely be numeric, and you couldn't assign a bool to a numeric type.
infixAssignmentOperator
  : '=' ('+' | '-' | '*' | '/' | '\\' | '%' | '**' | '&' | '|' | '&&' | '||' | '<<' | '>>')
  ;

multipleAssignmentLhs
  : (assignmentLhs LWS? ',')* LWS? assignmentLhs
  ;

assignmentLhs
  : identifier LWS? ':' LWS? destructuringTypePrimitive         # AssignmentLhsDestructuringPatternWithIdent
  | destructuringPrimitive                                      # AssignmentLhsDestructuringPatternOnly
  | functionCall                                                # AssignmentLhsIndexAssignment  // todo: change this from functionCall to indexAssignment
  ;

// id1
// id1, id2
// id1: type1
// id1: type1, id2: type2
// optionallyTypedIdentifierList
//   : (optionallyTypedIdentifier LWS? ',')* LWS? optionallyTypedIdentifier
//   ;

// id1
// id1: type1
optionallyTypedIdentifier
  : identifier (LWS? ':' LWS? typeName)?
  ;

// id: Int
// id: CallStackLocal Int
// id: CallStackLocal (Array Int)
variableDecl
  : identifier LWS? ':' LWS? typeName
  ;


// id1
// id1, id2
identifierList
  : identifier ( WS? ',' WS? identifier )*
  ;

// id1
// id1, id2
fqIdentifierList
  : fqIdentifier ( WS? ',' WS? fqIdentifier )*
  ;

// id1
// id1, id2
localIdentifierList
  : identifier ( WS? ',' WS? identifier )*
  ;

// <expr> ; <expr>
// <expr>
expressions
  : expression ( WS? eos WS? expression )*
  ;

// <expr> , <expr>
expressionList
  : expression ( WS? ',' WS? expression )*
  ;

// struct Foo T [T: Iterable] { Int, Float, T }
// struct Foo T { Int, Float, T, }
// struct Foo T {
//   Int,
//   Float,
//   T,
// }
// struct Foo T {
//   Int
//   Float
//   T
// }
// struct Foo T [T: Iterable] { x: Int, y: Float, z: T }
// struct Foo T { x: Int, y: Float, z: T, }
// struct Foo T {
//   x: Int,
//   x: Float,
//   y: T,
// }
// struct Foo T {
//   x: Int
//   y: Float
//   z: T
// }
structDefn
  : 'struct' LWS inlineStructDefn
  ;

inlineStructDefn
  : typeName (LWS typeParametersDecl)? LWS? '{' WS? commaDelimitedTypeNameList WS? '}'    # InlineStructDefnCommaDelimitedPositionalTypes
  | typeName (LWS typeParametersDecl)? LWS? '{' WS? newlineDelimitedTypeNameList WS? '}'  # InlineStructDefnNewlineDelimitedPositionalTypes
  | typeName (LWS typeParametersDecl)? LWS? '{' WS? commaDelimitedFieldList WS? '}'       # InlineStructDefnCommaDelimitedFieldList
  | typeName (LWS typeParametersDecl)? LWS? '{' WS? newlineDelimitedFieldList WS? '}'     # InlineStructDefnNewlineDelimitedFieldList
  ;

typeParametersDecl
  :	'[' (WS? freeTypeParameter WS? ',')* (WS? freeTypeParameter  WS?) ']'
  ;

// T : Foo
// T _ : Mappable
// T : Foo + Bar
// T supersetOf X
// T supersetOf X|Y|Z
freeTypeParameter
  :	typeName (LWS? typeConstraint)?
  ;

// : Foo
// : Foo & Bar
// supersetOf X
// supersetOf X|Y|Z
typeConstraint
  : ':' LWS? intersectionType         # TypeConstraintImplementsIntersectionType
  | ':' LWS? typeName                 # TypeConstraintImplementsInterface
//   | SUPERSET_OF LWS typeIdentifier (LWS? '|' LWS? typeIdentifier)*  # TypeConstraintSupersetOfConstraint
  ;

// Buildable & Mappable
intersectionType
  : typeName (LWS? '&' LWS? typeName)+
  ;

// T
// Array T
// Iterable T
// Iterable _
// Iterable.T
// Iterable._
// Foo A B C
// Foo A (B C)
// Foo A B.C
typeName
  :	typeIdentifier (LWS (freeTypeIdentifier | typeNameSubRule))*
  | typeNameSubRule
  ;

typeNameSubRule
  : typeIdentifier ('.' freeTypeIdentifier)* 
  | '(' LWS? typeName LWS? ')'
  ;

// Int, Float, T
// Int, Float, T,
//  Int,
//  Float,
//  T,
commaDelimitedTypeNameList
  : typeName ( WS? ',' WS? typeName )* (WS? ',')?
  ;

/*
  Int
  Float
  T
 */
newlineDelimitedTypeNameList 
  : typeName ( WS? NL+ WS? typeName )* (WS? NL+)?
  ;

// x: Int, y: Float, z: T
// x: Int, y: Float, z: T,
//   x: Int,
//   y: Float, 
//   z: T,
commaDelimitedFieldList
  : typedIdentifier ( WS? ',' WS? typedIdentifier )* (WS? ',')?
  ;

/*
  x: Int
  y: Float
  z: T
 */
newlineDelimitedFieldList
  : typedIdentifier ( WS? NL+ WS? typedIdentifier )* (WS? NL+)?
  ;


// id1: type1
typedIdentifier
  :	WS? identifier LWS? ':' LWS? typeName WS?;

// Union definition:
// union String? = String | Nil
// union House = SmallHouse { sqft: Float }
//  | MediumHouse { sqft: Float }
//  | LargeHouse { sqft: Float }
// union Foo T [T: Blah] = 
//  | Bar A [A: Stringable] { a: A, t: T }
//  | Baz B [B: Qux] { b: B, t: T }
// union Option A = Some A {A} | None A {}
// union Result A B = Success A {A} | Failure B {B}
// union ContrivedResult A B [A: Fooable, B: Barable] = Success A X [X: Stringable] {A, X} | Failure B Y [Y: Serializable] {B, Y}
unionDefn
  : 'union' LWS typeName (LWS typeParametersDecl)? LWS? '=' 
      WS? (NL+ WS? '|' WS?)? unionAlternative (WS? '|' WS? unionAlternative WS?)+
  ;

unionAlternative
  : typeName | inlineStructDefn
  ;

// fn <function name>[<optional type paramter list>] = <function object OR partial application expression OR placeholder lambda expression>
// fn <function name>[<optional type paramter list>](<parameter list>) -> <optional return type> { <function body> }
// fn <function name>[<optional type paramter list>](<parameter list>) -> <optional return type> => <function body>
functionDefn
  : 'fn' LWS identifier (LWS typeParametersDecl)? WS? '=>' WS? functionAliasExpression   # FunctionDefnAlias
  | 'fn' LWS identifier (LWS typeParametersDecl)? functionSignature WS? functionBody     # FunctionDefnFull
  ;

/*
 fn func1[<optional type paramter list>](<parameter list>) -> <optional return type> => <expr>
 fn func2[<optional type paramter list>](<parameter list>) -> <optional return type> { <expr>; <expr>; ... }
 */
functionDefnList
  : functionDefn ( WS? NL+ WS? functionDefn )* (WS? NL+)?
  ;

// fn <function name>[<optional type paramter list>](<parameter list>) -> <optional return type>
functionDecl
  : 'fn' LWS identifier (LWS typeParametersDecl)? functionSignature
  ;

/*
 fn func1[<optional type paramter list>](<parameter list>) -> <optional return type>
 fn func2[<optional type paramter list>](<parameter list>) -> <optional return type>
 */
functionDeclList
  : functionDecl ( WS? NL+ WS? functionDecl )* (WS? NL+)?
  ;

// fn[<optional type paramter list>](<parameter list>) -> <optional return type> { <function body> }
// fn[<optional type paramter list>](<parameter list>) -> <optional return type> => <function body>
// (<parameter list>) -> <optional return type> { <function body> }
// (<parameter list>) -> <optional return type> => <function body>
anonymousFunctionDefn
  : 'fn' typeParametersDecl? functionSignature WS? functionBody   # AnonymousFunctionDefnLong
  | functionSignature WS? functionBody                            # AnonymousFunctionDefnShort
  | lambdaExpression
  ;

// { <paramter list> -> <return type> => expression }
lambdaExpression
  : '{' WS? parameterList WS? '->' WS? typeName WS? '=>' WS? expressions WS? '}'    # LambdaParamsAndReturn
  | '{' WS? parameterList WS? '=>' WS? expressions WS? '}'                          # LambdaParamsOnly
	| '{' WS? '->' WS? typeName WS? '=>' WS? expressions WS? '}'                      # LambdaReturnOnly
  | '{' WS? expressions WS? '}'                                                     # LambdaExpressionsOnly
  ;

functionAliasExpression
  : identifier
  | anonymousFunctionDefn
  | partialApplicationExpr
  | placeholderLambdaExpr
  ;

functionSignature
  : '(' WS? parameterList WS? ')'                           # FunctionSignatureParameterList
  | '(' WS? parameterList WS? ')' WS? '->' WS? typeName     # FunctionSignatureParameterListAndReturnType
  ;

parameterList
  : (fnParameter WS? ',')* WS? fnParameter
  ;

fnParameter
  : identifier LWS? ':' LWS? destructuringTypePrimitive     # FnParameterDestructuringPatternWithIdent
	| destructuringPrimitive                                  # FnParameterDestructuringPatternOnly
  | functionCall                                            # FnParameterIndexAssignment  // todo: change this from functionCall to indexAssignment
  ;

destructuringPrimitive
  : destructuringTuple
  | destructuringPrimitive '::' destructuringPrimitive      # DestructuringPrimitiveTuplePair
  | destructuringSequence
  | destructuringStruct
  | identifier
  ;

destructuringTypePrimitive
  : destructuringTuple
  | destructuringTypePrimitive '::' destructuringTypePrimitive    # DestructuringTypePrimitiveTuplePair
  | destructuringSequence
  | destructuringStruct
  | typeName
  ;


functionBody
  : block
  | '=>' WS? expression
  ;

// { expressions }
// { expressions catch: patternMatchAlternatives ensure: expressions }
block
  : '{' WS? expressions (WS 'catch' WS patternMatchAlternatives (WS 'ensure' WS expressions)? )? WS? '}'
  ;

// interface Foo T [T: Iterable] {
//   fn foo() -> i32
// }
interfaceDefn
  : 'interface' LWS typeName (LWS typeParametersDecl)? LWS? '{' WS? functionDeclList? WS? '}'
  ;

// [implName =] impl [X, Y, Z, ...] A <B C ...> for D <E F ...> {
//   fn foo(T1, T2, ...) -> T3 { ... }
//   ...
// }
implDefn
  : identifier LWS? '=' LWS? anonymousImplDefn    # ImplDefnNamedImplDefn
  | anonymousImplDefn                             # ImplDefnAnonymousImplDefn
  ;

// impl [X, Y, Z, ...] A <B C ...> for D <E F ...> {
//   fn foo(T1, T2, ...) -> T3 { ... }
//   ...
// }
anonymousImplDefn
  : 'impl' (LWS typeParametersDecl)? LWS typeName LWS 'for' LWS typeName LWS? '{' WS? functionDefnList? WS? '}'
  ;

// macro <name of macro function>(<parameters>) {
//   <pre-template logic goes here>
//   `<template goes here>`
// }
macroDefn
  : 'macro' LWS identifier '(' WS? localIdentifierList WS? ')' WS? macroBlock
  ;

// { expressions }
// { expressions catch: patternMatchAlternatives ensure: expressions }
macroBlock
  : '{' WS? macroExpressions (WS 'catch' WS patternMatchAlternatives (WS 'ensure' WS expressions)? )? WS? '}'
  ;

// <expr> ; <expr>
// <expr>
macroExpressions
  : macroExpression ( WS? eos WS? macroExpression )*
  ;

macroExpression
  : expression
  | macroTemplate
  ;

macroTemplate
  : '`' WS? ( UNICODE_CHAR | NEWLINE )*? WS? '`'
  ;


// assuming h is of type House
// h match {
//   TinyHouse | SmallHouse => puts("build a small house")
//   m: MediumHouse => puts("build a modest house - $m")
//   LargeHouse{area} => puts ("build a large house of $area sq. ft.")
//   HugeHouse{poolCount=pools} => puts ("build a huge house with $poolCount pools!")
// }
patternMatchAlternatives
  : (caseClause eos)* caseClause
  ;

// TinyHouse | SmallHouse => puts("build a small house")
// m: MediumHouse => puts("build a modest house - $m")
// l: LargeHouse{area} => puts ("build a large house of $area sq. ft. - $l")
// HugeHouse{_, poolCount} => puts ("build a huge house with $poolCount pools!")
caseClause
  : patternWithOptionalIdentifier WS '=>' WS expression
  ;

patternWithOptionalIdentifier
  : identifier LWS? ':' LWS? pattern        # PatternWithOptionalIdentifierPatternWithIdent
  | pattern                                 # PatternWithOptionalIdentifierPatternOnly
  ;

pattern
  : typeName
  | unionOfTypes
  | destructuringTuple
  | destructuringPrimitive '::' destructuringPrimitive      # PatternDestructuringTuplePair
  | destructuringSequence
  | destructuringStruct
  | literal
  ;

// Array A | bool | String
unionOfTypes: typeName (WS '|' WS typeName)+ ;

// (v1)
// (v1, v2)
// v1::v2
destructuringTuple
  : '(' WS? destructuringPrimitive WS? ')'
  ;

// Array[x]
// Array[_, x]
// Array[x, y, zs*]
// Array Int [x, y, zs*]
// List[1, x, 3]
destructuringSequence
  : typeName '[' (WS? destructuringPrimitive WS? ',')* (WS? destructuringPrimitive WS?) ']'     # DestructuringSequence
	| typeName '[' (WS? destructuringPrimitive WS? ',')* (WS? identifier '*' WS?) ']'             # DestructuringSequenceWithSplatArgument
  ;

// MyStruct{v1, v2}
// MyStruct{v1=name, v2=age}
// MyStruct Int {v1, v2}
destructuringStruct
  : typeName LWS? '{' (WS? destructuringPrimitive WS? ',')* (WS? destructuringPrimitive WS?) '}'    # DestructuringStruct
  | typeName LWS? '{' (WS? identifier WS? '=' WS? destructuringPrimitive WS? ',')* (WS? identifier WS? '=' WS? destructuringPrimitive WS?) '}'    # DestructuringStructWithNamedIdentifiers
  ;

// references:
// - http://kotlinlang.org/docs/reference/grammar.html#expression
// - https://docs.julialang.org/en/release-0.4/manual/mathematical-operations/
expression
  : '(' WS? expression WS? ')'                                          # ExpressionSubExpression
  | fqIdentifier
  | literal
  | importDecl
  | functionDefn
  | doExpr
  | ifExpr
	| expression WS 'match' WS '{' WS? patternMatchAlternatives WS? '}'   # ExpressionMatchExpr
  | whileExpr
  | forExpr
  | returnStatement
  | 'break'                                                             # ExpressionBreakExpr
  | 'continue'                                                          # ExpressionContinueExpr
	| 'jumppoint' WS label WS '{' WS? expressions WS? '}'                 # ExpressionJumpPointDefn
	| 'jump' WS label WS expression                                       # ExpressionJumpExpr

  | functionCall
  | lhs=expression WS? op=('.' | '&.') WS? rhs=expression               # ExpressionNavigation
  | expression op=('!!')                                                # ExpressionUnaryPostfixExpr  // todo: may not be needed

	| op=('+' | '-' | '!' | '~') expression                               # ExpressionUnaryPrefixExpr
  | expression WS 'as' WS typeName                                      # ExpressionAsExpr            // todo: may not be needed
  | lhs=expression WS op=('**') WS rhs=expression                       # ExpressionExponentiationExpr
  | lhs=expression WS op=('*' | '/' | '\\' | '%') WS rhs=expression     # ExpressionMultiplicativeExpr
  | lhs=expression WS op=('+' | '-') WS rhs=expression                  # ExpressionAdditiveExpr
	| lhs=expression '..' rhs=expression                                  # ExpressionInclusiveRange
	| lhs=expression '...' rhs=expression                                 # ExpressionExclusiveRange
  | lhs=expression WS op=('<' | '>' | '>=' | '<=') WS rhs=expression    # ExpressionComparison
  | lhs=expression WS op=('==' | '!=') WS rhs=expression                # ExpressionEqualityComparison
  | lhs=expression WS '&&' WS rhs=expression                            # ExpressionConjunction
  | lhs=expression WS '||' WS rhs=expression                            # ExpressionDisjunction

  | assignmentExpr
  | variableDecl

  | expression WS? 'if' condition=expression                            # ExpressionIfSuffixExpr
  | expression WS? 'unless' condition=expression                        # ExpressionUnlessSuffixExpr
  | expression WS? 'while' condition=expression                         # ExpressionWhileSuffixExpr
  ;

// a label takes the form of a one-word Symbol in Ruby - e.g. :foo, :stop, :ThisLittlePiggyWentToMarket
label
  : ':' identifier
  ;

doExpr
  : 'do' WS? block
  ;

functionCall
  : fqIdentifier LWS? '(' WS? expressionList WS? ')'    // todo: the stuff prior to the open-paren needs to be redone
  ;

ifExpr
  : 'if' LWS expression WS? '{' WS? expressions WS? '}' WS? 'else' WS? '{' WS? expressions WS? '}'
  ;

literal
  : '()'                  # LiteralUnit
  | 'nil'                 # LiteralNil
  | ('true' | 'false')    # LiteralBool
  | INT_LITERAL           # LiteralInt
	| FLOAT_LITERAL         # LiteralFloat
  | STRING_LITERAL        # LiteralString
  ;

returnStatement
  : 'return' expression?
  ;

whileExpr
  : 'while' WS expression WS '{' WS? expressions WS? '}'
  ;

forExpr
  : 'for' WS forExprVariable WS 'in' WS expression WS '{' WS? expressions WS? '}'
  ;

forExprVariable
  : identifier LWS? ':' LWS? destructuringAssignmentLhsPattern        # ForExprVariableDestructuringPatternWithIdent
  | identifier LWS? ':' LWS? typeName                                 # ForExprVariableTypedIdent
  | destructuringAssignmentLhsPattern                                 # ForExprVariableDestructuringPatternOnly
  | identifier                                                        # ForExprVariableIdent
  ;

eol
  : (WS? NL)+
  ;

eos
  : (WS? (';' | NL))+ WS?
  ;


// Identifiers
packageIdentifier: LETTER ( LETTER | UNICODE_DIGIT)*;
fqPackageIdentifier: (packageIdentifier '.')* packageIdentifier;

identifier: LETTER ( LETTER | UNICODE_DIGIT)* QUESTION?;
fqIdentifier: (identifier '.')* identifier;   // fully qualified identifier

typeIdentifier: UNICODE_LETTER (LETTER | UNICODE_DIGIT)* QUESTION?;

freeTypeIdentifier: LETTER ( LETTER | UNICODE_DIGIT)*;


/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//                                                 LEXER
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// integer literals - similar to https://golang.org/ref/spec#Integer_literals
INT_LITERAL
  : DECIMAL_LITERAL INT_WIDTH?
	| BINARY_LITERAL INT_WIDTH?
	| OCTAL_LITERAL INT_WIDTH?
	| HEX_LITERAL INT_WIDTH?
  ;

fragment DECIMAL_LITERAL
  : DECIMAL_DIGIT+ ('_' | DECIMAL_DIGIT)*
  ;

fragment BINARY_LITERAL
  : '0b' BINARY_DIGIT+ ('_' | BINARY_DIGIT)*
  ;

fragment OCTAL_LITERAL
  : '0o' OCTAL_DIGIT+ ('_' | OCTAL_DIGIT)*
  ;

fragment HEX_LITERAL
  : '0x' HEX_DIGIT+ ('_' | HEX_DIGIT)*
  ;

fragment INT_WIDTH
  : 'i8' | 'i16' | 'i32' | 'i64' | 'u8' | 'u16' | 'u32' | 'u64'
  ;


// floating-point literals
FLOAT_LITERAL
  : DECIMAL_LITERAL '.' DECIMAL_LITERAL EXPONENT? FLOAT_WIDTH?
  | DECIMAL_LITERAL (EXPONENT | FLOAT_WIDTH | EXPONENT FLOAT_WIDTH)
	| '.' DECIMAL_LITERAL EXPONENT? FLOAT_WIDTH?
  ;

fragment EXPONENT
  : ( 'e' | 'E' ) ( '+' | '-' )? DECIMAL_LITERAL
  ;

fragment FLOAT_WIDTH
  : 'f32' | 'f64'
  ;

// string literals - similar to https://golang.org/ref/spec#String_literals
STRING_LITERAL
  : '"' (UNICODE_VALUE | NEWLINE | BYTE_VALUE)* '"'
  | '"""' (UNICODE_VALUE | NEWLINE | BYTE_VALUE)* '"""'   // heredoc behaves like Ruby's <<~ heredoc
  ;

UNICODE_VALUE    : UNICODE_CHAR | LITTLE_U_VALUE | BIG_U_VALUE | ESCAPED_CHAR ;
BYTE_VALUE       : OCTAL_BYTE_VALUE | HEX_BYTE_VALUE ;
OCTAL_BYTE_VALUE : '\\' OCTAL_DIGIT OCTAL_DIGIT OCTAL_DIGIT ;
HEX_BYTE_VALUE   : '\\x' HEX_DIGIT HEX_DIGIT ;
LITTLE_U_VALUE   : '\\u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT ;
BIG_U_VALUE      : '\\U' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT
                           HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT ;
ESCAPED_CHAR     : '\\' ( 'a' | 'b' | 'f' | 'n' | 'r' | 't' | 'v' | '\\' | '\'' | '"' ) ;

// Taken from https://golang.org/ref/spec#Characters

//newline = /* the Unicode code point U+000A */ .
fragment NEWLINE
  : [\u000A]
  ;

// unicode_char   = /* an arbitrary Unicode code point except newline */ .
fragment UNICODE_CHAR: ~[\u000A];

/*
The following helper program generates the UNICODE_LETTER lexer rule:
require 'nokogiri'

def unicode_min_ranges_for_general_categories(xml_doc, general_category_abbreviations = ["Lu", "Ll", "Lt", "Lm", "Lo"])
  node_set = general_category_abbreviations.map {|abbreviation| xml_doc.css("char[gc=#{abbreviation}]") }.reduce(:|)
  hex_encoded_code_points = node_set.map {|element| element.attr("cp") }
  int_encoded_code_points = hex_encoded_code_points.map(&:hex)
  int_encoded_code_points.sort!
  ranges = []
  contiguous_run = []
  prev = int_encoded_code_points.first - 1
  int_encoded_code_points.each do |i|
    if i != prev + 1
      ranges << [contiguous_run.first, contiguous_run.last]
      contiguous_run.clear
    end
    contiguous_run << i
    prev = i
  end
  ranges << [contiguous_run.first, contiguous_run.last]
  ranges
end

def convert_min_ranges_to_lexer_rules(ranges)
  ranges.map do |range|
    if range.first == range.last
      code_point = sprintf("%X", range.first).rjust(4, '0')
      "[\\u{#{code_point}}]"    # we output the extended unicode literal syntax in Antlr4 - e.g. \u{12345} - rather than the abbreviated syntax - e.g. \u04FF. See https://github.com/antlr/antlr4/blob/master/doc/unicode.md
    else
      code_point1 = sprintf("%X", range.first).rjust(4, '0')
      code_point2 = sprintf("%X", range.last).rjust(4, '0')
      "[\\u{#{code_point1}}-\\u{#{code_point2}}]"    # we output the extended unicode literal syntax in Antlr4 - e.g. \u{12345} - rather than the abbreviated syntax - e.g. \u04FF. See https://github.com/antlr/antlr4/blob/master/doc/unicode.md
    end
  end.join(" | ")
end

def min_ranges_for_numbers
  doc = File.open("ucd.all.flat.xml") { |f| Nokogiri::XML(f) }
  ranges = unicode_min_ranges_for_general_categories(doc, ["Nd"])   # Nd - see http://www.unicode.org/reports/tr44/#General_Category_Values
  puts convert_min_ranges_to_lexer_rules(ranges)
end

def min_ranges_for_letters
  doc = File.open("ucd.all.flat.xml") { |f| Nokogiri::XML(f) }
  ranges = unicode_min_ranges_for_general_categories(doc, ["Lu", "Ll", "Lt", "Lm", "Lo"])   # Lu, Ll, Lt, Lm, Lo - see http://www.unicode.org/reports/tr44/#General_Category_Values
  puts convert_min_ranges_to_lexer_rules(ranges)
end

 */

// unicode_letter = /* a Unicode code point classified as "Letter" */ .
// to generate the following list:
// 1. download https://www.unicode.org/Public/UCD/latest/ucdxml/ucd.all.flat.zip and unzip it, to produce ucd.all.flat.xml
// 2. cd to the directory where you put the ucd.all.flat.xml file, and then run `irb`
// 3. paste the ruby program noted above into irb
// 4. min_ranges_for_letters
// 5. paste the printed output in as this rule
fragment UNICODE_LETTER
  : [\u{0041}-\u{005A}] | [\u{0061}-\u{007A}] | [\u{00AA}] | [\u{00B5}] | [\u{00BA}] | [\u{00C0}-\u{00D6}] | [\u{00D8}-\u{00F6}] | [\u{00F8}-\u{02C1}] | [\u{02C6}-\u{02D1}] | [\u{02E0}-\u{02E4}] | [\u{02EC}] | [\u{02EE}] | [\u{0370}-\u{0374}] | [\u{0376}-\u{0377}] | [\u{037A}-\u{037D}] | [\u{037F}] | [\u{0386}] | [\u{0388}-\u{038A}] | [\u{038C}] | [\u{038E}-\u{03A1}] | [\u{03A3}-\u{03F5}] | [\u{03F7}-\u{0481}] | [\u{048A}-\u{052F}] | [\u{0531}-\u{0556}] | [\u{0559}] | [\u{0561}-\u{0587}] | [\u{05D0}-\u{05EA}] | [\u{05F0}-\u{05F2}] | [\u{0620}-\u{064A}] | [\u{066E}-\u{066F}] | [\u{0671}-\u{06D3}] | [\u{06D5}] | [\u{06E5}-\u{06E6}] | [\u{06EE}-\u{06EF}] | [\u{06FA}-\u{06FC}] | [\u{06FF}] | [\u{0710}] | [\u{0712}-\u{072F}] | [\u{074D}-\u{07A5}] | [\u{07B1}] | [\u{07CA}-\u{07EA}] | [\u{07F4}-\u{07F5}] | [\u{07FA}] | [\u{0800}-\u{0815}] | [\u{081A}] | [\u{0824}] | [\u{0828}] | [\u{0840}-\u{0858}] | [\u{0860}-\u{086A}] | [\u{08A0}-\u{08B4}] | [\u{08B6}-\u{08BD}] | [\u{0904}-\u{0939}] | [\u{093D}] | [\u{0950}] | [\u{0958}-\u{0961}] | [\u{0971}-\u{0980}] | [\u{0985}-\u{098C}] | [\u{098F}-\u{0990}] | [\u{0993}-\u{09A8}] | [\u{09AA}-\u{09B0}] | [\u{09B2}] | [\u{09B6}-\u{09B9}] | [\u{09BD}] | [\u{09CE}] | [\u{09DC}-\u{09DD}] | [\u{09DF}-\u{09E1}] | [\u{09F0}-\u{09F1}] | [\u{09FC}] | [\u{0A05}-\u{0A0A}] | [\u{0A0F}-\u{0A10}] | [\u{0A13}-\u{0A28}] | [\u{0A2A}-\u{0A30}] | [\u{0A32}-\u{0A33}] | [\u{0A35}-\u{0A36}] | [\u{0A38}-\u{0A39}] | [\u{0A59}-\u{0A5C}] | [\u{0A5E}] | [\u{0A72}-\u{0A74}] | [\u{0A85}-\u{0A8D}] | [\u{0A8F}-\u{0A91}] | [\u{0A93}-\u{0AA8}] | [\u{0AAA}-\u{0AB0}] | [\u{0AB2}-\u{0AB3}] | [\u{0AB5}-\u{0AB9}] | [\u{0ABD}] | [\u{0AD0}] | [\u{0AE0}-\u{0AE1}] | [\u{0AF9}] | [\u{0B05}-\u{0B0C}] | [\u{0B0F}-\u{0B10}] | [\u{0B13}-\u{0B28}] | [\u{0B2A}-\u{0B30}] | [\u{0B32}-\u{0B33}] | [\u{0B35}-\u{0B39}] | [\u{0B3D}] | [\u{0B5C}-\u{0B5D}] | [\u{0B5F}-\u{0B61}] | [\u{0B71}] | [\u{0B83}] | [\u{0B85}-\u{0B8A}] | [\u{0B8E}-\u{0B90}] | [\u{0B92}-\u{0B95}] | [\u{0B99}-\u{0B9A}] | [\u{0B9C}] | [\u{0B9E}-\u{0B9F}] | [\u{0BA3}-\u{0BA4}] | [\u{0BA8}-\u{0BAA}] | [\u{0BAE}-\u{0BB9}] | [\u{0BD0}] | [\u{0C05}-\u{0C0C}] | [\u{0C0E}-\u{0C10}] | [\u{0C12}-\u{0C28}] | [\u{0C2A}-\u{0C39}] | [\u{0C3D}] | [\u{0C58}-\u{0C5A}] | [\u{0C60}-\u{0C61}] | [\u{0C80}] | [\u{0C85}-\u{0C8C}] | [\u{0C8E}-\u{0C90}] | [\u{0C92}-\u{0CA8}] | [\u{0CAA}-\u{0CB3}] | [\u{0CB5}-\u{0CB9}] | [\u{0CBD}] | [\u{0CDE}] | [\u{0CE0}-\u{0CE1}] | [\u{0CF1}-\u{0CF2}] | [\u{0D05}-\u{0D0C}] | [\u{0D0E}-\u{0D10}] | [\u{0D12}-\u{0D3A}] | [\u{0D3D}] | [\u{0D4E}] | [\u{0D54}-\u{0D56}] | [\u{0D5F}-\u{0D61}] | [\u{0D7A}-\u{0D7F}] | [\u{0D85}-\u{0D96}] | [\u{0D9A}-\u{0DB1}] | [\u{0DB3}-\u{0DBB}] | [\u{0DBD}] | [\u{0DC0}-\u{0DC6}] | [\u{0E01}-\u{0E30}] | [\u{0E32}-\u{0E33}] | [\u{0E40}-\u{0E46}] | [\u{0E81}-\u{0E82}] | [\u{0E84}] | [\u{0E87}-\u{0E88}] | [\u{0E8A}] | [\u{0E8D}] | [\u{0E94}-\u{0E97}] | [\u{0E99}-\u{0E9F}] | [\u{0EA1}-\u{0EA3}] | [\u{0EA5}] | [\u{0EA7}] | [\u{0EAA}-\u{0EAB}] | [\u{0EAD}-\u{0EB0}] | [\u{0EB2}-\u{0EB3}] | [\u{0EBD}] | [\u{0EC0}-\u{0EC4}] | [\u{0EC6}] | [\u{0EDC}-\u{0EDF}] | [\u{0F00}] | [\u{0F40}-\u{0F47}] | [\u{0F49}-\u{0F6C}] | [\u{0F88}-\u{0F8C}] | [\u{1000}-\u{102A}] | [\u{103F}] | [\u{1050}-\u{1055}] | [\u{105A}-\u{105D}] | [\u{1061}] | [\u{1065}-\u{1066}] | [\u{106E}-\u{1070}] | [\u{1075}-\u{1081}] | [\u{108E}] | [\u{10A0}-\u{10C5}] | [\u{10C7}] | [\u{10CD}] | [\u{10D0}-\u{10FA}] | [\u{10FC}-\u{1248}] | [\u{124A}-\u{124D}] | [\u{1250}-\u{1256}] | [\u{1258}] | [\u{125A}-\u{125D}] | [\u{1260}-\u{1288}] | [\u{128A}-\u{128D}] | [\u{1290}-\u{12B0}] | [\u{12B2}-\u{12B5}] | [\u{12B8}-\u{12BE}] | [\u{12C0}] | [\u{12C2}-\u{12C5}] | [\u{12C8}-\u{12D6}] | [\u{12D8}-\u{1310}] | [\u{1312}-\u{1315}] | [\u{1318}-\u{135A}] | [\u{1380}-\u{138F}] | [\u{13A0}-\u{13F5}] | [\u{13F8}-\u{13FD}] | [\u{1401}-\u{166C}] | [\u{166F}-\u{167F}] | [\u{1681}-\u{169A}] | [\u{16A0}-\u{16EA}] | [\u{16F1}-\u{16F8}] | [\u{1700}-\u{170C}] | [\u{170E}-\u{1711}] | [\u{1720}-\u{1731}] | [\u{1740}-\u{1751}] | [\u{1760}-\u{176C}] | [\u{176E}-\u{1770}] | [\u{1780}-\u{17B3}] | [\u{17D7}] | [\u{17DC}] | [\u{1820}-\u{1877}] | [\u{1880}-\u{1884}] | [\u{1887}-\u{18A8}] | [\u{18AA}] | [\u{18B0}-\u{18F5}] | [\u{1900}-\u{191E}] | [\u{1950}-\u{196D}] | [\u{1970}-\u{1974}] | [\u{1980}-\u{19AB}] | [\u{19B0}-\u{19C9}] | [\u{1A00}-\u{1A16}] | [\u{1A20}-\u{1A54}] | [\u{1AA7}] | [\u{1B05}-\u{1B33}] | [\u{1B45}-\u{1B4B}] | [\u{1B83}-\u{1BA0}] | [\u{1BAE}-\u{1BAF}] | [\u{1BBA}-\u{1BE5}] | [\u{1C00}-\u{1C23}] | [\u{1C4D}-\u{1C4F}] | [\u{1C5A}-\u{1C7D}] | [\u{1C80}-\u{1C88}] | [\u{1CE9}-\u{1CEC}] | [\u{1CEE}-\u{1CF1}] | [\u{1CF5}-\u{1CF6}] | [\u{1D00}-\u{1DBF}] | [\u{1E00}-\u{1F15}] | [\u{1F18}-\u{1F1D}] | [\u{1F20}-\u{1F45}] | [\u{1F48}-\u{1F4D}] | [\u{1F50}-\u{1F57}] | [\u{1F59}] | [\u{1F5B}] | [\u{1F5D}] | [\u{1F5F}-\u{1F7D}] | [\u{1F80}-\u{1FB4}] | [\u{1FB6}-\u{1FBC}] | [\u{1FBE}] | [\u{1FC2}-\u{1FC4}] | [\u{1FC6}-\u{1FCC}] | [\u{1FD0}-\u{1FD3}] | [\u{1FD6}-\u{1FDB}] | [\u{1FE0}-\u{1FEC}] | [\u{1FF2}-\u{1FF4}] | [\u{1FF6}-\u{1FFC}] | [\u{2071}] | [\u{207F}] | [\u{2090}-\u{209C}] | [\u{2102}] | [\u{2107}] | [\u{210A}-\u{2113}] | [\u{2115}] | [\u{2119}-\u{211D}] | [\u{2124}] | [\u{2126}] | [\u{2128}] | [\u{212A}-\u{212D}] | [\u{212F}-\u{2139}] | [\u{213C}-\u{213F}] | [\u{2145}-\u{2149}] | [\u{214E}] | [\u{2183}-\u{2184}] | [\u{2C00}-\u{2C2E}] | [\u{2C30}-\u{2C5E}] | [\u{2C60}-\u{2CE4}] | [\u{2CEB}-\u{2CEE}] | [\u{2CF2}-\u{2CF3}] | [\u{2D00}-\u{2D25}] | [\u{2D27}] | [\u{2D2D}] | [\u{2D30}-\u{2D67}] | [\u{2D6F}] | [\u{2D80}-\u{2D96}] | [\u{2DA0}-\u{2DA6}] | [\u{2DA8}-\u{2DAE}] | [\u{2DB0}-\u{2DB6}] | [\u{2DB8}-\u{2DBE}] | [\u{2DC0}-\u{2DC6}] | [\u{2DC8}-\u{2DCE}] | [\u{2DD0}-\u{2DD6}] | [\u{2DD8}-\u{2DDE}] | [\u{2E2F}] | [\u{3005}-\u{3006}] | [\u{3031}-\u{3035}] | [\u{303B}-\u{303C}] | [\u{3041}-\u{3096}] | [\u{309D}-\u{309F}] | [\u{30A1}-\u{30FA}] | [\u{30FC}-\u{30FF}] | [\u{3105}-\u{312E}] | [\u{3131}-\u{318E}] | [\u{31A0}-\u{31BA}] | [\u{31F0}-\u{31FF}] | [\u{3400}-\u{4DB5}] | [\u{4E00}-\u{9FEA}] | [\u{A000}-\u{A48C}] | [\u{A4D0}-\u{A4FD}] | [\u{A500}-\u{A60C}] | [\u{A610}-\u{A61F}] | [\u{A62A}-\u{A62B}] | [\u{A640}-\u{A66E}] | [\u{A67F}-\u{A69D}] | [\u{A6A0}-\u{A6E5}] | [\u{A717}-\u{A71F}] | [\u{A722}-\u{A788}] | [\u{A78B}-\u{A7AE}] | [\u{A7B0}-\u{A7B7}] | [\u{A7F7}-\u{A801}] | [\u{A803}-\u{A805}] | [\u{A807}-\u{A80A}] | [\u{A80C}-\u{A822}] | [\u{A840}-\u{A873}] | [\u{A882}-\u{A8B3}] | [\u{A8F2}-\u{A8F7}] | [\u{A8FB}] | [\u{A8FD}] | [\u{A90A}-\u{A925}] | [\u{A930}-\u{A946}] | [\u{A960}-\u{A97C}] | [\u{A984}-\u{A9B2}] | [\u{A9CF}] | [\u{A9E0}-\u{A9E4}] | [\u{A9E6}-\u{A9EF}] | [\u{A9FA}-\u{A9FE}] | [\u{AA00}-\u{AA28}] | [\u{AA40}-\u{AA42}] | [\u{AA44}-\u{AA4B}] | [\u{AA60}-\u{AA76}] | [\u{AA7A}] | [\u{AA7E}-\u{AAAF}] | [\u{AAB1}] | [\u{AAB5}-\u{AAB6}] | [\u{AAB9}-\u{AABD}] | [\u{AAC0}] | [\u{AAC2}] | [\u{AADB}-\u{AADD}] | [\u{AAE0}-\u{AAEA}] | [\u{AAF2}-\u{AAF4}] | [\u{AB01}-\u{AB06}] | [\u{AB09}-\u{AB0E}] | [\u{AB11}-\u{AB16}] | [\u{AB20}-\u{AB26}] | [\u{AB28}-\u{AB2E}] | [\u{AB30}-\u{AB5A}] | [\u{AB5C}-\u{AB65}] | [\u{AB70}-\u{ABE2}] | [\u{AC00}-\u{D7A3}] | [\u{D7B0}-\u{D7C6}] | [\u{D7CB}-\u{D7FB}] | [\u{F900}-\u{FA6D}] | [\u{FA70}-\u{FAD9}] | [\u{FB00}-\u{FB06}] | [\u{FB13}-\u{FB17}] | [\u{FB1D}] | [\u{FB1F}-\u{FB28}] | [\u{FB2A}-\u{FB36}] | [\u{FB38}-\u{FB3C}] | [\u{FB3E}] | [\u{FB40}-\u{FB41}] | [\u{FB43}-\u{FB44}] | [\u{FB46}-\u{FBB1}] | [\u{FBD3}-\u{FD3D}] | [\u{FD50}-\u{FD8F}] | [\u{FD92}-\u{FDC7}] | [\u{FDF0}-\u{FDFB}] | [\u{FE70}-\u{FE74}] | [\u{FE76}-\u{FEFC}] | [\u{FF21}-\u{FF3A}] | [\u{FF41}-\u{FF5A}] | [\u{FF66}-\u{FFBE}] | [\u{FFC2}-\u{FFC7}] | [\u{FFCA}-\u{FFCF}] | [\u{FFD2}-\u{FFD7}] | [\u{FFDA}-\u{FFDC}] | [\u{10000}-\u{1000B}] | [\u{1000D}-\u{10026}] | [\u{10028}-\u{1003A}] | [\u{1003C}-\u{1003D}] | [\u{1003F}-\u{1004D}] | [\u{10050}-\u{1005D}] | [\u{10080}-\u{100FA}] | [\u{10280}-\u{1029C}] | [\u{102A0}-\u{102D0}] | [\u{10300}-\u{1031F}] | [\u{1032D}-\u{10340}] | [\u{10342}-\u{10349}] | [\u{10350}-\u{10375}] | [\u{10380}-\u{1039D}] | [\u{103A0}-\u{103C3}] | [\u{103C8}-\u{103CF}] | [\u{10400}-\u{1049D}] | [\u{104B0}-\u{104D3}] | [\u{104D8}-\u{104FB}] | [\u{10500}-\u{10527}] | [\u{10530}-\u{10563}] | [\u{10600}-\u{10736}] | [\u{10740}-\u{10755}] | [\u{10760}-\u{10767}] | [\u{10800}-\u{10805}] | [\u{10808}] | [\u{1080A}-\u{10835}] | [\u{10837}-\u{10838}] | [\u{1083C}] | [\u{1083F}-\u{10855}] | [\u{10860}-\u{10876}] | [\u{10880}-\u{1089E}] | [\u{108E0}-\u{108F2}] | [\u{108F4}-\u{108F5}] | [\u{10900}-\u{10915}] | [\u{10920}-\u{10939}] | [\u{10980}-\u{109B7}] | [\u{109BE}-\u{109BF}] | [\u{10A00}] | [\u{10A10}-\u{10A13}] | [\u{10A15}-\u{10A17}] | [\u{10A19}-\u{10A33}] | [\u{10A60}-\u{10A7C}] | [\u{10A80}-\u{10A9C}] | [\u{10AC0}-\u{10AC7}] | [\u{10AC9}-\u{10AE4}] | [\u{10B00}-\u{10B35}] | [\u{10B40}-\u{10B55}] | [\u{10B60}-\u{10B72}] | [\u{10B80}-\u{10B91}] | [\u{10C00}-\u{10C48}] | [\u{10C80}-\u{10CB2}] | [\u{10CC0}-\u{10CF2}] | [\u{11003}-\u{11037}] | [\u{11083}-\u{110AF}] | [\u{110D0}-\u{110E8}] | [\u{11103}-\u{11126}] | [\u{11150}-\u{11172}] | [\u{11176}] | [\u{11183}-\u{111B2}] | [\u{111C1}-\u{111C4}] | [\u{111DA}] | [\u{111DC}] | [\u{11200}-\u{11211}] | [\u{11213}-\u{1122B}] | [\u{11280}-\u{11286}] | [\u{11288}] | [\u{1128A}-\u{1128D}] | [\u{1128F}-\u{1129D}] | [\u{1129F}-\u{112A8}] | [\u{112B0}-\u{112DE}] | [\u{11305}-\u{1130C}] | [\u{1130F}-\u{11310}] | [\u{11313}-\u{11328}] | [\u{1132A}-\u{11330}] | [\u{11332}-\u{11333}] | [\u{11335}-\u{11339}] | [\u{1133D}] | [\u{11350}] | [\u{1135D}-\u{11361}] | [\u{11400}-\u{11434}] | [\u{11447}-\u{1144A}] | [\u{11480}-\u{114AF}] | [\u{114C4}-\u{114C5}] | [\u{114C7}] | [\u{11580}-\u{115AE}] | [\u{115D8}-\u{115DB}] | [\u{11600}-\u{1162F}] | [\u{11644}] | [\u{11680}-\u{116AA}] | [\u{11700}-\u{11719}] | [\u{118A0}-\u{118DF}] | [\u{118FF}] | [\u{11A00}] | [\u{11A0B}-\u{11A32}] | [\u{11A3A}] | [\u{11A50}] | [\u{11A5C}-\u{11A83}] | [\u{11A86}-\u{11A89}] | [\u{11AC0}-\u{11AF8}] | [\u{11C00}-\u{11C08}] | [\u{11C0A}-\u{11C2E}] | [\u{11C40}] | [\u{11C72}-\u{11C8F}] | [\u{11D00}-\u{11D06}] | [\u{11D08}-\u{11D09}] | [\u{11D0B}-\u{11D30}] | [\u{11D46}] | [\u{12000}-\u{12399}] | [\u{12480}-\u{12543}] | [\u{13000}-\u{1342E}] | [\u{14400}-\u{14646}] | [\u{16800}-\u{16A38}] | [\u{16A40}-\u{16A5E}] | [\u{16AD0}-\u{16AED}] | [\u{16B00}-\u{16B2F}] | [\u{16B40}-\u{16B43}] | [\u{16B63}-\u{16B77}] | [\u{16B7D}-\u{16B8F}] | [\u{16F00}-\u{16F44}] | [\u{16F50}] | [\u{16F93}-\u{16F9F}] | [\u{16FE0}-\u{16FE1}] | [\u{17000}-\u{187EC}] | [\u{18800}-\u{18AF2}] | [\u{1B000}-\u{1B11E}] | [\u{1B170}-\u{1B2FB}] | [\u{1BC00}-\u{1BC6A}] | [\u{1BC70}-\u{1BC7C}] | [\u{1BC80}-\u{1BC88}] | [\u{1BC90}-\u{1BC99}] | [\u{1D400}-\u{1D454}] | [\u{1D456}-\u{1D49C}] | [\u{1D49E}-\u{1D49F}] | [\u{1D4A2}] | [\u{1D4A5}-\u{1D4A6}] | [\u{1D4A9}-\u{1D4AC}] | [\u{1D4AE}-\u{1D4B9}] | [\u{1D4BB}] | [\u{1D4BD}-\u{1D4C3}] | [\u{1D4C5}-\u{1D505}] | [\u{1D507}-\u{1D50A}] | [\u{1D50D}-\u{1D514}] | [\u{1D516}-\u{1D51C}] | [\u{1D51E}-\u{1D539}] | [\u{1D53B}-\u{1D53E}] | [\u{1D540}-\u{1D544}] | [\u{1D546}] | [\u{1D54A}-\u{1D550}] | [\u{1D552}-\u{1D6A5}] | [\u{1D6A8}-\u{1D6C0}] | [\u{1D6C2}-\u{1D6DA}] | [\u{1D6DC}-\u{1D6FA}] | [\u{1D6FC}-\u{1D714}] | [\u{1D716}-\u{1D734}] | [\u{1D736}-\u{1D74E}] | [\u{1D750}-\u{1D76E}] | [\u{1D770}-\u{1D788}] | [\u{1D78A}-\u{1D7A8}] | [\u{1D7AA}-\u{1D7C2}] | [\u{1D7C4}-\u{1D7CB}] | [\u{1E800}-\u{1E8C4}] | [\u{1E900}-\u{1E943}] | [\u{1EE00}-\u{1EE03}] | [\u{1EE05}-\u{1EE1F}] | [\u{1EE21}-\u{1EE22}] | [\u{1EE24}] | [\u{1EE27}] | [\u{1EE29}-\u{1EE32}] | [\u{1EE34}-\u{1EE37}] | [\u{1EE39}] | [\u{1EE3B}] | [\u{1EE42}] | [\u{1EE47}] | [\u{1EE49}] | [\u{1EE4B}] | [\u{1EE4D}-\u{1EE4F}] | [\u{1EE51}-\u{1EE52}] | [\u{1EE54}] | [\u{1EE57}] | [\u{1EE59}] | [\u{1EE5B}] | [\u{1EE5D}] | [\u{1EE5F}] | [\u{1EE61}-\u{1EE62}] | [\u{1EE64}] | [\u{1EE67}-\u{1EE6A}] | [\u{1EE6C}-\u{1EE72}] | [\u{1EE74}-\u{1EE77}] | [\u{1EE79}-\u{1EE7C}] | [\u{1EE7E}] | [\u{1EE80}-\u{1EE89}] | [\u{1EE8B}-\u{1EE9B}] | [\u{1EEA1}-\u{1EEA3}] | [\u{1EEA5}-\u{1EEA9}] | [\u{1EEAB}-\u{1EEBB}] | [\u{20000}-\u{2A6D6}] | [\u{2A700}-\u{2B734}] | [\u{2B740}-\u{2B81D}] | [\u{2B820}-\u{2CEA1}] | [\u{2CEB0}-\u{2EBE0}] | [\u{2F800}-\u{2FA1D}]
  ;

// unicode_digit  = /* a Unicode code point classified as "Number, decimal digit" */ .
// to generate the following list:
// 1. download https://www.unicode.org/Public/UCD/latest/ucdxml/ucd.all.flat.zip and unzip it, to produce ucd.all.flat.xml
// 2. cd to the directory where you put the ucd.all.flat.xml file, and then run `irb`
// 3. paste the ruby program noted above into irb
// 4. min_ranges_for_numbers
// 5. paste the printed output in as this rule
fragment UNICODE_DIGIT
  : [\u{0030}-\u{0039}] | [\u{0660}-\u{0669}] | [\u{06F0}-\u{06F9}] | [\u{07C0}-\u{07C9}] | [\u{0966}-\u{096F}] | [\u{09E6}-\u{09EF}] | [\u{0A66}-\u{0A6F}] | [\u{0AE6}-\u{0AEF}] | [\u{0B66}-\u{0B6F}] | [\u{0BE6}-\u{0BEF}] | [\u{0C66}-\u{0C6F}] | [\u{0CE6}-\u{0CEF}] | [\u{0D66}-\u{0D6F}] | [\u{0DE6}-\u{0DEF}] | [\u{0E50}-\u{0E59}] | [\u{0ED0}-\u{0ED9}] | [\u{0F20}-\u{0F29}] | [\u{1040}-\u{1049}] | [\u{1090}-\u{1099}] | [\u{17E0}-\u{17E9}] | [\u{1810}-\u{1819}] | [\u{1946}-\u{194F}] | [\u{19D0}-\u{19D9}] | [\u{1A80}-\u{1A89}] | [\u{1A90}-\u{1A99}] | [\u{1B50}-\u{1B59}] | [\u{1BB0}-\u{1BB9}] | [\u{1C40}-\u{1C49}] | [\u{1C50}-\u{1C59}] | [\u{A620}-\u{A629}] | [\u{A8D0}-\u{A8D9}] | [\u{A900}-\u{A909}] | [\u{A9D0}-\u{A9D9}] | [\u{A9F0}-\u{A9F9}] | [\u{AA50}-\u{AA59}] | [\u{ABF0}-\u{ABF9}] | [\u{FF10}-\u{FF19}] | [\u{104A0}-\u{104A9}] | [\u{11066}-\u{1106F}] | [\u{110F0}-\u{110F9}] | [\u{11136}-\u{1113F}] | [\u{111D0}-\u{111D9}] | [\u{112F0}-\u{112F9}] | [\u{11450}-\u{11459}] | [\u{114D0}-\u{114D9}] | [\u{11650}-\u{11659}] | [\u{116C0}-\u{116C9}] | [\u{11730}-\u{11739}] | [\u{118E0}-\u{118E9}] | [\u{11C50}-\u{11C59}] | [\u{11D50}-\u{11D59}] | [\u{16A60}-\u{16A69}] | [\u{16B50}-\u{16B59}] | [\u{1D7CE}-\u{1D7FF}] | [\u{1E950}-\u{1E959}]
  ;

// taken from https://golang.org/ref/spec#Letters_and_digits
fragment LETTER: UNICODE_LETTER | '_';

fragment BINARY_DIGIT: [01];

fragment DECIMAL_DIGIT: [0-9];

fragment OCTAL_DIGIT: [0-7];

fragment HEX_DIGIT: [0-9A-Fa-f];

// Misc

fragment QUESTION: '?';

// Whitespace and comments

WS: (LWS | NL)+;

LWS: [ \t]+;

NL: [\r\n]+;

COMMENT
  : '/*' .*? '*/' -> skip
  ;

LINE_COMMENT
  : '//' ~[\r\n]* -> skip
  ;