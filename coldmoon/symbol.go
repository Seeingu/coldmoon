package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

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
	symbolObject.ref = symbolObject

	return symbolObject
}

func NewSymbolConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		description := pkg.SliceSafeGet(argumentsList, 0)
		if newTarget != nil {
			return co.ThrowTypeError(agent, "Symbol is not a constructor")
		}

		var descriptionString string
		if description != nil {
			descriptionString = description.String()
		}

		return agent.CreateSymbol(descriptionString)
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 0, CMString("Symbol"), builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	// 20.4.2.1
	object.defineBuiltinProperty(CMString("asyncIterator"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsAsyncIterator],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("hasInstance"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsHasInstance],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("isConcatSpreadable"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsIsConcatSpreadable],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("iterator"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsIterator],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("match"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsMatch],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("matchAll"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsMatchAll],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	object.defineBuiltinProperty(CMString("replace"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsReplace],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("search"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsSearch],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("species"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsSpecies],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("split"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsSplit],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("toPrimitive"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsToPrimitive],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("toStringTag"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsToStringTag],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	object.defineBuiltinProperty(CMString("unscopables"), &PropertyDescriptor{
		Value:        WellKnownSymbols[WellKnownSymbolsUnscopables],
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	var symbolFor BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		key := arguments[0]
		stringKey := key.String()
		if agent.GlobalSymbolRegistry[stringKey] != nil {
			return agent.GlobalSymbolRegistry[stringKey]
		}

		newSymbol := agent.CreateSymbol(stringKey)
		agent.GlobalSymbolRegistry[stringKey] = newSymbol

		return newSymbol
	}
	var keyFor BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		symbol := arguments[0]
		s, ok := symbol.(*SymbolValue)
		if !ok {
			panic("TypeError")
		}
		return NewStringValue(KeyForSymbol(agent, s))
	}

	object.defineBuiltinFunction(realm, CMString("for"), symbolFor, 1)
	object.defineBuiltinFunction(realm, CMString("keyFor"), keyFor, 1)

	BindPrototypeAndConstructor(realm.Intrinsics.SymbolPrototype, object)
	return object
}

func NewSymbolPrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "SymbolPrototype")

	toString := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		symbol := ThisSymbolValue(this)

		return NewStringValue(symbol.SymbolDescriptiveString())
	}
	valueOf := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		symbol := ThisSymbolValue(this)
		return symbol
	}
	toPrimitive := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return this
	}

	object.defineBuiltinFunction(realm, CMString("toString"), toString, 0)
	object.defineBuiltinFunction(realm, CMString("valueOf"), valueOf, 0)
	object.defineBuiltinFunctionWithAttributes(
		realm,
		WellKnownSymbolsToPrimitive, toPrimitive, 0,
		PropertyDescriptorAttributes{
			Writable:     false,
			Enumerable:   false,
			Configurable: true,
		})
	object.defineToStringTag("Symbol")

	return object
}
