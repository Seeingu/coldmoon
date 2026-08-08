import { readValue, armTimer } from "./module_strict_lib.js";

assert(readValue() === 42, "nested closure reads module binding");
armTimer();
