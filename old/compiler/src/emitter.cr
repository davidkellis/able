require "./ast"

module Emitter
end

class GoEmitter
  # this is intended to be called with an Ast::File argument
  def emit(ast : Ast::Node) : String
    case ast
    when Ast::File
      <<-EOF
      // package declaration
      #{emit(ast.package_declaration)}

      // package level declarations
      #{ast.package_level_declarations.map {|decl_node| emit(decl_node).as(String) }.join("\n")}
      EOF
    when Ast::PackageDecl
      "package #{ast.name}"
    when Ast::ImportDecl
      case ast.type
      when :all
        "import all"    # todo: finish these import lines
      when :some
        "import some"
      when :package
        "import package"
      else
        raise "unexpected ImportDecl type"
      end
    else
      raise "boom !"
    end
  end
end