package coldmoon

type StringObject struct {
	*Object
	Data string
}

func NewStringObject(agent *Agent, s string, prototype ObjectType) *StringObject {
	stringObject := &StringObject{
		Object: NewObject(agent, prototype),
		Data:   s,
	}

	length := uint64(len(s))
	stringObject.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(float64(length)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return stringObject
}

// 10.4.3.4
var StringCreate = NewStringObject

// MARK: - StringConstructor

func NewStringConstructor(realm *Realm) ObjectType {
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		value := argumentsList[0]
		var s string
		if len(argumentsList) > 0 {
			symbolValue, isSymbol := value.(*SymbolValue)
			if newTarget == nil && isSymbol {
				return NewStringValue(symbolValue.SymbolDescriptiveString())
			}
			s = value.String()
		}

		if newTarget == nil {
			return NewStringValue(s)
		}
		return NewValueFromObject(StringCreate(
			realm.Agent, s, GetPrototypeFromConstructor(newTarget, "%String.prototype%")),
		)
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 1, "String", builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.StringPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(realm.Intrinsics.StringPrototype, "constructor", NewValueFromObject(object))

	return object
}

func NewStringPrototype(realm *Realm) *StringObject {
	stringPrototype := &StringObject{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype),
		Data:   "",
	}

	var toString BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		s := thisStringValue(realm.Agent, thisArgument)
		return NewStringValue(s)
	}
	DefineBuiltinFunction(stringPrototype, "toString", toString, 0, realm)

	var valueOf BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return NewStringValue(thisStringValue(realm.Agent, thisArgument))
	}
	DefineBuiltinFunction(stringPrototype, "valueOf", valueOf, 0, realm)

	return stringPrototype
}

func thisStringValue(agent *Agent, v Value) string {
	switch v := v.(type) {
	case *StringValue:
		return v.Data
	case *ObjectValue:
		s, ok := v.Object.(*StringObject)
		if ok {
			return s.Data
		}

	}
	panic("TypeError")
}
