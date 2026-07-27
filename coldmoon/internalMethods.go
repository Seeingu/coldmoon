package coldmoon

type setThrowType int

const (
	setThrowTypeThrow setThrowType = iota
	setThrowTypeIgnore
)

type (
	SetFn               = func(o ObjectType, p PropertyKey, v Value, receiver Value) Completion[bool]
	DeleteFn            = func(o ObjectType, p PropertyKey) Completion[bool]
	HasPropertyFn       = func(o ObjectType, p PropertyKey) Completion[bool]
	GetOwnPropertyFn    = func(o ObjectType, p PropertyKey) *PropertyDescriptor
	DefineOwnPropertyFn = func(o ObjectType, p PropertyKey, desc *PropertyDescriptor) Completion[bool]
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
	HasProperty       HasPropertyFn
	Get               func(o ObjectType, p PropertyKey, receiver Value) CompletionValue
	Set               SetFn
	Delete            DeleteFn
	OwnPropertyKeys   func(o ObjectType) []PropertyKey
	Call              func(o ObjectType, this Value, arguments []Value) CompletionValue
	Construct         func(o ObjectType, arguments []Value, newTarget ObjectType) Completion[ObjectType]
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
