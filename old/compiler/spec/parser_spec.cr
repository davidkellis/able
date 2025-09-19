require "./spec_helper"
require "arborist"

GRAMMAR = Arborist::Grammar.new("./able.arborist")

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

  it "recognizes package declaration" do
    src = <<-SRC
    package foo
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes package level imports" do
    src = <<-SRC
    import io
    import io.*
    import io.{puts, gets}
    import io.{puts as p}
    import internationalization as i18n
    import internationalization as i18n.{Unicode}
    import davidkellis.matrix as mat.{Matrix, Vector as Vec}
    import davidkellis.matrix as mat.{Matrix, Vector as Vec,}
    import davidkellis.matrix as mat.{
      Matrix as Mat,
      Vector as Vec,
    }
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes package level variable declaration" do
    src = <<-SRC
    package foo

    a : i32
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes package level variable assignment" do
    src = <<-SRC
    a = ()
    a = nil
    a = true
    a = false
    a = 5
    a = 5i8
    a = 5u64
    a = 0b0101100
    a = 0b0101100u16
    a = 1.2
    a = 1.2e9f32
    a = 1e9_f32
    a = 8f64
    a = 8_f64
    a = ""
    a = "foo"
    a = "Here Be Dragons©"
    
    # operator shorthand assignment
    a += 1
    a /= 1
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes package level struct definitions" do
    src = <<-SRC
    struct Foo T [T: Iterable] { Int, Float, T }
    struct Foo T { Int, Float, T, }
    struct Foo T {
      Int,
      Float,
      T,
    }
    struct Foo T {
      Int
      Float
      T
    }
    struct Foo T [T: Iterable] { x: Int, y: Float, z: T }
    struct Foo T { x: Int, y: Float, z: T, }
    struct Foo T {
      x: Int,
      x: Float,
      y: T,
    }
    struct Foo T {
      x: Int
      y: Float
      z: T
    }
    struct None
    struct SmallHouse { sqft: Float }
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes package level union definitions" do
    src = <<-SRC
    union String? = String | Nil
    union House = SmallHouse { sqft: Float } | BigHouse { sqft: Float }
    union House = SmallHouse { sqft: Float }
    | MediumHouse { sqft: Float }
    | LargeHouse { sqft: Float }
    union Foo T [T: Blah] = 
    | Bar A [A: Stringable] { a: A, t: T }
    | Baz B [B: Qux] { b: B, t: T }
    union Option A = Some A {A} | None
    union Result A B = Success A {A} | Failure B {B}
    union ContrivedResult A B [A: Fooable, B: Barable] = Success A X [X: Stringable] {A, X} | Failure B Y [Y: Serializable] {B, Y}
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes package level function definitions" do
    src = <<-SRC
    fn foo() { a = 5 }
    fn foo() {
      a = 5
    }
    fn foo() -> i32 { a = 5 }
    fn foo() { puts(5) }
    fn foo() {
      puts(5)
    }
    fn foo(self: Self, f: T -> Unit) -> Unit { 5 }
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes package level interface definitions" do
    src = <<-SRC
    interface Stringable for S {
      fn to_s(S) -> String
    }
    interface Stringable for S {
      fn to_s(S) -> String
      fn inspect(S) -> String { "foo" }
    }
    interface Stringable for S {
      fn to_s(Self) -> String
    }
    interface Iterable T for I {
      fn each(self: Self, f: T -> Unit) -> Unit
    }
    interface Comparable for T {
      fn compare(T, T) -> i32
    }
    interface Mappable for M _ {
      fn map(m: M A, convert: A -> B) -> M B
    }
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes package level interface implementations" do
    src = <<-SRC
    impl Stringable for Foo { }
    impl Stringable for Foo {
      fn to_s(Self) -> String { "foo" }
    }
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes hello world program" do
    src = <<-SRC
    fn main() {
      io.puts("hello world")
    }
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes various forms of assignment within a function definition" do
    src = <<-SRC
    fn main() {
      a = ()
      a = nil
      a = true
      a = false
      a = 5
      a = 5i8
      a = 5u64
      a = 0b0101100
      a = 0b0101100u16
      a = 1.2
      a = 1.2e9f32
      a = 1e9_f32
      a = 8f64
      a = 8_f64
      a = ""
      a = "foo"
      a = "Here Be Dragons©"
      
      # operator shorthand assignment
      a += 1
      a /= 1

      # arr(5) = 123
    }
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  it "recognizes various expressions within a function definition" do
    src = <<-SRC
    fn main() {
      a while b
      a until b
      a if b
      a unless b
      a match {
        Foo => 1
        Bar => 2
        _ => 3
      }
      a |> b |> c
      a || b || c
      a && b && c
      a == b == c
      a <= b <= c
      a >= b >= c
      a..b
      a...b
    }
SRC
    parse_tree = GRAMMAR.parse(src)
    GRAMMAR.print_match_failure_error unless parse_tree
    parse_tree.should_not eq(nil)
  end

  describe "recognizes expressions" do
    it "recognizes expressions" do
      src = <<-SRC
      a while b
      a until b
      a if b
      a unless b
      a match {
        Foo => 1
        Bar => 2
        _ => 3
      }
      a = 1
      a |> b |> c
      a || b || c
      a && b && c
      a == b == c
      a <= b <= c
      a >= b >= c
      a..b
      a...b
      a+a+a
      a*a*a
      a^a^a
      foo()
      a.b.c()
      (a + b)
      (a && b)
SRC
      src = <<-SRC
(a && b)
SRC
      # Arborist::GlobalDebug.enable!
      parse_tree = GRAMMAR.parse(src, "unit_test_expressions")
      GRAMMAR.print_match_failure_error unless parse_tree
      # Arborist::GlobalDebug.disable!
      parse_tree.should_not eq(nil)
    end

    it "recognizes simple_expressions" do
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

# control flow
return
break
continue

# jumps and jumppoints
#jumppoint :foo { 123 }
#jump :foo 123
SRC
      parse_tree = GRAMMAR.parse(src, "unit_test_expressions")
      GRAMMAR.print_match_failure_error unless parse_tree
      parse_tree.should_not eq(nil)
    end
  end

end