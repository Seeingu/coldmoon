package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type MapValue struct {
	Value
	Data map[string]Value
}

func NewMapValue() *MapValue {
	return &MapValue{
		Data: make(map[string]Value),
	}
}

type MapObject struct {
	*Object
	MapValue *MapValue
}

// AddEntriesFromIterable
// spec: 24.1.1.2
func AddEntriesFromIterable(agent *Agent, target ObjectType, iterable Value, adder ObjectType) (co CompletionValue) {
	iteratorRecord, isAbrupt, rt := ReturnIfAbrupt(GetIterator(agent, iterable, IteratorKindSync), co)
	if isAbrupt {
		return rt
	}
	for {
		next, isDone := iteratorRecord.IteratorStepValue()
		nextItem, isAbrupt, rt := ReturnIfAbrupt(next, co)
		if isAbrupt {
			// Failures raised by the iterator itself are returned directly; only
			// failures while consuming an entry require IteratorClose.
			return rt
		}
		if isDone {
			co.value = target.ToValue()
			return
		}

		if nextItem == nil || !nextItem.IsObject() {
			return iteratorRecord.IteratorClose(
				co.ThrowTypeError(agent, "iterator entry must be an object"),
			)
		}
		entry := MustGetObject(nextItem)
		k, isAbrupt, rt := ReturnIfAbrupt(
			entry.internalMethods().Get(entry, NewStringPropertyKey("0"), nextItem),
			co,
		)
		if isAbrupt {
			return iteratorRecord.IteratorClose(rt)
		}
		v, isAbrupt, rt := ReturnIfAbrupt(
			entry.internalMethods().Get(entry, NewStringPropertyKey("1"), nextItem),
			co,
		)
		if isAbrupt {
			return iteratorRecord.IteratorClose(rt)
		}
		status := adder.Call(target.ToValue(), []Value{k, v})
		if status.IsAbrupt() {
			return iteratorRecord.IteratorClose(status)
		}
	}
}

func NewMapConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		iterable := pkg.SliceSafeGet(argumentsList, 0)
		if newTarget == nil {
			return co.ThrowTypeError(agent, "Map constructor requires new")
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%Map.prototype%", nil)
		m := &MapObject{
			Object:   o,
			MapValue: NewMapValue(),
		}
		m.ref = m
		if IsUndefinedOrNil(iterable) {
			return (m).ToValue()
		}
		adder, isAbrupt, rt := ReturnIfAbrupt(
			m.internalMethods().Get(m, NewStringPropertyKey("set"), m.ToValue()),
			co,
		)
		if isAbrupt {
			return rt
		}
		if !IsCallable(adder) {
			return co.ThrowTypeError(agent, "Map adder is not callable")
		}
		return AddEntriesFromIterable(agent, m, iterable, MustGetObject(adder))
	}
	object := CreateBuiltinFunction(agent, behavior, 0, CMString("Map"), builtinFunctionArgs{
		prototype:     realm.Intrinsics.FunctionPrototype,
		realm:         realm,
		isConstructor: true,
	})

	object.defineBuiltinAccessor(realm, WellKnownSymbolsSpecies, builtinAccessorParams{
		Getter: func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
			return this
		},
	})

	BindPrototypeAndConstructor(realm.Intrinsics.MapPrototype, object)

	return object
}

func NewMapPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "MapPrototype")

	var mapClear BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		m := RequireInternalSlot[*MapObject](this)
		m.MapValue.Data = make(map[string]Value)
		return UndefinedValue
	}
	var mapDelete BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		m := RequireInternalSlot[*MapObject](this)
		key := argumentAt(arguments, 0).Hash()
		if _, ok := m.MapValue.Data[key]; !ok {
			return FalseValue
		}
		delete(m.MapValue.Data, key)
		return TrueValue
	}
	var mapGet BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		m := RequireInternalSlot[*MapObject](this)
		key := argumentAt(arguments, 0).Hash()
		if v, ok := m.MapValue.Data[key]; ok {
			return v
		}
		return UndefinedValue
	}
	var mapHas BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		m := RequireInternalSlot[*MapObject](this)
		key := argumentAt(arguments, 0).Hash()
		_, ok := m.MapValue.Data[key]
		return NewBooleanValue(ok)
	}
	var mapSet BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		m := RequireInternalSlot[*MapObject](this)
		key := argumentAt(arguments, 0).Hash()
		value := argumentAt(arguments, 1)
		m.MapValue.Data[key] = value
		return this
	}
	var size BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		m := RequireInternalSlot[*MapObject](this)
		return NewNumberValue(JSNumber(len(m.MapValue.Data)))
	}
	var mapEntries BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return CreateMapIterator(agent, this, objectOwnPropertiesKindKeyAndValue).ToValue()
	}
	var mapKeys BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return CreateMapIterator(agent, this, objectOwnPropertiesKindKey).ToValue()
	}
	var mapValues BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return CreateMapIterator(agent, this, objectOwnPropertiesKindValue).ToValue()
	}
	var forEach BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		m := RequireInternalSlot[*MapObject](this)
		callbackFn := argumentAt(arguments, 0)
		thisArg := argumentAt(arguments, 1)
		if !IsCallable(callbackFn) {
			panic("TypeError")
		}
		entries := m.MapValue.Data
		numEntries := len(m.MapValue.Data)
		index := 0
		for ; index < numEntries; index++ {
			if v, ok := entries[NewNumberValue(JSNumber(index)).Hash()]; ok {
				callbackFn.Call(agent, thisArg, []Value{v, NewNumberValue(JSNumber(index)), this})
			}
			numEntries = len(m.MapValue.Data)
		}
		return UndefinedValue
	}

	object.defineBuiltinFunction(realm, CMString("clear"), mapClear, 0)
	object.defineBuiltinFunction(realm, CMString("delete"), mapDelete, 1)
	object.defineBuiltinFunction(realm, CMString("get"), mapGet, 1)
	object.defineBuiltinFunction(realm, CMString("has"), mapHas, 1)
	object.defineBuiltinFunction(realm, CMString("set"), mapSet, 2)
	object.defineBuiltinFunction(realm, CMString("entries"), mapEntries, 0)
	object.defineBuiltinFunction(realm, CMString("keys"), mapKeys, 0)
	object.defineBuiltinFunction(realm, CMString("values"), mapValues, 0)
	object.defineBuiltinFunction(realm, CMString("forEach"), forEach, 1)
	object.defineBuiltinProperty(WellKnownSymbolsIterator, object.propertyStorage().Get(NewStringPropertyKey("entries")))
	object.defineToStringTag("Map")
	object.defineBuiltinAccessor(realm, CMString("size"), builtinAccessorParams{
		Getter: size,
	})
	return object
}
