import AbleInterpreter from './interpreter';

const interpreter = new AbleInterpreter();

const source = `
42;
-123_i8;
0xff_u16;
3.14;
"hello";
true;
'a';
nil;
`;

const results = interpreter.interpret(source);
console.log(results);
