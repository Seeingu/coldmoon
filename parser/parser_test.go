package parser

import (
	"github.com/Seeingu/coldmoon/ast"
	"github.com/Seeingu/coldmoon/lexer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFunctionCalls(t *testing.T) {
	input := "a(1, 2 + 3)"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(t, 1, len(program.Statements))

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)
	exp, ok := stmt.Expression.(*ast.CallExpression)
	assert.True(t, ok)

	testIdentifier(t, exp.FunctionName, "a")

	assert.Equal(t, 2, len(exp.Arguments))
	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], infixExpected{
		2, "+", 3,
	})
}

func TestFunction(t *testing.T) {
	input := "function a(b, c) { return d + e; }"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	assert.Equal(t, 1, len(program.Statements))

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	testLiteralExpression(t, function.Parameters[0], "b")
	testLiteralExpression(t, function.Parameters[1], "c")

	returnStmt, ok := function.Body.Statements[0].(*ast.ReturnStatement)
	assert.True(t, ok)
	testInfixExpression(t, returnStmt.ReturnValue, infixExpected{
		leftValue:  "d",
		operator:   "+",
		rightValue: "e",
	})

}

func TestIndex(t *testing.T) {
	input := "myArray[1 + 1]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)
	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	assert.True(t, ok)

	testIdentifier(t, indexExp.Left, "myArray")
	testInfixExpression(t, indexExp.Index, infixExpected{1, "+", 1})
}

func TestEmptyObjectLiteral(t *testing.T) {
	input := "{}"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	o, ok := stmt.Expression.(*ast.ObjectLiteralExpression)
	assert.True(t, ok)
	assert.Equal(t, 0, len(o.Pairs))
}

func TestParsingObjects(t *testing.T) {
	input := `{"one": 1, "two": 2, three: 3}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.ObjectLiteralExpression)
	assert.True(t, ok)

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	assert.Equal(t, len(expected), len(hash.Pairs))

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		assert.True(t, ok)

		expectedValue := expected[literal.Value]
		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestArrayLiteral(t *testing.T) {
	input := "[1 + 2, 3- 4]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)
	array, ok := stmt.Expression.(*ast.ArrayLiteralExpression)
	assert.True(t, ok)
	assert.Equal(t, 2, len(array.Elements))
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(t, 1, len(program.Statements))
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	assert.True(t, ok)

	assert.Equal(t, "hello world", literal.Value)
}

func TestLetComplicated(t *testing.T) {
	input := `
	let one = 1;
	let two = one;
	let a = function() { 1 };
	two;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(t, 4, len(program.Statements))
}

func TestLet(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.LetStatement)
		testLet(t, stmt, tt.expectedIdentifier)
		val := stmt.Value
		testLiteralExpression(t, val, tt.expectedValue)
	}

}

func testLet(t *testing.T, s *ast.LetStatement, name string) bool {
	assert.Equal(t, s.Name.Value, name)
	return true
}

func TestIf(t *testing.T) {
	input := `if (x < y) { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok, "statements should be ExpressionStatement")

	exp, ok := stmt.Expression.(*ast.IfExpression)
	assert.True(t, ok, "expression should be if")

	testInfixExpression(t, exp.Condition, infixExpected{
		leftValue:  "x",
		operator:   "<",
		rightValue: "y",
	})

	body, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	testIdentifier(t, body.Expression, "x")

	assert.Nil(t, exp.Alternative, "alternative should be nil")
}

func TestPrefix(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"-5", "-", 5},
		{"!true", "!", true},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "statements should be ExpressionStatement")

		testPrefixExpression(t, stmt.Expression, tt.operator, tt.value)
	}
}

func TestInfix(t *testing.T) {
	tests := []struct {
		input    string
		expected infixExpected
	}{
		{"1 < 2", infixExpected{1, "<", 2}},
		{"1 > 2", infixExpected{1, ">", 2}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok, "statements should be ExpressionStatement")

		testInfixExpression(t, stmt.Expression, tt.expected)
	}

}

// MARK: Helpers

type infixExpected struct {
	leftValue  interface{}
	operator   string
	rightValue interface{}
}

func testInfixExpression(t *testing.T, e ast.Expression, expected infixExpected) {
	infix, ok := e.(*ast.InfixExpression)
	assert.True(t, ok, "expression should be InfixExpression")

	testLiteralExpression(t, infix.Left, expected.leftValue)
	assert.Equal(t, expected.operator, infix.Operator, "operator")
	testLiteralExpression(t, infix.Right, expected.rightValue)
}

func testPrefixExpression(t *testing.T, exp ast.Expression, operator string, rightValue interface{}) {
	prefix, ok := exp.(*ast.PrefixExpression)
	assert.True(t, ok, "expression should be PrefixExpression")

	assert.Equal(t, operator, prefix.Operator, "operator")
	testLiteralExpression(t, prefix.Right, rightValue)
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.IdentifierExpression)
	assert.True(t, ok, "exp expression should be ast.IdentifierExpression")
	assert.Equal(t, value, ident.Value)

	return true
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	assert.True(t, ok, "expression should be IntegerLiteral")
	assert.Equal(t, value, integ.Value)
	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.BooleanExpression)
	assert.True(t, ok, "exp expression should be ast.BooleanExpression")
	assert.Equal(t, value, bo.Value)

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
