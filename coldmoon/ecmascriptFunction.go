package coldmoon

import "fmt"

// 10.2.9
func SetFunctionName(function ObjectType, key PropertyKey, prefix string) {
	Assert(function.IsExtensible())
	Assert(!function.PropertyStorage().Has(NewStringPropertyKey("name")))

	var name string
	symbol, isSymbol := key.(*SymbolPropertyKey)
	if isSymbol {
		description := symbol.Value.Description

		if description == "" {
			name = description
		} else {
			name = fmt.Sprintf("[%s]", description)
		}
	} else {
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
