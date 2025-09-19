import os
from parglare import Grammar, GLRParser, Parser

def main(debug=False):
  this_folder = os.path.dirname(__file__)
  grammar_file = os.path.join(this_folder, 'able3.pg')
  g = Grammar.from_file(grammar_file, debug=debug, debug_colors=True)
  # parser = Parser(g, build_tree=True, debug=debug, debug_colors=True)
  parser = GLRParser(g, build_tree=True, debug=debug, debug_colors=True)

  with open(os.path.join(this_folder, 'able3.txt'), 'r') as f:
    results = parser.parse(f.read())
    print("{} sucessful parse(s).\n".format(len(results)))
    for result in results:
      print("***")
      print(result.tree_str())
    print("\n{} sucessful parse(s).".format(len(results)))


if __name__ == '__main__':
  main(debug=False)