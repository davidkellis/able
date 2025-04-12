import * as ast from './ast';

// Helper to create simple type expressions
const type = (name: string) => ast.simpleTypeExpression(name);
const i32 = type('i32');
const f64 = type('f64');
const string = type('string');
const bool = type('bool');
const nil = type('nil');
const voidType = type('void'); // Renamed to avoid conflict with void keyword

// --- Struct Definitions ---

const Point = ast.structDefinition(
    'Point',
    [
        ast.structFieldDefinition(f64, 'x'),
        ast.structFieldDefinition(f64, 'y'),
    ],
    'named'
);

const Color = ast.structDefinition('Color', [], 'singleton'); // Example singleton
const Red = ast.structDefinition('Red', [], 'singleton');
const Green = ast.structDefinition('Green', [], 'singleton');
const Blue = ast.structDefinition('Blue', [], 'singleton');

const IntPair = ast.structDefinition(
    'IntPair',
    [
        ast.structFieldDefinition(i32), // Positional 0
        ast.structFieldDefinition(i32), // Positional 1
    ],
    'positional'
);

const GenericContainer = ast.structDefinition(
    'GenericContainer',
    [ast.structFieldDefinition(ast.simpleTypeExpression('T'), 'value')],
    'named',
    [ast.genericParameter('T')]
);

// --- Union Definitions ---

const Shape = ast.unionDefinition(
    'Shape',
    [type('Circle'), type('Rectangle')] // Assuming Circle/Rectangle structs exist
);

const OptionString = ast.unionDefinition(
    'OptionString',
    [nil, string] // Represents ?string implicitly
);

// --- Interface Definitions ---

const Display = ast.interfaceDefinition(
    'Display',
    [
        ast.functionSignature(
            'to_string',
            [ast.functionParameter('self', type('Self'))],
            string
        )
    ],
    undefined, // No generic params for interface itself
    type('T') // `for T`
);

const Adder = ast.interfaceDefinition(
    'Adder',
    [
        ast.functionSignature(
            'add',
            [ast.functionParameter('self', type('Self')), ast.functionParameter('other', type('Self'))],
            type('Self')
        )
    ],
    undefined,
    type('T')
);

// --- Function Definitions ---

const addNumbers = ast.functionDefinition(
    'addNumbers',
    [ast.functionParameter('a', i32), ast.functionParameter('b', i32)],
    ast.blockExpression([
        ast.binaryExpression('+', ast.identifier('a'), ast.identifier('b'))
    ]),
    i32
);

const greet = ast.functionDefinition(
    'greet',
    [ast.functionParameter('name', string)],
    ast.blockExpression([
        ast.assignmentExpression( // Using := for declaration
            ':=',
            ast.identifier('message'),
            ast.stringInterpolation([
                ast.stringLiteral('Hello, '),
                ast.identifier('name'),
                ast.stringLiteral('!')
            ])
        ),
        ast.functionCall(ast.identifier('print'), [ast.identifier('message')]), // Assuming print exists
        ast.returnStatement() // Explicit return void
    ]),
    voidType
);

const genericIdentity = ast.functionDefinition(
    'identity',
    [ast.functionParameter('val', type('T'))],
    ast.blockExpression([ast.identifier('val')]),
    type('T'),
    [ast.genericParameter('T')]
);

// --- Methods Definition ---

const PointMethods = ast.methodsDefinition(
    type('Point'),
    [
        // Instance method using shorthand `fn #method`
        ast.functionDefinition(
            'distance_origin', // Original name before shorthand processing
            [], // Params excluding self
            ast.blockExpression([
                // Assuming sqrt and pow are available
                ast.functionCall(ast.identifier('sqrt'), [
                    ast.binaryExpression('+',
                        ast.functionCall(ast.identifier('pow'), [ast.memberAccessExpression(ast.identifier('self'), 'x'), ast.integerLiteral(2)]),
                        ast.functionCall(ast.identifier('pow'), [ast.memberAccessExpression(ast.identifier('self'), 'y'), ast.integerLiteral(2)])
                    )
                ])
            ]),
            f64,
            undefined, undefined,
            true // isMethodShorthand = true
        ),
        // Static method (constructor pattern)
        ast.functionDefinition(
            'origin',
            [],
            ast.blockExpression([
                ast.structLiteral(
                    [
                        ast.structFieldInitializer(ast.floatLiteral(0.0), 'x'),
                        ast.structFieldInitializer(ast.floatLiteral(0.0), 'y')
                    ],
                    false, // isPositional = false
                    'Point'
                )
            ]),
            type('Point') // Returns Self (Point)
        )
    ]
);

// --- Implementation Definition ---

const PointDisplayImpl = ast.implementationDefinition(
    'Display',
    type('Point'),
    [
        ast.functionDefinition(
            'to_string',
            [ast.functionParameter('self', type('Self'))], // Self is Point here
            ast.blockExpression([
                ast.stringInterpolation([
                    ast.stringLiteral('Point('),
                    ast.memberAccessExpression(ast.identifier('self'), 'x'),
                    ast.stringLiteral(', '),
                    ast.memberAccessExpression(ast.identifier('self'), 'y'),
                    ast.stringLiteral(')')
                ])
            ]),
            string
        )
    ]
);

const i32AdderImpl = ast.implementationDefinition(
    'Adder',
    i32,
    [
        ast.functionDefinition(
            'add',
            [ast.functionParameter('self', type('Self')), ast.functionParameter('other', type('Self'))],
            ast.blockExpression([
                ast.binaryExpression('+', ast.identifier('self'), ast.identifier('other'))
            ]),
            type('Self') // Returns i32
        )
    ]
);


// --- Main Program Logic (example within a 'main' function) ---

const mainFunction = ast.functionDefinition(
    'main',
    [],
    ast.blockExpression([
        // Literals
        ast.assignmentExpression(':=', ast.identifier('myInt'), ast.integerLiteral(42, 'i32')),
        ast.assignmentExpression(':=', ast.identifier('myFloat'), ast.floatLiteral(3.14)),
        ast.assignmentExpression(':=', ast.identifier('myBool'), ast.booleanLiteral(true)),
        ast.assignmentExpression(':=', ast.identifier('myNil'), ast.nilLiteral()),
        ast.assignmentExpression(':=', ast.identifier('myChar'), ast.charLiteral('A')),
        ast.assignmentExpression(':=', ast.identifier('myString'), ast.stringLiteral("Hello")),
        ast.assignmentExpression(':=', ast.identifier('myArray'), ast.arrayLiteral([ast.integerLiteral(1), ast.integerLiteral(2)])),

        // Struct Instantiation
        ast.assignmentExpression(':=', ast.identifier('p1'), ast.structLiteral(
            [
                ast.structFieldInitializer(ast.floatLiteral(1.0), 'x'),
                ast.structFieldInitializer(ast.floatLiteral(2.5), 'y')
            ],
            false, 'Point'
        )),
        ast.assignmentExpression(':=', ast.identifier('pair1'), ast.structLiteral(
            [
                ast.structFieldInitializer(ast.integerLiteral(10)),
                ast.structFieldInitializer(ast.integerLiteral(20))
            ],
            true, 'IntPair'
        )),
        // Functional Update
        ast.assignmentExpression(':=', ast.identifier('p2'), ast.structLiteral(
            [ast.structFieldInitializer(ast.floatLiteral(5.0), 'x')], // Override x
            false, 'Point', ast.identifier('p1') // Source ...p1
        )),

        // Member Access
        ast.assignmentExpression(':=', ast.identifier('p1_x'), ast.memberAccessExpression(ast.identifier('p1'), 'x')),
        ast.assignmentExpression(':=', ast.identifier('pair1_0'), ast.memberAccessExpression(ast.identifier('pair1'), ast.integerLiteral(0))), // Access positional

        // Mutation
        ast.assignmentExpression('=', ast.memberAccessExpression(ast.identifier('p1'), 'y'), ast.floatLiteral(3.0)),

        // Operators
        ast.assignmentExpression(':=', ast.identifier('sumResult'), ast.binaryExpression('+', ast.identifier('myInt'), ast.integerLiteral(8))),
        ast.assignmentExpression(':=', ast.identifier('isGreater'), ast.binaryExpression('>', ast.identifier('myFloat'), ast.floatLiteral(3.0))),
        ast.assignmentExpression(':=', ast.identifier('negated'), ast.unaryExpression('-', ast.identifier('myInt'))),
        ast.assignmentExpression(':=', ast.identifier('logicalNot'), ast.unaryExpression('!', ast.identifier('myBool'))),

        // Function Calls
        ast.assignmentExpression(':=', ast.identifier('addRes'), ast.functionCall(ast.identifier('addNumbers'), [ast.integerLiteral(5), ast.integerLiteral(3)])),
        ast.functionCall(ast.identifier('greet'), [ast.stringLiteral("World")]),
        ast.assignmentExpression(':=', ast.identifier('idRes'), ast.functionCall(ast.identifier('identity'), [ast.stringLiteral("test")], [string])), // Explicit generic arg

        // Method Call Syntax (Instance & Static)
        ast.assignmentExpression(':=', ast.identifier('originPt'), ast.functionCall(ast.memberAccessExpression(ast.identifier('Point'), 'origin'), [])), // Point.origin()
        ast.assignmentExpression(':=', ast.identifier('p1_display'), ast.functionCall(ast.memberAccessExpression(ast.identifier('p1'), 'to_string'), [])), // p1.to_string()
        ast.functionCall(ast.identifier('print'), [ast.identifier('p1_display')]), // Assuming print exists

        // Control Flow: if/or
        ast.assignmentExpression(':=', ast.identifier('grade'), ast.ifExpression(
            ast.binaryExpression('>=', ast.identifier('myInt'), ast.integerLiteral(90)),
            ast.blockExpression([ast.stringLiteral("A")]),
            [
                ast.orClause(
                    ast.blockExpression([ast.stringLiteral("B")]),
                    ast.binaryExpression('>=', ast.identifier('myInt'), ast.integerLiteral(80))
                ),
                ast.orClause( // Final else
                    ast.blockExpression([ast.stringLiteral("C or lower")])
                )
            ]
        )),

        // Control Flow: match
        ast.assignmentExpression(':=', ast.identifier('optionVal'), ast.identifier('myNil')), // Example Option value (nil)
        ast.assignmentExpression(':=', ast.identifier('matchResult'), ast.matchExpression(
            ast.identifier('optionVal'),
            [
                ast.matchClause( // case s: string => ... (represented by identifier pattern)
                    ast.identifier('s'),
                    ast.stringInterpolation([ast.stringLiteral("Got string: "), ast.identifier('s')])
                ),
                ast.matchClause( // case nil => ...
                    ast.literalPattern(ast.nilLiteral()),
                    ast.stringLiteral("Got nil")
                ),
                // Assuming Circle/Rectangle struct patterns
                // ast.matchClause(ast.structPattern([], false, 'Circle'), ast.stringLiteral("Circle")),
                // ast.matchClause(ast.structPattern([], false, 'Rectangle'), ast.stringLiteral("Rectangle")),
                ast.matchClause( // case _ => ...
                    ast.wildcardPattern(),
                    ast.stringLiteral("Unknown shape or value")
                )
            ]
        )),

        // Control Flow: while
        ast.assignmentExpression(':=', ast.identifier('counter'), ast.integerLiteral(0)),
        ast.whileLoop(
            ast.binaryExpression('<', ast.identifier('counter'), ast.integerLiteral(3)),
            ast.blockExpression([
                ast.functionCall(ast.identifier('print'), [ast.identifier('counter')]),
                ast.assignmentExpression('=', ast.identifier('counter'), ast.binaryExpression('+', ast.identifier('counter'), ast.integerLiteral(1)))
            ])
        ),

        // Control Flow: for
        ast.forLoop(
            ast.identifier('i'), // pattern
            ast.rangeExpression(ast.integerLiteral(1), ast.integerLiteral(3), true), // iterable (1..3)
            ast.blockExpression([
                ast.functionCall(ast.identifier('print'), [ast.stringInterpolation([ast.stringLiteral("Loop "), ast.identifier('i')])])
            ])
        ),

        // Control Flow: breakpoint/break
        ast.assignmentExpression(':=', ast.identifier('searchResult'), ast.breakpointExpression(
            'finder',
            ast.blockExpression([
                ast.assignmentExpression(':=', ast.identifier('data'), ast.arrayLiteral([ast.integerLiteral(1), ast.integerLiteral(-5), ast.integerLiteral(8)])),
                ast.forLoop(
                    ast.identifier('item'),
                    ast.identifier('data'),
                    ast.blockExpression([
                        ast.ifExpression(
                            ast.binaryExpression('<', ast.identifier('item'), ast.integerLiteral(0)),
                            ast.blockExpression([
                                ast.breakStatement('finder', ast.identifier('item')) // break 'finder item
                            ])
                        )
                    ])
                ),
                ast.nilLiteral() // Default value if loop completes
            ])
        )),

        // Error Handling: !, else
        // Assume `maybeGetString()` returns ?string (nil | string)
        // Assume `mustGetString()` returns !string (string | Error)
        ast.assignmentExpression(':=', ast.identifier('optStr'), ast.functionCall(ast.identifier('maybeGetString'), [])),
        ast.assignmentExpression(':=', ast.identifier('strOrDefault'), ast.orElseExpression( // optStr else { "default" }
            ast.identifier('optStr'),
            ast.blockExpression([ast.stringLiteral("default")])
        )),
        // Assume `processString` takes string, returns !bool
        ast.assignmentExpression(':=', ast.identifier('processResult'), ast.blockExpression([
             ast.assignmentExpression(':=', ast.identifier('s'), ast.propagationExpression( // mustGetString()!
                 ast.functionCall(ast.identifier('mustGetString'), [])
             )),
             ast.propagationExpression( // processString(s)!
                 ast.functionCall(ast.identifier('processString'), [ast.identifier('s')])
             )
             // Need return type compatible with propagated error
        ])),


        // Error Handling: raise/rescue (Conceptual - assumes Error types)
        ast.assignmentExpression(':=', ast.identifier('safeDivideResult'), ast.rescueExpression(
            ast.functionCall(ast.identifier('divide'), [ast.integerLiteral(10), ast.integerLiteral(0)]), // Assume divide raises error
            [
                ast.matchClause(
                    ast.identifier('e'), // Catch specific error type if possible, otherwise identifier/wildcard
                    ast.blockExpression([
                        ast.functionCall(ast.identifier('print'), [ast.stringLiteral("Caught error!")]),
                        ast.integerLiteral(-1) // Default value
                    ]),
                    // Guard example: ast.binaryExpression('==', ast.memberAccessExpression(ast.identifier('e'), 'type'), ast.stringLiteral("DivideByZero"))
                )
            ]
        )),

        // Lambda Expression
        ast.assignmentExpression(':=', ast.identifier('incrementer'), ast.lambdaExpression(
            [ast.functionParameter('x', i32)],
            ast.binaryExpression('+', ast.identifier('x'), ast.integerLiteral(1)),
            i32
        )),
        ast.assignmentExpression(':=', ast.identifier('incResult'), ast.functionCall(ast.identifier('incrementer'), [ast.integerLiteral(100)])),

        // Concurrency (Basic structure)
        ast.assignmentExpression(':=', ast.identifier('procHandle'), ast.procExpression(
            ast.functionCall(ast.identifier('someAsyncTask'), [])
        )),
        ast.assignmentExpression(':=', ast.identifier('spawnThunk'), ast.spawnExpression(
            ast.blockExpression([
                ast.functionCall(ast.identifier('print'), [ast.stringLiteral("Spawned task running...")]),
                ast.integerLiteral(500) // Result of spawned block
            ])
        )),
        // Getting result from spawn implicitly blocks:
        ast.assignmentExpression(':=', ast.identifier('spawnResult'), ast.identifier('spawnThunk')),


        ast.functionCall(ast.identifier('print'), [ast.stringLiteral("Sample program finished.")])
    ]),
    voidType // Main returns void
);

// --- Module Definition ---
// Combine all definitions into a module
const sampleModule = ast.module(
    [
        // Definitions first
        Point,
        Color, Red, Green, Blue, // Singletons
        IntPair,
        GenericContainer,
        Shape,
        OptionString,
        Display,
        Adder,
        addNumbers,
        greet,
        genericIdentity,
        PointMethods,
        PointDisplayImpl,
        i32AdderImpl,
        // Main execution logic
        mainFunction
    ],
    [
        // Example imports (assuming stdlib structure)
        ast.importStatement(['core']), // import core;
        ast.importStatement(['collections'], false, undefined, 'col'), // import collections as col;
        ast.importStatement(['io'], false, [ast.importSelector('print')]), // import io.{print};
    ]
    // Optional package statement: ast.packageStatement(['sample'])
);

// Export the final AST module (optional, useful for testing/inspection)
export default sampleModule;

// To use this file, you would typically import it and pass `sampleModule`
// to your interpreter or compiler.
// Example (in another file):
// import sampleProgramAst from './sample1';
// interpret(sampleProgramAst);
