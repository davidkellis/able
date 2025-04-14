import { interpret } from "../interpreter";
import interfacesModule from "./interfaces"; // Import the AST module

console.log("--- Running Interfaces Sample ---");
try {
  interpret(interfacesModule);
  console.log("--- Interfaces Sample Finished ---");
} catch (error) {
  console.error("--- Interfaces Sample Failed ---");
  if (error instanceof Error) {
    console.error("Error:", error.message);
    console.error("Stack:", error.stack);
  } else {
    console.error("Unknown error:", error);
  }
  process.exit(1); // Exit with error code
}
