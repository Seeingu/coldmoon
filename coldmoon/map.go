package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type MapValue struct {
	Value
	Data    map[string]*MapEntry
	Entries []*MapEntry
}

// MapEntry retains the original key alongside its value so iteration can
// preserve insertion order independently of the key's internal hash.
type MapEntry struct {
	Key     Value
	Value   Value
	Deleted bool
}

func NewMapValue() *MapValue {
	return &MapValue{
		Data: make(map[string]*MapEntry),
	}
}

// clear marks existing entries as deleted so live iterators can continue over
// entries appended after the clear operation.
func (m *MapValue) clear() {
	for _, entry := range m.Entries {
		entry.Deleted = true
	}
	m.Data = make(map[string]*MapEntry)
}

// delete removes a key from lookup while retaining its tombstone in the
// insertion-order sequence used by active iterators.
func (m *MapValue) delete(key Value) bool {
	hash := key.Hash()
	entry, ok := m.Data[hash]
	if !ok {
		return false
	}
	entry.Deleted = true
	delete(m.Data, hash)
	return true
}

// set updates an existing entry in place or appends a new insertion-order
// record when the key is not currently present.
func (m *MapValue) set(key, value Value) {
	hash := key.Hash()
	if entry, ok := m.Data[hash]; ok {
		entry.Value = value
		return
	}
	entry := &MapEntry{Key: key, Value: value}
	m.Data[hash] = entry
	m.Entries = append(m.Entries, entry)
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
		m.MapValue.clear()
		return UndefinedValue
	}
	var mapDelete BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		m := RequireInternalSlot[*MapObject](this)
		return NewBooleanValue(m.MapValue.delete(argumentAt(arguments, 0)))
	}
	var mapGet BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		m := RequireInternalSlot[*MapObject](this)
		key := argumentAt(arguments, 0).Hash()
		if entry, ok := m.MapValue.Data[key]; ok {
			return entry.Value
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
		m.MapValue.set(argumentAt(arguments, 0), argumentAt(arguments, 1))
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
		entries := m.MapValue.Entries
		numEntries := len(entries)
		index := 0
		for ; index < numEntries; index++ {
			entry := entries[index]
			if !entry.Deleted {
				callbackFn.Call(agent, thisArg, []Value{entry.Value, entry.Key, this})
			}
			numEntries = len(m.MapValue.Entries)
			entries = m.MapValue.Entries
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
