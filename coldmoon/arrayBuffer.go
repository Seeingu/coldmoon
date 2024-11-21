package coldmoon

type ArrayBufferObject struct {
	*Object
	ArrayBufferData       []byte
	ArrayBufferByteLength uint64
	ArrayBufferDetachKey  Value
}

func (a *ArrayBufferObject) ToValue() Value {
	return NewValueFromObject(a)
}

// 25.1.2.1
func AllocateArrayBuffer(agent *Agent, constructor ObjectType, byteLength uint64) *ArrayBufferObject {
	object := OrdinaryCreateFromConstructor(agent, constructor, "%ArrayBuffer.prototype%", nil)
	arrayBuffer := &ArrayBufferObject{
		Object:                object,
		ArrayBufferData:       make([]byte, byteLength),
		ArrayBufferByteLength: byteLength,
	}
	return arrayBuffer
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
		if newTarget == nil {
			panic("TypeError")
		}
		byteLength := ToIndex(agent, length)
		return AllocateArrayBuffer(agent, newTarget, byteLength).ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, "ArrayBuffer", builtinFunctionArgs{
		realm:         realm,
		isConstructor: true,
		prototype:     realm.Intrinsics.FunctionPrototype,
	})

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
	DefineBuiltinFunction(object, "isView", isView, 1, realm)
	DefineBuiltinAccessor(realm, object, "byteLength", byteLength, nil)

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
	return object
}
