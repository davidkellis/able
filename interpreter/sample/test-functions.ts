import { interpret } from "../interpreter";
import functionsSampleModule from "./functions";

console.log("--- Running Functions Sample ---");
try {
  interpret(functionsSampleModule);
  console.log("--- Functions Sample Finished ---");
} catch (error) {
  console.error("--- Functions Sample Failed ---");
  if (error instanceof Error) {
    console.error("Error:", error.message);
    console.error("Stack:", error.stack);
  } else {
    console.error("Unknown error:", error);
  }
  process.exit(1); // Exit with error code
}
