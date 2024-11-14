package coldmoon

type setThrowType int

const (
	setThrowTypeThrow setThrowType = iota
	setThrowTypeIgnore
)

type ObjectType interface {
	IsExtensible() bool
	LengthOfArrayLike() uint64
	PropertyStorage() *PropertyStorage
	DefinePropertyOrThrow(key PropertyKey, desc *PropertyDescriptor) bool
	CreateDataPropertyOrThrow(key PropertyKey, value Value) bool
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
type SetFn = func(o ObjectType, p PropertyKey, v Value, receiver Value) bool
type GetOwnPropertyFn = func(o ObjectType, p PropertyKey) *PropertyDescriptor
type InternalMethods struct {
	GetPrototypeOf    func(o ObjectType) ObjectType
	SetPrototypeOf    func(o ObjectType, v ObjectType) bool
	IsExtensible      func(o ObjectType) bool
	PreventExtensions func(o ObjectType) bool
	GetOwnProperty    GetOwnPropertyFn
	DefineOwnProperty func(o ObjectType, p PropertyKey, desc *PropertyDescriptor) bool
	HasProperty       func(o ObjectType, p PropertyKey) bool
	Get               func(o ObjectType, p PropertyKey, receiver Value) Value
	Set               SetFn
	Delete            func(o ObjectType, p PropertyKey) bool
	OwnPropertyKeys   func(o ObjectType) []PropertyKey
	Call              func(o ObjectType, this Value, arguments []Value) Value
	Construct         func(o ObjectType, arguments []Value, newTarget ObjectType) ObjectType
}

func NewInternalMethods() InternalMethods {
	return InternalMethods{
		GetPrototypeOf:    InternalGetPrototypeOf,
		SetPrototypeOf:    InternalSetPrototypeOf,
		IsExtensible:      InternalIsExtensible,
		PreventExtensions: InternalPreventExtensions,
		GetOwnProperty:    InternalGetOwnProperty,
		DefineOwnProperty: InternalDefineOwnProperty,
		HasProperty:       InternalHasProperty,
		Get:               InternalGet,
		Set:               InternalSet,
		Delete:            InternalDelete,
		OwnPropertyKeys:   InternalOwnPropertyKeys,
		Call:              nil,
		Construct:         nil,
	}
}
