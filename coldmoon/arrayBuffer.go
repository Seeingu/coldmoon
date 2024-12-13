package coldmoon

import (
	"bytes"
	"encoding/binary"
	"math"
)

// ArrayBufferLike Enum
type ArrayBufferLike struct {
	*Object
	ArrayBuffer       *ArrayBufferObject
	SharedArrayBuffer *SharedArrayBufferObject
}

func (a *ArrayBufferLike) Data() DataBlock {
	if a.ArrayBuffer != nil {
		return a.ArrayBuffer.ArrayBufferData
	}
	return a.SharedArrayBuffer.ArrayBufferData
}

func (a *ArrayBufferLike) ByteLength() JSInt {
	return JSInt(len(a.Data()))
}

func (a *ArrayBufferLike) MaxByteLength() JSInt {
	if a.ArrayBuffer != nil {
		return a.ArrayBuffer.ArrayBufferMaxByteLength
	}
	return a.SharedArrayBuffer.ArrayBufferMaxByteLength
}

func NewArrayBufferLike(object ObjectType) *ArrayBufferLike {
	switch o := object.(type) {
	case *ArrayBufferObject:
		a := &ArrayBufferLike{Object: o.Object, ArrayBuffer: o}
		a.ref = o
		return a
	case *SharedArrayBufferObject:
		a := &ArrayBufferLike{Object: o.Object, SharedArrayBuffer: o}
		a.ref = o
		return a
	default:
		panic("Unexpected object type in NewArrayBufferLike")
	}
}

type ArrayBufferObject struct {
	*Object
	ArrayBufferData          DataBlock
	ArrayBufferByteLength    JSInt
	ArrayBufferDetachKey     Value
	ArrayBufferMaxByteLength JSInt
}

// 25.1.3.1
func AllocateArrayBuffer(agent *Agent, constructor ObjectType, byteLength JSInt, maxByteLength JSInt) CompletionObject {
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

// 25.1.3.6
// return either an ArrayBufferObject or a throw completion
func CloneArrayBuffer(agent *Agent, srcBuffer *ArrayBufferLike, srcByteOffset JSInt, srcLength JSInt) CompletionObject {
	realm := agent.CurrentRealm()
	Assert(!IsDetachedBuffer(srcBuffer))
	targetBuffer := AllocateArrayBuffer(agent, realm.Intrinsics.ArrayBufferConstructor, srcLength, 0)
	srcBlock := srcBuffer.Data()
	targetBlock := targetBuffer.Data().(*ArrayBufferObject).ArrayBufferData
	CopyDataBlockBytes(targetBlock, 0, srcBlock, srcByteOffset, srcLength)
	return targetBuffer
}

// 25.1.3.7
func GetArrayBufferMaxByteLengthOption(agent *Agent, options Value) (l JSInt) {
	if !ValueIsObject(options) {
		return
	}
	maxByteLength := MustGetObject(options).Get(NewStringPropertyKey("maxByteLength"))
	if maxByteLength == UndefinedValue {
		return
	}

	return ToIndex(agent, maxByteLength)
}

func IsBigIntElementType(size JSInt) bool {
	return size == 8
}

// 25.1.2.2
func IsDetachedBuffer(buffer *ArrayBufferLike) bool {
	if buffer.ArrayBuffer != nil {
		return buffer.ArrayBuffer.ArrayBufferDetachKey != nil
	}
	return false
}

// 25.1.2.3
func DetachArrayBuffer(buffer *ArrayBufferObject) {
	buffer.ArrayBufferData = nil
	buffer.ArrayBufferByteLength = 0
	buffer.ArrayBufferDetachKey = nil
}

type MemoryOrder int

const (
	SeqCst MemoryOrder = iota
	Relaxed
)

func ArrayBufferByteLength(buffer *ArrayBufferLike, memoryOrder MemoryOrder) JSInt {
	Assert(!IsDetachedBuffer(buffer))
	if buffer.ArrayBuffer != nil {
		return buffer.ArrayBuffer.ArrayBufferByteLength
	}
	return buffer.SharedArrayBuffer.ArrayBufferByteLength
}

// 25.1.3.8
func IsFixedLengthArrayBuffer(buffer *ArrayBufferLike) bool {
	if buffer.ArrayBuffer != nil {
		return buffer.ArrayBuffer.ArrayBufferMaxByteLength == 0
	}
	return buffer.SharedArrayBuffer.ArrayBufferMaxByteLength == 0
}

func GetValueFromBuffer(agent *Agent, arrayBuffer *ArrayBufferLike, byteIndex JSInt, size JSInt, isTypedArray bool, order MemoryOrder, isLittleEndian bool) JSInt {
	Assert(!IsDetachedBuffer(arrayBuffer))
	block := arrayBuffer.Data()
	elementSize := size

	var rawValue []byte
	if IsSharedArrayBuffer(arrayBuffer) {
		// TODO
	} else {
		rawValue = block[byteIndex : byteIndex+elementSize]
	}

	return RawBytesToNumeric(rawValue, isLittleEndian)
}

func IsSharedArrayBuffer(buffer *ArrayBufferLike) bool {
	return buffer.SharedArrayBuffer != nil
}

// 25.1.3.17
func NumericToRawBytes(value Value, size JSInt, isLittleEndian bool) []byte {
	// TODO: convert Value to bytes
	buf := &bytes.Buffer{}
	var endian binary.ByteOrder
	if isLittleEndian {
		endian = binary.LittleEndian
	} else {
		endian = binary.BigEndian
	}
	err := binary.Write(buf, endian, value)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func SetValueInBuffer(
	agent *Agent,
	arrayBuffer *ArrayBufferLike,
	byteIndex JSInt,
	value Value,
	size JSInt,
	isTypedArray bool,
	order MemoryOrder,
	isLittleEndian bool,
) {
	Assert(!IsDetachedBuffer(arrayBuffer))
	Assert(byteIndex+size <= arrayBuffer.ByteLength())
	block := arrayBuffer.Data()
	elementSize := size
	rawBytes := NumericToRawBytes(value, elementSize, isLittleEndian)
	CopyDataBlockBytes(block, byteIndex, rawBytes, 0, size)
}

// 25.1.3.14
func RawBytesToNumeric(rawBytes []byte, isLittleEndian bool) JSInt {
	buf := &bytes.Buffer{}
	var endian binary.ByteOrder
	if isLittleEndian {
		endian = binary.LittleEndian
	} else {
		endian = binary.BigEndian
	}
	err := binary.Write(buf, endian, rawBytes)
	if err != nil {
		panic(err)
	}
	t := JSInt(0)
	err = binary.Read(buf, endian, &t)
	if err != nil {
		panic(err)
	}
	return t
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
		return NewValueFromObject(AllocateArrayBuffer(agent, newTarget, byteLength, requestedMaxByteLength).Data())
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
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "ArrayBufferPrototype")

	var isView BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		arg := arguments[0]
		if !ValueIsObject(arg) {
			return FalseValue
		}
		o := MustGetObject(arg)
		if ObjectIs[*DataView](o) {
			return TrueValue
		}

		return FalseValue
	}
	byteLength := func(this Value, arguments []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*ArrayBufferLike](this)
		if IsDetachedBuffer(o) {
			return NewNumberValue(0)
		}
		length := o.ByteLength()
		return NewNumberValue(length.ToNumber())
	}
	slice := func(this Value, arguments []Value, newTarget ObjectType) Value {
		start := arguments[0]
		end := arguments[1]
		o := RequireInternalSlot[*ArrayBufferLike](this)
		if IsDetachedBuffer(o) {
			panic("TypeError")
		}
		length := o.ByteLength()
		relativeStart := ToIntegerOrInfinity(agent, start)
		var first JSInt
		if relativeStart.IsNegInf() {
			first = 0
		} else if relativeStart < 0 {
			first = JSInt(math.Max(float64(length+relativeStart), 0))
		} else {
			first = JSInt(math.Min(float64(relativeStart), float64(length)))
		}

		relativeEnd := ToIntegerOrInfinity(agent, end)
		var final JSInt
		if math.IsInf(float64(relativeEnd), -1) {
			final = 0
		} else if relativeEnd < 0 {
			final = JSInt(math.Max(float64(length+relativeEnd), 0))
		} else {
			final = JSInt(math.Min(float64(relativeEnd), float64(length)))
		}

		newLen := JSInt(math.Max(float64(final-first), 0))
		ctor := o.SpeciesConstructor(realm.Intrinsics.ArrayBufferConstructor)
		newObject := ctor.Data().Construct([]Value{NewNumberValue(JSNumber(newLen))}, nil)
		_new := RequireInternalSlot[*ArrayBufferLike](NewValueFromObject(newObject))
		if IsDetachedBuffer(_new) {
			panic("TypeError")
		}
		if _new == o {
			panic("TypeError")
		}
		if _new.ByteLength() < newLen {
			panic("TypeError")
		}
		if IsDetachedBuffer(o) {
			panic("TypeError")
		}
		fromBuf := o.Data()
		toBuf := _new.Data()
		currentLen := o.ByteLength()
		if first < currentLen {
			count := JSInt(math.Min(float64(newLen), float64(currentLen-first)))
			CopyDataBlockBytes(toBuf, 0, fromBuf, first, count)
		}
		return NewValueFromObject(_new)
	}
	maxByteLength := func(this Value, arguments []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*ArrayBufferLike](this)
		if IsDetachedBuffer(o) {
			return NewNumberValue(0)
		}
		var length JSInt
		if IsFixedLengthArrayBuffer(o) {
			length = o.ByteLength()
		} else {
			length = o.MaxByteLength()
		}
		return NewNumberValue(length.ToNumber())
	}
	resizable := func(this Value, arguments []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*ArrayBufferLike](this)
		return NewBooleanValue(IsFixedLengthArrayBuffer(o))
	}
	resize := func(this Value, arguments []Value, newTarget ObjectType) Value {
		newByteLength := ToIndex(agent, arguments[0])
		o := RequireInternalSlot[*ArrayBufferLike](this)
		if o.MaxByteLength() == 0 {
			panic("TypeError")
		}
		if IsDetachedBuffer(o) {
			panic("TypeError")
		}
		if newByteLength > o.MaxByteLength() {
			panic("TypeError")
		}
		hostHandled := HostResizeArrayBuffer(o, newByteLength)
		if hostHandled == ResizeArrayBufferHandledHandled {
			return UndefinedValue
		}

		// TODO: Resize

		return UndefinedValue
	}

	DefineBuiltinFunction(object, "isView", isView, 1, realm)
	DefineBuiltinFunction(object, "slice", slice, 2, realm)
	DefineBuiltinFunction(object, "resize", resize, 1, realm)

	DefineBuiltinAccessor(realm, object, "byteLength", byteLength, nil)
	DefineBuiltinAccessor(realm, object, "maxByteLength", maxByteLength, nil)
	DefineBuiltinAccessor(realm, object, "resizable", resizable, nil)

	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("ArrayBuffer"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}
