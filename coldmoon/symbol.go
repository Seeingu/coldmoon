package coldmoon

import "fmt"

type SymbolValue struct {
	Value
	Id          uint64
	Description string
	IsPrivate   bool
}

var _ Value = (*SymbolValue)(nil)

func (s *SymbolValue) String() string {
	panic("TypeError")
}

func (s *SymbolValue) ToBoolean() bool {
	return true
}

// 20.4.3.3.1
func (s *SymbolValue) SymbolDescriptiveString() string {
	return fmt.Sprintf("Symbol(%s)", s.Description)
}

// MARK: - SymbolObject

type SymbolObject struct {
	*Object
	Data *SymbolValue
}

func NewSymbolObject(agent *Agent, s *SymbolValue, prototype ObjectType) *SymbolObject {
	symbolObject := &SymbolObject{
		Object: NewObject(agent, prototype, "Symbol"),
		Data:   s,
	}

	return symbolObject
}

func NewSymbolConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		description := argumentsList[0]
		if newTarget != nil {
			panic("TypeError")
		}

		var descriptionString string
		if description != nil {
			descriptionString = description.String()
		}

		return agent.CreateSymbol(descriptionString)
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 0, "Symbol", builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	// 20.4.2.1
	DefineBuiltinPropertyP(object, "asyncIterator", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsAsyncIterator],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "hasInstance", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsHasInstance],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "isConcatSpreadable", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsIsConcatSpreadable],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "iterator", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsIterator],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "match", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsMatch],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "matchAll", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsMatchAll],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.SymbolPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "replace", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsReplace],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "search", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsSearch],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "species", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsSpecies],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "split", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsSplit],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "toPrimitive", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsToPrimitive],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "toStringTag", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsToStringTag],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyP(object, "unscopables", &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsUnscopables],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinPropertyV(realm.Intrinsics.SymbolPrototype, "constructor",
		NewValueFromObject(object),
	)

	var symbolFor BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		key := arguments[0]
		stringKey := key.String()
		if agent.GlobalSymbolRegistry[stringKey] != nil {
			return agent.GlobalSymbolRegistry[stringKey]
		}

		newSymbol := agent.CreateSymbol(stringKey)
		agent.GlobalSymbolRegistry[stringKey] = newSymbol

		return newSymbol
	}
	var keyFor BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		symbol := arguments[0]
		s, ok := symbol.(*SymbolValue)
		if !ok {
			panic("TypeError")
		}
		return NewStringValue(keyForSymbol(agent, s))
	}

	DefineBuiltinFunction(object, "for", symbolFor, 1, realm)
	DefineBuiltinFunction(object, "keyFor", keyFor, 1, realm)

	return object
}

func NewSymbolPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "SymbolPrototype")

	toString := func(this Value, arguments []Value, newTarget ObjectType) Value {
		symbol := thisSymbolValue(this)

		return NewStringValue(symbol.SymbolDescriptiveString())
	}
	valueOf := func(this Value, arguments []Value, newTarget ObjectType) Value {
		symbol := thisSymbolValue(this)
		return symbol
	}
	toPrimitive := func(this Value, arguments []Value, newTarget ObjectType) Value {
		return this
	}

	DefineBuiltinFunction(object, "toString", toString, 0, realm)
	DefineBuiltinFunction(object, "valueOf", valueOf, 0, realm)
	DefineBuiltinFunctionWithAttributes(object, "@@toPrimitive", toPrimitive, 0, realm, PropertyDescriptorAttributes{
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("String"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}

func thisSymbolValue(v Value) *SymbolValue {
	if symbol, ok := v.(*SymbolValue); ok {
		return symbol
	}
	if object, ok := v.(*ObjectValue); ok {
		s, ok := object.Object.(*SymbolObject)
		if ok {
			return s.Data
		}
	}

	panic("TypeError")
}

func keyForSymbol(agent *Agent, symbol *SymbolValue) string {
	for _, value := range agent.GlobalSymbolRegistry {
		if value.Id == symbol.Id {
			return value.Description
		}
	}
	return ""
}
