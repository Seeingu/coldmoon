package coldmoon

type ArgumentsObject struct {
	*Object
	ParameterMap ObjectType
}

// 10.4.4.1
func argumentsGetOwnProperty(object ObjectType, key PropertyKey) (co Completion[*PropertyDescriptor]) {
	desc := OrdinaryGetOwnProperty(object, key)
	if desc == nil {
		return
	}
	_map := object.(*ArgumentsObject).ParameterMap
	isMapped := ObjectHasOwnProperty(_map, key)
	if isMapped {
		desc.Value = _map.Get(key)
	}
	co.value = desc
	return
}

// 10.4.4.2
func argumentsDefineOwnProperty(object ObjectType, key PropertyKey, desc *PropertyDescriptor) (co Completion[bool]) {
	_map := object.(*ArgumentsObject).ParameterMap
	isMapped := ObjectHasOwnProperty(_map, key)
	newArgDesc := desc
	if isMapped && desc.IsDataDescriptor() {
		if desc.Value == nil && !desc.Writable {
			newArgDesc.Value = _map.Get(key)
		}
	}
	allowed := OrdinaryDefineOwnProperty(object, key, newArgDesc)
	if !allowed {
		return
	}
	if isMapped {
		if desc.IsAccessorDescriptor() {
			_map.InternalMethods().Delete(_map, key)
		} else {
			if desc.Value != nil {
				_map.Set(key, desc.Value, setThrowTypeIgnore)
			}
			if !desc.Writable {
				_map.InternalMethods().Delete(_map, key)
			}
		}
	}
	co.value = true
	return
}

// 10.4.4.3
func argumentsGet(object ObjectType, key PropertyKey, receiver Value) (co CompletionValue) {
	_map := object.(*ArgumentsObject).ParameterMap
	isMapped := ObjectHasOwnProperty(_map, key)
	if !isMapped {
		return OrdinaryGet(object, key, receiver)
	} else {
		value := _map.Get(key)
		co.value = value
		return
	}
}

// 10.4.4.4
func argumentsSet(object ObjectType, key PropertyKey, value Value, receiver Value) (co Completion[bool]) {
	if SameValue(object.ToValue(), receiver) {
		_map := object.(*ArgumentsObject).ParameterMap
		isMapped := ObjectHasOwnProperty(_map, key)
		if isMapped {
			setResult := _map.Set(key, value, setThrowTypeIgnore)
			if setResult.IsAbrupt() {
				return CompletionFrom(co, setResult)
			}
		}
	}
	return OrdinarySet(object, key, value, receiver)
}

// 10.4.4.5
func argumentsDelete(object ObjectType, key PropertyKey) bool {
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

	length := JSInt(len(argumentsList))

	obj := &ArgumentsObject{
		Object: NewObject(agent, realm.Intrinsics.ObjectPrototype, "Arguments"),
	}
	obj.ref = obj

	obj.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(length.ToNumber()),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})

	for index, value := range argumentsList {
		pk := NewIntegerIndexPropertyKey(JSInt(index))
		obj.CreateDataPropertyOrThrow(pk, value)
	}

	obj.DefinePropertyOrThrow(NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]), &PropertyDescriptor{
		Value:        (realm.Intrinsics.ArrayPrototypeValues).ToValue(),
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
	length := JSInt(len(argumentsList))
	obj := &ArgumentsObject{
		Object: NewObject(agent, realm.Intrinsics.ObjectPrototype, "MappedArguments"),
	}
	obj.ref = obj
	internalMethods := obj.InternalMethods()
	internalMethods.GetOwnProperty = func(o ObjectType, p PropertyKey) *PropertyDescriptor {
		pp := argumentsGetOwnProperty(obj, p)
		return pp.Data()
	}
	internalMethods.DefineOwnProperty = argumentsDefineOwnProperty
	internalMethods.Get = func(o ObjectType, p PropertyKey, receiver Value) CompletionValue {
		completionValue := argumentsGet(obj, p, receiver)
		return completionValue
	}
	internalMethods.Set = argumentsSet
	internalMethods.Delete = argumentsDelete

	_map := OrdinaryObjectCreate(agent, nil, nil)
	obj.ParameterMap = _map
	parameterNames := formals.BoundNames()
	numberOfParameters := JSInt(len(parameterNames))
	for i := JSInt(0); i < length; i++ {
		value := argumentsList[i]
		obj.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(i), value)
	}

	obj.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(length.ToNumber()),
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
		index--
	}
	obj.DefinePropertyOrThrow(NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]), &PropertyDescriptor{
		Value:        (realm.Intrinsics.ArrayPrototypeValues).ToValue(),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
	obj.DefinePropertyOrThrow(NewStringPropertyKey("callee"), &PropertyDescriptor{
		Value:        (function).ToValue(),
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
	var getterClosure BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		function := agent.ActiveFunctionObject()
		_captures := function.(*BuiltinFunction).AdditionalFieldsV2.(*ArgGetterSetterCaptures)
		bindingValue := _captures.Env.GetBindingValue(agent, _captures.Name, false)
		return bindingValue
	}
	getter := CreateBuiltinFunction(agent, getterClosure, 1, CMString(""), builtinFunctionArgs{
		additionalFieldsV2: captures,
	})
	return getter
}

// 10.4.4.7.2
func MakeArgSetter(agent *Agent, name string, env EnvironmentRecord) ObjectType {
	captures := &ArgGetterSetterCaptures{
		Name: name,
		Env:  env,
	}
	var setterClosure BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		function := agent.ActiveFunctionObject()
		_captures := function.(*BuiltinFunction).AdditionalFieldsV2.(*ArgGetterSetterCaptures)
		_captures.Env.SetMutableBinding(_captures.Name, arguments[0], false)
		return UndefinedValue.ToCompletion()
	}

	setter := CreateBuiltinFunction(agent, setterClosure, 1, CMString(""), builtinFunctionArgs{
		additionalFieldsV2: captures,
	})
	return setter
}
