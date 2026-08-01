package tests

import "testing"

func TestSwitchCasesAfterDefaultPreserveSourceOrder(t *testing.T) {
	testSource(t, `
let matchedBefore = "";
switch (1) {
case 1: matchedBefore += "a";
default: matchedBefore += "d";
case 2: matchedBefore += "b";
case 3: matchedBefore += "c";
}
assertEqual(matchedBefore, "adbc");

let matchedAfter = "";
switch (2) {
case 1: matchedAfter += "a";
default: matchedAfter += "d";
case 2: matchedAfter += "b";
case 3: matchedAfter += "c";
}
assertEqual(matchedAfter, "bc");

let matchedDefault = "";
switch (9) {
case 1: matchedDefault += "a";
default: matchedDefault += "d";
case 2: matchedDefault += "b";
case 3: matchedDefault += "c";
}
assertEqual(matchedDefault, "dbc");

let stopped = "";
switch (2) {
default: stopped += "d";
case 2: stopped += "b"; break;
case 3: stopped += "c";
}
assertEqual(stopped, "b");
`)
}
