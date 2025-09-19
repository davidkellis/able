require "arborist"
require "./ast"

module Ast
  def self.build(parse_tree : Arborist::ParseTree) : Node
    matcher = Arborist::Matcher.new

    visitor = Arborist::Visitor(Node).new

    visitor.on("file") do |node|
      file_ast = File.new
      file_ast.package_declaration = node.capture("package_decl").visit(visitor).as(PackageDecl)
      
      # file_ast.package_level_declarations = node.captures("package_level_decl").map(&.visit(visitor).as(ImportDecl | AssignmentExpr))
      node.captures("package_level_decl").each do |node|
        package_level_decl_node = node.visit(visitor)
        file_ast.add_package_level_declaration(package_level_decl_node)
      end
      file_ast
    end

    visitor.on("package_decl") do |node|
      package_decl_ast = PackageDecl.new
      package_decl_ast.name = node.capture("package_ident").text
      package_decl_ast
    end

    visitor.on("package_level_decl") do |node|
      # puts node.capture.capture_name
      node.capture.visit(visitor)
    end

    # visitor.on("package_level_decl_import_decl") do |node|
    #   puts "*" * 80
    #   # todo: this is needed if we go with Option 2 in arborist/src/grammar_semantics.cr:355
    #   # node.capture("import_decl").visit(visitor)    

    #   # todo: this is needed if we go with Option 1 in arborist/src/grammar_semantics.cr:351
    #   import_decl_seq_node = node.capture("import_decl")
    #   import_decl_apply_node = import_decl_seq_node.capture("import_decl")
    #   import_decl_apply_node.visit(visitor)
    # end

    # visitor.on("package_level_decl_function_defn") do |node|
    #   # type_parameters_decl = node.capture?("type_parameters_decl")
    #   function_defn_seq_node = node.capture("function_defn")
    #   function_defn_apply_node = function_defn_seq_node.capture("function_defn")
    #   function_defn_apply_node.visit(visitor)
    # end

    
    #######################################################################################################################################
    # import_decl

    visitor.on("import_decl_all") do |node|
      package_name = node.capture("package_name").text
      package_alias = node.capture?("package_alias").try(&.text)
      ImportDecl.all(package_name, package_alias)
    end

    visitor.on("import_decl_some") do |node|
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
      package_name = node.capture("package_name").text
      package_alias = node.capture?("package_alias").try(&.text)
      ImportDecl.package(package_name, package_alias)
    end

    #######################################################################################################################################
    # function_defn

    # "fn" lws ident (lws type_parameters_decl)? function_signature ws? function_body  -- full
    visitor.on("function_defn_full") do |node|
      function_name = node.capture("ident").text
      type_parameters = node.capture?("type_parameters_decl")
      function_signature = node.capture("function_signature")
      function_body = node.capture("function_body").visit(visitor).as(Expression)

      fn_defn = FunctionDefn.new(function_name, function_body)
      fn_defn.type_parameters = get_type_parameters_from_type_parameters_decl(visitor, type_parameters) if type_parameters
      fn_defn.parameters = get_parameters_from_function_signature(visitor, function_signature)
      fn_defn
    end

    # "fn" lws ident (lws type_parameters_decl)? ws? "=>" ws? functionAliasExpression   -- alias
    visitor.on("function_defn_alias") do |node|
      raise "boom"
    end

    visitor.visit(parse_tree)
  end

  # type_parameters_decl <- "[" ws? free_type_parameter (comma_sep free_type_parameter) ws? "]"
  def self.get_type_parameters_from_type_parameters_decl(ast_visitor : Arborist::Visitor, type_parameters_decl_node : Arborist::ParseTree) : Array(TypeParameter)
    free_type_parameter_nodes = type_parameters_decl_node.captures("free_type_parameter")
    free_type_parameter_nodes.map do |free_type_parameter_node|
      type_name = free_type_parameter_node.capture("type_name").visit(ast_visitor).as(TypeName)
      type_constraint_node = free_type_parameter_node.capture?("type_constraint")
      type_constraint = type_constraint_node.visit(ast_visitor).as(TypeConstraint) if type_constraint_node
      TypeParameter.new(type_name, type_constraint)
    end
  end

  # function_signature <- "(" ws? parameter_list? ws? ")" (ws? "->" ws? type_name)?
  def self.get_parameters_from_function_signature(ast_visitor : Arborist::Visitor, function_signature_node : Arborist::ParseTree) : {parameters: Array(FnParameter), return_type: TypeName?}
    parameter_list_node = function_signature_node.capture?("parameter_list")
    parameters = Array(FnParameter).new
    if parameter_list_node
      parameters = parameter_list_node.captures("fn_parameter").map {|node| fn_parameter_to_ast(node) }
    end
    return_type = function_signature_node.capture("type_name").visit(ast_visitor).as(TypeName)
    {parameters: parameters, return_type: return_type}
  end

  def self.fn_parameter_to_ast(fn_parameter_node : Arborist::ParseTree) : FnParameter
    case top_level_alternative_label(fn_parameter_node)
    when ""
      raise "boom!"
    else
      raise ""
    end
  end

  def self.top_level_alternative_label(parse_tree : Arborist::ParseTree) : String
    parse_tree = parse_tree.child if parse_tree.is_a?(Arborist::ApplyTree)
    if child_tree_label = parse_tree.child.label()
      child_tree_label
    else
      raise "ParseTree does not have a top-level named alternative."
    end
  end
end