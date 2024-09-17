package parser

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Interpreter struct {
	expressions []Expression
}

var environment *Environment

func NewInterpreter(expressions []Expression) *Interpreter {
	environment = NewEnvironment(nil)
	return &Interpreter{expressions: expressions}
}

func NewInterpreterWithSource(source string) *Interpreter {
	scanner := NewScanner(source)
	parser := NewParser(scanner)
	expressions, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
	return NewInterpreter(expressions)
}

// MARK: Global state

var (
	undefinedObject = UndefinedObject{}
	nullObject      = NullObject{}
	trueObject      = BooleanObject{value: true}
	falseObject     = BooleanObject{value: false}
)

// MARK: utils

func (i *Interpreter) error(m string) {
	log.Fatal(m)
}

// MARK: Eval

func (i *Interpreter) expression(expression Expression, env *Environment) Object {
	switch expr := expression.(type) {
	// TODO: Store type info
	case LetExpression:
		if env.Exists(expr.identifier.toString()) {
			i.error("variable redeclared")
		}
		environment.Set(expr.identifier.toString(), expr.value)
	case NumberExpression:
		return NumberObject{value: expr.value}
	case StringExpression:
		return StringObject{value: expr.value}
	case BooleanExpression:
		if expr.value {
			return trueObject
		} else {
			return falseObject
		}
	}

	return undefinedObject

}

func (i *Interpreter) block(block BlockExpression, env *Environment) {
	for _, expression := range block.expressions {
		i.expression(expression, env)
	}
}

func (i *Interpreter) function(expression FunctionExpression, args []Expression) {
	localEnv := NewEnvironment(environment)
	for index, arg := range expression.args {
		localEnv.Set(arg.toString(), args[index])
	}

	i.block(expression.body, localEnv)

}

func (i *Interpreter) nativeFunction(expression NativeFunctionExpression, args []Expression) {
	var objectArgs []Object

	for _, arg := range args {
		objectArgs = append(objectArgs, i.expression(arg, environment))
	}
	expression.fn(objectArgs...)
}

func (i *Interpreter) call(expression CallExpression) {
	switch callerExpr := expression.caller.(type) {
	case IdentifierExpression:
		id := callerExpr.toString()
		value, ok := environment.Get(id)
		if !ok {
			i.error("identifier of function caller is not defined")
		}
		switch valueExpr := value.(type) {
		case FunctionExpression:
			i.function(valueExpr, expression.args)
		case NativeFunctionExpression:
			i.nativeFunction(valueExpr, expression.args)
		default:
			i.error("function caller is not of type FunctionExpression")

		}
	case ChainExpression:
		i.error("chain call not implemented")

	}

}

// MARK: Native

func (i *Interpreter) registerNativeFunctions() {
	environment.Set("print", NativeFunctionExpression{
		fn: func(args ...Object) {
			var sb strings.Builder
			for _, arg := range args {
				switch arg := arg.(type) {
				case NumberObject:
					sb.WriteString(strconv.Itoa(arg.value))
				case BooleanObject:
					if arg.value {
						sb.WriteString("true")
					} else {
						sb.WriteString("false")
					}
				case StringObject:
					sb.WriteString(arg.value)
				}
			}
			fmt.Println(sb.String())
		},
	})
}

func (i *Interpreter) Eval() {
	i.registerNativeFunctions()
	for _, expression := range i.expressions {
		switch expr := expression.(type) {
		case CallExpression:
			i.call(expr)
		default:
			i.expression(expr, environment)
		}
	}
}
