import def, { value, extra } from "./module_import_forms_lib.js";
import * as ns from "./module_import_forms_lib.js";
import def2, * as ns2 from "./module_import_forms_lib.js";

assert(def() === 99, "default binding");
assert(value === 42 && extra === 7, "named bindings");
assert(ns.value === 42 && ns.default() === 99, "namespace import");
assert(def2() === 99 && ns2.extra === 7, "default + namespace import");
