package coldmoon

type ExpressionValue struct {
	Value
	Expression Expression
}

func NewExpressionValue(e Expression) *ExpressionValue {
	v := &ExpressionValue{
		Expression: e,
	}
	v.Value = NewBaseValue(v)
	return v
}

func (v *ExpressionValue) String() string {
	return v.Expression.String()
}

// IsAnonymousFunctionDefinition implements the matching static semantic used
// by named evaluation. Arrow functions are always anonymous; function, class,
// generator, and async forms are anonymous only when they omit a binding name.
func IsAnonymousFunctionDefinition(expr Expression) bool {
	switch e := expr.(type) {
	case *ArrowFunction, *AsyncArrowFunction:
		return true
	case *FunctionExpression:
		return !e.astHasIdentifier()
	case *GeneratorExpression:
		return !e.astHasIdentifierName()
	case *AsyncFunctionExpression:
		return !e.astHasBindingIdentifier()
	case *PrimaryExpressionAsyncGeneratorExpression:
		return e.IdentifierName == ""
	case *ClassExpression:
		return !e.astHasIdentifier()
	case *ParenthesizedExpression:
		return IsAnonymousFunctionDefinition(e.Expression)
	default:
		return false
	}
}

// EvaluateNamedExpression performs the runtime NamedEvaluation operation for
// anonymous function and class definitions. The supplied property key becomes
// the inferred function name used by assignments, bindings, and object fields.
func EvaluateNamedExpression(vm *VM, expr Expression, name PropertyKeyOrPrivateName) (co CompletionValue) {
	switch e := expr.(type) {
	case *FunctionExpression:
		return e.NamedEvaluation(vm, name).ToCompletion()
	case *GeneratorExpression:
		return e.InstantiateGeneratorFunctionExpression(vm, name).ToValue().ToCompletion()
	case *AsyncFunctionExpression:
		return e.InstantiateAsyncFunctionExpression(vm, name).ToValue().ToCompletion()
	case *PrimaryExpressionAsyncGeneratorExpression:
		return e.InstantiateAsyncGeneratorFunctionExpression(vm, name).ToValue().ToCompletion()
	case *ArrowFunction:
		return e.InstantiateArrowFunctionExpression(vm, name).ToCompletion()
	case *AsyncArrowFunction:
		return e.InstantiateAsyncArrowFunctionExpression(vm, name).ToValue().ToCompletion()
	case *ClassExpression:
		class, isAbrupt, rt := ReturnIfAbrupt(
			e.ClassTail.ClassDefinitionEvaluation(vm, "", name),
			co,
		)
		if isAbrupt {
			return rt
		}
		SetClassSourceText(class, e.SourceText)
		co.value = class.ToValue()
		return
	case *ParenthesizedExpression:
		return EvaluateNamedExpression(vm, e.Expression, name)
	default:
		panic("named evaluation requires an anonymous function or class definition")
	}
}

// ReferencedNameKey converts a Reference Record's name union to the key form
// accepted by SetFunctionName and the named-evaluation algorithms.
func ReferencedNameKey(name *ReferencedName) PropertyKeyOrPrivateName {
	switch {
	case name.PrivateName != nil:
		return PropertyKeyOrPrivateNameName{PrivateName: *name.PrivateName}
	case name.Symbol != nil:
		return NewSymbolPropertyKey(name.Symbol)
	default:
		return NewStringPropertyKey(name.String)
	}
}
