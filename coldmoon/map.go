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
	iteratorRecord := GetIterator(agent, iterable, GetIteratorKindSync)
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
		NewValueFromObject(adder).CallAssumeCallable(NewValueFromObject(target), []Value{k, v})
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
		if iterable == UndefinedValue || iterable == NullValue {
			return NewValueFromObject(m)
		}
		adder := m.Get(NewStringPropertyKey("set"))
		if !IsCallable(adder) {
			panic("TypeError")
		}
		return NewValueFromObject(AddEntriesFromIterable(agent, m, iterable, MustGetObject(adder)))

	}
	object := CreateBuiltinFunction(agent, behavior, 0, "Map", builtinFunctionArgs{
		prototype: realm.Intrinsics.FunctionPrototype,
		realm:     realm,
	})
	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.MapPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(realm.Intrinsics.MapPrototype, "constructor", NewValueFromObject(object))

	return object
}

func NewMapPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype)
	return object
}
