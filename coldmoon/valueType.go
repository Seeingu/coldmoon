package coldmoon

type Value interface {
	ToBoolean() bool
	ToNumber(agent *Agent) *NumberValue

	Call(this Value, argumentsList ArgumentsList) Value
	ToString() CMString
	CallNoArgs(this Value) Value
	// 7.1.18
	ToObject(agent *Agent) ObjectType
	AOGetValue
	// --- internal methods ---

	TypeString() string
	String() string
	ToCompletion() CompletionValue
	ToPropertyDescriptor() *PropertyDescriptor
	Hash() string
}
