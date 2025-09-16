import * as AST from "../ast";

// Interface Definition
const formatterInterface = AST.interfaceDefinition(
    "Formatter",
    [
        AST.functionSignature(
            "format",
            [AST.functionParameter("self", AST.simpleTypeExpression("Self"))], // self: Self
            AST.simpleTypeExpression("string") // -> string
        )
    ],
    undefined, // No generic params for interface itself
    AST.simpleTypeExpression("T") // for T (generic over implementing type)
);

// Struct User
const userStruct = AST.structDefinition(
    "User",
    [
        AST.structFieldDefinition(AST.simpleTypeExpression("string"), "name"),
        AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "age")
    ],
    'named'
);

// Struct Product
const productStruct = AST.structDefinition(
    "Product",
    [
        AST.structFieldDefinition(AST.simpleTypeExpression("string"), "id"),
        AST.structFieldDefinition(AST.simpleTypeExpression("f64"), "price")
    ],
    'named'
);

// Implementation for User
const userImpl = AST.implementationDefinition(
    "Formatter", // Interface name
    AST.simpleTypeExpression("User"), // Target type
    [ // Definitions
        AST.functionDefinition(
            "format", // Method name
            [AST.functionParameter("self", AST.simpleTypeExpression("Self"))], // Params
            AST.blockExpression([ // Body
                AST.stringInterpolation([
                    AST.stringLiteral("User(name="),
                    AST.memberAccessExpression(AST.identifier("self"), "name"), // self.name
                    AST.stringLiteral(", age="),
                    AST.memberAccessExpression(AST.identifier("self"), "age"), // self.age
                    AST.stringLiteral(")")
                ])
            ]),
            AST.simpleTypeExpression("string") // Return type
        )
    ]
);

// Implementation for Product
const productImpl = AST.implementationDefinition(
    "Formatter", // Interface name
    AST.simpleTypeExpression("Product"), // Target type
    [ // Definitions
        AST.functionDefinition(
            "format", // Method name
            [AST.functionParameter("self", AST.simpleTypeExpression("Self"))], // Params
            AST.blockExpression([ // Body
                AST.stringInterpolation([
                    AST.stringLiteral("Product(id="),
                    AST.memberAccessExpression(AST.identifier("self"), "id"), // self.id
                    AST.stringLiteral(", price="),
                    AST.memberAccessExpression(AST.identifier("self"), "price"), // self.price
                    AST.stringLiteral(")")
                ])
            ]),
            AST.simpleTypeExpression("string") // Return type
        )
    ]
);

// Main program logic
const program = AST.module([
    formatterInterface,
    userStruct,
    productStruct,
    userImpl,
    productImpl,
    // Create instances
    AST.assignmentExpression(
        ":=",
        AST.identifier("user1"),
        AST.structLiteral(
            [
                AST.structFieldInitializer(AST.stringLiteral("Alice"), "name"),
                AST.structFieldInitializer(AST.integerLiteral(30), "age")
            ],
            false, // Not positional
            "User"
        )
    ),
    AST.assignmentExpression(
        ":=",
        AST.identifier("prod1"),
        AST.structLiteral(
            [
                AST.structFieldInitializer(AST.stringLiteral("abc-123"), "id"),
                AST.structFieldInitializer(AST.floatLiteral(99.95), "price")
            ],
            false, // Not positional
            "Product"
        )
    ),
    // Call format methods directly
    AST.assignmentExpression(
        ":=",
        AST.identifier("user_str"),
        AST.functionCall(
            AST.memberAccessExpression(AST.identifier("user1"), "format"), // user1.format()
            []
        )
    ),
    AST.assignmentExpression(
        ":=",
        AST.identifier("prod_str"),
        AST.functionCall(
            AST.memberAccessExpression(AST.identifier("prod1"), "format"), // prod1.format()
            []
        )
    ),
    // Print results
    AST.functionCall(AST.identifier("print"), [AST.identifier("user_str")]),
    AST.functionCall(AST.identifier("print"), [AST.identifier("prod_str")]),
]);

export default program;
