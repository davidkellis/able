import { interpret } from "../interpreter";
import exceptionsModule from "./exceptions"; // Import the AST module

console.log("--- Running Exceptions Sample ---");
try {
  interpret(exceptionsModule);
  console.log("--- Exceptions Sample Finished ---");
} catch (error) {
  // This top-level catch will handle unrescued errors from the interpreter
  console.error("--- Exceptions Sample Failed (or Unrescued Error) ---");
  if (error instanceof Error) {
    console.error("Error Type:", error.name); // e.g., RaiseSignal, Error
    console.error("Error Message:", error.message);
    // If it's a RaiseSignal, you might want to inspect error.value
    if ((error as any).name === 'RaiseSignal' && (error as any).value) {
        // NOTE: Need a proper way to print the AbleValue here
        console.error("Raised Value:", JSON.stringify((error as any).value, null, 2));
    }
    console.error("Stack:", error.stack);
  } else {
    console.error("Unknown error object:", error);
  }
  process.exit(1); // Exit with error code
}
