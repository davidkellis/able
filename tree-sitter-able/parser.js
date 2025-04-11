const Parser = require('tree-sitter');
// const Able = require('tree-sitter-able');
const Able = require('./bindings/node');

const parser = new Parser();
parser.setLanguage(Able);

const sourceCode = '32';
const tree = parser.parse(sourceCode);

console.log(tree.rootNode.toString());
