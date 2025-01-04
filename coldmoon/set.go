package coldmoon

import (
	"github.com/Seeingu/coldmoon/pkg"
	goset "github.com/hashicorp/go-set/v3"
)

type SetValue struct {
	Value
	Data *goset.HashSet[Value, string]
}

func NewSetValue() *SetValue {
	return &SetValue{
		Data: goset.NewHashSet[Value, string](0),
	}
}

type SetObject struct {
	*Object
	SetValue *SetValue
}

func NewSetConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	behavior := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		iterable := pkg.SliceSafeGet(argumentsList, 0)
		if newTarget == nil {
			return agent.ThrowTypeError("new target is nil")
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%SetObject.prototype%", nil)
		s := &SetObject{
			Object:   o,
			SetValue: NewSetValue(),
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
		for item := range set.SetValue.Data.Items() {
			set.SetValue.Data.Remove(item)
		}
		return UndefinedValue
	}
	// 24.2.3.4
	var setDelete BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		set.SetValue.Data.Remove(value)
		return TrueValue
	}
	// 24.2.3.7
	var setHas BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		return NewBooleanValue(set.SetValue.Data.Contains(value))
	}
	var setSize BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		set := RequireInternalSlot[*SetObject](thisValue)
		return NewNumberValue(JSNumber(set.SetValue.Data.Size()))
	}
	var setAdd BehaviorFn = func(thisValue Value, argumentsList []Value, _newTarget ObjectType) Value {
		value := argumentsList[0]
		set := RequireInternalSlot[*SetObject](thisValue)
		set.SetValue.Data.Insert(value)
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
		entries := set.SetValue.Data.Items()
		for item := range entries {
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
