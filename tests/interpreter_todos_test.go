package tests

import (
	"testing"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

func TestForLetCreatesDistinctPerIterationBindings(t *testing.T) {
	testSource(t, `
const callbacks = [];
for (let index = 0; index < 3; index++) {
  callbacks.push(() => index);
}
assertEqual(callbacks[0](), 0);
assertEqual(callbacks[1](), 1);
assertEqual(callbacks[2](), 2);
`)
}

func TestForInRestoresLexicalEnvironmentAfterAbruptHead(t *testing.T) {
	testSource(t, `
let value = "outer";
let caught = false;
try {
  for (let value in (() => { throw "boom"; })()) {}
} catch (error) {
  caught = error === "boom";
}
assertEqual(caught, true);
assertEqual(value, "outer");
`)
}

func TestForOfClosesIteratorOnBreakAndThrow(t *testing.T) {
	testSource(t, `
let breakCloseCount = 0;
const breakIterator = {
  [Symbol.iterator]: function () { return this; },
  next: function () { return { value: 1, done: false }; },
  return: function () {
    breakCloseCount += 1;
    return { value: undefined, done: true };
  }
};
for (const value of breakIterator) {
  break;
}
assertEqual(breakCloseCount, 1, "for-of break closes the iterator");

let throwCloseCount = 0;
const thrown = {};
let caught;
const throwIterator = {
  [Symbol.iterator]: function () { return this; },
  next: function () { return { value: 1, done: false }; },
  return: function () {
    throwCloseCount += 1;
    return { value: undefined, done: true };
  }
};
try {
  for (const value of throwIterator) {
    throw thrown;
  }
} catch (error) {
  caught = error;
}
assertEqual(throwCloseCount, 1, "for-of throw closes the iterator");
assertEqual(caught, thrown, "iterator close preserves the body throw");
`)
}

func TestForAwaitOfClosesAsyncIteratorOnBreakAndThrow(t *testing.T) {
	realm := evaluateAsyncSource(t, `
let asyncBreakCloseCount = 0;
let asyncThrowCloseCount = 0;
let asyncBreakCompleted = false;
let asyncCaught;
const asyncThrown = {};

const asyncBreakIterator = {
  [Symbol.asyncIterator]: function () { return this; },
  next: function () { return Promise.resolve({ value: 1, done: false }); },
  return: function () {
    asyncBreakCloseCount += 1;
    return Promise.resolve({ value: undefined, done: true });
  }
};

const asyncThrowIterator = {
  [Symbol.asyncIterator]: function () { return this; },
  next: function () { return Promise.resolve({ value: 1, done: false }); },
  return: function () {
    asyncThrowCloseCount += 1;
    return Promise.resolve({ value: undefined, done: true });
  }
};

async function runCloseChecks() {
  for await (const value of asyncBreakIterator) {
    break;
  }
  asyncBreakCompleted = true;

  try {
    for await (const value of asyncThrowIterator) {
      throw asyncThrown;
    }
  } catch (error) {
    asyncCaught = error;
  }
}
runCloseChecks();
`)

	Evaluate(`
assertEqual(asyncBreakCloseCount, 1, "for-await break closes the iterator");
assertEqual(asyncBreakCompleted, true, "for-await resumes after async close");
assertEqual(asyncThrowCloseCount, 1, "for-await throw closes the iterator");
assertEqual(asyncCaught, asyncThrown, "async iterator close preserves the body throw");
`, realm)
}

func TestStrictDeleteFailureIsCatchableCompletion(t *testing.T) {
	testSource(t, `
let caughtDeleteError = false;
try {
  (function () {
    "use strict";
    const target = {};
    Object.defineProperty(target, "fixed", {
      value: 1,
      configurable: false
    });
    delete target.fixed;
  })();
} catch (error) {
  caughtDeleteError = error instanceof TypeError;
}
assertEqual(caughtDeleteError, true, "strict delete reports a catchable TypeError completion");
`)
}
