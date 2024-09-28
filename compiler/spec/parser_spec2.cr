require "./spec_helper"
require "arborist"

GRAMMAR = Arborist::Grammar.new("./able2.arborist")

describe "Parser" do
  it "recognizes comments" do
    src = <<-SRC
    # foo
    # bar
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)

    src = <<-SRC
    # foo
    /*
    blah blah
    */
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  describe "expressions rule" do
    it "recognizes primary_expression" do
      src = <<-SRC
# parenthesized expression
(123)

# literal
()
nil
true
false
123.123
123
"foo"
SRC
      parse_tree = GRAMMAR.parse(src, "unit_test_expressions")
      GRAMMAR.print_match_failure_error unless parse_tree
      parse_tree.should_not eq(nil)
    end

    it "recognizes variable declarations" do
      src = <<-SRC
a : Int
b : Foo Bar
SRC
      parse_tree = GRAMMAR.parse(src, "unit_test_expressions")
      GRAMMAR.print_match_failure_error unless parse_tree
      parse_tree.should_not eq(nil)
    end

    it "recognizes operator shorthand assignment" do
      src = <<-SRC
a += 5
b /= 10
SRC
      parse_tree = GRAMMAR.parse(src, "unit_test_expressions")
      GRAMMAR.print_match_failure_error unless parse_tree
      parse_tree.should_not eq(nil)
    end

    it "recognizes simple assignment" do
      src = <<-SRC
a = 5
b = Array("foo", "bar")
a, b = 1, 2
a, b = 1, Array(1,2,3)
v = expr
v: type = expr
v1: type, v2: type = expr, expr
v: i32
SRC
      # Arborist::GlobalDebug.enable!
      parse_tree = GRAMMAR.parse(src, "unit_test_expressions")
      GRAMMAR.print_match_failure_error unless parse_tree
      # Arborist::GlobalDebug.disable!
      parse_tree.should_not eq(nil)
    end

    it "recognizes complex assignment" do
      src = <<-SRC
v1: Int, v2: Array Int = 5, Array(1,2,3)
v1: Int, v2: Array Int = 5, Array Int(1,2,3)
(v1, v2) = (1, 2)
Array[v1, v2, v3] = expr
Array[v1, v2, v3*] = expr
MyStruct{v1, v2} = expr
MyStruct{age = v1, date = v2} = expr
v += expr
(v1, v2) = expr
SRC
      parse_tree = GRAMMAR.parse(src, "unit_test_expressions")
      GRAMMAR.print_match_failure_error unless parse_tree
      parse_tree.should_not eq(nil)
    end

    it "recognizes function calls" do
      src = <<-SRC
a(b,c(d))
a(b,c(d)).e()
a(b,c(d)).e(f())
a(b).e(f())
SRC
      # Arborist::GlobalDebug.enable!
      parse_tree = GRAMMAR.parse(src, "unit_test_expressions")
      GRAMMAR.print_match_failure_error unless parse_tree
      # Arborist::GlobalDebug.disable!
      parse_tree.should_not eq(nil)
    end

  end

end