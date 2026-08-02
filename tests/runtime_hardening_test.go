package tests

import "testing"

// TestMathSQRT1_2 verifies the standard constant is initialized without Go's
// integer-division truncation.
func TestMathSQRT1_2(t *testing.T) {
	testSource(t, `assertEqual(Math.SQRT1_2, Math.SQRT2 / 2);`)
}

// TestMathRandom verifies the default Realm owns an initialized random source
// whose results have the ECMAScript-required numeric range.
func TestMathRandom(t *testing.T) {
	testSource(t, `
for (var i = 0; i < 16; i++) {
  var value = Math.random();
  assert(typeof value === "number");
  assert(value >= 0 && value < 1);
}
`)
}

// TestRelationalOperatorsPreserveDirectionAndNaN verifies that evaluation order
// does not accidentally swap the operands used by abstract relational comparison.
func TestRelationalOperatorsPreserveDirectionAndNaN(t *testing.T) {
	testSource(t, `
assert(1 > 0);
assert(1 >= 0);
assert(0 < 1);
assert(0 <= 1);
assert(!(0 > 1));
assert(!(0 >= 1));
assert(!(1 < 0));
assert(!(1 <= 0));
assert("a" < "b");
assert("b" >= "a");
assert("𐀀" < "");
assert(!("" < "𐀀"));
assert(1n < 2n);
assert(1n < 2);
assert(1 < 2n);
assert(1.5 < 2n);
assert(2n > 1.5);
assert(1n < "2");
assert(!(1n < "not-a-bigint"));
assert(!(1n >= "not-a-bigint"));
assert(!(NaN < 0));
assert(!(NaN > 0));
assert(!(NaN <= 0));
assert(!(NaN >= 0));
assert(!(Infinity < Infinity));
assert(!(Infinity > Infinity));
assert(Infinity <= Infinity);
assert(Infinity >= Infinity);
assert(!(-Infinity < -Infinity));
assert(!(-Infinity > -Infinity));
assert(-Infinity <= -Infinity);
assert(-Infinity >= -Infinity);

var coercionOrder = "";
var left = { valueOf: function () { coercionOrder += "L"; return 2; } };
var right = { valueOf: function () { coercionOrder += "R"; return 1; } };
left < right;
assertEqual(coercionOrder, "LR");
coercionOrder = "";
assert(left > right);
assertEqual(coercionOrder, "LR");
coercionOrder = "";
left <= right;
assertEqual(coercionOrder, "LR");
coercionOrder = "";
left >= right;
assertEqual(coercionOrder, "LR");

var equalityAbrupt = false;
try {
  1 === missingEqualityOperand;
} catch (error) {
  equalityAbrupt = true;
}
assert(equalityAbrupt);
`)
}

// TestClassHeritageFailuresAreCatchable verifies that expected class-definition
// failures remain JavaScript abrupt completions instead of escaping as Go panics.
func TestClassHeritageFailuresAreCatchable(t *testing.T) {
	testSource(t, `
var nonConstructorCaught = false;
try {
  class Invalid extends 1 {}
} catch (error) {
  nonConstructorCaught = error instanceof TypeError;
}
assert(nonConstructorCaught);

var missingHeritageCaught = false;
try {
  class MissingHeritage extends missingHeritageValue {}
} catch (error) {
  missingHeritageCaught = error instanceof ReferenceError;
}
assert(missingHeritageCaught);

var prototypeMarker = {};
var ProxiedBase = new Proxy(function () {}, {
  get: function (target, key) {
    if (key === "prototype") throw prototypeMarker;
    return target[key];
  }
});
var prototypeGetCaught = false;
try {
  class ProxyDerived extends ProxiedBase {}
} catch (error) {
  prototypeGetCaught = error === prototypeMarker;
}
assert(prototypeGetCaught);

function Base() {}
Base.prototype = 1;
var invalidPrototypeCaught = false;
try {
  class InvalidPrototype extends Base {}
} catch (error) {
  invalidPrototypeCaught = error instanceof TypeError;
}
assert(invalidPrototypeCaught);

class NullBase extends null {}
assert(Object.getPrototypeOf(NullBase.prototype) === null);
assert(Object.getPrototypeOf(NullBase) === Function.prototype);

class BaseWithField { baseField; }
var baseWithField = new BaseWithField();
assert(Object.hasOwn(baseWithField, "baseField"));

class DerivedWithField extends BaseWithField { derivedField; }
var derivedWithField = new DerivedWithField();
assert(Object.hasOwn(derivedWithField, "baseField"));
assert(Object.hasOwn(derivedWithField, "derivedField"));

var constructMarker = {};
class ThrowingBase { constructor() { throw constructMarker; } }
class ImplicitThrowingDerived extends ThrowingBase {}
var implicitConstructCaught = false;
try {
  new ImplicitThrowingDerived();
} catch (error) {
  implicitConstructCaught = error === constructMarker;
}
assert(implicitConstructCaught);

class ExplicitThrowingDerived extends ThrowingBase { constructor() { super(); } }
var explicitConstructCaught = false;
try {
  new ExplicitThrowingDerived();
} catch (error) {
  explicitConstructCaught = error === constructMarker;
}
assert(explicitConstructCaught);

class BaseClass {}
class DerivedClass extends BaseClass {}
Object.setPrototypeOf(DerivedClass, null);
var missingConstructorCaught = false;
try {
  new DerivedClass();
} catch (error) {
  missingConstructorCaught = error instanceof TypeError;
}
assert(missingConstructorCaught);
`)
}
