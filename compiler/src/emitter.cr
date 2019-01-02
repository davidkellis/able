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

      // top level declarations
      // go here...
      EOF
    when Ast::PackageDecl
      "package #{ast.name}"
    else
      raise "boom !"
    end
  end
end