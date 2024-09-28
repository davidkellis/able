import os
import sys
from pprint import pprint
import inspect
import parglare
from parglare import Grammar, GLRParser, visitor
from parglare.trees import tree_node_iterator



def main(debug=False):
  this_folder = os.path.dirname(__file__)
  grammar_file = os.path.join(this_folder, sys.argv[1])
  g = Grammar.from_file(grammar_file, debug=debug, debug_colors=True)
  parser = GLRParser(g, debug=debug, debug_colors=True)

  # pprint(inspect.getmembers(g.classes["package_decl"]))
  # PackageDeclClass = g.classes["package_decl"]


  # def node_to_str(root):
  #     from parglare.glr import Parent

  #     def visit(n, subresults, depth):
  #         indent = '  ' * depth
  #         if isinstance(n, Parent):
  #             s = f'{indent}{n.head.symbol} - ambiguity[{n.ambiguity}]'
  #             for idx, p in enumerate(subresults):
  #                 s += f'\n{indent}{idx+1}:{p}'
  #         elif n.is_nonterm():
  #             s = '{}{}[{}->{}, {}]'.format(indent, n.production.symbol,
  #                                           n.start_position,
  #                                           n.end_position, isinstance(n.production, PackageDeclClass))
  #             # if hasattr(n, "production"):
  #             #   s = '{}{}[{}->{}, {}]'.format(indent, n.production.symbol,
  #             #                                 n.start_position,
  #             #                                 n.end_position, n._pg_attrs)
  #             # else:
  #             #   s = '{}{}[{}->{}, None]'.format(indent, n.production.symbol,
  #             #                             n.start_position,
  #             #                             n.end_position)

  #             if subresults:
  #                 s = '{}\n{}'.format(s, '\n'.join(subresults))
  #         else:
  #             s = '{}{}[{}->{}, "{}"]'.format(indent, n.symbol,
  #                                             n.start_position,
  #                                             n.end_position,
  #                                             n.value)
  #         return s
  #     return visitor(root, tree_node_iterator, visit)

  # parglare.trees.to_str = node_to_str



  input_file_path = sys.argv[2]
  forest = parser.parse_file(input_file_path)
  parse_tree = forest.get_first_tree()
  # ast = parser.call_actions(parse_tree)
  print(parse_tree.to_str())
  # print(ast.to_str())

if __name__ == '__main__':
  main(debug=False)
