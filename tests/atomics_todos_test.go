package tests

import "testing"

func TestAtomicsIntegerOperationsAndWaitResults(t *testing.T) {
	testSource(t, `
const shared = new SharedArrayBuffer(16);
const values = new Int32Array(shared);

assertEqual(Atomics.store(values, 0, -1), -1, "store result");
assertEqual(Atomics.load(values, 0), -1, "signed load");
assertEqual(Atomics.add(values, 0, 3), -1, "add returns previous");
assertEqual(Atomics.load(values, 0), 2, "add stores result");
assertEqual(Atomics.sub(values, 0, 5), 2, "sub returns previous");
assertEqual(Atomics.load(values, 0), -3, "sub stores result");
assertEqual(Atomics.and(values, 0, 6), -3, "and returns previous");
assertEqual(Atomics.or(values, 0, 8), 4, "or returns previous");
assertEqual(Atomics.xor(values, 0, 3), 12, "xor returns previous");
assertEqual(Atomics.exchange(values, 0, 10), 15, "exchange returns previous");
assertEqual(Atomics.compareExchange(values, 0, 10, 20), 10, "compareExchange match");
assertEqual(Atomics.compareExchange(values, 0, 10, 30), 20, "compareExchange mismatch");
assertEqual(Atomics.load(values, 0), 20, "compareExchange final value");

// The engine runs a single non-blockable agent, so Atomics.wait must throw a
// TypeError instead of blocking forever with nobody able to notify it.
let waitThrew = false;
try {
  Atomics.wait(values, 0, 19, 0);
} catch (error) {
  waitThrew = error instanceof TypeError;
}
assertEqual(waitThrew, true, "main-thread Atomics.wait throws TypeError");
assertEqual(Atomics.notify(values, 0), 0, "empty waiter queue");
assertEqual(Atomics.notify(new Int32Array(1), 0), 0, "non-shared notify");

let rejectedFloatArray = false;
try {
  Atomics.load(new Float32Array(1), 0);
} catch (error) {
  rejectedFloatArray = true;
}
assertEqual(rejectedFloatArray, true, "floating typed arrays rejected");
`)
}

func TestAtomicsBigIntOperations(t *testing.T) {
	testSource(t, `
const shared = new SharedArrayBuffer(16);
const values = new BigInt64Array(shared);

assertEqual(Atomics.store(values, 0, -1n), -1n, "BigInt store result");
assertEqual(Atomics.load(values, 0), -1n, "BigInt signed load");
assertEqual(Atomics.add(values, 0, 3n), -1n, "BigInt add previous");
assertEqual(Atomics.load(values, 0), 2n, "BigInt add result");
assertEqual(Atomics.compareExchange(values, 0, 2n, -4n), 2n, "BigInt compareExchange");
assertEqual(Atomics.load(values, 0), -4n, "BigInt compareExchange result");
let waitThrew = false;
try {
  Atomics.wait(values, 0, -3n, 0);
} catch (error) {
  waitThrew = error instanceof TypeError;
}
assertEqual(waitThrew, true, "main-thread BigInt Atomics.wait throws TypeError");
`)
}

// TestAtomicsBigIntReadModifyWritePreservesAll64Bits covers values that cannot
// be represented exactly as ECMAScript Numbers. The atomic implementation must
// keep the read-modify-write result as raw bits instead of routing it through a
// float64 conversion.
func TestAtomicsBigIntReadModifyWritePreservesAll64Bits(t *testing.T) {
	testSource(t, `
const signed = new BigInt64Array(new SharedArrayBuffer(8));
const aboveSafeInteger = 9007199254740993n;
Atomics.store(signed, 0, aboveSafeInteger);
assertEqual(Atomics.add(signed, 0, 2n), aboveSafeInteger, "signed RMW previous value");
assertEqual(Atomics.load(signed, 0), 9007199254740995n, "signed RMW result above 2^53");

Atomics.store(signed, 0, 9223372036854775807n);
assertEqual(Atomics.sub(signed, 0, 0n), 9223372036854775807n, "MaxInt64 remains exact");

const unsigned = new BigUint64Array(new SharedArrayBuffer(8));
Atomics.store(unsigned, 0, 18446744073709551614n);
assertEqual(Atomics.add(unsigned, 0, 1n), 18446744073709551614n, "unsigned RMW previous value");
assertEqual(Atomics.load(unsigned, 0), 18446744073709551615n, "unsigned RMW result");
`)
}
