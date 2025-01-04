package coldmoon

import (
	"github.com/Seeingu/coldmoon/pkg"
)

type SetObject struct {
	*Object
	data            map[string]Value
	orderedHashKeys []string
}

// MARK: - Set Value utils

// items is a helper function that returns the values of the set in insertion order.
func (s *SetObject) items() []Value {
	ordered := make([]Value, 0, len(s.orderedHashKeys))
	for _, key := range s.orderedHashKeys {
		ordered = append(ordered, s.data[key])
	}
	return ordered
}

func (s *SetObject) get(value Value) Value {
	return s.data[value.Hash()]
}

func (s *SetObject) has(value Value) bool {
	_, ok := s.data[value.Hash()]
	return ok
}

func (s *SetObject) size() int {
	return len(s.data)
}

func (s *SetObject) add(value Value) {
	s.data[value.Hash()] = value
	s.orderedHashKeys = append(s.orderedHashKeys, value.Hash())
}

func (s *SetObject) delete(value Value) {
	s.orderedHashKeys = pkg.SliceDelete(s.orderedHashKeys, value.Hash())
	delete(s.data, value.Hash())
}

func (s *SetObject) clear() {
	s.data = make(map[string]Value)
	s.orderedHashKeys = make([]string, 0)
}

// MARK: - Set

func NewSetConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	behavior := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		iterable := pkg.SliceSafeGet(argumentsList, 0)
		if newTarget == nil {
			return agent.ThrowTypeError("new target is nil")
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%SetObject.prototype%", nil)
		s := &SetObject{
			Object:          o,
			data:            make(map[string]Value),
			orderedHashKeys: make([]string, 0),
		}
		s.ref = s
		if IsUndefinedOrNil(iterable) {
			return s.ToValue()
		}
		adder := s.Get(NewStringPropertyKey("add"))
		if !IsCallable(adder) {
			return agent.ThrowTypeError("adder is not callable")
		}
		iteratorRecord := GetIterator(agent, iterable, IteratorKindSync)
		for {
			next := iteratorRecord.Data().IteratorStep()
			if next == nil {
				return s.ToValue()
			}
			nextItem := IteratorValue(next)
			adder.Call(s.ToValue(), []Value{nextItem})
		}
	}
	object := CreateBuiltinFunctionV2(agent, behavior, 0, CMString("SetObject"), builtinFunctionArgs{
		prototype:     realm.Intrinsics.FunctionPrototype,
		realm:         realm,
		isConstructor: true,
	})

	species := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		return this
	}
	object.defineBuiltinAccessor(realm, WellKnownSymbolsSpecies, builtinAccessorParams{
		Getter: species,
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
		set.clear()
		return UndefinedValue
	}
	// 24.2.3.4
	var setDelete BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		set.delete(value)
		return TrueValue
	}
	// 24.2.3.7
	var setHas BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		return NewBooleanValue(set.has(value))
	}
	var setSize BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		set := RequireInternalSlot[*SetObject](thisValue)
		return NewNumberValue(JSNumber(set.size()))
	}
	var setAdd BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		set.add(value)
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
		entries := set.items()
		for _, item := range entries {
			callbackFn.Call(thisArg, []Value{item, item, thisValue})
		}
		return UndefinedValue
	}

	object.defineBuiltinFunction(realm, CMString("add"), setAdd, 1)
	object.defineBuiltinFunction(realm, CMString("clear"), setClear, 0)
	object.defineBuiltinFunction(realm, CMString("delete"), setDelete, 1)
	object.defineBuiltinFunction(realm, CMString("has"), setHas, 1)
	object.defineBuiltinFunction(realm, CMString("size"), setSize, 0)
	object.defineBuiltinFunction(realm, CMString("entries"), setEntries, 0)
	object.defineBuiltinFunction(realm, CMString("values"), setValues, 0)
	object.defineBuiltinFunction(realm, CMString("forEach"), forEach, 1)

	object.defineBuiltinProperty(CMString("keys"), object.PropertyStorage().Get(NewStringPropertyKey("values")))
	object.defineBuiltinProperty(WellKnownSymbolsIterator, object.PropertyStorage().Get(NewStringPropertyKey("values")))
	object.defineToStringTag("Set")
	return object
}
