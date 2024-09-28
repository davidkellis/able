// import Parser from 'tree-sitter';
// import * as tsj from 'tree-sitter-javascript';

const Parser = require('tree-sitter');
const JavaScript = require('tree-sitter-javascript');

// const JavaScript = tsj.language;

const parser = new Parser();
parser.setLanguage(JavaScript);

const sourceCode = 'let x = 1; console.log(x);';
const tree = parser.parse(sourceCode);

console.log(tree.rootNode.toString());
