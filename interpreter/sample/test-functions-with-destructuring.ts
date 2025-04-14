import { interpret } from "../interpreter";
import functionsWithDestructuringModule from "./functions-with-destructuring";

console.log("--- Running Functions with Destructuring Sample ---");
try {
  interpret(functionsWithDestructuringModule);
  console.log("--- Functions with Destructuring Sample Finished ---");
} catch (error) {
  console.error("--- Functions with Destructuring Sample Failed ---");
  if (error instanceof Error) {
    console.error("Error:", error.message);
    console.error("Stack:", error.stack);
  } else {
    console.error("Unknown error:", error);
  }
  process.exit(1); // Exit with error code
}
