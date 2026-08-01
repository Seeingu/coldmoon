package tests

import "testing"

func TestArrayAndArgumentSpread(t *testing.T) {
	testSource(t, `
const array = [0, ...[1, 2], 3];
assertEqual(array.length, 4);
assertEqual(array[0], 0);
assertEqual(array[1], 1);
assertEqual(array[2], 2);
assertEqual(array[3], 3);

function digits(a, b, c, d) {
    return a * 1000 + b * 100 + c * 10 + d;
}
assertEqual(digits(1, ...[2, 3], 4), 1234);

const receiver = {
    base: 10,
    add: function(a, b) { return this.base + a + b; }
};
assertEqual(receiver.add(...[2, 3]), 15);

const constructed = new Array(...[2]);
assertEqual(constructed.length, 2);
`)
}
