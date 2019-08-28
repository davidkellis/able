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

  class FunctionDefn < Node
    @name : String
    @type_parameters : Array(TypeParameter)
    @parameters : Array(FnParameter)
    @body : Expression
    @return_type : TypeName?

    def initialize(@name, @body)
      @type_parameters = Array(TypeParameter).new
      @parameters = Array(FnParameter).new
    end

    def type_parameters=(@type_parameters : Array(TypeParameter))
      self
    end

    def parameters=(@parameters : Array(FnParameter))
      self
    end
  end

  class TypeParameter < Node
    @type_name : TypeName
    @type_constraint : TypeConstraint?

    def initialize(@type_name, @type_constraint)
    end
  end

  # abstract
  class TypeConstraint < Node
  end

  class ImplementsTypeConstraint < TypeConstraint
    @type_name : TypeName

    def initialize(@type_name : TypeName)
    end
  end

  class TypeName < Node
    @name : String
    @type_parameters : Array(TypeName)
    
    def initialize(@name : String, @type_parameters : Array(TypeName))
    end

    def placeholder?
      @name == "_"
    end
  end

  class FnParameter < Node
    @name : String
    @type : TypeName

    def initialize(@name, @type)
    end
  end

  # abstract
  class Expression < Node
  end

  class BlockExpr < Expression
    @expressions : Array(Expression)

    def initialize(@expressions)
    end
  end

end