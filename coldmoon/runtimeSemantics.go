package coldmoon

// 13.1.3
type RuntimeSemanticsEvaluation interface {
	Evaluation(i *IR, b *BytecodeContext)
}
