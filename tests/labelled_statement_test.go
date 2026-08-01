package tests

import "testing"

func TestLabelledBreakAndContinue(t *testing.T) {
	testSource(t, `
let reached = false;
blockLabel: {
  break blockLabel;
  reached = true;
}
assertEqual(reached, false, "labelled block break");

let visits = "";
outer: for (let i = 0; i < 3; i++) {
  for (let j = 0; j < 3; j++) {
    if (j === 1) continue outer;
    visits += "" + i + j;
  }
}
assertEqual(visits, "001020", "continue targets labelled outer loop");

let nestedLabels = 0;
first: second: for (let i = 0; i < 3; i++) {
  nestedLabels++;
  continue first;
}
assertEqual(nestedLabels, 3, "nested labels are passed to the iteration");
`)
}

func TestDoWhileEvaluationHandlesContinueAndBreak(t *testing.T) {
	testSource(t, `
let value = 0;
do {
  value++;
  if (value < 3) continue;
  break;
} while (true);
assertEqual(value, 3);
`)
}
