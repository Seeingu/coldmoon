package coldmoon

type ErrorObject struct {
	*Object
	Name    string
	Message string
}

func NewErrorConstructor(realm *Realm) ObjectType {
	agent := realm.Agent

	var internalSet SetFn = func(o ObjectType, p PropertyKey, v Value, receiver Value) bool {
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
		errorObject.InternalMethods().Set = internalSet

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

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.ErrorPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinProperty(realm.Intrinsics.ErrorPrototype, "constructor", NewValueFromObject(object))

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

	DefineBuiltinProperty(object, "name", NewStringValue("Error"))

	DefineBuiltinProperty(object, "message", NewStringValue(""))

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
