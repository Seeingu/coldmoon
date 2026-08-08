// Module code is strict mode code; strictness propagates into every nested
// function body. These helpers exercise the two paths that used to crash:
// a synchronous nested closure reading a module-level binding, and a timer
// callback crossing the job boundary.
let moduleValue = 42;
let timerFired = false;

export function readValue() {
  return (() => moduleValue)();
}

export function armTimer() {
  setTimeout(() => {
    timerFired = true;
    assert(timerFired === true);
    assert(readValue() === 42);
  }, 0);
}
