package coldmoon

type SetValue struct {
	Value
	Data map[Value]Value
}
type SetObject struct {
	*Object
	SetValue *SetValue
}

func NewSetConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	behavior := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		iterable := argumentsList[0]
		if newTarget == nil {
			return agent.ThrowTypeError("new target is nil")
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%SetObject.prototype%", nil)
		s := &SetObject{
			Object:   o,
			SetValue: &SetValue{},
		}
		s.ref = s
		if iterable == UndefinedValue || iterable == NullValue {
			return (s).ToValue()
		}
		adder := s.Get(NewStringPropertyKey("add"))
		if !IsCallable(adder) {
			return agent.ThrowTypeError("adder is not callable")
		}
		iteratorRecord := GetIterator(agent, iterable, IteratorKindSync)
		for {
			next := iteratorRecord.Data().IteratorStep()
			if next.(*BooleanObject).Data == false {
				return (s).ToValue()
			}
			nextItem := IteratorValue(next)
			(MustGetObject(adder)).ToValue().Call((s).ToValue(), []Value{nextItem})
		}
	}
	object := CreateBuiltinFunction(agent, behavior, 0, "SetObject", builtinFunctionArgs{
		prototype: realm.Intrinsics.FunctionPrototype,
		realm:     realm,
	})

	DefineBuiltinAccessorV2(realm, object, BuiltinAccessorParams{
		Getter: func(this Value, argumentsList []Value, newTarget ObjectType) Value {
			return this
		},
		WellKnownSymbolsKey: WellKnownSymbolsSpecies,
	})

	BindPrototypeAndConstructor(realm.Intrinsics.SetPrototype, object)

	return object
}

func NewSetPrototype(realm *Realm) ObjectType {
	agent := realm.Agent

	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "SetPrototype")

	// 24.2.3.2
	var setClear BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		set := RequireInternalSlot[*SetObject](thisValue)
		set.SetValue.Data = make(map[Value]Value)
		return UndefinedValue
	}
	// 24.2.3.4
	var setDelete BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		delete(set.SetValue.Data, value)
		return TrueValue
	}
	// 24.2.3.7
	var setHas BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		_, ok := set.SetValue.Data[value]
		return NewBooleanValue(ok)
	}
	var setSize BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		set := RequireInternalSlot[*SetObject](thisValue)
		return NewNumberValue(JSNumber(len(set.SetValue.Data)))
	}
	var setAdd BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		set.SetValue.Data[value] = value
		return thisValue
	}
	var setEntries BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		iterator := CreateSetIterator(agent, thisValue, objectOwnPropertiesKindKeyAndValue)
		return iterator.ToValue()
	}
	var setValues BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		iterator := CreateSetIterator(agent, thisValue, objectOwnPropertiesKindValue)
		return iterator.ToValue()
	}
	var forEach BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		callbackFn := argumentsList[0]
		thisArg := argumentsList[1]
		set := RequireInternalSlot[*SetObject](thisValue)
		if !IsCallable(callbackFn) {
			return agent.ThrowTypeError("callback is not callable")
		}
		entries := set.SetValue.Data
		numEntries := uint64(len(entries))
		index := uint64(0)
		for index < numEntries {
			if v, ok := entries[NewNumberValue(JSNumber(index))]; ok {
				callbackFn.Call(thisArg, []Value{v, v, thisValue})
			}
			numEntries = uint64(len(entries))
			index++
		}
		return UndefinedValue
	}

	DefineBuiltinFunction(object, "add", setAdd, 1, realm)
	DefineBuiltinFunction(object, "clear", setClear, 0, realm)
	DefineBuiltinFunction(object, "delete", setDelete, 1, realm)
	DefineBuiltinFunction(object, "has", setHas, 1, realm)
	DefineBuiltinFunction(object, "size", setSize, 0, realm)
	DefineBuiltinFunction(object, "entries", setEntries, 0, realm)
	DefineBuiltinFunction(object, "values", setValues, 0, realm)
	DefineBuiltinFunction(object, "forEach", forEach, 1, realm)

	DefineBuiltinPropertyP(object, "keys", object.PropertyStorage().Get(NewStringPropertyKey("values")))
	DefineBuiltinPropertyP(object, "@@iterator", object.PropertyStorage().Get(NewStringPropertyKey("values")))
	DefineToStringTagBuiltinProperty(object, "Set")
	return object
}
