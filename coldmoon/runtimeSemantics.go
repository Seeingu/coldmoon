package coldmoon

// 13.1.3
type RuntimeSemanticsEvaluation interface {
	Evaluation(vm *VM2) Value
}

// 8.4.5
type RuntimeSemanticsNamedEvaluation interface {
	NamedEvaluation(name string, expr Expression) ObjectType
}
