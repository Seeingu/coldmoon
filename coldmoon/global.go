package coldmoon

type constructorProperties struct {
	Name               string
	PropertyDescriptor *PropertyDescriptor
}

// 19.1
func GlobalObjectProperties(r *Realm) []constructorProperties {
	properties := []constructorProperties{
		{
			Name: "globalThis",
			PropertyDescriptor: &PropertyDescriptor{
				Value:        NewValueFromObject(r.GlobalEnv.GlobalThisValue),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			Name: "Infinity",
			PropertyDescriptor: &PropertyDescriptor{
				Value:        InfinityValue,
				Writable:     false,
				Enumerable:   false,
				Configurable: false,
			},
		},
		{
			Name: "NaN",
			PropertyDescriptor: &PropertyDescriptor{
				Value:        NaNValue,
				Writable:     false,
				Enumerable:   false,
				Configurable: false,
			},
		},
		{
			Name: "undefined",
			PropertyDescriptor: &PropertyDescriptor{
				Value:        UndefinedValue,
				Writable:     false,
				Enumerable:   false,
				Configurable: false,
			},
		},
		{
			Name: "Boolean",
			PropertyDescriptor: &PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.BooleanConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			Name: "isFinite",
			PropertyDescriptor: &PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.IsFinite),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			Name: "isNaN",
			PropertyDescriptor: &PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.IsNaN),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			Name: "eval",
			PropertyDescriptor: &PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Eval),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			Name: "Object",
			PropertyDescriptor: &PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.ObjectConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Function",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.FunctionConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Array",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.ArrayConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
	}
	return properties
}

func NewIsFinite(realm *Realm) ObjectType {
	var isFinite BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		number := ToNumber(realm.Agent, args[0])
		return NewBooleanValue(number.IsFinite())
	}

	return CreateBuiltinFunction(realm.Agent, isFinite, 1, "isFinite", builtinFunctionArgs{
		realm: realm,
	})
}

func NewIsNaN(realm *Realm) ObjectType {
	var isNaN BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		number := ToNumber(realm.Agent, args[0])
		return NewBooleanValue(number.IsNaN())
	}

	return CreateBuiltinFunction(realm.Agent, isNaN, 1, "isNaN", builtinFunctionArgs{
		realm: realm,
	})
}

func NewEval(realm *Realm) ObjectType {
	var eval BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		if len(args) == 0 {
			return nil
		}
		return PerformEval(realm.Agent, args[0], false, false)
	}
	return CreateBuiltinFunction(realm.Agent, eval, 1, "eval", builtinFunctionArgs{
		realm: realm,
	})
}
