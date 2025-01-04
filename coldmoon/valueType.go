package coldmoon

type Value interface {
	ToBoolean() bool
	ToNumber(agent *Agent) *NumberValue

	Call(this Value, argumentsList ArgumentsList) Value
	CallNoArgs(this Value) Value
	// --- internal methods ---

	String() string
	// 7.1.18
	ToObject(agent *Agent) ObjectType
	ToCompletion() CompletionValue
	ToPropertyDescriptor() *PropertyDescriptor
	Hash() string
}
