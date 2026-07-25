package tests

import (
	"fmt"
	"os"
	"path"
	"testing"

	cr "github.com/Seeingu/coldmoon/runtime"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

func testSource(t *testing.T, source string) {
	agent := NewAgent()
	InitializeConstants()
	el := cr.NewEventLoop(agent)
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	cr.RegisterTerminalRuntime(el, realm)
	Evaluate(source, realm)
}

func testModule(t *testing.T, f string) {
	agent := NewAgent()
	InitializeConstants()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	el := cr.NewEventLoop(agent)
	cr.RegisterTerminalRuntime(el, realm)
	EvaluateModule(resolveTestdataPath(f), realm)
}

// TestObjectLiteralDataPropertyNamedGet verifies that contextual accessor
// keywords remain valid ordinary property names when followed by a colon.
func TestObjectLiteralDataPropertyNamedGet(t *testing.T) {
	testSource(t, `const descriptor = { get: function() { return 1; } };
assertEqual(descriptor.get(), 1);`)
}

// TestObjectLiteralDataPropertyNamedReservedWord verifies that IdentifierName
// property keys may use reserved words.
func TestObjectLiteralDataPropertyNamedReservedWord(t *testing.T) {
	testSource(t, `const iterator = { return: function() { return 1; } };
assertEqual(iterator.return(), 1);`)
}

func TestArrayPredicateMethodsDefaultThisArgument(t *testing.T) {
	testSource(t, `const arrayLike = { 0: 1, length: 1 };
assert(Array.prototype.every.call(arrayLike, function(value, index, object) {
  return value === 1 && index === 0 && object === arrayLike;
}));
assert(Array.prototype.some.call(arrayLike, function(value) { return value === 1; }));
assert(Array.prototype.filter.call(arrayLike, function(value) { return value === 1; })[0] === 1);`)
}

func TestArrayIndexMethodsOptionalFromIndex(t *testing.T) {
	testSource(t, `const arrayLike = { 0: "a", 1: "b", 2: "a", length: 3 };
assert(Array.prototype.indexOf.call(arrayLike, "b") === 1);
assert(Array.prototype.indexOf.call(arrayLike, "a", -1) === 2);
assert(Array.prototype.lastIndexOf.call(arrayLike, "a") === 2);
assert(Array.prototype.lastIndexOf.call(arrayLike, "a", -2) === 0);`)
}

func TestArrayReduceDistinguishesMissingInitialValue(t *testing.T) {
	testSource(t, `assert([1, 2, 3].reduce(function(a, b) { return a + b; }) === 6);
assert([1].reduce(function(a) { return a; }, undefined) === undefined);
assert([1, 2, 3].reduceRight(function(a, b) { return a - b; }) === 0);
assert([1].reduceRight(function(a) { return a; }, undefined) === undefined);`)
}

func TestArrayFindMethodsAdvanceAndDefaultThisArgument(t *testing.T) {
	testSource(t, `const values = [1, 2, 3];
assert(values.find(function(value) { return value === 2; }) === 2);
assert(values.findIndex(function(value) { return value === 3; }) === 2);
assert(values.findLast(function(value) { return value < 3; }) === 2);
assert(values.findLastIndex(function(value) { return value < 3; }) === 1);`)
}

func TestArrayMapPreservesHoles(t *testing.T) {
	testSource(t, `const values = [1, , 3];
let calls = 0;
const mapped = values.map(function(value) {
  calls++;
  return value * 2;
});
assert(calls === 2);
assert(mapped.length === 3);
assert(mapped[0] === 2);
assert(!(1 in mapped));
assert(mapped[2] === 6);`)
}

func TestArrayForEachSupportsGenericReceiversAndAbruptCallbacks(t *testing.T) {
	testSource(t, `const arrayLike = { 0: "first", 2: "third", length: 3 };
const seen = [];
Array.prototype.forEach.call(arrayLike, function(value, index, object) {
  seen.push(value + index);
  assert(object === arrayLike);
});
assert(seen.length === 2);
assert(seen[0] === "first0");
assert(seen[1] === "third2");
const sentinel = {};
let caught = null;
try {
  [1].forEach(function() { throw sentinel; });
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

func TestFunctionInvocationHelpersDefaultArguments(t *testing.T) {
	testSource(t, `function count() { return arguments.length; }
assert(count.apply() === 0);
assert(count.call() === 0);
const bound = count.bind();
assert(bound() === 0);`)
}

func TestPropertyDescriptorAllowsUndefinedAccessors(t *testing.T) {
	testSource(t, `const object = {};
Object.defineProperty(object, "getter", { get: undefined });
Object.defineProperty(object, "setter", { set: undefined });
assert(object.getter === undefined);
assert(object.setter === undefined);
assert(Object.prototype.hasOwnProperty.call(object, "getter"));
assert(Object.prototype.hasOwnProperty.call(object, "setter"));`)
}

func TestPropertyDescriptorErrorsAreCatchableAndAtomic(t *testing.T) {
	testSource(t, `const object = {};
let caught = null;
try {
  Object.defineProperty(object, "invalid", { get: 1 });
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);
try {
  Object.defineProperties(object, {
    first: { value: 1 },
    second: { get: 1 }
  });
} catch (error) {}
assert(!Object.prototype.hasOwnProperty.call(object, "first"));`)
}

func TestObjectIntegrityLevelsUseConcreteObjectIdentity(t *testing.T) {
	testSource(t, `const array = [1, 2];
Object.seal(array);
assert(Object.isSealed(array));
assert(!Object.isFrozen(array));
Object.freeze(array);
assert(Object.isFrozen(array));
assert(!Object.isExtensible(array));`)
}

func TestStringTrimCoercesReceiverAndUsesECMAScriptWhitespace(t *testing.T) {
	tests := map[string]string{
		"boolean":    `assert(String.prototype.trim.call(false) === "false");`,
		"number":     `assert(String.prototype.trim.call(42) === "42");`,
		"both":       `assert("  value 　".trim() === "value");`,
		"start only": `assert(" value ".trimStart() === "value ");`,
		"end only":   `assert(" value ".trimEnd() === " value");`,
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			testSource(t, source)
		})
	}
}

// TestArrayFromPropagatesIteratorGetterError covers the abrupt GetMethod path
// before Array.from chooses between iterator and array-like processing.
func TestArrayFromPropagatesIteratorGetterError(t *testing.T) {
	testSource(t, `const sentinel = {};
const items = {};
Object.defineProperty(items, Symbol.iterator, {
  get: function() {
    throw sentinel;
  }
});
let caught = null;
try {
  Array.from(items);
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayFromPropagatesIteratorValueGetterError verifies that IteratorValue
// does not turn a JavaScript accessor exception into a host assertion panic.
func TestArrayFromPropagatesIteratorValueGetterError(t *testing.T) {
	testSource(t, `const sentinel = {};
const poisonedValue = {};
Object.defineProperty(poisonedValue, "value", {
  get: function() {
    throw sentinel;
  }
});
const items = {};
items[Symbol.iterator] = function() {
  return {
    next: function() {
      return poisonedValue;
    }
  };
};
let caught = null;
try {
  Array.from(items);
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayFromClosesIteratorOnMapperError verifies that an abrupt mapper
// completion is preserved after invoking the iterator's return method.
func TestArrayFromClosesIteratorOnMapperError(t *testing.T) {
	testSource(t, `const sentinel = {};
let closeCount = 0;
const items = {};
items[Symbol.iterator] = function() {
  return {
    return: function() {
      closeCount += 1;
    },
    next: function() {
      return { done: false };
    }
  };
};
let caught = null;
try {
  Array.from(items, function() {
    throw sentinel;
  });
} catch (error) {
  caught = error;
}
assert(caught === sentinel);
assertEqual(closeCount, 1);`)
}

// TestArrayFromPropagatesCustomConstructorError verifies that the iterable
// path stops before iteration when construction completes abruptly.
func TestArrayFromPropagatesCustomConstructorError(t *testing.T) {
	testSource(t, `const sentinel = {};
const C = function() {
  throw sentinel;
};
const items = {};
items[Symbol.iterator] = function() {};
let caught = null;
try {
  Array.from.call(C, items);
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayFromPropagatesLengthSetterError verifies that the iterable path
// observes an abrupt completion while finalizing a custom result object.
func TestArrayFromPropagatesLengthSetterError(t *testing.T) {
	testSource(t, `const sentinel = {};
const C = function() {};
Object.defineProperty(C.prototype, "length", {
  set: function(_) {
    throw sentinel;
  }
});
const items = {};
items[Symbol.iterator] = function() {
  return {
    next: function() {
      return { done: true };
    }
  };
};
let caught = null;
try {
  Array.from.call(C, items);
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayFromClosesIteratorOnElementDefinitionError verifies that failure to
// create an indexed result property becomes a TypeError after iterator close.
func TestArrayFromClosesIteratorOnElementDefinitionError(t *testing.T) {
	testSource(t, `const C = function() {
  Object.defineProperty(this, "0", {
    writable: true,
    configurable: false
  });
};
let closeCount = 0;
let nextCount = 0;
const items = {};
items[Symbol.iterator] = function() {
  return {
    return: function() {
      closeCount += 1;
    },
    next: function() {
      nextCount += 1;
      return { done: nextCount > 1 };
    }
  };
};
let caught = null;
try {
  Array.from.call(C, items);
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);
assertEqual(closeCount, 1);`)
}

// TestArrayFromArrayBuffer verifies that omitting ArrayBuffer's options
// argument is valid and that the buffer is treated as an empty array-like.
func TestArrayFromArrayBuffer(t *testing.T) {
	testSource(t, `const buffer = new ArrayBuffer(7);
const result = Array.from(buffer);
assertEqual(result.length, 0);
assertEqual(new Array(0).length, 0);
assertEqual(new Array("0").length, 1);`)
}

// TestArrayFromUsesCrossRealmConstructorPrototype verifies the test262 realm
// hook and GetPrototypeFromConstructor's realm fallback.
func TestArrayFromUsesCrossRealmConstructorPrototype(t *testing.T) {
	agent := NewAgent()
	InitializeConstants()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	cr.RegisterTest262Runtime(realm)
	Evaluate(`const other = $262.createRealm().global;
const C = new other.Function();
C.prototype = null;
const result = Array.from.call(C, []);
assert(Object.getPrototypeOf(result) === other.Object.prototype);`, realm)
}

// TestForInVarBinding verifies that enumeration assigns each key to the
// existing var binding instead of evaluating the binding node as an expression.
func TestForInVarBinding(t *testing.T) {
	testSource(t, `const object = { a: 1, b: 2 };
let count = 0;
for (var key in object) {
  if (key === "a" || key === "b") {
    count += 1;
  }
}
assertEqual(count, 2);`)
}

// TestArrayFromOverwritesConfigurableNonWritableElement verifies that
// CreateDataProperty replaces a configurable indexed data property.
func TestArrayFromOverwritesConfigurableNonWritableElement(t *testing.T) {
	testSource(t, `const items = { "0": 2, length: 1 };
const C = function() {
  Object.defineProperty(this, "0", {
    value: 1,
    writable: false,
    enumerable: false,
    configurable: true
  });
};
const result = Array.from.call(C, items);
const descriptor = Object.getOwnPropertyDescriptor(result, "0");
assertEqual(descriptor.value, 2);
assertEqual(descriptor.writable, true);
assertEqual(descriptor.enumerable, true);
assertEqual(descriptor.configurable, true);`)
}

// TestGeneratorResumeSequence verifies the next/yield handshake for both a
// suspended yield and normal generator completion.
func TestGeneratorResumeSequence(t *testing.T) {
	testSource(t, `const iterator = function*() {
  yield 2;
}();
const first = iterator.next();
assertEqual(first.value, 2);
assertEqual(first.done, false);
const second = iterator.next();
assertEqual(second.value, undefined);
assertEqual(second.done, true);`)
}

// TestArrayFromGeneratorOverwritesConfigurableNonWritableElement covers the
// generator-backed iterable path through CreateDataProperty.
func TestArrayFromGeneratorOverwritesConfigurableNonWritableElement(t *testing.T) {
	testSource(t, `const items = function*() {
  yield 2;
};
const C = function() {
  Object.defineProperty(this, "0", {
    value: 1,
    writable: false,
    enumerable: false,
    configurable: true
  });
};
const result = Array.from.call(C, items());
const descriptor = Object.getOwnPropertyDescriptor(result, "0");
assertEqual(descriptor.value, 2);
assertEqual(descriptor.writable, true);
assertEqual(descriptor.enumerable, true);
assertEqual(descriptor.configurable, true);`)
}

// TestArrayOfPropagatesLengthSetterError verifies that Array.of forwards the
// final Set operation's abrupt completion.
func TestArrayOfPropagatesLengthSetterError(t *testing.T) {
	testSource(t, `const sentinel = {};
const C = function(length) {
  assertEqual(length, 0);
  Object.defineProperty(this, "length", {
    set: function() {
      throw sentinel;
    }
  });
};
let caught = null;
try {
  Array.of.call(C);
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayOfPropagatesProxyDefinePropertyError verifies that an abrupt
// defineProperty trap completion is not converted into a host panic.
func TestArrayOfPropagatesProxyDefinePropertyError(t *testing.T) {
	testSource(t, `const sentinel = {};
const C = function() {
  return new Proxy({}, {
    defineProperty: function() {
      throw sentinel;
    }
  });
};
let caught = null;
try {
  Array.of.call(C, "Bob");
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayLengthCoercesPrimitiveAndBoxedValues verifies that array exotic
// length assignment performs numeric conversion instead of requiring a number.
func TestArrayLengthCoercesPrimitiveAndBoxedValues(t *testing.T) {
	testSource(t, `let array = [];
array.length = true;
assertEqual(array.length, 1);
array = [0];
array.length = null;
assertEqual(array.length, 0);
array = [0];
array.length = new Boolean(false);
assertEqual(array.length, 0);
array = [];
array.length = new Number(1);
assertEqual(array.length, 1);
array = [];
array.length = "1";
assertEqual(array.length, 1);
array = [];
array.length = new String("1");
assertEqual(array.length, 1);
array = [0, 1];
Object.defineProperty(array, "length", {
  value: 1,
  writable: false
});
assertEqual(array.length, 1);
assertEqual(Object.getOwnPropertyDescriptor(array, "length").writable, false);
array.length = 2;
assertEqual(array.length, 1);`)
}

// TestArrayLengthPropagatesNumericCoercionError verifies that a valueOf
// exception remains catchable JavaScript state during ToUint32.
func TestArrayLengthPropagatesNumericCoercionError(t *testing.T) {
	testSource(t, `const sentinel = {};
const array = [];
let caught = null;
try {
  array.length = {
    valueOf: function() {
      throw sentinel;
    }
  };
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayLengthCoercionPrecedesWritableCheck verifies both numeric coercions
// happen before a side effect makes the current length non-writable.
func TestArrayLengthCoercionPrecedesWritableCheck(t *testing.T) {
	testSource(t, `let array = [1, 2, 3];
let coercions = 0;
const length = {};
length[Symbol.toPrimitive] = function(hint) {
  assertEqual(hint, "number");
  coercions += 1;
  Object.defineProperty(array, "length", { writable: false });
  return 0;
};
assertEqual(Reflect.set(array, "length", length), false);
assertEqual(coercions, 2);
assertEqual(array.length, 3);
array = [1, 2, 3];
coercions = 0;
let caught = null;
try {
  (function() {
    "use strict";
    array.length = length;
  })();
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);
assertEqual(coercions, 2);
assertEqual(array.length, 3);`)
}

// TestArrayLengthDescriptorValidationAfterCoercion verifies Object and Reflect
// surface the same failed [[DefineOwnProperty]] with their distinct APIs.
func TestArrayLengthDescriptorValidationAfterCoercion(t *testing.T) {
	testSource(t, `let array = [1, 2];
let calls = 0;
const length = {
  valueOf: function() {
    calls += 1;
    if (calls === 2) {
      Object.defineProperty(array, "length", { writable: false });
    }
    return array.length;
  }
};
let caught = null;
try {
  Object.defineProperty(array, "length", { value: length, writable: true });
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);
assertEqual(calls, 2);
array = [1, 2];
calls = 0;
assertEqual(
  Reflect.defineProperty(array, "length", { value: length, writable: true }),
  false
);
assertEqual(calls, 2);`)
}

// TestArrayLengthRejectsAccessorDescriptor verifies the non-configurable data
// property cannot be converted to an accessor and its getter is never called.
func TestArrayLengthRejectsAccessorDescriptor(t *testing.T) {
	testSource(t, `let getterCalled = false;
let caught = null;
try {
  Object.defineProperty([], "length", {
    get: function() {
      getterCalled = true;
      return 0;
    }
  });
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);
assertEqual(getterCalled, false);
assertEqual(
  Reflect.defineProperty([], "length", { set: function(_) {} }),
  false
);
const array = [];
Object.defineProperty(array, "length", { writable: false });
caught = null;
try {
  Object.defineProperty(array, "length", { writable: true });
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);`)
}

// TestObjectGetPrototypeOfReturnsNull verifies a null internal prototype is
// exposed as the JavaScript null value.
func TestObjectGetPrototypeOfReturnsNull(t *testing.T) {
	testSource(t, `const object = Object.create(null);
assert(Object.getPrototypeOf(object) === null);`)
}

// TestArrayCopyWithinBoxesBooleanWithoutArguments verifies omitted parameters
// are treated as undefined and the boxed receiver is returned.
func TestArrayCopyWithinBoxesBooleanWithoutArguments(t *testing.T) {
	testSource(t, `assert(Array.prototype.copyWithin.call(true) instanceof Boolean);
assert(Array.prototype.copyWithin.call(false) instanceof Boolean);`)
}

// TestArraySetPrototypeOf verifies array exotic objects use the ordinary
// prototype mutation algorithm without assuming the concrete base object type.
func TestArraySetPrototypeOf(t *testing.T) {
	testSource(t, `const array = [];
const prototype = [1, 2, 3];
Object.setPrototypeOf(array, prototype);
assert(Object.getPrototypeOf(array) === prototype);
assertEqual(array[1], 2);`)
}

// TestArrayCopyWithinPropagatesElementGetterError verifies Get(O, fromKey)
// preserves an abrupt accessor completion.
func TestArrayCopyWithinPropagatesElementGetterError(t *testing.T) {
	testSource(t, `const sentinel = {};
const object = { length: 1 };
Object.defineProperty(object, "0", {
  get: function() {
    throw sentinel;
  }
});
let caught = null;
try {
  Array.prototype.copyWithin.call(object, 0, 0);
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayCopyWithinPropagatesLengthConversionError verifies an invalid
// symbolic length remains a catchable TypeError completion.
func TestArrayCopyWithinPropagatesLengthConversionError(t *testing.T) {
	testSource(t, `let caught = null;
try {
  Array.prototype.copyWithin.call({ length: Symbol() }, 0, 0);
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);`)
}

// TestArrayCopyWithinThrowsWhenTargetCannotBeDeleted verifies the
// DeletePropertyOrThrow branch turns a false internal delete into TypeError.
func TestArrayCopyWithinThrowsWhenTargetCannotBeDeleted(t *testing.T) {
	testSource(t, `const object = { length: 2 };
Object.defineProperty(object, "1", {
  configurable: false,
  writable: true
});
let caught = null;
try {
  Array.prototype.copyWithin.call(object, 1, 0);
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);`)
}

// TestArrayCopyWithinPropagatesProxyDeleteError verifies an abrupt
// deleteProperty trap completion is preserved by [[Delete]].
func TestArrayCopyWithinPropagatesProxyDeleteError(t *testing.T) {
	testSource(t, `const sentinel = {};
const proxy = new Proxy({ "1": true, length: 2 }, {
  deleteProperty: function(target, property) {
    if (property === "1") {
      throw sentinel;
    }
  }
});
let caught = null;
try {
  Array.prototype.copyWithin.call(proxy, 1, 0);
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayCopyWithinPropagatesProxyHasError verifies an abrupt has trap
// completion is preserved by [[HasProperty]].
func TestArrayCopyWithinPropagatesProxyHasError(t *testing.T) {
	testSource(t, `const sentinel = {};
const proxy = new Proxy({ length: 1 }, {
  has: function(target, property) {
    if (property === "0") {
      throw sentinel;
    }
    return property in target;
  }
});
let caught = null;
try {
  Array.prototype.copyWithin.call(proxy, 0, 0);
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestInOperatorUsesHasProperty verifies property-key conversion and prototype
// lookup through the language-level relational operator.
func TestInOperatorUsesHasProperty(t *testing.T) {
	testSource(t, `const prototype = { inherited: true };
const object = Object.create(prototype);
object.own = true;
assert("own" in object);
assert("inherited" in object);
assertEqual("missing" in object, false);`)
}

// TestObjectConstructorWithoutArguments verifies both call and construct forms
// allocate an ordinary object when no value is supplied.
func TestObjectConstructorWithoutArguments(t *testing.T) {
	testSource(t, `const called = Object();
const constructed = new Object();
assert(Object.getPrototypeOf(called) === Object.prototype);
assert(Object.getPrototypeOf(constructed) === Object.prototype);`)
}

// TestVoidAndSequenceExpressionEvaluation verifies void preserves operand side
// effects and a comma expression yields its final value.
func TestVoidAndSequenceExpressionEvaluation(t *testing.T) {
	testSource(t, `let value = 0;
assert((value = 1, value = 2, value) === 2);
assert((void (value = 3)) === undefined);
assertEqual(value, 3);
const values = [void 0, void 0, void 0];
assertEqual(values.length, 3);`)
}

// TestFunctionLengthDoesNotChangeWithIndexedProperties verifies the formal
// parameter count remains the array-like length used by concat.
func TestFunctionLengthDoesNotChangeWithIndexedProperties(t *testing.T) {
	testSource(t, `const functionObject = function(a, b, c) {};
assertEqual(functionObject.length, 3);
functionObject[0] = 1;
functionObject[1] = 2;
functionObject[2] = 3;
assertEqual(functionObject.length, 3);
assertEqual((function(a, b, c) {}).length, 3);
Function.prototype[Symbol.isConcatSpreadable] = true;
const inherited = function(a, b, c) {};
assertEqual(inherited[Symbol.isConcatSpreadable], true);
const result = [].concat(inherited);
assertEqual(result.length, 3);
assertEqual(0 in result, false);
assertEqual(1 in result, false);
assertEqual(2 in result, false);
delete Function.prototype[Symbol.isConcatSpreadable];`)
}

// TestRegularExpressionLiteralEvaluation verifies a regex literal creates a
// RegExp instance and can participate in ordinary property operations.
func TestRegularExpressionLiteralEvaluation(t *testing.T) {
	testSource(t, `const regexp = /abc/;
assert(regexp instanceof RegExp);
regexp.extra = 1;
assertEqual(regexp.extra, 1);`)
}

// TestArrayConcatPropagatesSpreadableGetterError verifies the
// Symbol.isConcatSpreadable lookup preserves an abrupt accessor completion.
func TestArrayConcatPropagatesSpreadableGetterError(t *testing.T) {
	testSource(t, `const sentinel = {};
const value = {};
Object.defineProperty(value, Symbol.isConcatSpreadable, {
  get: function() {
    throw sentinel;
  }
});
let caught = null;
try {
  [].concat(value);
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayConcatPropagatesLengthGetterError verifies LengthOfArrayLike
// preserves an abrupt length accessor completion.
func TestArrayConcatPropagatesLengthGetterError(t *testing.T) {
	testSource(t, `const sentinel = {};
const value = {};
value[Symbol.isConcatSpreadable] = true;
Object.defineProperty(value, "length", {
  get: function() {
    throw sentinel;
  }
});
let caught = null;
try {
  [].concat(value);
} catch (error) {
  caught = error;
}
assert(caught === sentinel);`)
}

// TestArrayForEachDefaultsThisArgument verifies the optional thisArg is
// undefined and all present elements are visited.
func TestArrayForEachDefaultsThisArgument(t *testing.T) {
	testSource(t, `let count = 0;
[1, 2, 3].forEach(function(value, index) {
  assertEqual(value, index + 1);
  count += 1;
});
assertEqual(count, 3);`)
}

// TestSloppyDuplicateParameters verifies the shared binding takes the last
// argument while the arguments object retains every indexed value.
func TestSloppyDuplicateParameters(t *testing.T) {
	testSource(t, `const result = (function(a, a, a) {
  return {
    binding: a,
    values: [arguments[0], arguments[1], arguments[2]]
  };
})(1, 2, 3);
assertEqual(result.binding, 3);
assertEqual(result.values[0], 1);
assertEqual(result.values[1], 2);
assertEqual(result.values[2], 3);`)
}

// TestArrayConcatPropagatesArgumentsGetterError verifies element access on a
// mapped arguments exotic object preserves an abrupt getter completion.
func TestArrayConcatPropagatesArgumentsGetterError(t *testing.T) {
	testSource(t, `const sentinel = {};
const args = (function(a) {
  return arguments;
})(1);
let calls = 0;
const getter = function() {
  calls++;
  throw sentinel;
};
Object.defineProperty(args, "0", {
  get: getter
});
const descriptor = Object.getOwnPropertyDescriptor(args, "0");
assert(descriptor.get === getter, "mapped argument descriptor getter");
let direct = null;
try {
  direct = args[0];
} catch (error) {
  direct = error;
}
assert(calls === 1, "mapped argument getter call count");
assert(direct !== null, "direct mapped argument getter throws");
assert(direct === sentinel, "direct mapped argument getter");
args[Symbol.isConcatSpreadable] = true;
let caught = null;
try {
  [].concat(args);
} catch (error) {
  caught = error;
}
assert(caught === sentinel, "concat mapped argument getter");`)
}

// TestArrayConcatRejectsNonObjectConstructor verifies ArraySpeciesCreate
// reports a catchable TypeError when an array constructor is not an object.
func TestArrayConcatRejectsNonObjectConstructor(t *testing.T) {
	testSource(t, `const values = [null, 1, "string", true];
values.forEach(function(value) {
  const array = [];
  array.constructor = value;
  let caught = null;
  try {
    array.concat();
  } catch (error) {
    caught = error;
  }
  assert(caught instanceof TypeError);
});`)
}

// TestArrayConcatRejectsRevokedProxy verifies IsArray propagates the
// catchable TypeError produced by a revoked Proxy.
func TestArrayConcatRejectsRevokedProxy(t *testing.T) {
	testSource(t, `const record = Proxy.revocable([], {});
let constructorCalls = 0;
Object.defineProperty(record.proxy, "constructor", {
  get: function() {
    constructorCalls++;
  }
});
record.revoke();
let caught = null;
try {
  Array.prototype.concat.call(record.proxy);
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);
assert(constructorCalls === 0);`)
}

// TestArrayConcatRejectsRevokedSpreadabilityProxy verifies a revoked Proxy
// throws while concat reads Symbol.isConcatSpreadable.
func TestArrayConcatRejectsRevokedSpreadabilityProxy(t *testing.T) {
	testSource(t, `const record = Proxy.revocable({}, {});
record.revoke();
let caught = null;
try {
  [].concat(record.proxy);
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);`)
}

// TestObjectBooleanCoercionIsAlwaysTrue verifies wrapper objects do not use
// their wrapped primitive value during ToBoolean.
func TestObjectBooleanCoercionIsAlwaysTrue(t *testing.T) {
	testSource(t, `assert(Boolean(new Boolean(false)) === true);
assert(Boolean(new String("")) === true);
assert(Boolean(new String()) === true);
assert(Boolean(new Number(0)) === true);
assert(Boolean(new Number(NaN)) === true);`)
}

// TestBooleanPrototypeMethodsRejectOtherObjects verifies the Boolean
// prototype methods return catchable TypeErrors for incompatible receivers.
func TestBooleanPrototypeMethodsRejectOtherObjects(t *testing.T) {
	testSource(t, `const values = [new String(), new Date(), {}, { x: 1 }];
values.forEach(function(value) {
  let toStringError = null;
  let valueOfError = null;
  try {
    Boolean.prototype.toString.call(value);
  } catch (error) {
    toStringError = error;
  }
  try {
    Boolean.prototype.valueOf.call(value);
  } catch (error) {
    valueOfError = error;
  }
  assert(toStringError instanceof TypeError);
  assert(valueOfError instanceof TypeError);
});`)
}

// TestForInAssignmentTargetAndEscapedString verifies a bare assignment target
// in for-in loops and escaped delimiters in string literals.
func TestForInAssignmentTargetAndEscapedString(t *testing.T) {
	testSource(t, `var key;
var seen = false;
for (key in { visible: 1 }) {
  if (key === "visible") {
    seen = true;
  }
}
assert(seen);
assert('single quote: \'' === "single quote: '");`)
}

// TestBooleanWrapperKeepsItsExoticIdentity verifies fallback
// Object.prototype.toString observes the Boolean wrapper brand.
func TestBooleanWrapperKeepsItsExoticIdentity(t *testing.T) {
	testSource(t, `const original = Boolean.prototype.toString;
delete Boolean.prototype.toString;
const object = new Boolean();
assert(object.toString() === "[object Boolean]");
Boolean.prototype.toString = original;`)
}

// TestNumberPrototypeMethodsRejectOtherObjects verifies incompatible receiver
// failures are catchable JavaScript TypeErrors.
func TestNumberPrototypeMethodsRejectOtherObjects(t *testing.T) {
	testSource(t, `const values = [new String(), new Boolean(), new Date(), {}, []];
values.forEach(function(value) {
  let caught = null;
  try {
    Number.prototype.valueOf.call(value);
  } catch (error) {
    caught = error;
  }
  assert(caught instanceof TypeError);
});`)
}

// TestNumberPredicatesWithoutArguments verifies missing arguments are treated
// as undefined and return false.
func TestNumberPredicatesWithoutArguments(t *testing.T) {
	testSource(t, `assert(Number.isFinite() === false);
assert(Number.isInteger() === false);
assert(Number.isNaN() === false);
assert(Number.isSafeInteger() === false);`)
}

// TestNumberToFixed verifies fixed-point formatting, coercion, and range
// validation for Number primitives and wrappers.
func TestNumberToFixed(t *testing.T) {
	testSource(t, `assert(Number.prototype.toFixed() === "0");
assert((new Number(1)).toFixed(1) === "1.0");
assert((123.456).toFixed(2) === "123.46");
assert((-0).toFixed(2) === "0.00");
assert(NaN.toFixed(1) === "NaN");
assert(String(1e21) === "1e+21");
assert(String(-0) === "0");
let caught = null;
try {
  (3).toFixed(101);
} catch (error) {
  caught = error;
}
assert(caught instanceof RangeError);`)
}

// TestNumberToExponential verifies exponential formatting and non-finite
// handling.
func TestNumberToExponential(t *testing.T) {
	testSource(t, `assert((123.456).toExponential(2) === "1.23e+2");
assert((0).toExponential(2) === "0.00e+0");
assert((Infinity).toExponential(1000) === "Infinity");
assert((NaN).toExponential(Infinity) === "NaN");
let caught = null;
try {
  (3).toExponential(101);
} catch (error) {
  caught = error;
}
assert(caught instanceof RangeError);`)
}

// TestNumberToPrecision verifies significant-digit formatting and its
// fixed/exponential threshold.
func TestNumberToPrecision(t *testing.T) {
	testSource(t, `assert((7).toPrecision(3) === "7.00");
assert((10).toPrecision(1) === "1e+1");
assert((0.000001).toPrecision(2) === "0.0000010");
assert((0.0000001).toPrecision(2) === "1.0e-7");
assert((Infinity).toPrecision(1000) === "Infinity");
assert((42).toPrecision() === "42");
let caught = null;
try {
  (3).toPrecision(0);
} catch (error) {
  caught = error;
}
assert(caught instanceof RangeError);`)
}

// TestNumericOverflowProducesInfinity verifies decimal parsing accepts
// strconv range results as JavaScript infinities.
func TestNumericOverflowProducesInfinity(t *testing.T) {
	testSource(t, `assert(10e10000 === Infinity);
assert(Number("10e10000") === Infinity);
assert(parseFloat("-10e10000") === -Infinity);`)
}

// TestBigIntToNumberConversion verifies unary BigInt operations allocate a
// result and preserve Number conversion rounding.
func TestBigIntToNumberConversion(t *testing.T) {
	testSource(t, `assert(Number(2n ** 53n + 1n) === 9007199254740992);
assert(Number(-(2n ** 53n + 3n)) === -9007199254740996);
assert(~0n === -1n);`)
}

func TestBigIntWidthTruncation(t *testing.T) {
	testSource(t, `assert(BigInt.asIntN(8, 0xabcdef0123456789abcdef0183n) === -0x7dn);
assert(BigInt.asUintN(8, 0xabcdef0123456789abcdef0183n) === 0x83n);`)
}

func TestBigIntStringAndReceiverConversions(t *testing.T) {
	testSource(t, `assert(BigInt("") === 0n);
assert(BigInt("  0o12  ") === 10n);
assert((100n).toString() === "100");
let rangeCaught = null;
try {
  BigInt(1.1);
} catch (error) {
  rangeCaught = error;
}
assert(rangeCaught instanceof RangeError);
let caught = null;
try {
  BigInt.prototype.valueOf.call({});
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);`)
}

// TestNumberToStringRadixDigits verifies lowercase digits a-z are used for
// integer radices through 36.
func TestNumberToStringRadixDigits(t *testing.T) {
	testSource(t, `assert((10).toString(11) === "a");
assert((35).toString(36) === "z");
assert((31).toString(16) === "1f");
assert((-0).toString(2) === "0");
let total = 0;
for (let index = 0; index < 3; index++) {
  total += index;
}
assert(total === 3);`)
}

// TestSymbolConstructionThrows verifies Symbol rejects construction with a
// catchable TypeError.
func TestSymbolConstructionThrows(t *testing.T) {
	testSource(t, `let caught = null;
try {
  new Symbol("description");
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);`)
}

// TestSymbolKeyForRejectsNonSymbols verifies non-symbol inputs produce
// catchable TypeErrors.
func TestSymbolKeyForRejectsNonSymbols(t *testing.T) {
	testSource(t, `const symbolObject = Object(Symbol("s"));
assert(typeof symbolObject === "object");
const values = [null, undefined, "key", {}, [], symbolObject];
values.forEach(function(value) {
  let caught = null;
  try {
    Symbol.keyFor(value);
  } catch (error) {
    caught = error;
  }
  assert(caught instanceof TypeError);
});`)
}

// TestSymbolDescriptionAccessor verifies descriptor shape and the distinction
// between an absent description and an empty string.
func TestSymbolDescriptionAccessor(t *testing.T) {
	testSource(t, `const descriptor = Object.getOwnPropertyDescriptor(Symbol.prototype, "description");
assert(typeof descriptor.get === "function");
assert(descriptor.set === undefined);
assert(descriptor.writable === undefined);
assert(Symbol().description === undefined);
assert(Symbol(undefined).description === undefined);
assert(Symbol("").description === "");
assert(Object(Symbol("test")).description === "test");`)
}

// TestSymbolWrapperOrdinaryStringConversion verifies deleting @@toPrimitive
// falls back to the ordinary toString/valueOf path.
func TestSymbolWrapperOrdinaryStringConversion(t *testing.T) {
	testSource(t, `const method = Symbol.prototype[Symbol.toPrimitive];
delete Symbol.prototype[Symbol.toPrimitive];
assert("".concat(Object(Symbol())) === "Symbol()");
assert(String(Object(Symbol("test"))) === "Symbol(test)");
Symbol.prototype[Symbol.toPrimitive] = method;`)
}

func TestBaselineNew(t *testing.T) {
	sourceTexts := []string{
		`
const p = new Promise((res, rej) => {
    res(1);
});

async function asyncReturn() {
    return p;
}

function basicReturn() {
    return Promise.resolve(p);
}

assertEqual(p === basicReturn(), true); // true
assertEqual(p === asyncReturn(), false); // false
`,
		`
async function b() {
    return 'b'
}

async function asyncCall() {
    const result = await b()
    assertEqual(result, 'b')
}
asyncCall();
`,
		fmt.Sprintf(
			`
const a = 5;
const b = 10;
assertEqual(%[1]sFifteen is ${a + b} and\nnot ${2 * a + b}.%[1]s,
'Fifteen is 15 and\nnot 20.');
`, "`"),
		`
let a = 1
setTimeout(() => {
    a = 2
    assert(a === 2)
}, 0);
a=3`,
		`2 == 1;
2 > 1;
false || 1;
true && 1;
true ? 2 : 1;
2 ** 3;
() => 123;
`,
		`class A {
	static sa = 'a';
}
const B = class {}`,
		`function a() {
}
function* b() {
}
async function* c() {
}
const e = async () => {
}`,
		"``",
		`let a = 1;
let b = a + 1;
let c = a + b + b;
'hello' + a + "world";
c += a + b;
a ? b : c;
var d = 1 ?	2 : 3;`,
		`function toString(a) {
	return a;
}
let a = toString(10);
a + 'hello';`,
		`let a = 1;
for (var i = 0; i < 3; i++) {
	a += i;
}
assert(a === 4);`,
		`var a = [];
var a = [1,2,3];
a[4294967295] = "not an array element";
a[4294967295];`,
		`const a = {
  b: 8,
  writable: false,
  enumerable: false,
  configurable: false
};
delete a.b;
let b = 2;
a.b = b;
a.b;`,
		`const desc = Object.getOwnPropertyDescriptor(BigInt64Array, 'BYTES_PER_ELEMENT');
let a = 0;
if (Object.prototype.hasOwnProperty.call(desc, 'enumerable')) {
	a = 1;
}
a;`,
		`
assert(Boolean(function() {}()) === false);
assert(typeof Boolean(void 0) === "boolean");
void 0;
`,
		`const sab = new SharedArrayBuffer(1024);
const ta = new Uint8Array(sab);
ta[0] = 5; // 5
ta[123] = 12;
Atomics.add(ta, 0, 12); // 5
Atomics.and(ta, 0, 1); 
Atomics.compareExchange(ta, 0, 5, 12); // 1
Atomics.exchange(ta, 0, 12); // 1
Atomics.isLockFree(1); // true
Atomics.isLockFree(2); // true
Atomics.isLockFree(3); // false
Atomics.isLockFree(4); // true
Atomics.or(ta, 0, 1); // 12
Atomics.store(ta, 0, 12); // 12
Atomics.sub(ta, 0, 2); // 12
Atomics.xor(ta, 0, 1); // 10
Atomics.load(ta, 0); 
`,
		`const a = {};
const b = a?.b ?? 1;
assert(b === 1, 'b should be 1')`,
		`
class ValidatorClass {
  get [Symbol.toStringTag]() {
    return 'Validator';
  }
}
const v = new ValidatorClass();
assert(Object.prototype.toString.call(v) === '[object Validator]', 'should be [object Validator]');
`,
		// TODO: String locale compare
		"const string3 = `Yet another string primitive`;",
		`
const string1 = "A string primitive";
const string2 = 'Also a string primitive';
const string4 = new String("A String object");
assert("cat".charAt(1) === 'a');
assert("cat"[1] === 'a');
const strPrim = "foo"; // A literal is a string primitive
const strPrim2 = String(1); // Coerced into the string primitive "1"
const strPrim3 = String(true); // Coerced into the string primitive "true"
const strObj = new String(strPrim); // String with new returns a string wrapper object.

assert(typeof strPrim === "string", "typeof: expected string");
assert(typeof strPrim2 === "string", "typeof: expected string");
assert(typeof strPrim3 === "string", "typeof: expected string");
assert(typeof strObj === "object", "typeof: expected object"); 
`,
		`
const foo = function* () {
  yield 'a';
  yield 'b';
  yield 'c';
};

let str = '';
for (const val of foo()) {
  str = str + val;
}
assert(str === 'abc');

const gen = foo();
assert(gen.next().value === 'a');
assert(gen.next().value === 'b');
assert(gen.next().value === 'c');

const infinite = function*() {
  let index = 0;

  while (true) {
    yield index++;
  }
}

const generator = infinite(); // "Generator { }"

assert(generator.next().value === 0); // 0
assert(generator.next().value === 1); // 1
assert(generator.next().value === 2); // 2
`,
		`
const a = [1,2,3]
for (const i in a) {
	assert(i === '0' || i === '1' || i === '2');
}

const array1 = ['a', 'b', 'c'];
const iterator1 = array1[Symbol.iterator]();

for (const value of iterator1) {
    assert(value === 'a' || value === 'b' || value === 'c');
}
`,
		`
const object1 = {
    [Symbol.toPrimitive](hint) {
        if (hint === 'number') {
            return 42;
        }
        return null;
    },
};

assert(+object1 === 42);
`,
		`
const map1 = new Map();
map1.set('a', 1);
map1.set('b', 2);
map1.set('c', 3);
assert(map1.get('a') === 1);
map1.set('a', 97);
assert(map1.get('a') === 97);
assert(map1.size === 3);
map1.delete('b');
assert(map1.size === 2);
`,
		`
const map1 = new Map();

map1.set('0', 'foo');
map1.set(1, 'bar');

const iterator1 = map1[Symbol.iterator]();

assertEqual(iterator1.next().value[0], 0);
assertEqual(iterator1.next().value[1], 'bar');
`,
		`
const set1 = new Set();

set1.add(42);
set1.add('forty two');

const iterator1 = set1[Symbol.iterator]();

assertEqual(iterator1.next().value, 42);
assertEqual(iterator1.next().value, 'forty two');
`,
		`
const num1 = 42;
const num2 = 3.14;
const num3 = Number('123');
const num4 = parseInt('123', 10);
const num5 = parseFloat('3.14');
const num6 = 0b1010; // binary
const num7 = 0o52; // octal
const num8 = 0x2A; // hexadecimal
const num9 = 8.64e15;

assert(num1 === 42);
assert(num2 === 3.14);
assert(num3 === 123);
assert(num4 === 123);
assert(num5 === 3.14);
assert(num6 === 10);
assert(num7 === 42);
assert(num8 === 42);
assert(num9 === 8640000000000000);
`,
		`
const utcDate1 = new Date(Date.UTC(96, 1, 2, 3, 4, 5));
const utcDate2 = new Date(Date.UTC(0, 0, 0, 0, 0, 0));

assertEqual(utcDate1.toUTCString(), "Fri, 02 Feb 1996 03:04:05 UTC");
assertEqual(utcDate2.toUTCString(), "Sun, 31 Dec 1899 00:00:00 UTC");

// check time zone
//assertEqual(new Date(8.64e15).toString(), "Sat Sep 13 275760 08:00:00 CST+0800"); 
assertEqual(new Date(8.64e15 + 1).toString(), "Invalid Date"); 

const date = new Date('December 17, 1995 03:24:00');

date[Symbol.toPrimitive]('string');
assertEqual(date[Symbol.toPrimitive]('number'), 819170640000);
`,
		`
const str = 'table football';
const regex = new RegExp('foo*');
const globalRegex = new RegExp('foo*', 'g');
assertEqual(regex.test(str), true);
assertEqual(globalRegex.lastIndex, 0);
assertEqual(globalRegex.test(str), true);
assertEqual(globalRegex.lastIndex, 9);
assertEqual(globalRegex.test(str), false);
`,
		fmt.Sprintf(`
class Animal {
    constructor(name) {
        this.name = name;
    }

    speak() {
        return this.name + ' makes a noise.';
    }
}

class Dog extends Animal {
    constructor(name) {
        super(name); // call the super class constructor and pass in the name parameter
    }

    speak() {
        return this.name + ' barks.';
    }
}

const d = new Dog("Mitzie");
assertEqual(d.name, "Mitzie");
assert(d instanceof Dog);
assert(d instanceof Animal);
assertEqual(d.speak(), "Mitzie barks.");
assertEqual(%[1]sstring text line 1\nstring text line 2%[1]s, 'string text line 1\nstring text line 2');
`, "`"),
		`
const previouslyMaxSafeInteger = 9007199254740991n;
const alsoHuge = BigInt(9007199254740991);
const hugeString = BigInt("9007199254740991");
const hugeHex = BigInt("0x1fffffffffffff");
const hugeOctal = BigInt("0o377777777777777777");
const hugeBin = BigInt(
    "0b11111111111111111111111111111111111111111111111111111",
);

assertEqual(previouslyMaxSafeInteger.toString(), "9007199254740991n");
assertEqual(alsoHuge.toString(), "9007199254740991n");
assertEqual(hugeString.toString(), "9007199254740991n");
assertEqual(hugeHex.toString(), "9007199254740991n");
assertEqual(hugeOctal.toString(), "9007199254740991n");
assertEqual(hugeBin.toString(), "9007199254740991n");
assert(typeof 1n === "bigint");
assert(typeof BigInt("1") === "bigint")
`,
		`
const duck = {
    name: "Maurice",
    color: "white",
    greeting() {
        console.log("Quaaaack! My name is" + this.name);
    },
};

assert(Reflect.has(duck, "color"));
assert(Reflect.has(duck, "haircut") === false);
assert(Reflect.ownKeys(duck).length===3);
assert(Reflect.set(duck, "eyes", "black"));
assertEqual(Reflect.get(duck, "eyes"), "black");
`,
		`
let a = ''
const promiseA = new Promise((resolve, reject) => {
    resolve(777);
});
promiseA.then((val) => {
	a += val
	assert(a === '1777');
});
a += '1'
assert(a === '1');
`,
		`
// proxy1
const target = {
    message1: "hello",
    message2: "everyone",
};

const handler1 = {};

const proxy1 = new Proxy(target, handler1);
assertEqual(proxy1.message1, "hello"); 
assertEqual(proxy1.message2, "everyone"); 
`, `
// proxy2
const target = {
  message1: "hello",
  message2: "everyone",
};

const handler2 = {
  get(target, prop, receiver) {
    return "world";
  },
};

const proxy2 = new Proxy(target, handler2);
assertEqual(proxy2.message1, "world"); // world
assertEqual(proxy2.message2, "world"); // world
`,
		`
let a = 1;
for (var i = 0; i < 3; i++) {
	a += i;
}
assert(a === 4);
`,
		`
function toString(a) {
    return a;
}
let a = toString(10);
assert(a + 'hello' === '10hello');
`,
		`
true;
false;
null;
2;
assert(2 != 1);
assert(2 > 1);
assert(2 >= 2);
assert(2 < 31);
`, `const expr = "Papayas";
let a = 1;
switch (expr) {
    case "Oranges":
        a = 2
        break;
    case "Mangoes":
    case "Papayas":
        a = 3
        break;
    default:
        a = 4
}

assert(a === 3);
`,
	}
	testNewSources(t, sourceTexts)
}

func testNewSources(t *testing.T, sourceTexts []string) {
	for _, sourceText := range sourceTexts {
		fmt.Println("Testing source:" + sourceText)
		testSource(t, sourceText)
	}
}

func resolveTestdataPath(f string) string {
	dir, _ := os.Getwd()
	return path.Join(dir, "..", "testdata", f)
}

func TestModule(t *testing.T) {
	sourceFiles := []string{
		"simple_import.js",
	}
	for _, f := range sourceFiles {
		testModule(t, f)
	}
}
