package coldmoon

// 13.1.3
type RuntimeSemanticsEvaluation interface {
	Evaluation(i *IR, b *BytecodeContext)
}

// 8.4.5
type RuntimeSemanticsNamedEvaluation interface {
	NamedEvaluation(name string, expr Expression) ObjectType
}
