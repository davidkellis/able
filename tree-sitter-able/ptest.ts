import { AbleParser } from './parser';

const sourceCode = '32';
const parser = new AbleParser();
const tree = parser.parse(sourceCode);

console.log(tree.rootNode.toString());

// we could run parser.ts like this:
// ❯ tsx ptest.ts
// (source_file (expression_statement (integer_literal)))
