package coldmoon

type constructorProperties struct {
	Name               string
	PropertyDescriptor *PropertyDescriptor
}

// 19.1
func GlobalObjectProperties(r *Realm) []constructorProperties {
	properties := []constructorProperties{
		{
			"globalThis",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.GlobalEnv.GlobalThisValue),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Infinity",
			&PropertyDescriptor{
				Value:        InfinityValue,
				Writable:     false,
				Enumerable:   false,
				Configurable: false,
			},
		},
		{
			"NaN",
			&PropertyDescriptor{
				Value:        NaNValue,
				Writable:     false,
				Enumerable:   false,
				Configurable: false,
			},
		},
		{
			"undefined",
			&PropertyDescriptor{
				Value:        UndefinedValue,
				Writable:     false,
				Enumerable:   false,
				Configurable: false,
			},
		},
		{
			"Boolean",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.BooleanConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"isFinite",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.IsFinite),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"isNaN",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.IsNaN),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"eval",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Eval),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Object",
			&PropertyDescriptor{
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
		{
			"String",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.StringConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Number",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.NumberConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Symbol",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.SymbolConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"BigInt",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.BigIntConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Math",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Math),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Error",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.ErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"EvalError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.EvalErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"RangeError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.RangeErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"ReferenceError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.ReferenceErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"SyntaxError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.SyntaxErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"TypeError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.TypeErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"URIError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.URIErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Reflect",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Reflect),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Proxy",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Proxy),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"AggregateError",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.AggregateErrorConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Date",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.DateConstructor),
				Writable:     true,
				Enumerable:   false,
				Configurable: true,
			},
		},
		{
			"Map",
			&PropertyDescriptor{
				Value:        NewValueFromObject(r.Intrinsics.Map),
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
