package coldmoon

type ErrorObject struct {
	*Object
	Name    string
	Message string
}

var errorInternalSet SetFn = func(o ObjectType, p PropertyKey, v Value, receiver Value) bool {
	switch pp := p.(type) {
	case StringPropertyKey:
		if pp.Value == "name" {
			o.(*ErrorObject).Name = v.String()
		}
		if pp.Value == "message" {
			o.(*ErrorObject).Message = v.String()
		}
	}
	return OrdinarySet(o.ToObject(), p, v, receiver)
}

func NewErrorConstructor(realm *Realm) ObjectType {
	agent := realm.Agent

	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, _newTarget ObjectType) Value {
		message := argumentsList[0]
		options := argumentsList[1]

		newTarget := _newTarget
		if newTarget == nil {
			newTarget = agent.ActiveFunctionObject()
		}

		object := OrdinaryCreateFromConstructor(agent, newTarget, "%Error.prototype%", []string{})
		errorObject := &ErrorObject{
			Object: object,
		}
		errorObject.InternalMethods().Set = errorInternalSet

		if message != UndefinedValue {
			msg := message.String()

			errorObject.CreateNonEnumerableDataProperty(NewStringPropertyKey("message"), NewStringValue(msg))

			errorObject.Message = msg
		}

		InstallErrorCause(agent, errorObject, options)

		return NewValueFromObject(errorObject)
	}

	object := CreateBuiltinFunction(
		realm.Agent,
		behavior,
		1,
		"Error",
		builtinFunctionArgs{
			realm:         realm,
			prototype:     realm.Intrinsics.FunctionPrototype,
			isConstructor: true,
		},
	)

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.ErrorPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinPropertyV(realm.Intrinsics.ErrorPrototype, "constructor", NewValueFromObject(object))

	return object
}

// 20.5.8.1
func InstallErrorCause(agent *Agent, errorObject *ErrorObject, options Value) {
	if optionsValue, ok := options.(*ObjectValue); ok {
		causeKey := NewStringPropertyKey("cause")
		if optionsValue.Object.HasProperty(causeKey) {
			cause := optionsValue.Object.Get(causeKey)

			errorObject.CreateNonEnumerableDataProperty(causeKey, cause)
		}
	}
}

func NewErrorPrototype(realm *Realm) ObjectType {
	agent := realm.Agent

	object := NewObject(agent, realm.Intrinsics.ObjectPrototype)

	DefineBuiltinPropertyV(object, "name", NewStringValue("Error"))

	DefineBuiltinPropertyV(object, "message", NewStringValue(""))

	var toString BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		O, ok := thisValue.(*ObjectValue)
		if !ok {
			panic("TypeError")
		}
		object := O.Object
		name := object.Get(NewStringPropertyKey("name"))
		var nameString string
		if name == nil {
			nameString = "Error"
		} else {
			nameString = name.String()
		}

		msg := object.Get(NewStringPropertyKey("message"))
		var msgString string
		if msg == nil {
			msgString = ""
		} else {
			msgString = msg.String()
		}

		if nameString == "" {
			return NewStringValue(msgString)
		}

		if msgString == "" {
			return NewStringValue(nameString)
		}

		return NewStringValue(nameString + ": " + msgString)
	}

	DefineBuiltinFunction(object, "toString", toString, 0, realm)

	return object
}

// 20.5.6.4
func NativeError() ObjectType {
	errorObject := &ErrorObject{
		Object: NewObject(nil, nil),
	}
	return errorObject
}

func NewNativeErrorConstructor(realm *Realm, name string) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, _newTarget ObjectType) Value {
		message := argumentsList[0]
		options := argumentsList[1]

		newTarget := _newTarget
		if newTarget == nil {
			newTarget = agent.ActiveFunctionObject()
		}

		protoName := "%" + name + ".prototype%"
		object := OrdinaryCreateFromConstructor(agent, newTarget, protoName, []string{})
		errorObject := &ErrorObject{
			Object: object,
			Name:   name,
		}
		errorObject.InternalMethods().Set = errorInternalSet

		if message != UndefinedValue {
			msg := message.String()

			errorObject.CreateNonEnumerableDataProperty(NewStringPropertyKey("message"), NewStringValue(msg))

			errorObject.Message = msg
		}

		InstallErrorCause(agent, errorObject, options)

		return NewValueFromObject(errorObject)
	}
	object := CreateBuiltinFunction(agent, behavior, 1, name, builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.ErrorConstructor,
		isConstructor: true,
	})

	protoName := "%" + name + ".prototype%"
	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.Get(protoName)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.Get(protoName), "constructor", NewValueFromObject(object))

	return object
}

func NewNativeErrorPrototype(realm *Realm, name string) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ErrorPrototype)

	DefineBuiltinPropertyV(object, "name", NewStringValue(name))
	DefineBuiltinPropertyV(object, "message", NewStringValue(""))

	return object
}

func NewAggregateErrorConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, _newTarget ObjectType) Value {
		errors := argumentsList[0]
		message := argumentsList[1]
		options := argumentsList[2]
		newTarget := _newTarget
		if newTarget == nil {
			newTarget = agent.ActiveFunctionObject()
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%AggregateError.prototype%", []string{})
		errorObject := &ErrorObject{
			Object:  o,
			Name:    "AggregateError",
			Message: "",
		}
		errorObject.InternalMethods().Set = errorInternalSet
		if message != UndefinedValue {
			msg := message.String()
			errorObject.CreateNonEnumerableDataProperty(NewStringPropertyKey("message"), NewStringValue(msg))
			errorObject.Message = msg
		}

		InstallErrorCause(agent, errorObject, options)
		errorsList := GetIterator(agent, errors, GetIteratorKindSync).IteratorToList()
		errorObject.DefinePropertyOrThrow(NewStringPropertyKey("errors"), &PropertyDescriptor{
			Value:        NewValueFromObject(CreateArrayFromList(agent, errorsList)),
			Writable:     true,
			Enumerable:   false,
			Configurable: true,
		})
		return NewValueFromObject(errorObject)
	}
	object := CreateBuiltinFunction(agent, behavior, 2, "AggregateError", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.ErrorConstructor,
	})
	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.AggregateErrorPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.AggregateErrorPrototype, "constructor", NewValueFromObject(object))
	return object
}

func NewAggregateErrorPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ErrorPrototype)
	DefineBuiltinPropertyV(object, "name", NewStringValue("AggregateError"))
	DefineBuiltinPropertyV(object, "message", NewStringValue(""))
	return object
}
