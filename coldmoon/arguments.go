package coldmoon

type ArgumentsObject struct {
	*Object
	ParameterMap ObjectType
}

// 10.4.4.1
func GetOwnProperty(object ObjectType, key PropertyKey) *CompletionPropertyDescriptor {
	desc := OrdinaryGetOwnProperty(object, key)
	if desc == nil {
		return NewCompletionPropertyDescriptorUndefined()
	}
	_map := object.(*ArgumentsObject).ParameterMap
	isMapped := ObjectHasOwnProperty(_map, key)
	if isMapped {
		desc.Value = _map.Get(key)
	}
	return NewCompletionPropertyDescriptor(desc)
}

// 10.4.4.2
func DefineOwnProperty(object ObjectType, key PropertyKey, desc *PropertyDescriptor) bool {
	_map := object.(*ArgumentsObject).ParameterMap
	isMapped := ObjectHasOwnProperty(_map, key)
	newArgDesc := desc
	if isMapped && desc.IsDataDescriptor() {
		if desc.Value == nil && desc.Writable == false {
			newArgDesc.Value = _map.Get(key)
		}
	}
	allowed := OrdinaryDefineOwnProperty(object, key, newArgDesc)
	if !allowed {
		return false
	}
	if isMapped {
		if desc.IsAccessorDescriptor() {
			_map.InternalMethods().Delete(_map, key)
		} else {
			if desc.Value != nil {
				_map.Set(key, desc.Value, setThrowTypeIgnore)
			}
			if desc.Writable == false {
				_map.InternalMethods().Delete(_map, key)
			}
		}
	}
	return true
}

// 10.4.4.3
func Get(object ObjectType, key PropertyKey, receiver Value) *CompletionValue {
	_map := object.(*ArgumentsObject).ParameterMap
	isMapped := ObjectHasOwnProperty(_map, key)
	if !isMapped {
		return OrdinaryGet(object, key, receiver).ToCompletion()
	} else {
		value := _map.Get(key)
		return NewCompletionValue(value)
	}
}

// 10.4.4.4
func Set(object ObjectType, key PropertyKey, value Value, receiver Value) bool {
	if SameValue(object.ToValue(), receiver) {
		_map := object.(*ArgumentsObject).ParameterMap
		isMapped := ObjectHasOwnProperty(_map, key)
		if isMapped {
			_map.Set(key, value, setThrowTypeIgnore)
		}
	}
	return OrdinarySet(object, key, value, receiver)
}

// 10.4.4.5
func Delete(object ObjectType, key PropertyKey) bool {
	_map := object.(*ArgumentsObject).ParameterMap
	isMapped := ObjectHasOwnProperty(_map, key)
	result := OrdinaryDelete(object, key)
	if result && isMapped {
		_map.InternalMethods().Delete(_map, key)
	}
	return result
}

// 10.4.4.6
func CreateUnmappedArgumentsObject(agent *Agent, argumentsList []Value) ObjectType {
	realm := agent.CurrentRealm()

	length := len(argumentsList)

	obj := &ArgumentsObject{
		Object: NewObject(agent, realm.Intrinsics.ObjectPrototype),
	}

	obj.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(float64(length)),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})

	for index, value := range argumentsList {
		pk := NewIntegerIndexPropertyKey(index)
		obj.CreateDataPropertyOrThrow(pk, value)
	}

	obj.DefinePropertyOrThrow(NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]), &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.ArrayPrototypeValues),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
	obj.DefinePropertyOrThrow(NewStringPropertyKey("callee"), &PropertyDescriptor{
		Get:          realm.Intrinsics.ThrowTypeError,
		Set:          realm.Intrinsics.ThrowTypeError,
		Enumerable:   false,
		Configurable: false,
	})

	return obj
}

// 10.4.4.7
func CreateMappedArgumentsObject(agent *Agent, function ObjectType, formals *FormalParameters, argumentsList []Value, env EnvironmentRecord) ObjectType {
	realm := agent.CurrentRealm()
	length := len(argumentsList)
	obj := &ArgumentsObject{
		Object: NewObject(agent, realm.Intrinsics.ObjectPrototype),
	}
	internalMethods := obj.InternalMethods()
	internalMethods.GetOwnProperty = func(o ObjectType, p PropertyKey) *PropertyDescriptor {
		return GetOwnProperty(obj, p).Data()
	}
	internalMethods.DefineOwnProperty = DefineOwnProperty
	internalMethods.Get = func(o ObjectType, p PropertyKey, receiver Value) Value {
		return Get(obj, p, receiver).Value
	}
	internalMethods.Set = Set
	internalMethods.Delete = Delete

	_map := OrdinaryObjectCreate(agent, nil, nil)
	obj.ParameterMap = _map
	parameterNames := formals.BoundNames()
	numberOfParameters := len(parameterNames)
	for i := 0; i < numberOfParameters; i++ {
		value := argumentsList[i]
		obj.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(i), value)
	}

	obj.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(float64(length)),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})

	mappedNames := make(map[IdentifierName]bool)
	index := numberOfParameters - 1
	for index >= 0 {
		name := parameterNames[index]
		if !mappedNames[name] {
			mappedNames[name] = true
			if index < length {
				g := MakeArgGetter(agent, string(name), env)
				p := MakeArgSetter(agent, string(name), env)
				_map.InternalMethods().DefineOwnProperty(
					_map,
					NewIntegerIndexPropertyKey(index),
					&PropertyDescriptor{
						Get:          g,
						Set:          p,
						Enumerable:   false,
						Configurable: true,
					},
				)
			}
		}
	}
	obj.DefinePropertyOrThrow(NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]), &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.ArrayPrototypeValues),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
	obj.DefinePropertyOrThrow(NewStringPropertyKey("callee"), &PropertyDescriptor{
		Value:        NewValueFromObject(function),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
	return obj
}

// 10.4.4.7.1
func MakeArgGetter(agent *Agent, name string, env EnvironmentRecord) ObjectType {
	captures := &ArgGetterSetterCaptures{
		Name: name,
		Env:  env,
	}
	var getterClosure BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		function := agent.ActiveFunctionObject()
		_captures := function.(*BuiltinFunction).AdditionalFields.ArgGetterSetterCaptures
		return _captures.Env.GetBindingValue(agent, _captures.Name, false).Value
	}
	getter := CreateBuiltinFunction(agent, getterClosure, 1, "", builtinFunctionArgs{
		additionalFields: &AdditionalFields{
			ArgGetterSetterCaptures: captures,
		},
	})
	return getter
}

// 10.4.4.7.2
func MakeArgSetter(agent *Agent, name string, env EnvironmentRecord) ObjectType {
	captures := &ArgGetterSetterCaptures{
		Name: name,
		Env:  env,
	}
	var setterClosure BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		function := agent.ActiveFunctionObject()
		_captures := function.(*BuiltinFunction).AdditionalFields.ArgGetterSetterCaptures
		_captures.Env.SetMutableBinding(_captures.Name, arguments[0], false)
		return UndefinedValue
	}

	setter := CreateBuiltinFunction(agent, setterClosure, 1, "", builtinFunctionArgs{
		additionalFields: &AdditionalFields{
			ArgGetterSetterCaptures: captures,
		},
	})
	return setter
}
