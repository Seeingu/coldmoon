package tests

import (
	"testing"

	. "github.com/Seeingu/coldmoon/coldmoon"
	cr "github.com/Seeingu/coldmoon/runtime"
)

func evaluateAsyncSource(t *testing.T, source string) *Realm {
	t.Helper()
	agent := NewAgent()
	InitializeConstants()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	cr.RegisterTerminalRuntime(realm)
	Evaluate(source, realm)
	return realm
}

func TestYieldStarForwardsNextThrowAndReturn(t *testing.T) {
	testSource(t, `
function* forwardingNext() {
  return yield* (function* () {
    const sent = yield 1;
    return sent + 1;
  })();
}
const nextGenerator = forwardingNext();
let result = nextGenerator.next();
assertEqual(result.value, 1, "yield* next value");
assertEqual(result.done, false, "yield* next done");
result = nextGenerator.next(41);
assertEqual(result.value, 42, "yield* completion value");
assertEqual(result.done, true, "yield* completion done");

let thrownReason;
const throwIterator = {
  [Symbol.iterator]: function () { return this; },
  next: function () { return { value: "pending", done: false }; },
  throw: function (reason) {
    thrownReason = reason;
    return { value: "recovered", done: true };
  }
};
const throwGenerator = (function* () { return yield* throwIterator; })();
throwGenerator.next();
result = throwGenerator.throw("boom");
assertEqual(thrownReason, "boom", "yield* forwards throw reason");
assertEqual(result.value, "recovered", "yield* throw completion value");
assertEqual(result.done, true, "yield* throw completion done");

let returnedValue;
const returnIterator = {
  [Symbol.iterator]: function () { return this; },
  next: function () { return { value: "pending", done: false }; },
  return: function (value) {
    returnedValue = value;
    return { value: value + 1, done: true };
  }
};
const returnGenerator = (function* () { return yield* returnIterator; })();
returnGenerator.next();
result = returnGenerator.return(9);
assertEqual(returnedValue, 9, "yield* forwards return value");
assertEqual(result.value, 10, "yield* return completion value");
assertEqual(result.done, true, "yield* return completion done");
`)
}

func TestYieldStarClosesIteratorWhenThrowIsMissing(t *testing.T) {
	testSource(t, `
let closed = false;
const iterator = {
  [Symbol.iterator]: function () { return this; },
  next: function () { return { value: 1, done: false }; },
  return: function () {
    closed = true;
    return { value: undefined, done: true };
  }
};
const generator = (function* () { yield* iterator; })();
generator.next();
let threwTypeError = false;
try {
  generator.throw("boom");
} catch (error) {
  threwTypeError = error instanceof TypeError;
}
assertEqual(closed, true, "yield* closes delegate without throw");
assertEqual(threwTypeError, true, "yield* reports missing throw method");
`)
}

func TestAwaitRejectionResumesWithThrowCompletion(t *testing.T) {
	realm := evaluateAsyncSource(t, `
let caught;
async function run() {
  try {
    await Promise.reject("rejected");
  } catch (error) {
    caught = error;
  }
}
run();
`)
	Evaluate(`assertEqual(caught, "rejected");`, realm)
}

func TestAsyncFromSyncIteratorAndAsyncYieldStar(t *testing.T) {
	realm := evaluateAsyncSource(t, `
let first;
let second;
let completion;
async function* delegate() {
  return yield* [Promise.resolve(1), 2];
}
async function run() {
  const generator = delegate();
  first = await generator.next();
  second = await generator.next();
  completion = await generator.next();
}
run();
`)
	Evaluate(`
assertEqual(first.value, 1, "async-from-sync unwraps values");
assertEqual(first.done, false, "first delegated result");
assertEqual(second.value, 2, "second delegated result");
assertEqual(second.done, false, "second delegated result done");
assertEqual(completion.value, undefined, "delegated completion value");
assertEqual(completion.done, true, "delegated completion done");
`, realm)
}
