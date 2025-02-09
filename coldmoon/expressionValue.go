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

// TODO: implement this method
func IsAnonymousFunctionDefinition(expr Expression) bool {
	return false
}
