import { interpret } from "../interpreter";
import interfacesModule2 from "./interfaces2"; // Import the AST module

console.log("--- Running Interfaces Sample ---");
try {
  interpret(interfacesModule2);
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
