//////////////////////////////////////////////////////////////////////////////////////////////////////////
// parser.js originally was this:

// const Parser = require('tree-sitter');
// const Able = require('./bindings/node');
// const parser = new Parser();
// parser.setLanguage(Able);

// const sourceCode = '32';
// const tree = parser.parse(sourceCode);

// console.log(tree.rootNode.toString());

// we could run parser.js like this:
// ❯ node parser.js
// (source_file (expression_statement (integer_literal)))

//////////////////////////////////////////////////////////////////////////////////////////////////////////
// then we revised the parser.js to this typescript version:

import Parser from 'tree-sitter';
import Able from './bindings/node';



export class AbleParser {
  private parser: Parser;

  constructor() {
    this.parser = new Parser();
    this.parser.setLanguage(Able);
  }

  parse(source) {
    return this.parser.parse(source);
  }
}

// const sourceCode = '32';
// const parser = new AbleParser();
// const tree = parser.parse(sourceCode);

// console.log(tree.rootNode.toString());

// we could run parser.ts like this:
// ❯ tsx parser.ts
// (source_file (expression_statement (integer_literal)))


export default AbleParser;


// module.exports = AbleParser;
