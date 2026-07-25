package coldmoon

import (
	"encoding/binary"
	"math"
)

// ArrayBufferLike Enum
type ArrayBufferLike struct {
	*Object
	ArrayBuffer       *ArrayBufferObject
	SharedArrayBuffer *SharedArrayBufferObject
}

func (a *ArrayBufferLike) Data() *DataBlock {
	if a.ArrayBuffer != nil {
		return a.ArrayBuffer.ArrayBufferData
	}
	return a.SharedArrayBuffer.ArrayBufferData
}

func (a *ArrayBufferLike) ByteLength() JSInt {
	return a.Data().Size()
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
	ArrayBufferData          *DataBlock
	ArrayBufferByteLength    JSInt
	ArrayBufferDetachKey     Value
	ArrayBufferMaxByteLength JSInt
}

// 25.1.3.1
func AllocateArrayBuffer(agent *Agent, constructor ObjectType, byteLength JSInt, maxByteLength JSInt) (co Completion[ObjectType]) {
	var allocatingResizableBuffer bool
	if maxByteLength != 0 {
		allocatingResizableBuffer = true
	}
	if allocatingResizableBuffer {
		if byteLength > maxByteLength {
			co.err = agent.ThrowException(RangeError, "byteLength > maxByteLength")
			return
		}
	}
	object := OrdinaryCreateFromConstructor(agent, constructor, "%ArrayBuffer.prototype%", nil)
	arrayBuffer := &ArrayBufferObject{
		Object:                object,
		ArrayBufferData:       CreateByteDataBlock(agent, byteLength),
		ArrayBufferByteLength: byteLength,
	}
	co.value = arrayBuffer
	return
}

// 25.1.3.6
// return either an ArrayBufferObject or a throw completion
func CloneArrayBuffer(agent *Agent, srcBuffer *ArrayBufferLike, srcByteOffset JSInt, srcLength JSInt) Completion[ObjectType] {
	realm := agent.CurrentRealm()
	Assert(!IsDetachedBuffer(srcBuffer))
	targetBuffer := AllocateArrayBuffer(agent, realm.Intrinsics.ArrayBufferConstructor, srcLength, 0)
	srcBlock := srcBuffer.Data()
	targetBlock := targetBuffer.Data().(*ArrayBufferObject).ArrayBufferData
	CopyDataBlockBytes(targetBlock, 0, srcBlock, srcByteOffset, srcLength)
	return targetBuffer
}

// GetArrayBufferMaxByteLengthOption
// spec: 25.1.3.7
func GetArrayBufferMaxByteLengthOption(agent *Agent, options Value) (co Completion[JSInt]) {
	if !options.IsObject() {
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

func IsUnclampedIntegerElementType(size JSInt) bool {
	return size == 1 || size == 2 || size == 4 || size == 8
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

// 25.1.3.2
func ArrayBufferByteLength(buffer *ArrayBufferLike, memoryOrder MemoryOrder) JSInt {
	Assert(!IsDetachedBuffer(buffer))
	return buffer.Data().Size()
}

// 25.1.3.8
func IsFixedLengthArrayBuffer(buffer *ArrayBufferLike) bool {
	if buffer.ArrayBuffer != nil {
		return buffer.ArrayBuffer.ArrayBufferMaxByteLength == 0
	}
	return buffer.SharedArrayBuffer.ArrayBufferMaxByteLength == 0
}

func GetRawBytesFromSharedBlock(
	block *DataBlock,
	byteIndex JSInt,
	size JSInt,
	isTypedArray bool,
	order MemoryOrder,
) []byte {
	return block.Slice(byteIndex, byteIndex+size)
}

// 25.1.3.16
func GetValueFromBuffer(
	agent *Agent,
	arrayBuffer *ArrayBufferLike,
	byteIndex JSInt,
	size JSInt,
	isTypedArray bool,
	order MemoryOrder,
) uint64 {
	Assert(!IsDetachedBuffer(arrayBuffer))
	block := arrayBuffer.Data()
	elementSize := size

	var rawValue []byte
	if IsSharedArrayBuffer(arrayBuffer) {
		rawValue = GetRawBytesFromSharedBlock(block, byteIndex, elementSize, isTypedArray, order)
	} else {
		rawValue = block.data[byteIndex : byteIndex+elementSize]
	}

	return RawBytesToNumeric(elementSize, rawValue, agent.IsLittleEndian)
}

func IsSharedArrayBuffer(buffer *ArrayBufferLike) bool {
	return buffer.SharedArrayBuffer != nil
}

// 25.1.3.17
func NumericToRawBytes(value Value, size JSInt, isLittleEndian bool) []byte {
	rawBytes := make([]byte, size)
	n := value.(*NumberValue).Data
	switch size {
	case 1:
		rawBytes[0] = byte(n)
	case 2:
		if isLittleEndian {
			binary.LittleEndian.PutUint16(rawBytes, uint16(n))
		} else {
			binary.BigEndian.PutUint16(rawBytes, uint16(n))
		}
	case 4:
		if isLittleEndian {
			binary.LittleEndian.PutUint32(rawBytes, uint32(n))
		} else {
			binary.BigEndian.PutUint32(rawBytes, uint32(n))
		}
	case 8:
		if isLittleEndian {
			binary.LittleEndian.PutUint64(rawBytes, uint64(n))
		} else {
			binary.BigEndian.PutUint64(rawBytes, uint64(n))
		}
	}
	return rawBytes
}

func SetValueInBuffer(
	agent *Agent,
	arrayBuffer *ArrayBufferLike,
	byteIndex JSInt,
	value Value,
	size JSInt,
	isTypedArray bool,
	order MemoryOrder,
) {
	Assert(!IsDetachedBuffer(arrayBuffer))
	Assert(byteIndex+size <= arrayBuffer.ByteLength())
	block := arrayBuffer.Data()
	elementSize := size
	rawBytes := NumericToRawBytes(value, elementSize, agent.IsLittleEndian)
	block.Set(byteIndex, rawBytes)
}

// 25.1.3.14
func RawBytesToNumeric(size JSInt, rawBytes []byte, isLittleEndian bool) uint64 {
	switch size {
	case 1:
		return uint64(rawBytes[0])
	case 2:
		if isLittleEndian {
			return uint64(binary.LittleEndian.Uint16(rawBytes))
		} else {
			return uint64(binary.BigEndian.Uint16(rawBytes))
		}
	case 4:
		if isLittleEndian {
			return uint64(binary.LittleEndian.Uint32(rawBytes))
		} else {
			return uint64(binary.BigEndian.Uint32(rawBytes))
		}
	case 8:
		if isLittleEndian {
			return binary.LittleEndian.Uint64(rawBytes)
		} else {
			return binary.BigEndian.Uint64(rawBytes)
		}
	}
	panic("unreachable")
}

func NewArrayBufferConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		length := argumentsList[0]
		var options Value = UndefinedValue
		if len(argumentsList) > 1 {
			options = argumentsList[1]
		}
		if newTarget == nil {
			return agent.ThrowTypeError("TypeError")
		}
		byteLength, isAbrupt, rt := ReturnIfAbrupt(ToIndex(agent, length), co)
		if isAbrupt {
			return rt
		}
		requestedMaxByteLength, isAbrupt, rt := ReturnIfAbrupt(GetArrayBufferMaxByteLengthOption(agent, options), co)
		if isAbrupt {
			return rt
		}
		arrayBuffer := AllocateArrayBuffer(agent, newTarget, byteLength, requestedMaxByteLength)
		return arrayBuffer.Data().ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, CMString("ArrayBuffer"), builtinFunctionArgs{
		realm:         realm,
		isConstructor: true,
		prototype:     realm.Intrinsics.FunctionPrototype,
	})

	BindPrototypeAndConstructor(realm.Intrinsics.ArrayBufferPrototype, object)

	object.defineBuiltinAccessor(realm, WellKnownSymbolsSpecies, builtinAccessorParams{
		Getter: func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
			return this
		},
	})
	return object
}

func NewArrayBufferPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "ArrayBufferPrototype")

	var isView BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		arg := arguments[0]
		if !arg.IsObject() {
			return FalseValue
		}
		o := MustGetObject(arg)
		if ObjectIs[*DataView](o) {
			return TrueValue
		}

		return FalseValue
	}
	byteLength := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := RequireInternalSlot[*ArrayBufferLike](this)
		if IsDetachedBuffer(o) {
			return NewNumberValue(0)
		}
		length := o.ByteLength()
		return NewNumberValue(length.ToNumber())
	}
	slice := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		start := arguments[0]
		end := arguments[1]
		o := RequireInternalSlot[*ArrayBufferLike](this)
		if IsDetachedBuffer(o) {
			panic("TypeError")
		}
		length := o.ByteLength()
		relativeStart, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, start), CompletionValue{})
		if isAbrupt {
			return rt
		}
		var first JSInt
		if relativeStart.IsNegInf() {
			first = 0
		} else if relativeStart < 0 {
			first = JSInt(math.Max(float64(length+relativeStart), 0))
		} else {
			first = JSInt(math.Min(float64(relativeStart), float64(length)))
		}

		relativeEnd, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, end), CompletionValue{})
		if isAbrupt {
			return rt
		}
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
		_new := RequireInternalSlot[*ArrayBufferLike](newObject.value.ToValue())
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
		return (_new).ToValue()
	}
	maxByteLength := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
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
	resizable := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := RequireInternalSlot[*ArrayBufferLike](this)
		return NewBooleanValue(IsFixedLengthArrayBuffer(o))
	}
	resize := func(this Value, arguments []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		newByteLength, isAbrupt, rt := ReturnIfAbrupt(ToIndex(agent, arguments[0]), co)
		if isAbrupt {
			return rt
		}
		o := RequireInternalSlot[*ArrayBufferLike](this)
		if IsDetachedBuffer(o) {
			return co.ThrowTypeError(agent, "is detached")
		}
		if newByteLength > o.MaxByteLength() {
			return co.ThrowError(agent, RangeError, "newByteLength > maxByteLength")
		}
		hostHandled := HostResizeArrayBuffer(o, newByteLength)
		if hostHandled == ResizeArrayBufferHandledHandled {
			return UndefinedValue
		}

		// TODO: Resize

		return UndefinedValue
	}

	object.defineBuiltinFunction(realm, CMString("isView"), isView, 1)
	object.defineBuiltinFunction(realm, CMString("slice"), slice, 2)
	object.defineBuiltinFunction(realm, CMString("resize"), resize, 1)

	object.defineBuiltinAccessor(realm, CMString("byteLength"), builtinAccessorParams{
		Getter: byteLength,
	})
	object.defineBuiltinAccessor(realm, CMString("maxByteLength"), builtinAccessorParams{
		Getter: maxByteLength,
	})
	object.defineBuiltinAccessor(realm, CMString("resizable"), builtinAccessorParams{
		Getter: resizable,
	})

	object.defineToStringTag("ArrayBuffer")
	return object
}
