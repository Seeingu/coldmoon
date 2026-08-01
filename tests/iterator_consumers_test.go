package tests

import "testing"

// TestCollectionConstructorsConsumeIterables verifies the successful iterable
// path shared by Map, Set, and Object.fromEntries.
func TestCollectionConstructorsConsumeIterables(t *testing.T) {
	testSource(t, `
const map = new Map([["first", 1], ["second", 2]]);
assertEqual(map.get("first"), 1);
assertEqual(map.get("second"), 2);

const set = new Set([1, 2, 1]);
assert(set.has(1));
assert(set.has(2));

const object = Object.fromEntries([["answer", 42]]);
assertEqual(object.answer, 42);
`)
}

// TestObjectFromEntriesClosesIteratorForPrimitiveEntry verifies that entry
// validation becomes a catchable TypeError after closing exactly once.
func TestObjectFromEntriesClosesIteratorForPrimitiveEntry(t *testing.T) {
	testSource(t, `
let nextCount = 0;
let closeCount = 0;
const iterable = {};
iterable[Symbol.iterator] = function() {
  return {
    next: function() {
      nextCount += 1;
      return { done: false, value: null };
    },
    return: function() {
      closeCount += 1;
      return {};
    }
  };
};

let caught = null;
try {
  Object.fromEntries(iterable);
} catch (error) {
  caught = error;
}
assert(caught instanceof TypeError);
assertEqual(nextCount, 1);
assertEqual(closeCount, 1);
`)
}

// TestObjectFromEntriesPreservesKeyGetterError verifies that IteratorClose
// still observes the return getter while preserving the original throw.
func TestObjectFromEntriesPreservesKeyGetterError(t *testing.T) {
	testSource(t, `
const original = {};
const closeFailure = {};
const entry = {};
Object.defineProperty(entry, "0", {
  get: function() { throw original; }
});

let closeCount = 0;
const iterable = {};
iterable[Symbol.iterator] = function() {
  const iterator = {
    next: function() { return { done: false, value: entry }; }
  };
  Object.defineProperty(iterator, "return", {
    get: function() {
      closeCount += 1;
      throw closeFailure;
    }
  });
  return iterator;
};

let caught = null;
try {
  Object.fromEntries(iterable);
} catch (error) {
  caught = error;
}
assert(caught === original);
assertEqual(closeCount, 1);
`)
}

// TestObjectFromEntriesPreservesValueGetterError covers the second entry
// property lookup and the original-throw precedence over return() failures.
func TestObjectFromEntriesPreservesValueGetterError(t *testing.T) {
	testSource(t, `
const original = {};
const closeFailure = {};
const entry = { 0: "key" };
Object.defineProperty(entry, "1", {
  get: function() { throw original; }
});

let closeCount = 0;
const iterable = {};
iterable[Symbol.iterator] = function() {
  return {
    next: function() { return { done: false, value: entry }; },
    return: function() {
      closeCount += 1;
      throw closeFailure;
    }
  };
};

let caught = null;
try {
  Object.fromEntries(iterable);
} catch (error) {
  caught = error;
}
assert(caught === original);
assertEqual(closeCount, 1);
`)
}

// TestMapConstructorClosesIteratorWhenAdderThrows verifies that a custom Map
// adder cannot be ignored and that iteration stops before a second next call.
func TestMapConstructorClosesIteratorWhenAdderThrows(t *testing.T) {
	testSource(t, `
const original = {};
const closeFailure = {};
let nextCount = 0;
let closeCount = 0;
const iterable = {};
iterable[Symbol.iterator] = function() {
  return {
    next: function() {
      nextCount += 1;
      if (nextCount === 1) {
        return { done: false, value: ["key", "value"] };
      }
      return { done: true };
    },
    return: function() {
      closeCount += 1;
      throw closeFailure;
    }
  };
};
Map.prototype.set = function() { throw original; };

let caught = null;
try {
  new Map(iterable);
} catch (error) {
  caught = error;
}
assert(caught === original);
assertEqual(nextCount, 1);
assertEqual(closeCount, 1);
`)
}

// TestSetConstructorClosesIteratorWhenAdderThrows applies the same completion
// and close-once contract to Set's single-value ingestion path.
func TestSetConstructorClosesIteratorWhenAdderThrows(t *testing.T) {
	testSource(t, `
const original = {};
const closeFailure = {};
let nextCount = 0;
let closeCount = 0;
const iterable = {};
iterable[Symbol.iterator] = function() {
  return {
    next: function() {
      nextCount += 1;
      if (nextCount === 1) {
        return { done: false, value: "value" };
      }
      return { done: true };
    },
    return: function() {
      closeCount += 1;
      throw closeFailure;
    }
  };
};
Set.prototype.add = function() { throw original; };

let caught = null;
try {
  new Set(iterable);
} catch (error) {
  caught = error;
}
assert(caught === original);
assertEqual(nextCount, 1);
assertEqual(closeCount, 1);
`)
}

// TestObjectFromEntriesClosesIteratorWhenAdderThrows verifies that failures
// produced while coercing an entry key also close the shared iterator.
func TestObjectFromEntriesClosesIteratorWhenAdderThrows(t *testing.T) {
	testSource(t, `
const original = {};
const closeFailure = {};
const key = {
  toString: function() { throw original; }
};
let closeCount = 0;
const iterable = {};
let nextCount = 0;
iterable[Symbol.iterator] = function() {
  return {
    next: function() {
      nextCount += 1;
      if (nextCount === 1) {
        return { done: false, value: [key, 1] };
      }
      return { done: true };
    },
    return: function() {
      closeCount += 1;
      throw closeFailure;
    }
  };
};

let caught = null;
try {
  Object.fromEntries(iterable);
} catch (error) {
  caught = error;
}
assert(caught === original);
assertEqual(closeCount, 1);
assertEqual(nextCount, 1);
`)
}

// TestObjectFromEntriesDoesNotCloseForIteratorStepErrors verifies the boundary
// between iterator-owned failures and entry-consumption failures.
func TestObjectFromEntriesDoesNotCloseForIteratorStepErrors(t *testing.T) {
	testSource(t, `
const nextFailure = {};
let nextCloseCount = 0;
const nextIterable = {};
nextIterable[Symbol.iterator] = function() {
  return {
    next: function() { throw nextFailure; },
    return: function() {
      nextCloseCount += 1;
      return {};
    }
  };
};

let nextCaught = null;
try {
  Object.fromEntries(nextIterable);
} catch (error) {
  nextCaught = error;
}
assert(nextCaught === nextFailure);
assertEqual(nextCloseCount, 0);

const doneFailure = {};
let doneCloseCount = 0;
const doneIterable = {};
doneIterable[Symbol.iterator] = function() {
  const result = {};
  Object.defineProperty(result, "done", {
    get: function() { throw doneFailure; }
  });
  return {
    next: function() { return result; },
    return: function() {
      doneCloseCount += 1;
      return {};
    }
  };
};

let doneCaught = null;
try {
  Object.fromEntries(doneIterable);
} catch (error) {
  doneCaught = error;
}
assert(doneCaught === doneFailure);
assertEqual(doneCloseCount, 0);
`)
}

// TestObjectFromEntriesPropagatesIteratorAcquisitionError verifies that an
// abrupt Symbol.iterator getter is returned before an iterator record exists.
func TestObjectFromEntriesPropagatesIteratorAcquisitionError(t *testing.T) {
	testSource(t, `
const getterFailure = {};
const getterIterable = {};
Object.defineProperty(getterIterable, Symbol.iterator, {
  get: function() { throw getterFailure; }
});

let getterCaught = null;
try {
  Object.fromEntries(getterIterable);
} catch (error) {
  getterCaught = error;
}
assert(getterCaught === getterFailure);

const callFailure = {};
const callIterable = {};
callIterable[Symbol.iterator] = function() { throw callFailure; };
let callCaught = null;
try {
  Object.fromEntries(callIterable);
} catch (error) {
  callCaught = error;
}
assert(callCaught === callFailure);

const nextGetterFailure = {};
const nextGetterIterable = {};
nextGetterIterable[Symbol.iterator] = function() {
  const iterator = {};
  Object.defineProperty(iterator, "next", {
    get: function() { throw nextGetterFailure; }
  });
  return iterator;
};
let nextGetterCaught = null;
try {
  Object.fromEntries(nextGetterIterable);
} catch (error) {
  nextGetterCaught = error;
}
assert(nextGetterCaught === nextGetterFailure);
`)
}

// TestCollectionConstructorsPropagateAdderLookupErrors verifies that Map and
// Set resolve their adders before acquiring or advancing the iterable.
func TestCollectionConstructorsPropagateAdderLookupErrors(t *testing.T) {
	testSource(t, `
let iteratorLookupCount = 0;
const iterable = {};
Object.defineProperty(iterable, Symbol.iterator, {
  get: function() {
    iteratorLookupCount += 1;
    return function() { return { next: function() { return { done: true }; } }; };
  }
});

const mapFailure = {};
Object.defineProperty(Map.prototype, "set", {
  get: function() { throw mapFailure; }
});
let mapCaught = null;
try {
  new Map(iterable);
} catch (error) {
  mapCaught = error;
}
assert(mapCaught === mapFailure);
assertEqual(iteratorLookupCount, 0);

const setFailure = {};
Object.defineProperty(Set.prototype, "add", {
  get: function() { throw setFailure; }
});
let setCaught = null;
try {
  new Set(iterable);
} catch (error) {
  setCaught = error;
}
assert(setCaught === setFailure);
assertEqual(iteratorLookupCount, 0);
`)
}

// TestIteratorClosePreservesOriginalErrorThroughProxyLookup verifies that a
// Proxy handler getter cannot escape the close boundary and replace the throw.
func TestIteratorClosePreservesOriginalErrorThroughProxyLookup(t *testing.T) {
	testSource(t, `
const original = {};
const closeFailure = {};
const entry = {};
Object.defineProperty(entry, "0", {
  get: function() { throw original; }
});

const targetIterator = {
  next: function() { return { done: false, value: entry }; },
  return: function() { return {}; }
};
let trapLookupCount = 0;
const handler = {};
Object.defineProperty(handler, "get", {
  get: function() {
    trapLookupCount += 1;
    if (trapLookupCount === 1) {
      return function(target, key) { return target[key]; };
    }
    throw closeFailure;
  }
});
const iterator = new Proxy(targetIterator, handler);
const iterable = {};
iterable[Symbol.iterator] = function() { return iterator; };

let caught = null;
try {
  Object.fromEntries(iterable);
} catch (error) {
  caught = error;
}
assert(caught === original);
assertEqual(trapLookupCount, 2);
`)
}

// TestArrayElisionPropagatesIteratorStepError protects destructuring callers
// that consume the now completion-aware IteratorStep result directly.
func TestArrayElisionPropagatesIteratorStepError(t *testing.T) {
	testSource(t, `
const original = {};
const result = {};
Object.defineProperty(result, "done", {
  get: function() { throw original; }
});
const iterable = {};
iterable[Symbol.iterator] = function() {
  return { next: function() { return result; } };
};

let caught = null;
try {
  [,] = iterable;
} catch (error) {
  caught = error;
}
assert(caught === original);
`)
}
