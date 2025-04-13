import { interpret } from '../interpreter';
import structsModule from './structs';

console.log("--- Running Structs Sample ---");
try {
    interpret(structsModule);
    console.log("--- Structs Sample Finished ---");
} catch (error) {
    console.error("--- Structs Sample Failed ---");
    if (error instanceof Error) {
        console.error("Error:", error.message);
        console.error("Stack:", error.stack);
    } else {
        console.error("Unknown error:", error);
    }
    process.exit(1); // Exit with error code
}
