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

  {
      name: "stdlib/channel_iterator",
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
              [AST.functionParameter("capacity", AST.simpleTypeExpression("i32"))],
              AST.blockExpression([
                AST.assignmentExpression(
                  ":=",
                  AST.identifier("handle"),
                  AST.functionCall(AST.identifier("__able_channel_new"), [
                    AST.identifier("capacity"),
                  ]),
                ),
                AST.returnStatement(
                  AST.structLiteral(
                    [
                      AST.structFieldInitializer(AST.identifier("capacity"), "capacity"),
                      AST.structFieldInitializer(AST.identifier("handle"), "handle"),
                    ],
                    false,
                    "Channel",
                  ),
                ),
              ]),
              AST.simpleTypeExpression("Channel"),
            ),
            AST.functionDefinition(
              "send",
              [AST.functionParameter("self"), AST.functionParameter("value")],
              AST.blockExpression([
                AST.functionCall(
                  AST.identifier("__able_channel_send"),
                  [
                    AST.memberAccessExpression(AST.identifier("self"), "handle"),
                    AST.identifier("value"),
                  ],
                ),
              ]),
            ),
            AST.functionDefinition(
              "receive",
              [AST.functionParameter("self")],
              AST.blockExpression([
                AST.returnStatement(
                  AST.functionCall(
                    AST.identifier("__able_channel_receive"),
                    [AST.memberAccessExpression(AST.identifier("self"), "handle")],
                  ),
                ),
              ]),
            ),
            AST.functionDefinition(
              "close",
              [AST.functionParameter("self")],
              AST.blockExpression([
                AST.functionCall(
                  AST.identifier("__able_channel_close"),
                  [AST.memberAccessExpression(AST.identifier("self"), "handle")],
                ),
              ]),
            ),
            AST.functionDefinition(
              "iterator",
              [AST.functionParameter("self")],
              AST.blockExpression([
                AST.returnStatement(
                  AST.iteratorLiteral(
                    [
                      AST.whileLoop(
                        AST.booleanLiteral(true),
                        AST.blockExpression([
                          AST.assignmentExpression(
                            ":=",
                            AST.identifier("received"),
                            AST.functionCall(
                              AST.memberAccessExpression(AST.identifier("self"), "receive"),
                              [],
                            ),
                          ),
                          AST.matchExpression(
                            AST.identifier("received"),
                            [
                              AST.matchClause(
                                AST.typedPattern(
                                  AST.identifier("value"),
                                  AST.simpleTypeExpression("i32"),
                                ),
                                AST.blockExpression([
                                  AST.functionCall(
                                    AST.memberAccessExpression(AST.identifier("gen"), "yield"),
                                    [AST.identifier("value")],
                                  ),
                                ]),
                              ),
                              AST.matchClause(
                                AST.literalPattern(AST.nilLiteral()),
                                AST.blockExpression([
                                  AST.functionCall(
                                    AST.memberAccessExpression(AST.identifier("gen"), "stop"),
                                    [],
                                  ),
                                ]),
                              ),
                            ],
                          ),
                        ]),
                      ),
                    ],
                    undefined,
                    AST.simpleTypeExpression("i32"),
                  ),
                ),
              ]),
            ),
          ],
        ),
        AST.assign(
          "channel",
          AST.call(
            AST.memberAccessExpression(AST.identifier("Channel"), "new"),
            AST.integerLiteral(2),
          ),
        ),
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "send"),
          [AST.integerLiteral(10)],
        ),
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "send"),
          [AST.integerLiteral(20)],
        ),
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("channel"), "close"),
          [],
        ),
        AST.assign(
          "iter",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("channel"), "iterator"),
            [],
          ),
        ),
        AST.assign("sum", AST.integerLiteral(0)),
        AST.assign(
          "current",
          AST.functionCall(AST.memberAccessExpression(AST.identifier("iter"), "next"), []),
        ),
        AST.whileLoop(
          AST.booleanLiteral(true),
          AST.blockExpression([
            AST.matchExpression(
              AST.identifier("current"),
              [
                AST.matchClause(
                  AST.structPattern([], false, "IteratorEnd"),
                  AST.blockExpression([AST.breakStatement()]),
                ),
                AST.matchClause(
                  AST.typedPattern(
                    AST.identifier("value"),
                    AST.simpleTypeExpression("i32"),
                  ),
                  AST.blockExpression([
                    AST.assignmentExpression(
                      "+=",
                      AST.identifier("sum"),
                      AST.identifier("value"),
                    ),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("current"),
                      AST.functionCall(
                        AST.memberAccessExpression(AST.identifier("iter"), "next"),
                        [],
                      ),
                    ),
                  ]),
                ),
                AST.matchClause(
                  AST.wildcardPattern(),
                  AST.blockExpression([AST.breakStatement()]),
                ),
              ],
            ),
          ]),
        ),
        AST.assign(
          "second_channel",
          AST.call(
            AST.memberAccessExpression(AST.identifier("Channel"), "new"),
            AST.integerLiteral(1),
          ),
        ),
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("second_channel"), "send"),
          [AST.integerLiteral(42)],
        ),
        AST.functionCall(
          AST.memberAccessExpression(AST.identifier("second_channel"), "close"),
          [],
        ),
        AST.assign(
          "iter_single",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("second_channel"), "iterator"),
            [],
          ),
        ),
        AST.assign(
          "first",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("iter_single"), "next"),
            [],
          ),
        ),
        AST.assign(
          "second",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("iter_single"), "next"),
            [],
          ),
        ),
        AST.assign(
          "third",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("iter_single"), "next"),
            [],
          ),
        ),
        AST.arrayLiteral([
          AST.identifier("sum"),
          AST.identifier("first"),
          AST.identifier("second"),
          AST.identifier("third"),
        ]),
      ]),
      manifest: {
        description: "Channel.iterator yields lazily until receive() returns nil",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "i32", value: 30 },
              { kind: "i32", value: 42 },
              { kind: "iterator_end" },
              { kind: "iterator_end" },
            ],
          },
        },
      },
    },

  {
      name: "stdlib/range_iterator",
      module: AST.module([
        AST.structDefinition(
          "IntRange",
          [
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "start"),
            AST.structFieldDefinition(AST.simpleTypeExpression("i32"), "end"),
            AST.structFieldDefinition(AST.simpleTypeExpression("bool"), "inclusive"),
          ],
          "named",
        ),
        AST.methodsDefinition(
          AST.simpleTypeExpression("IntRange"),
          [
            AST.functionDefinition(
              "iterator",
              [AST.functionParameter("self")],
              AST.blockExpression([
                AST.assignmentExpression(
                  ":=",
                  AST.identifier("current"),
                  AST.memberAccessExpression(AST.identifier("self"), "start"),
                ),
                AST.returnStatement(
                  AST.iteratorLiteral(
                    [
                      AST.ifExpression(
                        AST.binaryExpression(
                          ">=",
                          AST.memberAccessExpression(AST.identifier("self"), "end"),
                          AST.memberAccessExpression(AST.identifier("self"), "start"),
                        ),
                        AST.blockExpression([
                          AST.ifExpression(
                            AST.memberAccessExpression(AST.identifier("self"), "inclusive"),
                            AST.blockExpression([
                              AST.whileLoop(
                                AST.binaryExpression(
                                  "<=",
                                  AST.identifier("current"),
                                  AST.memberAccessExpression(AST.identifier("self"), "end"),
                                ),
                                AST.blockExpression([
                                  AST.functionCall(
                                    AST.memberAccessExpression(AST.identifier("gen"), "yield"),
                                    [AST.identifier("current")],
                                  ),
                                  AST.assignmentExpression(
                                    "+=",
                                    AST.identifier("current"),
                                    AST.integerLiteral(1),
                                  ),
                                ]),
                              ),
                            ]),
                            [
                              AST.orClause(
                                AST.blockExpression([
                                  AST.whileLoop(
                                    AST.binaryExpression(
                                      "<",
                                      AST.identifier("current"),
                                      AST.memberAccessExpression(
                                        AST.identifier("self"),
                                        "end",
                                      ),
                                    ),
                                    AST.blockExpression([
                                      AST.functionCall(
                                        AST.memberAccessExpression(
                                          AST.identifier("gen"),
                                          "yield",
                                        ),
                                        [AST.identifier("current")],
                                      ),
                                      AST.assignmentExpression(
                                        "+=",
                                        AST.identifier("current"),
                                        AST.integerLiteral(1),
                                      ),
                                    ]),
                                  ),
                                ]),
                              ),
                            ],
                          ),
                        ]),
                        [
                          AST.orClause(
                            AST.blockExpression([
                              AST.ifExpression(
                                AST.memberAccessExpression(
                                  AST.identifier("self"),
                                  "inclusive",
                                ),
                                AST.blockExpression([
                                  AST.whileLoop(
                                    AST.binaryExpression(
                                      ">=",
                                      AST.identifier("current"),
                                      AST.memberAccessExpression(
                                        AST.identifier("self"),
                                        "end",
                                      ),
                                    ),
                                    AST.blockExpression([
                                      AST.functionCall(
                                        AST.memberAccessExpression(
                                          AST.identifier("gen"),
                                          "yield",
                                        ),
                                        [AST.identifier("current")],
                                      ),
                                      AST.assignmentExpression(
                                        "-=",
                                        AST.identifier("current"),
                                        AST.integerLiteral(1),
                                      ),
                                    ]),
                                  ),
                                ]),
                                [
                                  AST.orClause(
                                    AST.blockExpression([
                                      AST.whileLoop(
                                        AST.binaryExpression(
                                          ">",
                                          AST.identifier("current"),
                                          AST.memberAccessExpression(
                                            AST.identifier("self"),
                                            "end",
                                          ),
                                        ),
                                        AST.blockExpression([
                                          AST.functionCall(
                                            AST.memberAccessExpression(
                                              AST.identifier("gen"),
                                              "yield",
                                            ),
                                            [AST.identifier("current")],
                                          ),
                                          AST.assignmentExpression(
                                            "-=",
                                            AST.identifier("current"),
                                            AST.integerLiteral(1),
                                          ),
                                        ]),
                                      ),
                                    ]),
                                  ),
                                ],
                              ),
                            ]),
                          ),
                        ],
                      ),
                    ],
                    undefined,
                    AST.simpleTypeExpression("i32"),
                  ),
                ),
              ]),
            ),
          ],
        ),
        AST.assign(
          "forward",
          AST.structLiteral(
            [
              AST.structFieldInitializer(AST.integerLiteral(1), "start"),
              AST.structFieldInitializer(AST.integerLiteral(3), "end"),
              AST.structFieldInitializer(AST.booleanLiteral(true), "inclusive"),
            ],
            false,
            "IntRange",
          ),
        ),
        AST.assign(
          "forward_iter",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("forward"), "iterator"),
            [],
          ),
        ),
        AST.assign("forward_sum", AST.integerLiteral(0)),
        AST.assign(
          "forward_value",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("forward_iter"), "next"),
            [],
          ),
        ),
        AST.whileLoop(
          AST.booleanLiteral(true),
          AST.blockExpression([
            AST.matchExpression(
              AST.identifier("forward_value"),
              [
                AST.matchClause(
                  AST.structPattern([], false, "IteratorEnd"),
                  AST.blockExpression([AST.breakStatement()]),
                ),
                AST.matchClause(
                  AST.typedPattern(
                    AST.identifier("value"),
                    AST.simpleTypeExpression("i32"),
                  ),
                  AST.blockExpression([
                    AST.assignmentExpression(
                      "+=",
                      AST.identifier("forward_sum"),
                      AST.identifier("value"),
                    ),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("forward_value"),
                      AST.functionCall(
                        AST.memberAccessExpression(AST.identifier("forward_iter"), "next"),
                        [],
                      ),
                    ),
                  ]),
                ),
                AST.matchClause(
                  AST.wildcardPattern(),
                  AST.blockExpression([AST.breakStatement()]),
                ),
              ],
            ),
          ]),
        ),
        AST.assign(
          "exclusive",
          AST.structLiteral(
            [
              AST.structFieldInitializer(AST.integerLiteral(1), "start"),
              AST.structFieldInitializer(AST.integerLiteral(1), "end"),
              AST.structFieldInitializer(AST.booleanLiteral(false), "inclusive"),
            ],
            false,
            "IntRange",
          ),
        ),
        AST.assign(
          "exclusive_iter",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("exclusive"), "iterator"),
            [],
          ),
        ),
        AST.assign("exclusive_sum", AST.integerLiteral(0)),
        AST.assign(
          "exclusive_value",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("exclusive_iter"), "next"),
            [],
          ),
        ),
        AST.whileLoop(
          AST.booleanLiteral(true),
          AST.blockExpression([
            AST.matchExpression(
              AST.identifier("exclusive_value"),
              [
                AST.matchClause(
                  AST.structPattern([], false, "IteratorEnd"),
                  AST.blockExpression([AST.breakStatement()]),
                ),
                AST.matchClause(
                  AST.typedPattern(
                    AST.identifier("value"),
                    AST.simpleTypeExpression("i32"),
                  ),
                  AST.blockExpression([
                    AST.assignmentExpression(
                      "+=",
                      AST.identifier("exclusive_sum"),
                      AST.identifier("value"),
                    ),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("exclusive_value"),
                      AST.functionCall(
                        AST.memberAccessExpression(AST.identifier("exclusive_iter"), "next"),
                        [],
                      ),
                    ),
                  ]),
                ),
                AST.matchClause(
                  AST.wildcardPattern(),
                  AST.blockExpression([AST.breakStatement()]),
                ),
              ],
            ),
          ]),
        ),
        AST.assign(
          "descending",
          AST.structLiteral(
            [
              AST.structFieldInitializer(AST.integerLiteral(3), "start"),
              AST.structFieldInitializer(AST.integerLiteral(1), "end"),
              AST.structFieldInitializer(AST.booleanLiteral(true), "inclusive"),
            ],
            false,
            "IntRange",
          ),
        ),
        AST.assign(
          "descending_iter",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("descending"), "iterator"),
            [],
          ),
        ),
        AST.assign("descending_sum", AST.integerLiteral(0)),
        AST.assign(
          "descending_value",
          AST.functionCall(
            AST.memberAccessExpression(AST.identifier("descending_iter"), "next"),
            [],
          ),
        ),
        AST.whileLoop(
          AST.booleanLiteral(true),
          AST.blockExpression([
            AST.matchExpression(
              AST.identifier("descending_value"),
              [
                AST.matchClause(
                  AST.structPattern([], false, "IteratorEnd"),
                  AST.blockExpression([AST.breakStatement()]),
                ),
                AST.matchClause(
                  AST.typedPattern(
                    AST.identifier("value"),
                    AST.simpleTypeExpression("i32"),
                  ),
                  AST.blockExpression([
                    AST.assignmentExpression(
                      "+=",
                      AST.identifier("descending_sum"),
                      AST.identifier("value"),
                    ),
                    AST.assignmentExpression(
                      "=",
                      AST.identifier("descending_value"),
                      AST.functionCall(
                        AST.memberAccessExpression(AST.identifier("descending_iter"), "next"),
                        [],
                      ),
                    ),
                  ]),
                ),
                AST.matchClause(
                  AST.wildcardPattern(),
                  AST.blockExpression([AST.breakStatement()]),
                ),
              ],
            ),
          ]),
        ),
        AST.arrayLiteral([
          AST.identifier("forward_sum"),
          AST.identifier("exclusive_sum"),
          AST.identifier("descending_sum"),
        ]),
      ]),
      manifest: {
        description: "Generator-based IntRange iterator produces inclusive/exclusive and descending sequences lazily",
        expect: {
          result: {
            kind: "array",
            elements: [
              { kind: "i32", value: 6 },
              { kind: "i32", value: 0 },
              { kind: "i32", value: 6 },
            ],
          },
        },
      },
    },
];

export default stdlibFixtures;
