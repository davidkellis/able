import { AST } from "../../context";
import type { Fixture } from "../../types";

const stdlibFixtures: Fixture[] = [
  {
      name: "stdlib/channel_mutex_helpers",
      module: AST.module([
        AST.structDefinition(
          "Channel",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "capacity"),
            AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "handle"),
          ],
          "named",
        ),
        AST.methodsDefinition(
          AST.simpleTypeExpression("Channel"),
          [
            AST.functionDefinition(
              "new",
              [],
              AST.blockExpression([
                AST.assignmentExpression(
                  ":=",
                  AST.identifier("handle"),
                  AST.functionCall(AST.identifier("__able_channel_new"), [
                    AST.integerLiteral(0),
                  ]),
                ),
                AST.returnStatement(
                  AST.structLiteral(
                    [
                      AST.structFieldInitializer(AST.integerLiteral(0), "capacity"),
                      AST.structFieldInitializer(AST.identifier("handle"), "handle"),
                    ],
                    false,
                    "Channel",
                  ),
                ),
              ]),
              AST.simpleTypeExpression("Channel"),
            ),
          ],
        ),
        AST.structDefinition(
          "Mutex",
          [AST.structFieldDefinition(AST.simpleTypeExpression("i64"), "handle")],
          "named",
        ),
        AST.methodsDefinition(
          AST.simpleTypeExpression("Mutex"),
          [
            AST.functionDefinition(
              "new",
              [],
              AST.blockExpression([
                AST.assignmentExpression(
                  ":=",
                  AST.identifier("handle"),
                  AST.functionCall(AST.identifier("__able_mutex_new"), []),
                ),
                AST.returnStatement(
                  AST.structLiteral(
                    [AST.structFieldInitializer(AST.identifier("handle"), "handle")],
                    false,
                    "Mutex",
                  ),
                ),
              ]),
              AST.simpleTypeExpression("Mutex"),
            ),
          ],
        ),
        AST.fn(
          "channel_handle",
          [AST.param("capacity", AST.simpleTypeExpression("i32"))],
          [AST.ret(AST.call("__able_channel_new", AST.id("capacity")))],
          AST.simpleTypeExpression("i64"),
        ),
        AST.fn(
          "mutex_handle",
          [],
          [AST.ret(AST.call("__able_mutex_new"))],
          AST.simpleTypeExpression("i64"),
        ),
        AST.assign(
          "channel_handle_value",
          AST.call("channel_handle", AST.integerLiteral(0)),
        ),
        AST.assign(
          "channel_instance",
          AST.call(
            AST.memberAccessExpression(AST.identifier("Channel"), "new"),
          ),
        ),
        AST.assign(
          "mutex_instance",
          AST.call(
            AST.memberAccessExpression(AST.identifier("Mutex"), "new"),
          ),
        ),
        AST.assign(
          "mutex_handle_value",
          AST.call("mutex_handle"),
        ),
        AST.assign("score", AST.integerLiteral(0)),
        AST.ifExpression(
          AST.binaryExpression(
            "!=",
            AST.memberAccessExpression(AST.identifier("channel_instance"), "handle"),
            AST.integerLiteral(0),
          ),
          AST.blockExpression([
            AST.assignmentExpression(
              "+=",
              AST.identifier("score"),
              AST.integerLiteral(1),
            ),
          ]),
          [],
        ),
        AST.ifExpression(
          AST.binaryExpression(
            "==",
            AST.memberAccessExpression(AST.identifier("channel_instance"), "capacity"),
            AST.integerLiteral(0),
          ),
          AST.blockExpression([
            AST.assignmentExpression(
              "+=",
              AST.identifier("score"),
              AST.integerLiteral(1),
            ),
          ]),
          [],
        ),
        AST.ifExpression(
          AST.binaryExpression(
            "!=",
            AST.identifier("channel_handle_value"),
            AST.integerLiteral(0),
          ),
          AST.blockExpression([
            AST.assignmentExpression(
              "+=",
              AST.identifier("score"),
              AST.integerLiteral(1),
            ),
          ]),
          [],
        ),
        AST.ifExpression(
          AST.binaryExpression(
            "!=",
            AST.memberAccessExpression(AST.identifier("mutex_instance"), "handle"),
            AST.integerLiteral(0),
          ),
          AST.blockExpression([
            AST.assignmentExpression(
              "+=",
              AST.identifier("score"),
              AST.integerLiteral(1),
            ),
          ]),
          [],
        ),
        AST.ifExpression(
          AST.binaryExpression(
            "!=",
            AST.identifier("mutex_handle_value"),
            AST.integerLiteral(0),
          ),
          AST.blockExpression([
            AST.assignmentExpression(
              "+=",
              AST.identifier("score"),
              AST.integerLiteral(1),
            ),
          ]),
          [],
        ),
        AST.binaryExpression(
          "==",
          AST.identifier("score"),
          AST.integerLiteral(5),
        ),
      ]),
      manifest: {
        description: "Channel.new and Mutex.new expose runtime handles via stdlib helpers",
        expect: {
          result: { kind: "bool", value: true },
        },
      },
    },
];

export default stdlibFixtures;
