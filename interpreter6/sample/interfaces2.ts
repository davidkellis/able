import * as AST from "../ast";
import { interpret } from "../interpreter";

// Create a module that demonstrates interfaces and implementations
function createInterfacesModule(): AST.Module {
  // Define a Shape interface with methods for area and perimeter
  const shapeInterface = AST.interfaceDefinition(
    "Shape",
    [
      // Method signature for area
      AST.functionSignature(
        "area",
        [AST.functionParameter(AST.identifier("self"))],
        AST.simpleTypeExpression("f64")
      ),
      // Method signature for perimeter
      AST.functionSignature(
        "perimeter",
        [AST.functionParameter(AST.identifier("self"))],
        AST.simpleTypeExpression("f64")
      ),
      // Method signature for description
      AST.functionSignature(
        "description",
        [AST.functionParameter(AST.identifier("self"))],
        AST.simpleTypeExpression("string")
      )
    ]
  );

  // Define a Circle struct
  const circleStruct = AST.structDefinition(
    "Circle",
    [
      AST.structFieldDefinition(AST.simpleTypeExpression("f64"), "radius"),
      AST.structFieldDefinition(AST.simpleTypeExpression("f64"), "x"),
      AST.structFieldDefinition(AST.simpleTypeExpression("f64"), "y")
    ],
    "named"
  );

  // Define a Rectangle struct
  const rectangleStruct = AST.structDefinition(
    "Rectangle",
    [
      AST.structFieldDefinition(AST.simpleTypeExpression("f64"), "width"),
      AST.structFieldDefinition(AST.simpleTypeExpression("f64"), "height")
    ],
    "named"
  );

  // Implement Shape for Circle
  const circleShapeImpl = AST.implementationDefinition(
    "Shape",
    AST.simpleTypeExpression("Circle"),
    [
      // Area implementation
      AST.functionDefinition(
        "area",
        [AST.functionParameter(AST.identifier("self"))],
        AST.blockExpression([
          AST.returnStatement(
            AST.binaryExpression(
              "*",
              AST.binaryExpression(
                "*",
                AST.memberAccessExpression(AST.identifier("self"), AST.identifier("radius")),
                AST.memberAccessExpression(AST.identifier("self"), AST.identifier("radius"))
              ),
              AST.floatLiteral(3.14159)
            )
          )
        ]),
        AST.simpleTypeExpression("f64")
      ),
      // Perimeter implementation
      AST.functionDefinition(
        "perimeter",
        [AST.functionParameter(AST.identifier("self"))],
        AST.blockExpression([
          AST.returnStatement(
            AST.binaryExpression(
              "*",
              AST.binaryExpression(
                "*",
                AST.memberAccessExpression(AST.identifier("self"), AST.identifier("radius")),
                AST.floatLiteral(2)
              ),
              AST.floatLiteral(3.14159)
            )
          )
        ]),
        AST.simpleTypeExpression("f64")
      ),
      // Description implementation
      AST.functionDefinition(
        "description",
        [AST.functionParameter(AST.identifier("self"))],
        AST.blockExpression([
          AST.returnStatement(
            AST.stringInterpolation([
              AST.stringLiteral("Circle with radius "),
              AST.memberAccessExpression(AST.identifier("self"), AST.identifier("radius"))
            ])
          )
        ]),
        AST.simpleTypeExpression("string")
      )
    ]
  );

  // Implement Shape for Rectangle
  const rectangleShapeImpl = AST.implementationDefinition(
    "Shape",
    AST.simpleTypeExpression("Rectangle"),
    [
      // Area implementation
      AST.functionDefinition(
        "area",
        [AST.functionParameter(AST.identifier("self"))],
        AST.blockExpression([
          AST.returnStatement(
            AST.binaryExpression(
              "*",
              AST.memberAccessExpression(AST.identifier("self"), AST.identifier("width")),
              AST.memberAccessExpression(AST.identifier("self"), AST.identifier("height"))
            )
          )
        ]),
        AST.simpleTypeExpression("f64")
      ),
      // Perimeter implementation
      AST.functionDefinition(
        "perimeter",
        [AST.functionParameter(AST.identifier("self"))],
        AST.blockExpression([
          AST.returnStatement(
            AST.binaryExpression(
              "*",
              AST.binaryExpression(
                "+",
                AST.memberAccessExpression(AST.identifier("self"), AST.identifier("width")),
                AST.memberAccessExpression(AST.identifier("self"), AST.identifier("height"))
              ),
              AST.floatLiteral(2)
            )
          )
        ]),
        AST.simpleTypeExpression("f64")
      ),
      // Description implementation
      AST.functionDefinition(
        "description",
        [AST.functionParameter(AST.identifier("self"))],
        AST.blockExpression([
          AST.returnStatement(
            AST.stringInterpolation([
              AST.stringLiteral("Rectangle with width "),
              AST.memberAccessExpression(AST.identifier("self"), AST.identifier("width")),
              AST.stringLiteral(" and height "),
              AST.memberAccessExpression(AST.identifier("self"), AST.identifier("height"))
            ])
          )
        ]),
        AST.simpleTypeExpression("string")
      )
    ]
  );

  // Define a function that works with any Shape
  const printShapeInfo = AST.functionDefinition(
    "printShapeInfo",
    [AST.functionParameter(AST.identifier("shape"))],
    AST.blockExpression([
      AST.functionCall(
        AST.identifier("print"),
        [
          AST.stringInterpolation([
            AST.stringLiteral("Shape info: "),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("shape"), AST.identifier("description")),
              []
            )
          ])
        ]
      ),
      AST.functionCall(
        AST.identifier("print"),
        [
          AST.stringInterpolation([
            AST.stringLiteral("Area: "),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("shape"), AST.identifier("area")),
              []
            )
          ])
        ]
      ),
      AST.functionCall(
        AST.identifier("print"),
        [
          AST.stringInterpolation([
            AST.stringLiteral("Perimeter: "),
            AST.functionCall(
              AST.memberAccessExpression(AST.identifier("shape"), AST.identifier("perimeter")),
              []
            )
          ])
        ]
      )
    ])
  );

  // Main function to demonstrate interfaces and implementations
  const main = AST.functionDefinition(
    "main",
    [],
    AST.blockExpression([
      // Create a circle
      AST.assignmentExpression(
        ":=",
        AST.identifier("circle"),
        AST.structLiteral(
          [
            AST.structFieldInitializer(AST.floatLiteral(5), "radius"),
            AST.structFieldInitializer(AST.floatLiteral(0), "x"),
            AST.structFieldInitializer(AST.floatLiteral(0), "y")
          ],
          false,
          "Circle"
        )
      ),
      // Create a rectangle
      AST.assignmentExpression(
        ":=",
        AST.identifier("rectangle"),
        AST.structLiteral(
          [
            AST.structFieldInitializer(AST.floatLiteral(10), "width"),
            AST.structFieldInitializer(AST.floatLiteral(5), "height")
          ],
          false,
          "Rectangle"
        )
      ),
      // Print info about the circle
      AST.functionCall(
        AST.identifier("print"),
        [AST.stringLiteral("Circle info:")]
      ),
      AST.functionCall(
        AST.identifier("printShapeInfo"),
        [AST.identifier("circle")]
      ),
      // Print info about the rectangle
      AST.functionCall(
        AST.identifier("print"),
        [AST.stringLiteral("\nRectangle info:")]
      ),
      AST.functionCall(
        AST.identifier("printShapeInfo"),
        [AST.identifier("rectangle")]
      )
    ])
  );

  // Create the module with all definitions and the main function
  return AST.module(
    [
      shapeInterface,
      circleStruct,
      rectangleStruct,
      circleShapeImpl,
      rectangleShapeImpl,
      printShapeInfo,
      main
    ]
  );
}

// Run the program
export default createInterfacesModule();
