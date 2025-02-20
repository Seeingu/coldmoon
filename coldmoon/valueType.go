package coldmoon

type Value interface {
	ToBoolean() bool
	ToNumber(agent *Agent) *NumberValue

	Call(this Value, argumentsList ArgumentsList) Value
	ToString() CMString
	// ToPropertyDescriptor
	// spec: 6.2.6.5
	ToPropertyDescriptor(agent *Agent) *PropertyDescriptor
	CallNoArgs(this Value) Value
	// 7.1.18
	ToObject(agent *Agent) ObjectType
	// ThisStringValue implemented in BaseValue
	// 22.1.3.35.1
	ThisStringValue() string
	AOGetValue
	// --- internal methods ---

	// TypeString is aligned with spec's Type()
	// also 13.5.3.1 typeof
	TypeString() string
	String() string
	ToCompletion() CompletionValue
	ToBuiltinPropertyDescriptor() *PropertyDescriptor
	ToPropertyKey() PropertyKey
	Hash() string

	NumberOrBigInt() (*NumberValue, *BigIntValue, bool)
	ReferenceRecord() (*ReferenceRecord, bool)
}
