import { interpret } from "../interpreter";
import patternsModule from "./patterns"; // Import the AST module

console.log("--- Running Patterns Sample ---");
try {
  interpret(patternsModule);
  console.log("--- Patterns Sample Finished ---");
} catch (error) {
  console.error("--- Patterns Sample Failed ---");
  if (error instanceof Error) {
    console.error("Error:", error.message);
    console.error("Stack:", error.stack);
  } else {
    console.error("Unknown error:", error);
  }
  process.exit(1); // Exit with error code
}
