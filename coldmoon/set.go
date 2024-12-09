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
			panic("TypeError")
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%SetObject.prototype%", nil)
		s := &SetObject{
			Object:   o,
			SetValue: &SetValue{},
		}
		if iterable == UndefinedValue || iterable == NullValue {
			return NewValueFromObject(s)
		}
		adder := s.Get(NewStringPropertyKey("add"))
		if !IsCallable(adder) {
			panic("TypeError")
		}
		iteratorRecord := GetIterator(agent, iterable, IteratorKindSync)
		for {
			next := iteratorRecord.Data().IteratorStep()
			if next.(*BooleanObject).Data == false {
				return NewValueFromObject(s)
			}
			nextItem := IteratorValue(next)
			NewValueFromObject(MustGetObject(adder)).CallAssumeCallable(NewValueFromObject(s), []Value{nextItem})
		}
		return NewValueFromObject(s)
	}
	object := CreateBuiltinFunction(agent, behavior, 0, "SetObject", builtinFunctionArgs{
		prototype: realm.Intrinsics.FunctionPrototype,
		realm:     realm,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.SetPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	var getter BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		return thisValue
	}
	DefineBuiltinAccessor(realm, object, "@@species", getter, nil)

	DefineBuiltinPropertyV(realm.Intrinsics.SetPrototype, "constructor", NewValueFromObject(object))

	return object
}

func NewSetPrototype(realm *Realm) ObjectType {
	agent := realm.Agent

	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "SetPrototype")

	var setClear BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		set := RequireInternalSlot[*SetObject](thisValue)
		if set.SetValue == nil {
			panic("TypeError")
		}
		set.SetValue.Data = make(map[Value]Value)
		return UndefinedValue
	}
	var setDelete BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		if set.SetValue == nil {
			panic("TypeError")
		}
		delete(set.SetValue.Data, value)
		return TrueValue
	}
	var setHas BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		if set.SetValue == nil {
			panic("TypeError")
		}
		_, ok := set.SetValue.Data[value]
		return NewBooleanValue(ok)
	}
	var setSize BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		set := RequireInternalSlot[*SetObject](thisValue)
		if set.SetValue == nil {
			panic("TypeError")
		}
		return NewNumberValue(JSNumber(len(set.SetValue.Data)))
	}
	var setAdd BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		if set.SetValue == nil {
			panic("TypeError")
		}
		set.SetValue.Data[value] = value
		return thisValue
	}
	var setEntries BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		iterator := CreateSetIterator(agent, thisValue, objectOwnPropertiesKindKeyAndValue)
		return NewValueFromObject(iterator)
	}
	var setValues BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		iterator := CreateSetIterator(agent, thisValue, objectOwnPropertiesKindValue)
		return NewValueFromObject(iterator)
	}
	var forEach BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		callbackFn := argumentsList[0]
		thisArg := argumentsList[1]
		set := RequireInternalSlot[*SetObject](thisValue)
		if !IsCallable(callbackFn) {
			panic("TypeError")
		}
		entries := set.SetValue.Data
		numEntries := uint64(len(entries))
		index := uint64(0)
		for index < numEntries {
			if v, ok := entries[NewNumberValue(JSNumber(index))]; ok {
				callbackFn.CallAssumeCallable(thisArg, []Value{v, v, thisValue})
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
	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("SetObject"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}
