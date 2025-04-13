import { interpret } from '../interpreter';
import arithmeticModule from './arithmetic';

console.log("--- Running Arithmetic Sample ---");
try {
    interpret(arithmeticModule);
    console.log("--- Arithmetic Sample Finished ---");
} catch (error) {
    console.error("--- Arithmetic Sample Failed ---");
    if (error instanceof Error) {
        console.error("Error:", error.message);
        console.error("Stack:", error.stack);
    } else {
        console.error("Unknown error:", error);
    }
    process.exit(1); // Exit with error code
}
