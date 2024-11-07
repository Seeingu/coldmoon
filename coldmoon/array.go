package coldmoon

import "math"

type ArrayObject struct {
	*Object
}

func getArrayLength(array ObjectType) float64 {
	lengthDesc := OrdinaryGetOwnProperty(array, NewStringPropertyKey("length"))
	Assert(lengthDesc.IsDataDescriptor())
	return lengthDesc.Value.(*NumberValue).Data

}

// 10.4.2.2
func ArrayCreate(agent *Agent, length float64, proto ObjectType) ObjectType {
	// 10.4.2.1
	var defineOwnProperty = func(array ObjectType, p PropertyKey, desc *PropertyDescriptor) bool {
		propertyKeyString, ok := p.(*StringPropertyKey)
		if ok && propertyKeyString.Value == "length" {
			return ArraySetLength(agent, array, desc)
		}
		propertyKeyIndex, ok := p.(*IntegerIndexPropertyKey)
		if ok {
			lengthDesc := OrdinaryGetOwnProperty(array, NewStringPropertyKey("length"))
			Assert(lengthDesc.IsDataDescriptor())
			Assert(!lengthDesc.Configurable)

			lengthValue := lengthDesc.Value
			length := lengthValue.(*NumberValue).Data
			Assert(math.IsInf(length, 0) || length >= 0)

			index := propertyKeyIndex.Value

			if index >= int(length) && !lengthDesc.Writable {
				return false
			}

			succeeded := OrdinaryDefineOwnProperty(array, p, desc)

			if !succeeded {
				return false
			}

			if index >= int(length) {
				lengthDesc.Value = NewNumberValue(float64(index) + 1)

				succeeded = OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), lengthDesc)
				Assert(succeeded)
			}
		}
		return true
	}

	realm := agent.CurrentRealm()

	if length > POW_2_32-1 {
		panic("RangeError")
	}

	if proto == nil {
		proto = realm.Intrinsics.ArrayPrototype
	}

	arr := &ArrayObject{
		Object: &Object{
			data: &Data{
				prototype: proto,
			},
		},
	}
	arr.Object.InternalMethods().DefineOwnProperty = defineOwnProperty
	OrdinaryDefineOwnProperty(arr, NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(length),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	return arr
}

// 10.4.2.4
func ArraySetLength(agent *Agent, array ObjectType, desc *PropertyDescriptor) bool {
	if desc.Value == nil {
		return OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), desc)
	}
	newLenDesc := desc

	newLenValue := desc.Value
	newLen := newLenValue.(*NumberValue).Data
	numberLen := ToNumber(agent, newLenValue)

	if numberLen.Data != newLen {
		panic("RangeError")
	}

	newLenDesc.Value = NewNumberValue(newLen)

	oldLenDesc := OrdinaryGetOwnProperty(array, NewStringPropertyKey("length"))
	Assert(oldLenDesc.IsDataDescriptor())
	Assert(!oldLenDesc.Configurable)
	oldLen := oldLenDesc.Value.(*NumberValue).Data

	if newLen >= oldLen {
		return OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), newLenDesc)
	}

	if !desc.Writable {
		return false
	}

	var newWritable = false
	// TODO: nil
	if newLenDesc.Writable {
		newWritable = true
	} else {
		newWritable = false
		newLenDesc.Writable = true
	}

	succeeded := OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), newLenDesc)
	if !succeeded {
		return false
	}

	if !newWritable {
		succeeded = OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), &PropertyDescriptor{
			Writable: false,
		})
		Assert(succeeded)
	}

	return true
}

type ArrayConstructor struct {
	*Object
}

func NewArrayConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		if newTarget == nil {
			newTarget = agent.ActiveFunctionObject()
		}

		proto := GetPrototypeFromConstructor(newTarget, "%Array.prototype%")

		numberOfArgs := len(args)
		if numberOfArgs == 0 {
			return NewValueFromObject(ArrayCreate(agent, 0, proto))
		} else if numberOfArgs == 1 {
			length := args[0]
			array := ArrayCreate(agent, 0, proto)

			var intLen uint32
			if _, ok := length.(*NumberValue); ok {
				array.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(0), length)
				intLen = 1
			} else {
				intLen = ToUint32(agent, length)
			}

			array.Set(
				NewStringPropertyKey("length"),
				NewNumberValue(float64(intLen)),
				setThrowTypeThrow,
			)

			return NewValueFromObject(array)
		} else {
			Assert(numberOfArgs >= 2)
			array := ArrayCreate(agent, float64(numberOfArgs), proto)

			for k := range numberOfArgs {
				propertyKey := NewIntegerIndexPropertyKey(k)
				array.CreateDataPropertyOrThrow(propertyKey, args[k])
			}

			Assert(getArrayLength(array) == float64(numberOfArgs))
			return NewValueFromObject(array)
		}
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 1, "Array", builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.ArrayPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinProperty(realm.Intrinsics.ArrayPrototype, "constructor", NewValueFromObject(object))

	return object
}

type ArrayPrototype struct {
	*Object
}

func NewArrayPrototype(realm *Realm) *ArrayPrototype {
	object := &ArrayPrototype{
		Object: &Object{
			data: &Data{
				prototype: realm.Intrinsics.ObjectPrototype,
			},
		},
	}

	DefineBuiltinProperty(object, "length", &PropertyDescriptor{
		Value:        NewNumberValue(0),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})

	return object

}
