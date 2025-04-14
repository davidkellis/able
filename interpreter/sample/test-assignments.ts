import { interpret } from "../interpreter";
import assignmentsModule from "./assignments";

console.log("--- Running Assignments Sample ---");
try {
  interpret(assignmentsModule);
  console.log("--- Assignments Sample Finished ---");
} catch (error) {
  console.error("--- Assignments Sample Failed ---");
  if (error instanceof Error) {
    console.error("Error:", error.message);
    console.error("Stack:", error.stack);
  } else {
    console.error("Unknown error:", error);
  }
  process.exit(1); // Exit with error code
}
