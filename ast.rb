class PackageDefinition
  attr_accessor :name, :expressions

  def initialize(name, expressions)
    @name = name
    @expressions = expressions
  end

  def imports
    @imports ||= expressions.select {|e| e.is_a? Import }
  end
end

class Import
  attr_accessor :package_name, :package_alias, :identifiers

  # identifiers is either an array of string identifiers, or a map of string => string identifiers and corresponding aliases
  def initialize(package_name, package_alias, identifiers)
    @package_name, @package_alias, @identifiers = package_name, package_alias, identifiers
  end
end

class FunctionDefinition
  attr_accessor :type_parameters, :function_name, :parameters, :return_type, :expressions

  def initialize(type_parameters, function_name, parameters, return_type, expressions)
    @type_parameters, @function_name, @parameters, @return_type, @expressions = type_parameters, function_name, parameters, return_type, expressions
  end
end

class FunctionCall
  attr_accessor :function_name, :arguments

  def initialize(function_name, arguments)
    @function_name, @arguments = function_name, arguments
  end
end

class InferredType
end

class StringLiteral
  attr_accessor :value

  def initialize(value)
    @type = type
    @value = value
  end
end

class Identifier
  attr_accessor :name

  def initialize(name)
    @name = name
  end
end

class RubyEmitter
  attr_accessor :ast, :out

  def emit(ast, out = $stdout)
    @ast, @out = ast, out

    case ast
    when PackageDefinition
      build_package_definition(ast)
    end
  end

  def emit_package_definition(package_definition)
    buffer = <<-EOF
package #{package_definition.name}

#{emit_import()}
    EOF
  end

  def emit_import(import, indent_level = 0)
  end
end


class GoEmitter
  attr_accessor :ast, :out

  def emit(ast, out = $stdout)
    @ast, @out = ast, out

    case ast
    when PackageDefinition
      build_package_definition(ast)
    end
  end

  def emit_package_definition(package_definition)
    buffer = <<-EOF
package #{package_definition.name}

#{emit_import()}
    EOF
  end

  def emit_import(import, indent_level = 0)
  end
end


def hello_world_ast
# package main
# import io.puts
# fn main() {
#   puts("hello world")
# } keyb

  ast = PackageDefinition.new("main", [
    Import.new("io", nil, ["puts"]),
    FunctionDefinition.new("main", [], InferredType.new, [
      FunctionCall.new("puts", [
        Literal.new("hello world")
      ])
    ])
  ])
  emitter = GoEmitter.new
  emitter.emit(ast)
end

hello_world_ast
