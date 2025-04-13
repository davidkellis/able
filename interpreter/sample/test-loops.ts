import { interpret } from '../interpreter';
import loopsModule from './loops';

console.log("--- Running Loops Sample ---");
try {
    interpret(loopsModule);
    console.log("--- Loops Sample Finished ---");
} catch (error) {
    console.error("--- Loops Sample Failed ---");
    if (error instanceof Error) {
        console.error("Error:", error.message);
        console.error("Stack:", error.stack);
    } else {
        console.error("Unknown error:", error);
    }
    process.exit(1); // Exit with error code
}
