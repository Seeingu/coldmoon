package coldmoon

// 13.1.3
type RuntimeSemanticsEvaluation interface {
	Evaluation(vm *VM2) Value
}

// 8.4.5
type RuntimeSemanticsNamedEvaluation interface {
	NamedEvaluation(name string, expr Expression) ObjectType
}

// 13.2.5.5
type RuntimeSemanticsPropertyDefinitionEvaluation interface {
	// PropertyDefinitionEvaluation returns UNUSED
	PropertyDefinitionEvaluation(vm *VM2, obj ObjectType)
}

// 13.3.8.1
type RuntimeSemanticsArgumentListEvaluation interface {
	ArgumentListEvaluation(vm *VM2) []Value
}

// 15.3.4
type RuntimeSemanticsInstantiateArrowFunctionExpression interface {
	InstantiateArrowFunctionExpression(vm *VM2, name string) Value
}
