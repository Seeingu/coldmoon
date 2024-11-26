package coldmoon

type ObjectType interface {
	IsExtensible() bool
	LengthOfArrayLike() uint64
	PropertyStorage() *PropertyStorage
	DefinePropertyOrThrow(key PropertyKey, desc *PropertyDescriptor) bool
	DefineField(field *ClassFieldDefinition)
	InitializeInstanceElements(constructor ObjectType)
	CreateDataProperty(key PropertyKey, value Value) bool
	DeletePropertyOrThrow(key PropertyKey) bool
	CreateDataPropertyOrThrow(key PropertyKey, value Value) bool
	EnumerableOwnProperties(kind objectOwnPropertiesKind) []Value
	OrdinaryToPrimitive(hint PreferredType) Value
	CopyDataProperties(source Value, excludedItems []PropertyKey)
	ToObject() *Object
	Get(key PropertyKey) Value
	Set(key PropertyKey, value Value, throw setThrowType)
	Agent() *Agent
	HasProperty(key PropertyKey) bool
	GetFunctionRealm() *Realm
	Construct(
		argumentLists []Value,
		newTarget ObjectType,
	) ObjectType
	InternalMethods() *InternalMethods
	Prototype() ObjectType
	Extensible() bool
	SpeciesConstructor(defaultConstructor ObjectType) *CompletionObject
	ToCompletion() *CompletionValue
	ToValue() Value
}

func ObjectIs[O ObjectType](o ObjectType) bool {
	_, ok := o.(O)
	return ok
}
