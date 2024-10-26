package coldmoon

type InternalMethods struct {
	GetPrototypeOf    func(o *Object) *Object
	SetPrototypeOf    func(o *Object, v *Object) bool
	IsExtensible      func(o *Object) bool
	PreventExtensions func(o *Object) bool
	GetOwnProperty    func(o *Object, p PropertyKey) *PropertyDescriptor
	DefineOwnProperty func(o *Object, p PropertyKey, desc *PropertyDescriptor) bool
	HasProperty       func(o *Object, p PropertyKey) bool
	Get               func(o *Object, p PropertyKey, receiver Value) Value
	Set               func(o *Object, p PropertyKey, v Value, receiver Value) bool
	Delete            func(o *Object, p PropertyKey) bool
	OwnPropertyKeys   func(o *Object) []PropertyKey
	Call              func(o *Object, this Value, arguments []Value) Value
	Construct         func(o *Object, arguments []Value) *Object
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
