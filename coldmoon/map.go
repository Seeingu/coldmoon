package coldmoon

type MapValue struct {
	Value
	Data map[Value]Value
}

type MapObject struct {
	*Object
	MapValue *MapValue
}

func AddEntriesFromIterable(agent *Agent, target ObjectType, iterable Value, adder ObjectType) ObjectType {
	iteratorRecord := GetIterator(agent, iterable, IteratorKindSync).Data()
	for {
		next := iteratorRecord.IteratorStep()
		if next.(*BooleanObject).Data == false {
			return target
		}

		nextItem := IteratorValue(next)
		if !ValueIs[*ObjectValue](nextItem) {
			panic("TypeError")
			// TODO IteratorClose
		}
		k := MustGetObject(nextItem).Get(NewStringPropertyKey("0"))
		v := MustGetObject(nextItem).Get(NewStringPropertyKey("1"))
		(adder).ToValue().CallAssumeCallable((target).ToValue(), []Value{k, v})
	}
	return target
}

func NewMapConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		iterable := argumentsList[0]
		if newTarget == nil {
			panic("TypeError")
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%Map.prototype%", nil)
		m := &MapObject{
			Object:   o,
			MapValue: &MapValue{},
		}
		m.ref = m
		if iterable == UndefinedValue || iterable == NullValue {
			return (m).ToValue()
		}
		adder := m.Get(NewStringPropertyKey("set"))
		if !IsCallable(adder) {
			panic("TypeError")
		}
		return (AddEntriesFromIterable(agent, m, iterable, MustGetObject(adder))).ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 0, "Map", builtinFunctionArgs{
		prototype: realm.Intrinsics.FunctionPrototype,
		realm:     realm,
	})

	DefineBuiltinAccessorV2(realm, object, BuiltinAccessorParams{
		Getter: func(this Value, argumentsList []Value, newTarget ObjectType) Value {
			return this
		},
		WellKnownSymbolsKey: WellKnownSymbolsSpecies,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        (realm.Intrinsics.MapPrototype).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.MapPrototype, "constructor", (object).ToValue())

	return object
}

func NewMapPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "MapPrototype")

	var mapClear BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		m := RequireInternalSlot[*MapObject](this)
		m.MapValue.Data = make(map[Value]Value)
		return UndefinedValue
	}
	var mapDelete BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		m := RequireInternalSlot[*MapObject](this)
		key := arguments[0]
		if _, ok := m.MapValue.Data[key]; !ok {
			return FalseValue
		}
		delete(m.MapValue.Data, key)
		return TrueValue
	}
	var mapGet BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		m := RequireInternalSlot[*MapObject](this)
		key := arguments[0]
		if v, ok := m.MapValue.Data[key]; ok {
			return v
		}
		return UndefinedValue
	}
	var mapHas BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		m := RequireInternalSlot[*MapObject](this)
		key := arguments[0]
		_, ok := m.MapValue.Data[key]
		return NewBooleanValue(ok)
	}
	var mapSet BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		m := RequireInternalSlot[*MapObject](this)
		key := arguments[0]
		value := arguments[1]
		m.MapValue.Data[key] = value
		return this
	}
	var size BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		m := RequireInternalSlot[*MapObject](this)
		return NewNumberValue(JSNumber(len(m.MapValue.Data)))
	}
	var mapEntries BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		return (CreateMapIterator(agent, this, objectOwnPropertiesKindKeyAndValue)).ToValue()
	}
	var mapKeys BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		return (CreateMapIterator(agent, this, objectOwnPropertiesKindKey)).ToValue()
	}
	var mapValues BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		return (CreateMapIterator(agent, this, objectOwnPropertiesKindValue)).ToValue()
	}
	var forEach BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		m := RequireInternalSlot[*MapObject](this)
		callbackFn := arguments[0]
		thisArg := arguments[1]
		if !IsCallable(callbackFn) {
			panic("TypeError")
		}
		entries := m.MapValue.Data
		numEntries := len(m.MapValue.Data)
		index := 0
		for ; index < numEntries; index++ {
			if v, ok := entries[NewNumberValue(JSNumber(index))]; ok {
				callbackFn.CallAssumeCallable(thisArg, []Value{v, NewNumberValue(JSNumber(index)), this})
			}
			numEntries = len(m.MapValue.Data)
		}
		return UndefinedValue
	}

	DefineBuiltinFunction(object, "clear", mapClear, 0, realm)
	DefineBuiltinFunction(object, "delete", mapDelete, 1, realm)
	DefineBuiltinFunction(object, "get", mapGet, 1, realm)
	DefineBuiltinFunction(object, "has", mapHas, 1, realm)
	DefineBuiltinFunction(object, "set", mapSet, 2, realm)
	DefineBuiltinFunction(object, "size", size, 0, realm)
	DefineBuiltinFunction(object, "entries", mapEntries, 0, realm)
	DefineBuiltinFunction(object, "keys", mapKeys, 0, realm)
	DefineBuiltinFunction(object, "values", mapValues, 0, realm)
	DefineBuiltinFunction(object, "forEach", forEach, 1, realm)

	DefineBuiltinPropertyP(object, "@@iterator", object.PropertyStorage().Get(NewStringPropertyKey("entries")))
	DefineToStringTagBuiltinProperty(object, "Map")
	return object
}
