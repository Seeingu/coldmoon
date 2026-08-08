package tests

import (
	"testing"

	cr "github.com/Seeingu/coldmoon/runtime"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

// This file locks in the P0 crash fixes from docs/code-review-2026-08.md:
// valid JavaScript must never take down the host process with a Go panic or a
// scheduler deadlock.

// TestP0GenericArrayBuiltinsOnPrimitiveReceivers covers generic
// Array.prototype methods called with primitive receivers: they must box the
// receiver with ToObject instead of asserting it to an object.
func TestP0GenericArrayBuiltinsOnPrimitiveReceivers(t *testing.T) {
	testSource(t, `
assertEqual(Array.prototype.push.call(2, 5), 1, "push boxes a primitive receiver");
assertEqual(Array.prototype.join.call("ab"), "a,b", "join boxes a primitive receiver");
assertEqual(Array.prototype.pop.call(2), undefined, "pop boxes a primitive receiver");
assertEqual(Array.prototype.toLocaleString.call(2), "", "toLocaleString boxes a primitive receiver");
`)
}

// TestP0TypedArraySortWithoutComparator covers %TypedArray%.prototype.sort
// (and toSorted) with no comparator, which used to assert a nil compareFn
// into an object.
func TestP0TypedArraySortWithoutComparator(t *testing.T) {
	testSource(t, `
var u8 = new Uint8Array([3, 1, 2]);
u8.sort();
assertEqual(u8.join(","), "1,2,3", "sort without comparator");
var sorted = new Float64Array([2.5, 0.5, -1]).toSorted();
assertEqual(sorted.join(","), "-1,0.5,2.5", "toSorted without comparator");
`)
}

// TestP0TypedArrayFromPrimitiveAndZeroLength covers %TypedArray%.from on a
// primitive source and zero-length typed arrays, both of which used to crash:
// from() asserted the primitive source to an object, and zero-length arrays
// were allocated without an ArrayBuffer.
func TestP0TypedArrayFromPrimitiveAndZeroLength(t *testing.T) {
	testSource(t, `
assertEqual(Uint8Array.from(5)[0], undefined, "from(5) produces an empty array");
assertEqual(Uint8Array.from({})[0], undefined, "from({}) produces an empty array");
var zero = new Uint8Array(0);
assertEqual(zero.join(""), "", "zero-length typed array is usable");
`)
}

// TestP0ObjectIsBigInt covers SameValue/SameValueZero on BigInt values, which
// used to panic by asserting the BigInt through the *NumberValue branch.
func TestP0ObjectIsBigInt(t *testing.T) {
	testSource(t, `
assert(Object.is(1n, 1n), "Object.is same BigInt");
assert(!Object.is(1n, 2n), "Object.is different BigInt");
assert(!Object.is(1n, 1), "Object.is does not cross Number/BigInt");
var set = new Set();
set.add(1n);
assert(set.has(1n), "Set uses SameValueZero for BigInt");
assert(!set.has(2n), "Set distinguishes BigInt values");
`)
}

// TestP0NeverSettlingAwaitDoesNotDeadlock: awaiting a promise that never
// settles used to leave the scheduler waiting on a wake that could never
// arrive, killing the process with "all goroutines are asleep". Await-suspended
// continuations are now dropped from the keepalive count, so the script
// completes normally while the pending promise stays unresolved.
func TestP0NeverSettlingAwaitDoesNotDeadlock(t *testing.T) {
	testSource(t, `
var completed = false;
async function f() { await new Promise(function () {}); }
f();
completed = true;
assert(completed === true, "script finished while a promise stays pending");
`)
}

// TestP0AwaitResumesFromTimer: removing await-suspended continuations from the
// keepalive count must not break the case where a timer is the only wake
// source. Evaluate drains the scheduler, so the promise must be settled and
// the async function finished by the time it returns.
func TestP0AwaitResumesFromTimer(t *testing.T) {
	agent := NewAgent()
	InitializeConstants()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	cr.RegisterTerminalRuntime(realm)

	Evaluate(`var resolved = false;
async function f() { await new Promise(function (r) { setTimeout(function () { r(1); }, 0); }); resolved = true; }
f();`, realm)

	resolved := realm.GlobalObject.Get(NewStringPropertyKey("resolved"))
	if !resolved.ToBoolean() {
		t.Fatal("timer did not wake the awaited continuation before Evaluate returned")
	}
}

// TestP0ModuleFunctionsReadModuleBindings covers strictness propagation into
// nested function bodies inside module code. Without it, reading a module-level
// binding from a nested function hit the ModuleEnvironment's strict assertion
// and panicked the host.
func TestP0ModuleFunctionsReadModuleBindings(t *testing.T) {
	testModule(t, "module_strict_main.js")
}

// TestP0AssertThrowsCatchableError: assert failures used to escape as bare Go
// string panics that crashed the process. They now throw a real Error object
// that JavaScript can observe.
func TestP0AssertThrowsCatchableError(t *testing.T) {
	testSource(t, `
var caught = null;
try {
  assert(false, "boom");
} catch (e) {
  caught = e;
}
assert(caught instanceof Error, "assert throws a catchable Error");
`)
}

// TestP0DataViewConstructor covers the constructor fixes: it must be a
// constructor, honor an explicit byteLength instead of discarding it, apply
// the bounds check unconditionally, and compute the remainder view length for
// resizable buffers.
func TestP0DataViewConstructor(t *testing.T) {
	testSource(t, `
var ab = new ArrayBuffer(8);
var dv = new DataView(ab);
assertEqual(dv.byteLength, 8, "auto byteLength over fixed-length buffer");
assertEqual(dv.byteOffset, 0, "byteOffset");
assert(dv.buffer === ab, "viewed buffer");
var dv2 = new DataView(ab, 2, 4);
assertEqual(dv2.byteLength, 4, "explicit byteLength");
assertEqual(dv2.byteOffset, 2, "explicit byteOffset");
var threw = false;
try { new DataView(ab, 6, 4); } catch (e) { threw = e instanceof RangeError; }
assert(threw === true, "out-of-bounds byteLength throws RangeError");
var rb = new ArrayBuffer(8, { maxByteLength: 16 });
var rdv = new DataView(rb, 2);
assertEqual(rdv.byteLength, 6, "resizable buffer defaults to the remainder");
`)
}
