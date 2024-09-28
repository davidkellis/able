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

  class Package < Node
    alias Declaration = ImportDecl | FunctionDefn

    property path : String
    property declarations : Array(Declaration)

    def initialize
      @declarations = Array(Declaration).new
    end

    def add_declaration(declaration : Declaration)
      @declarations.push(declaration)
    end

    def s_exp(indent : Int32 = 0) : String
      prefix = " " * indent
      decl_sexps = declarations.map(&.s_exp(indent+2)).join("\n")
      "#{prefix}(package (name #{path.to_s})\n#{decl_sexps})"
    end
  end

  class ImportDecl < Node
    getter package_name : String            # a fully-qualified package name
    getter package_alias : String?          # a single-component package name
    getter members : Hash(String, String)   # a list of members to import from `@package_name`; the key is the member name to import, the value is the alias (in the case that no member alias is defined, the member name is the alias)

    def initialize(@package_name, @package_alias, @members)
    end

    def s_exp(indent : Int32 = 0) : String
      prefix = " " * indent
      "#{prefix}(import #{@package_name} #{@package_alias || "nil"} (members #{@members.map{|k,v| "(#{k} -> #{v})" }.join(",")})"
    end
  end

  class FunctionDefn < Node
    property name : String
    property type_parameters : Array(TypeParameter)
    property parameters : Array(FnParameter)
    property body : Block
    property return_type : Type

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
    @type_constraint : ImplementsTypeConstraint?

    def initialize(@type_name, @type_constraint)
    end
  end

  class ImplementsTypeConstraint < Node
    property type_names : Array(TypeName)

    def initialize(@type_names : Array(TypeName))
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