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
	}
	return properties

}
