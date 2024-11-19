package coldmoon

type ObjectType interface {
	IsExtensible() bool
	LengthOfArrayLike() uint64
	PropertyStorage() *PropertyStorage
	DefinePropertyOrThrow(key PropertyKey, desc *PropertyDescriptor) bool
	DeletePropertyOrThrow(key PropertyKey) bool
	CreateDataPropertyOrThrow(key PropertyKey, value Value) bool
	EnumerableOwnProperties(kind objectOwnPropertiesKind) []Value
	OrdinaryToPrimitive(hint PreferredType) Value
	ToObject() *Object
	Get(key PropertyKey) Value
	Set(key PropertyKey, value Value, throw setThrowType)
	Agent() *Agent
	HasProperty(key PropertyKey) bool
	GetFunctionRealm() *Realm
	Construct(
		argumentLists []Value,
		newTarget *Object,
	) ObjectType
	InternalMethods() *InternalMethods
	Prototype() ObjectType
	Extensible() bool
}

func ObjectIs[O ObjectType](o ObjectType) bool {
	_, ok := o.(O)
	return ok
}
