package tests

import "testing"

// This file locks in the P1 spec-semantic fixes from docs/code-review-2026-08.md
// (the high-frequency semantic batch).

// TestP1EqualityTrio covers strict equality on NaN/-0, loose equality across
// Number and BigInt, and the removed dead NaN branch.
func TestP1EqualityTrio(t *testing.T) {
	testSource(t, `
assert(NaN !== NaN, "NaN is not strictly equal to itself");
assert(0 === -0, "-0 equals +0 strictly");
assert(0 == -0, "-0 equals +0 loosely");
assert(1 == 1n, "Number equals BigInt loosely");
assert(2n == 2, "BigInt equals Number loosely");
assert(!(1.5 == 1n), "fractional Number does not equal BigInt");
assert(!(Infinity == 1n), "Infinity does not equal BigInt");
assert(!(NaN == NaN), "NaN is not loosely equal to itself");
assert("5" == 5n, "String to Number conversion before BigInt compare");
assert(true == 1n, "Boolean to Number conversion before BigInt compare");
`)
}

// TestP1CallEvaluationOrder verifies the callee member expression (including
// its getters) is evaluated before the argument expressions.
func TestP1CallEvaluationOrder(t *testing.T) {
	testSource(t, `
var log = [];
function arg(x) { log.push(x); return x; }
var obj = {};
Object.defineProperty(obj, "m", {
  get: function () { log.push("getter"); return function (a, b) { return a + b; }; }
});
var result = obj.m(arg(1), arg(2));
assert(result === 3, "call result");
assert(log.join(",") === "getter,1,2", "member expression evaluated before arguments");
`)
}

// TestP1ArrayIndexParsing covers canonical array-index detection: strings like
// "1e2", "0.5", and "01" are ordinary properties, not indices.
func TestP1ArrayIndexParsing(t *testing.T) {
	testSource(t, `
var a = [];
a["1e2"] = 1;
assert(a.length === 0, "1e2 is not an array index");
a["0.5"] = 2;
assert(a.length === 0, "0.5 is not an array index");
a["01"] = 3;
assert(a.length === 0, "01 is not an array index");
a[0] = 5;
a["1"] = 6;
assert(a.length === 2 && a[1] === 6, "canonical index still works");
`)
}

// TestP1StringEscapes covers \uXXXX, \u{...}, \xXX, \0, and legacy octal
// escapes in string literals.
func TestP1StringEscapes(t *testing.T) {
	testSource(t, `
assert("\u0041" === "A", "unicode escape");
assert("\x42" === "B", "hex escape");
assert("\0".charCodeAt(0) === 0, "null escape");
assert("\101" === "A", "legacy octal escape");
assert("\u{1F600}" === "😀", "braced code point escape");
assert("pre\u4E2Dpost" === "pre中post", "BMP escape in context");
`)
}

// TestP1ProxyGetOwnPropertyDescriptor covers the reworked [[GetOwnProperty]]
// trap: descriptor conversion, undefined results, and invariant validation.
func TestP1ProxyGetOwnPropertyDescriptor(t *testing.T) {
	testSource(t, `
var proxy = new Proxy({}, {
  getOwnPropertyDescriptor: function (t, k) { return { value: 42, writable: true, enumerable: true, configurable: true }; }
});
var desc = Object.getOwnPropertyDescriptor(proxy, "x");
assert(desc !== undefined && desc.value === 42, "trap descriptor returned");

var undef = new Proxy({}, {
  getOwnPropertyDescriptor: function () { return undefined; }
});
assert(Object.getOwnPropertyDescriptor(undef, "x") === undefined, "undefined trap result");

var frozen = Object.freeze({ x: 1 });
var violating = new Proxy(frozen, {
  getOwnPropertyDescriptor: function () { return { value: 2, writable: true, enumerable: true, configurable: true }; }
});
var threw = false;
try { Object.getOwnPropertyDescriptor(violating, "x"); } catch (e) { threw = e instanceof TypeError; }
assert(threw, "non-configurable invariant enforced");
`)
}

// TestP1ProxyGetPrototypeOfNull covers the null-prototype and invariant paths
// of the getPrototypeOf trap.
func TestP1ProxyGetPrototypeOfNull(t *testing.T) {
	testSource(t, `
var target = Object.create(null);
var proxy = new Proxy(target, {
  getPrototypeOf: function () { return null; }
});
assert(Object.getPrototypeOf(proxy) === null, "null prototype via trap");

var frozen = Object.freeze({});
var invariant = new Proxy(frozen, {
  getPrototypeOf: function () { return {}; }
});
var threw = false;
try { Object.getPrototypeOf(invariant); } catch (e) { threw = e instanceof TypeError; }
assert(threw, "non-extensible invariant enforced");
`)
}

// TestP1ProxyDefinePropertyPlainDescriptor covers defineProperty with a plain
// {value: 1} descriptor, which must not be treated as setting configurable
// false.
func TestP1ProxyDefinePropertyPlainDescriptor(t *testing.T) {
	testSource(t, `
var target = {};
var proxy = new Proxy(target, {
  defineProperty: function (t, k, d) { return Reflect.defineProperty(t, k, d); }
});
Object.defineProperty(proxy, "x", { value: 1 });
assert(proxy.x === 1, "plain descriptor defined");
`)
}

// TestP1ParameterInstantiationAbrupt covers a throwing default parameter
// value: the function body must not run.
func TestP1ParameterInstantiationAbrupt(t *testing.T) {
	testSource(t, `
var bodyRan = false;
function f({x} = (function () { throw new Error("boom"); })()) { bodyRan = true; }
var caught = false;
try { f(); } catch (e) { caught = e instanceof Error; }
assert(caught, "parameter error thrown");
assert(!bodyRan, "body did not run after parameter failure");
`)
}

// TestP1ArrayJoinToString covers join/toLocaleString stringification with
// ToString instead of the Go debug representation.
func TestP1ArrayJoinToString(t *testing.T) {
	testSource(t, `
assert([1, undefined, 3].join() === "1,,3", "undefined element becomes empty");
assert([1, null, 3].join("-") === "1--3", "null element becomes empty");
var sep = { toString: function () { return "|"; } };
assert([1, 2].join(sep) === "1|2", "object separator uses toString");
var elem = { toString: function () { return "X"; } };
assert([elem, 2].join(",") === "X,2", "object element uses toString");
assert([NaN].join() === "NaN", "NaN stringifies");
`)
}

// TestP1TypedArrayAccessors covers buffer/byteLength/byteOffset/length being
// accessor properties on %TypedArray%.prototype rather than methods.
func TestP1TypedArrayAccessors(t *testing.T) {
	testSource(t, `
var taProto = Object.getPrototypeOf(Uint8Array.prototype);
assert(typeof Object.getOwnPropertyDescriptor(taProto, "length").get === "function", "length is an accessor");
assert(typeof Object.getOwnPropertyDescriptor(taProto, "buffer").get === "function", "buffer is an accessor");
var t = new Uint8Array(4);
assert(t.length === 4, "length accessor");
assert(t.byteLength === 4, "byteLength accessor");
assert(t.byteOffset === 0, "byteOffset accessor");
assert(t.buffer instanceof ArrayBuffer, "buffer accessor");
`)
}

// TestP1SetSizeAccessorAndDedupe covers Set.prototype.size as an accessor and
// repeated add calls not duplicating iteration entries.
func TestP1SetSizeAccessorAndDedupe(t *testing.T) {
	testSource(t, `
assert(typeof Object.getOwnPropertyDescriptor(Set.prototype, "size").get === "function", "Set size is an accessor");
var s = new Set();
s.add(1); s.add(2); s.add(1); s.add(2);
assert(s.size === 2, "duplicate adds do not grow the set");
assert(Array.from(s).join(",") === "1,2", "iteration has no duplicates");
`)
}

// TestP1ArrayBufferIsView covers ArrayBuffer.isView living on the constructor
// and recognizing typed arrays.
func TestP1ArrayBufferIsView(t *testing.T) {
	testSource(t, `
assert(ArrayBuffer.isView(new Uint8Array(4)), "typed array is a view");
assert(!ArrayBuffer.isView([]), "plain array is not a view");
assert(!ArrayBuffer.isView(5), "primitive is not a view");
assert(typeof ArrayBuffer.prototype.isView === "undefined", "isView is not on the prototype");
`)
}

// TestP1LoopBreakAndContinue covers break/continue propagation out of loop
// bodies, which previously escaped as abrupt completions.
func TestP1LoopBreakAndContinue(t *testing.T) {
	testSource(t, `
var n = 0;
for (;;) { n++; if (n > 2) break; }
assert(n === 3, "for(;;) with break");
var after = false;
while (true) { after = true; break; }
assert(after === true, "while break is absorbed");
var evens = [];
for (var k = 0; k < 6; k++) { if (k % 2) continue; evens.push(k); }
assert(evens.join(",") === "0,2,4", "continue works");
var out = [];
outer: for (var i = 0; i < 3; i++) {
  for (var j = 0; j < 3; j++) {
    if (j === 1) break outer;
    out.push(i + "," + j);
  }
}
assert(out.join(" ") === "0,0", "labeled break");
var rows = [];
outer2: for (var a = 0; a < 3; a++) {
  for (var b = 0; b < 3; b++) {
    if (b === 1) continue outer2;
    rows.push(a + "" + b);
  }
}
assert(rows.join(" ") === "00 10 20", "labeled continue");
`)
}

// TestP1SuperPropertyEvaluation covers super.m(), super["m"](), and super.x
// assignment, including the this binding observed by accessors.
func TestP1SuperPropertyEvaluation(t *testing.T) {
	testSource(t, `
class A { constructor() { this.x = 10; } m() { return this.x; } get v() { return this.x + 100; } }
class B extends A {
  constructor() { super(); }
  m() { return super.m() + 1; }
  bracket() { return super["v"]; }
  assign() { super.x = 99; return this.x; }
}
assert(new B().m() === 11, "super.m()");
assert(new B().bracket() === 110, "super[expr]()");
assert(new B().assign() === 99, "super.x assignment");
`)
}

// TestP1ArrayBoundMethods covers at() out-of-range, with() RangeError, and
// toSpliced length growth.
func TestP1ArrayBoundMethods(t *testing.T) {
	testSource(t, `
assert([1, 2, 3].at(-1) === 3, "at negative index");
assert([1, 2, 3].at(5) === undefined, "at out of range");
var withThrew = false;
try { [1, 2, 3].with(5, 9); } catch (e) { withThrew = e instanceof RangeError; }
assert(withThrew, "with out of range throws");
assert([1, 2, 3, 4].toSpliced(1, 2, 8, 9).join(",") === "1,8,9,4", "toSpliced works");
`)
}

// TestP1ImportForms covers default+named, namespace, and default+namespace
// import clauses, which previously panicked the parser.
func TestP1ImportForms(t *testing.T) {
	testModule(t, "module_import_forms_main.js")
}
