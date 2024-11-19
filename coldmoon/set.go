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
	var behavior = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		iterable := argumentsList[0]
		if newTarget == nil {
			panic("TypeError")
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%Set.prototype%", nil)
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
		iteratorRecord := GetIterator(agent, iterable, GetIteratorKindSync)
		for {
			next := iteratorRecord.IteratorStep()
			if next.(*BooleanObject).Data == false {
				return NewValueFromObject(s)
			}
			nextItem := IteratorValue(next)
			NewValueFromObject(MustGetObject(adder)).CallAssumeCallable(NewValueFromObject(s), []Value{nextItem})
		}
		return NewValueFromObject(s)

	}
	object := CreateBuiltinFunction(agent, behavior, 0, "Set", builtinFunctionArgs{
		prototype: realm.Intrinsics.FunctionPrototype,
		realm:     realm,
	})

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.SetPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinProperty(realm.Intrinsics.SetPrototype, "constructor", NewValueFromObject(object))

	return object
}

func NewSetPrototype(realm *Realm) ObjectType {
	agent := realm.Agent

	object := NewObject(agent, realm.Intrinsics.ObjectPrototype)
	return object
}
