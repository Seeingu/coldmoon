package coldmoon

import "math"

type MathObject struct {
	*Object
}

func NewMathObject(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype)

	DefineBuiltinProperty(object, "E", &PropertyDescriptor{
		Value:        NewNumberValue(math.E),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "LN10", &PropertyDescriptor{
		Value:        NewNumberValue(math.Ln10),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "LN2", &PropertyDescriptor{
		Value:        NewNumberValue(math.Ln2),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "LOG2E", &PropertyDescriptor{
		Value:        NewNumberValue(math.Log2E),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "LOG10E", &PropertyDescriptor{
		Value:        NewNumberValue(math.Log10E),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "PI", &PropertyDescriptor{
		Value:        NewNumberValue(math.Pi),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "SQRT1_2", &PropertyDescriptor{
		Value:        NewNumberValue(math.Sqrt(1 / 2)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "SQRT2", &PropertyDescriptor{
		Value:        NewNumberValue(math.Sqrt2),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("Math"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}
