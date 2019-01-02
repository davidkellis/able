require "arborist"

module Ast
  class Node
    property parent : Node?
    
    # children returns the list of all actual/literal child Node nodes in AST tree structure, 
    def children : Array(Node)
      raise "Node#children is an abstract method"
    end

    def child : Node
      children.first
    end

    def child? : Node?
      children.first?
    end

    # returns all descendants listed out in a pre-order traversal
    def descendants : Array(Node)
      nodes = [] of Node
      accumulate = ->(node : Node) { nodes << node }
      children.each {|child| child.preorder_traverse(accumulate) }
      nodes
    end

    def s_exp(indent : Int32 = 0) : String
      raise "Node#s_exp is an abstract method"
    end

    def preorder_traverse(visit : Node -> _)
      visit.call(self)
      children.each {|child| child.preorder_traverse(visit) }
    end

    def postorder_traverse(&visit : Node -> _)
      postorder_traverse(visit)
    end

    def postorder_traverse(visit : Node -> _)
      children.each {|child| child.postorder_traverse(visit) }
      visit.call(self)
    end

    def root?
      @parent.nil?
    end

    # returns all nodes listed out in a pre-order traversal
    def self_and_descendants : Array(Node)
      ([self] of Node).concat(descendants)
    end

  end

  class File < Node
    @package_declaration : PackageDecl?

    def package_declaration : PackageDecl
      @package_declaration || raise "File#package_declaration undefined"
    end

    def package_declaration=(package_decl : PackageDecl)
      @package_declaration = package_decl
      self
    end

    def children : Array(Node)
      [package_declaration]
    end

    def s_exp(indent : Int32 = 0) : String
      prefix = " " * indent
      "#{prefix}(file\n#{package_declaration.s_exp(indent+2)})"
    end
  end

  class PackageDecl < Node
    @name : String?

    def name : String
      @name || raise "PackageDecl#name undefined"
    end

    def name=(name : String)
      @name = name
      self
    end

    def s_exp(indent : Int32 = 0) : String
      prefix = " " * indent
      "#{prefix}(package_decl #{name})"
    end
  end

  def self.build(parse_tree : Arborist::ParseTree) : Node
    matcher = Arborist::Matcher.new

    visitor = Arborist::Visitor(Node).new

    visitor.on("File") do |node|
      file_ast = File.new
      file_ast.package_declaration = node.capture("PackageDecl").visit(visitor).as(PackageDecl)
      file_ast
    end

    visitor.on("PackageDecl") do |node|
      package_decl_ast = PackageDecl.new
      package_decl_ast.name = node.capture("package_ident").text
      package_decl_ast
    end

    visitor.visit(parse_tree)
  end
end