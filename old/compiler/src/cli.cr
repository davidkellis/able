t1 = Time.now
puts "starting up"
require "option_parser"
require "arborist"

require "./ast"
require "./ast_builder"
require "./emitter"
require "./compiler"
t2 = Time.now; puts "#{t2-t1} seconds elapsed"; t1 = t2;

def run(input_file_path)
  t1 = Time.now
  puts "loading grammar:"
  grammar = Arborist::Grammar.new("./able.arborist")
  t2 = Time.now; puts "#{t2-t1} seconds elapsed"; t1 = t2;

  puts "parsing input file:"
  parse_tree = grammar.parse_file(input_file_path)
  t2 = Time.now; puts "#{t2-t1} seconds elapsed"; t1 = t2;

  if parse_tree
    puts "parse tree:"
    puts parse_tree.s_exp
    t2 = Time.now; puts "#{t2-t1} seconds elapsed"; t1 = t2;

    ast = Ast.build(parse_tree)
    puts "ast:"
    puts ast.s_exp
    t2 = Time.now; puts "#{t2-t1} seconds elapsed"; t1 = t2;

    emitter = GoEmitter.new
    puts
    puts "golang:"
    puts emitter.emit(ast)
    t2 = Time.now; puts "#{t2-t1} seconds elapsed"; t1 = t2;
  else
    grammar.print_match_failure_error
    STDERR.puts "Failed parsing."
    exit(1)
  end
rescue e
  STDERR.puts "Failed to parse: #{e.message}"
  STDERR.puts e.backtrace.join("\n") if Config.debug
  exit(1)
end

class Config
  class_property debug : Bool = false
  class_property input_file : String?
  class_property args : Array(String) = [] of String
end

def main
  OptionParser.parse! do |parser|
    parser.banner = "Usage: able -f my_program.able"
    parser.on("-d", "Enable debug mode.") { Config.debug = true }
    parser.on("-f my_program.able", "Specifies the source code file") { |file_name| Config.input_file = file_name }
    parser.on("-h", "--help", "Show this help") { puts parser; exit }
    parser.invalid_option do |flag|
      STDERR.puts "ERROR: #{flag} is not a valid option."
      STDERR.puts parser
      exit(1)
    end
    parser.unknown_args do |unknown_args|
      Config.args = unknown_args
    end
  end

  (STDERR.puts("Input file must be specified") ; exit(1)) unless Config.input_file
  (STDERR.puts("Input file \"#{Config.input_file}\" does not exist") ; exit(1)) unless File.exists?(Config.input_file.as(String))

  run(Config.input_file.as(String))
end

main