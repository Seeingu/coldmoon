package vm

import (
	"fmt"
	"github.com/Seeingu/coldmoon/ast"
	"github.com/Seeingu/coldmoon/compiler"
	"github.com/Seeingu/coldmoon/lexer"
	"github.com/Seeingu/coldmoon/object"
	"github.com/Seeingu/coldmoon/parser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIndexExpression(t *testing.T) {
	tests := []vmTest{
		{"[1, 2, 3][1]", 2},
		{"[][0]", JSUndefined},
		{`{a: 1}["a"]`, 1},
		{`{a: 1}["b"]`, JSUndefined},
	}
	runVMTests(t, tests)
}

func TestObjectLiterals(t *testing.T) {
	tests := []vmTest{
		{
			"{}", map[object.HashKey]int64{},
		},
		{
			"{a: 1, b: 2}",
			map[object.HashKey]int64{
				(&object.StringObject{Value: "a"}).HashKey(): 1,
				(&object.StringObject{Value: "b"}).HashKey(): 2,
			},
		},
	}
	runVMTests(t, tests)
}

func TestArray(t *testing.T) {
	tests := []vmTest{
		{"[]", []int{}},
		{"[1,2,3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4]", []int{3, 12}},
	}
	runVMTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTest{
		{`"a"`, "a"},
		{`"a" + "b"`, "ab"},
		{`"a" + "b" + "c"`, "abc"},
	}
	runVMTests(t, tests)
}

func TestGlobalLet(t *testing.T) {
	tests := []vmTest{
		{
			`
	let one = 1;
	let two = one;
	two;`, 1,
		},
	}

	runVMTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTest{
		{"if (true) { 10 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", JSNull},
	}
	runVMTests(t, tests)
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTest{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"3 * 2 - 1", 5},
		{"6 - 2 * 2", 2},
		{"6 - 2 / 2", 5},
		{"(6 - 2) / 2", 2},
	}
	runVMTests(t, tests)
}

func TestBoolean(t *testing.T) {
	tests := []vmTest{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 == 1", true},
		{"!true", false},
		{"!!true", true},
		{"!5", false},
	}
	runVMTests(t, tests)
}

// MARK: Helpers

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.StringObject)
	if !ok {
		return fmt.Errorf("expected *object.StringObject, got %T", actual)
	}
	if result.Value != expected {
		return fmt.Errorf("expected *object.StringObject, got %s", result.Value)
	}
	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not integer. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object expected %d, got %d", expected, result.Value)
	}

	return nil
}

type vmTest struct {
	input    string
	expected interface{}
}

func runVMTests(t *testing.T, tests []vmTest) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()
		err := comp.Compile(program)
		assert.NoError(t, err)

		vm := New(comp.Bytecode())
		err = vm.Run()
		assert.NoError(t, err)

		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.BooleanObject)
	if !ok {
		return fmt.Errorf("object is not boolean. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object expected %t, got %t", expected, result.Value)
	}
	return nil
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case map[object.HashKey]int64:
		o, ok := actual.(*object.ObjectObject)
		assert.True(t, ok)
		assert.Equal(t, len(expected), len(o.Pairs))
		for k, v := range expected {
			pair, ok := o.Pairs[k]
			assert.True(t, ok)
			err := testIntegerObject(v, pair.Value)
			assert.NoError(t, err)
		}
	case []int:
		array, ok := actual.(*object.ArrayObject)
		assert.True(t, ok)
		assert.Equal(t, len(expected), len(array.Elements))
		for i, elem := range expected {
			err := testIntegerObject(int64(elem), array.Elements[i])
			assert.NoError(t, err)
		}
	case int:
		err := testIntegerObject(int64(expected), actual)
		assert.NoError(t, err)
	case string:
		err := testStringObject(expected, actual)
		assert.NoError(t, err)
	case bool:
		err := testBooleanObject(expected, actual)
		assert.NoError(t, err)
	case *object.NullObject:
		assert.Equal(t, JSNull, actual)
	}
}
