package coldmoon

type Value interface {
	ToBoolean() bool
	ToNumber(agent *Agent) *NumberValue

	Call(agent *Agent, this Value, argumentsList ArgumentsList) CompletionValue
	CallNoArgs(this Value) CompletionValue
	ToString() CMString
	// ToPropertyDescriptor
	// spec: 6.2.6.5
	ToPropertyDescriptor(agent *Agent) *PropertyDescriptor
	// ToPrimitive
	// spec: 7.1.1
	// TODO(BM): throw completion handling
	ToPrimitive(agent *Agent, hint PreferredType) Value
	// ToObject
	// spec: 7.1.18
	ToObject(agent *Agent) Completion[ObjectType]
	// ThisStringValue implemented in BaseValue
	// 22.1.3.35.1
	ThisStringValue() string
	AOGetValue
	// --- internal methods ---

	// TypeString is aligned with spec's Type()
	// also 13.5.3.1 typeof
	TypeString() string
	String() string
	CompletionConvertable[Value]
	ToBuiltinPropertyDescriptor() *PropertyDescriptor
	ToPropertyKey() PropertyKey
	Hash() string

	IsObject() bool
	IsPromise() bool
	GetObject() (object ObjectType, ok bool)
	NumberOrBigInt() (*NumberValue, *BigIntValue, bool)
	ReferenceRecord() (*ReferenceRecord, bool)
}
