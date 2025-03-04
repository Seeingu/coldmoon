package coldmoon

type ObjectType interface {
	IsExtensible() bool
	LengthOfArrayLike() Completion[JSInt]
	PropertyStorage() *PropertyStorage
	DefinePropertyOrThrow(key PropertyKey, desc *PropertyDescriptor) bool
	DefineField(field *ClassFieldDefinition)
	SetPrototype(p ObjectType)
	// 14.7.5.9
	EnumerateObjectProperties() ObjectType
	// TODO: assert constructor is a function
	InitializeInstanceElements(constructor ObjectType)
	CreateDataProperty(key PropertyKey, value Value) bool
	DeletePropertyOrThrow(key PropertyKey) bool
	CreateDataPropertyOrThrow(key PropertyKey, value Value) bool
	EnumerableOwnProperties(kind objectOwnPropertiesKind) []Value
	OrdinaryToPrimitive(hint PreferredType) (co CompletionValue)
	CopyDataProperties(source Value, excludedItems []PropertyKey)
	FindViaPredicate(
		len JSInt,
		direction direction,
		predicate Value,
		thisArg Value,
	) FoundResult
	Get(key PropertyKey) Value
	Set(key PropertyKey, value Value, throw setThrowType) CompletionValue
	Agent() *Agent
	HasProperty(key PropertyKey) bool
	GetFunctionRealm() *Realm
	Construct(
		argumentLists []Value,
		newTarget ObjectType,
	) Completion[ObjectType]
	InternalMethods() *InternalMethods
	Prototype() ObjectType
	Extensible() bool
	SpeciesConstructor(defaultConstructor ObjectType) Completion[ObjectType]
	// PrivateMethodOrAccessorAdd 7.3.28
	PrivateMethodOrAccessorAdd(method *PrivateElement)
	PrivateGet(privateName PrivateName) CompletionValue
	ToCompletion() Completion[ObjectType]
	ToValue() Value
	// --- internal methods ---

	HasSlot(name string) bool
	SetSlot(name string, value any)
	GetSlot(name string) (value any, ok bool)
	String() string
	Ref() ObjectType
	IsOrdinary() bool
	GetId() uint64
	Call(this Value, argumentsList ArgumentsList) CompletionValue
	// defineBuiltinProperty is an alias of DefinePropertyOrThrow
	// used to define internal property
	defineBuiltinProperty(name PropertyConvertable, desc *PropertyDescriptor)
	// defineToStringTag defines @@toStringTag
	defineToStringTag(name string)
	defineUnscopables(value Value)
	defineBuiltinFunction(
		realm *Realm,
		name PropertyConvertable,
		fn BehaviorFn,
		length JSInt,
	)
	defineBuiltinAccessor(
		realm *Realm, pname PropertyConvertable, params builtinAccessorParams,
	)

	defineBuiltinFunctionWithAttributes(
		realm *Realm,
		name PropertyConvertable,
		fn BehaviorFn,
		length JSInt,
		attr PropertyDescriptorAttributes,
	)
}

func ObjectIs[O ObjectType](o ObjectType) bool {
	_, ok := o.Ref().(O)
	return ok
}

func ObjectAs[O ObjectType](o ObjectType) O {
	return o.Ref().(O)
}
