import { interpret } from '../interpreter';
import conditionalsModule from './conditionals';

console.log("--- Running Conditionals Sample ---");
try {
    interpret(conditionalsModule);
    console.log("--- Conditionals Sample Finished ---");
} catch (error) {
    console.error("--- Conditionals Sample Failed ---");
    if (error instanceof Error) {
        console.error("Error:", error.message);
        console.error("Stack:", error.stack);
    } else {
        console.error("Unknown error:", error);
    }
    process.exit(1); // Exit with error code
}
