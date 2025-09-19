class Interpreter
  module ObjectInterface
    def inspect : String
      "Object{id=#{object_id}}"
    end

    def type : String
      raise "Object#type not implemented"
    end
  end

  class Bindings
    property @bindings : Hash(string, Object)
    
    def initialize()
      @bindings = Hash(string, Object).new
    end

    def add(name, value)
      @bindings[name] = value
    end

    def remove(name)
      @bindings.delete(name)
    end
  end

  alias Object = Bool | U8 | I8 | I32 | I64 | F32 | F64 | Bool | String | Function

  class Bool
    include ObjectInterface

    property value : ::Bool

    def initialize(@value)
    end

    def inspect
      "#{@value}"
    end

    def type : String
      "bool"
    end
  end

  class Number(T)
    include ObjectInterface

    property value : T

    def initialize(@value)
    end

    def inspect
      "#{@value}"
    end
  end

  class Int(T) < Number(T)
  end

  class Float(T) < Number(T)
  end

  class U8 < Int(::UInt8)
    def type
      "u8"
    end
  end

  class I8 < Int(::Int8)
    def type
      "i8"
    end
  end

  class I32 < Int(::Int32)
    def type
      "i32"
    end
  end

  class I64 < Int(::Int64)
    def type
      "i64"
    end
  end

  class F32 < Float(::Float32)
    def type
      "f32"
    end
  end

  class F64 < Float(::Float64)
    def type
      "f64"
    end
  end

  class String
    include ObjectInterface

    property value : ::String

    def initialize(@value)
    end

    def inspect
      "\"#{@value}\""
    end

    def type : String
      "string"
    end
  end

  class Block
    property expressions : Array(Ast::Expression)

    def initialize(@expressions)
    end
  end

  class Function
    include ObjectInterface

    property parameter_names : Array(string)
    property bindings : Bindings
    property body : Block

    def initialize(@parameter_names, @bindings, @body)
    end

    def inspect
      "\"#{@value}\""
    end

    def type : String
      "string"
    end
  end


  property global_bindings : Bindings

  def initialize
    @global_bindings = Bindings.new

  end

  def eval(node : Ast::Node) : Object
    case node
    when Ast::Package
    when Ast::Integer
    when Ast::Float
    when Ast::Branch
    when Ast::Loop
    when Ast::Return
    when Ast::Break
      
    end
  end
end
