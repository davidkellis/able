import { interpret } from "../interpreter";
import assignmentsAndClosuresModule from "./assignments-and-closures";

console.log("--- Running Assignments and Closures Sample ---");
try {
  interpret(assignmentsAndClosuresModule);
  console.log("--- Assignments and Closures Sample Finished ---");
} catch (error) {
  console.error("--- Assignments and Closures Sample Failed ---");
  if (error instanceof Error) {
    console.error("Error:", error.message);
    console.error("Stack:", error.stack);
  } else {
    console.error("Unknown error:", error);
  }
  process.exit(1); // Exit with error code
}
