import * as AST from "../ast";
import type { Interpreter } from "./index";
import type { AwaitWakerPayload } from "./concurrency_shared";

export function applyConcurrencyBuiltins(cls: typeof Interpreter): void {
  cls.prototype.initConcurrencyBuiltins = function initConcurrencyBuiltins(this: Interpreter): void {
    if (this.concurrencyBuiltinsInitialized) return;
    this.concurrencyBuiltinsInitialized = true;

    const procErrorDefAst = AST.structDefinition(
      "ProcError",
      [AST.structFieldDefinition(AST.simpleTypeExpression("String"), "details")],
      "named"
    );
    const pendingDefAst = AST.structDefinition("Pending", [], "named");
    const resolvedDefAst = AST.structDefinition("Resolved", [], "named");
    const cancelledDefAst = AST.structDefinition("Cancelled", [], "named");
    const failedDefAst = AST.structDefinition(
      "Failed",
      [AST.structFieldDefinition(AST.simpleTypeExpression("ProcError"), "error")],
      "named"
    );

    this.evaluate(procErrorDefAst, this.globals);
    this.evaluate(pendingDefAst, this.globals);
    this.evaluate(resolvedDefAst, this.globals);
    this.evaluate(cancelledDefAst, this.globals);
    this.evaluate(failedDefAst, this.globals);
    this.evaluate(
      AST.unionDefinition(
        "ProcStatus",
        [
          AST.simpleTypeExpression("Pending"),
          AST.simpleTypeExpression("Resolved"),
          AST.simpleTypeExpression("Cancelled"),
          AST.simpleTypeExpression("Failed"),
        ],
        undefined,
        undefined,
        false
      ),
      this.globals,
    );

    const getStructDef = (name: string): AST.StructDefinition => {
      const val = this.globals.get(name);
      if (val.kind !== "struct_def") throw new Error(`Failed to initialize struct '${name}'`);
      return val.def;
    };

    this.procErrorStruct = getStructDef("ProcError");
    this.procStatusStructs = {
      Pending: getStructDef("Pending"),
      Resolved: getStructDef("Resolved"),
      Cancelled: getStructDef("Cancelled"),
      Failed: getStructDef("Failed"),
    };

    this.procStatusPendingValue = this.makeNamedStructInstance(this.procStatusStructs.Pending, []);
    this.procStatusResolvedValue = this.makeNamedStructInstance(this.procStatusStructs.Resolved, []);
    this.procStatusCancelledValue = this.makeNamedStructInstance(this.procStatusStructs.Cancelled, []);

    const awaitWakerDefAst = AST.structDefinition("AwaitWaker", [], "named");
    const awaitRegistrationDefAst = AST.structDefinition("AwaitRegistration", [], "named");
    this.evaluate(awaitWakerDefAst, this.globals);
    this.evaluate(awaitRegistrationDefAst, this.globals);
    const awaitWakerVal = this.globals.get("AwaitWaker");
    if (!awaitWakerVal || awaitWakerVal.kind !== "struct_def") {
      throw new Error("Failed to initialize struct 'AwaitWaker'");
    }
    this.awaitWakerStruct = awaitWakerVal.def;
    const awaitRegistrationVal = this.globals.get("AwaitRegistration");
    if (!awaitRegistrationVal || awaitRegistrationVal.kind !== "struct_def") {
      throw new Error("Failed to initialize struct 'AwaitRegistration'");
    }
    this.awaitRegistrationStruct = awaitRegistrationVal.def;
    const wakeNative = this.makeNativeFunction("AwaitWaker.wake", 1, (_interp, args) => {
      const self = args[0];
      if (!self || self.kind !== "struct_instance") {
        return { kind: "nil", value: null };
      }
      const payload = (self as any).__awaitPayload as AwaitWakerPayload | undefined;
      if (payload) {
        payload.wake();
      }
      return { kind: "nil", value: null };
    });
    this.awaitWakerNativeMethods = { wake: wakeNative };
    this.ensureAwaitHelperBuiltins();
  };
}
