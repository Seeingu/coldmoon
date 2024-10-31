package coldmoon

import "fmt"

// 10.2.4
func AddRestrictedFunctionProperties(F ObjectType, realm *Realm) {
	// TODO: Assert
	thrower := realm.Intrinsics.ThrowTypeError
	F.DefinePropertyOrThrow(NewStringPropertyKey("caller"), &PropertyDescriptor{
		Get:          thrower,
		Set:          thrower,
		Enumerable:   false,
		Configurable: true,
	})

	F.DefinePropertyOrThrow(NewStringPropertyKey("arguments"), &PropertyDescriptor{
		Get:          thrower,
		Set:          thrower,
		Enumerable:   false,
		Configurable: true,
	})

}

// 10.2.9
func SetFunctionName(function ObjectType, key PropertyKey, prefix string) {
	Assert(function.IsExtensible())
	Assert(!function.PropertyStorage().Has(NewStringPropertyKey("name")))

	var name string
	switch k := key.(type) {
	case SymbolPropertyKey:
		description := k.Value.Description

		if description == "" {
			name = description
		} else {
			name = fmt.Sprintf("[%s]", description)
		}
	case StringPropertyKey:
		name = k.Value
	default:
		panic("unimplemented")
	}

	function.(*BuiltinFunction).InitialName = name
	if prefix != "" {
		name = prefix + " " + name
		function.(*BuiltinFunction).InitialName = name
	}

	function.DefinePropertyOrThrow(NewStringPropertyKey("name"), &PropertyDescriptor{
		Value:        NewStringValue(name),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

}

// 10.2.10
func SetFunctionLength(function *Object, length float64) {
	Assert(function.IsExtensible())
	Assert(!function.PropertyStorage().Has(NewStringPropertyKey("length")))

	function.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(length),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
}
