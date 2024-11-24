package coldmoon

import "math"

type ArrayBufferObject struct {
	*Object
	ArrayBufferData          DataBlock
	ArrayBufferByteLength    uint64
	ArrayBufferDetachKey     Value
	ArrayBufferMaxByteLength uint64
}

func (a *ArrayBufferObject) ToValue() Value {
	return NewValueFromObject(a)
}

// 25.1.3.1
func AllocateArrayBuffer(agent *Agent, constructor ObjectType, byteLength uint64, maxByteLength uint64) *CompletionObject {
	var allocatingResizableBuffer bool
	if maxByteLength != 0 {
		allocatingResizableBuffer = true
	}
	if allocatingResizableBuffer {
		if byteLength > maxByteLength {
			return NewCompletionObjectError(agent.ThrowException(RangeError, "byteLength > maxByteLength"))
		}
	}
	object := OrdinaryCreateFromConstructor(agent, constructor, "%ArrayBuffer.prototype%", nil)
	arrayBuffer := &ArrayBufferObject{
		Object:                object,
		ArrayBufferData:       make([]byte, byteLength),
		ArrayBufferByteLength: byteLength,
	}
	return NewCompletionObject(arrayBuffer)
}

func GetArrayBufferMaxByteLengthOption(agent *Agent, options Value) (l uint64) {
	if !ValueIsObject(options) {
		return
	}
	maxByteLength := MustGetObject(options).Get(NewStringPropertyKey("maxByteLength"))
	if maxByteLength == UndefinedValue {
		return
	}

	return ToIndex(agent, maxByteLength)
}

// 25.1.2.2
func IsDetachedBuffer(buffer *ArrayBufferObject) bool {
	return buffer.ArrayBufferDetachKey != nil
}

// 25.1.2.3
func DetachArrayBuffer(buffer *ArrayBufferObject) {
	buffer.ArrayBufferData = nil
	buffer.ArrayBufferByteLength = 0
	buffer.ArrayBufferDetachKey = nil
}

func NewArrayBufferConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		length := argumentsList[0]
		options := argumentsList[1]
		if newTarget == nil {
			panic("TypeError")
		}
		byteLength := ToIndex(agent, length)
		requestedMaxByteLength := GetArrayBufferMaxByteLengthOption(agent, options)
		return AllocateArrayBuffer(agent, newTarget, byteLength, requestedMaxByteLength).Object.ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, "ArrayBuffer", builtinFunctionArgs{
		realm:         realm,
		isConstructor: true,
		prototype:     realm.Intrinsics.FunctionPrototype,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value: NewValueFromObject(realm.Intrinsics.ArrayBufferPrototype),
	})
	DefineBuiltinPropertyV(realm.Intrinsics.ArrayBufferPrototype, "constructor", NewValueFromObject(object))

	var getter BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		return this
	}
	DefineBuiltinAccessor(realm, object, "@@species", getter, nil)
	return object
}

func NewArrayBufferPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype)

	var isView BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		arg := arguments[0]
		if !ValueIsObject(arg) {
			return FalseValue
		}
		// TODO

		return FalseValue
	}
	var byteLength = func(this Value, arguments []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*ArrayBufferObject](this)
		if IsDetachedBuffer(o) {
			return NewNumberValue(0)
		}
		length := o.ArrayBufferByteLength
		return NewNumberValue(float64(length))
	}
	var slice = func(this Value, arguments []Value, newTarget ObjectType) Value {
		start := arguments[0]
		end := arguments[1]
		o := RequireInternalSlot[*ArrayBufferObject](this)
		if IsDetachedBuffer(o) {
			panic("TypeError")
		}
		length := o.ArrayBufferByteLength
		relativeStart := ToIntegerOrInfinity(agent, start)
		var first float64
		if math.IsInf(relativeStart, -1) {
			first = 0
		} else if relativeStart < 0 {
			first = math.Max(float64(length)+relativeStart, 0)
		} else {
			first = math.Min(relativeStart, float64(length))
		}

		relativeEnd := ToIntegerOrInfinity(agent, end)
		var final float64
		if math.IsInf(relativeEnd, -1) {
			final = 0
		} else if relativeEnd < 0 {
			final = math.Max(float64(length)+relativeEnd, 0)
		} else {
			final = math.Min(relativeEnd, float64(length))
		}

		newLen := math.Max(final-first, 0)
		ctor := o.SpeciesConstructor(realm.Intrinsics.ArrayBufferConstructor)
		newObject := MustGetObject(ctor.Value).Construct([]Value{NewNumberValue(newLen)}, nil)
		_new := RequireInternalSlot[*ArrayBufferObject](NewValueFromObject(newObject))
		if IsDetachedBuffer(_new) {
			panic("TypeError")
		}
		if _new == o {
			panic("TypeError")
		}
		if _new.ArrayBufferByteLength < uint64(newLen) {
			panic("TypeError")
		}
		if IsDetachedBuffer(o) {
			panic("TypeError")
		}
		fromBuf := o.ArrayBufferData
		toBuf := _new.ArrayBufferData
		currentLen := o.ArrayBufferByteLength
		if first < float64(currentLen) {
			count := math.Min(newLen, float64(currentLen)-first)
			CopyDataBlockBytes(toBuf, 0, fromBuf, int(first), int(count))
		}
		return NewValueFromObject(_new)
	}
	DefineBuiltinFunction(object, "isView", isView, 1, realm)
	DefineBuiltinAccessor(realm, object, "byteLength", byteLength, nil)
	DefineBuiltinFunction(object, "slice", slice, 2, realm)

	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("ArrayBuffer"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}
