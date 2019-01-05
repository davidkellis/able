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
    property package_level_declarations : Array(ImportDecl | AssignmentExpr)

    def initialize
      @package_level_declarations = Array(ImportDecl | AssignmentExpr).new
    end

    def package_declaration : PackageDecl
      @package_declaration || raise "File#package_declaration undefined"
    end

    def package_declaration=(package_decl : PackageDecl)
      @package_declaration = package_decl
      self
    end

    def add_package_level_declaration(package_level_declaration : Node)
      case package_level_declaration
      when ImportDecl
        @package_level_declarations.push(package_level_declaration)
      when AssignmentExpr
        @package_level_declarations.push(package_level_declaration)
      else
        raise "Unexpected package level declaration: #{typeof(package_level_declaration)}"
      end
    end

    def children : Array(Node)
      [package_declaration].concat(@package_level_declarations)
    end

    def s_exp(indent : Int32 = 0) : String
      prefix = " " * indent
      package_decl_sexp = package_declaration.s_exp(indent+2)
      package_level_decl_sexp = package_level_declarations.map(&.s_exp(indent+2)).join("\n")
      "#{prefix}(file\n#{package_decl_sexp}\n#{package_level_decl_sexp})"
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

  class ImportDecl < Node
    getter type : Symbol    # :all (import all members from package), :some (import specified members from package), :package (import only package name)
    getter package_name : String    # a fully-qualified package name
    getter package_alias : String?  # a single-component package name
    getter members : Hash(String, String)  # a list of members to import from `@package_name`; the key is the member name to import, the value is the alias (in the case that no member alias is defined, the member name is the alias)

    def self.all(fq_package_name : String, package_alias : String?)
      new(:all, fq_package_name, package_alias, Hash(String, String).new)
    end

    def self.some(fq_package_name : String, package_alias : String?, members : Hash(String, String))
      new(:some, fq_package_name, package_alias, members)
    end

    def self.package(fq_package_name : String, package_alias : String?)
      new(:package, fq_package_name, package_alias, Hash(String, String).new)
    end

    def initialize(@type, @package_name, @package_alias, @members)
    end

    def s_exp(indent : Int32 = 0) : String
      prefix = " " * indent
      "#{prefix}(import_decl #{@type} #{@package_name} #{@package_alias || "nil"} [#{@members.map{|k,v| "#{k}=>#{v}" }.join(",")}])"
    end
  end

  class AssignmentExpr < Node
  end

  def self.build(parse_tree : Arborist::ParseTree) : Node
    matcher = Arborist::Matcher.new

    visitor = Arborist::Visitor(Node).new

    visitor.on("file") do |node|
      file_ast = File.new
      file_ast.package_declaration = node.capture("PackageDecl").visit(visitor).as(PackageDecl)
      # file_ast.package_level_declarations = node.captures("PackageLevelDecl").map(&.visit(visitor).as(ImportDecl | AssignmentExpr))
      node.captures("PackageLevelDecl").each do |node|
        # todo: fix this; we want to visit the PackageLevelDecl node, which should return an ImportDecl node or a AssignmentExpr node, which we then want to add to the 
        package_level_decl_node = node.visit(visitor)
        file_ast.add_package_level_declaration(package_level_decl_node)
      end
      file_ast
    end

    visitor.on("PackageDecl") do |node|
      package_decl_ast = PackageDecl.new
      package_decl_ast.name = node.capture("package_ident").text
      package_decl_ast
    end

    visitor.on("PackageLevelDecl_import_decl") do |node|
      puts "*" * 80
      node.capture("import_decl").visit(visitor)
      # import_decl_seq_node = node.capture("import_decl")
      # import_decl_apply_node = import_decl_seq_node.capture("import_decl")
      # import_decl_apply_node.visit(visitor)
    end

    visitor.on("import_decl_all") do |node|
      puts "1" * 80
      package_name = node.capture("package_name").text
      package_alias = node.capture?("package_alias").try(&.text)
      ImportDecl.all(package_name, package_alias)
    end

    visitor.on("import_decl_some") do |node|
      puts "2" * 80
      package_name = node.capture("package_name").text
      package_alias = node.capture?("package_alias").try(&.text)
      members : Hash(String,String) = node.captures("members").map do |member_node|
        member_name = member_node.capture("member_name").text
        member_alias = member_node.capture?("member_alias").try(&.text) || member_name    # if no member alias is defined, then we define the alias to be the original member name
        [member_name, member_alias]
      end.to_h
      ImportDecl.some(package_name, package_alias, members)
    end

    visitor.on("import_decl_package") do |node|
      puts "3" * 80
      package_name = node.capture("package_name").text
      package_alias = node.capture?("package_alias").try(&.text)
      ImportDecl.package(package_name, package_alias)
    end

    visitor.visit(parse_tree)
  end
end