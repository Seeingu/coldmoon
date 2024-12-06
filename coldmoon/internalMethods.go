package coldmoon

type setThrowType int

const (
	setThrowTypeThrow setThrowType = iota
	setThrowTypeIgnore
)

type (
	SetFn               = func(o ObjectType, p PropertyKey, v Value, receiver Value) bool
	GetOwnPropertyFn    = func(o ObjectType, p PropertyKey) *PropertyDescriptor
	DefineOwnPropertyFn = func(o ObjectType, p PropertyKey, desc *PropertyDescriptor) bool
	GetPrototypeOfFn    = func(o ObjectType) ObjectType
	SetPrototypeOfFn    = func(o ObjectType, v ObjectType) bool
)

// TODO: Update method signature
type InternalMethods struct {
	GetPrototypeOf    GetPrototypeOfFn
	SetPrototypeOf    SetPrototypeOfFn
	IsExtensible      func(o ObjectType) bool
	PreventExtensions func(o ObjectType) bool
	GetOwnProperty    GetOwnPropertyFn
	DefineOwnProperty DefineOwnPropertyFn
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
