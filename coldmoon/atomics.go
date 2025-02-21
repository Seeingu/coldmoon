package coldmoon

// 25.4.3.1
func ValidateIntegerTypedArray(agent *Agent, typedArray Value, waitable bool) (c Completion[*TypedArrayWithBufferWitnessRecord]) {
	taRecord := ValidateTypedArray(agent, typedArray, Relaxed)
	ta := taRecord.TypedArray
	name := ta.TypedArrayName
	if waitable {
		if name != TypedArrayNameInt32 && name != TypedArrayNameBigInt64 {
			c.err = agent.ThrowTypeError("waitable typed array")
			return
		}
	} else {
		size := getTypedArraySizeFromName(name)
		if !IsUnclampedIntegerElementType(size) &&
			!IsBigIntElementType(size) {
			c.err = agent.ThrowTypeError("non-waitable typed array")
			return
		}
	}
	c.value = taRecord
	return
}

type AtomicOp int

const (
	AtomicOpAdd AtomicOp = iota
	AtomicOpAnd
	AtomicOpCompareExchange
	AtomicOpExchange
	AtomicOpIsLockFree
	AtomicOpLoad
	AtomicOpOr
	AtomicOpStore
	AtomicOpSub
	AtomicOpWait
	AtomicOpNotify
	AtomicOpXor
)

// 25.4.3.3
func ValidateAtomicAccessOnIntegerTypedArray(
	agent *Agent,
	typedArrayValue, requestedIndex Value,
	waitable bool,
) (co Completion[JSInt]) {
	taRecord := ValidateIntegerTypedArray(agent, typedArrayValue, waitable)
	if taRecord.IsError() {
		co.err = taRecord.Error()
		return
	}
	return ValidateAtomicAccess(agent, taRecord.Data(), requestedIndex)
}

func ValidateAtomicAccess(agent *Agent, taRecord *TypedArrayWithBufferWitnessRecord, requestIndex Value) (co Completion[JSInt]) {
	length := TypedArrayLength(taRecord)
	accessIndex := ToIndex(agent, requestIndex)
	if accessIndex >= length {
		co.err = agent.ThrowRangeError("out of range")
		return
	}

	typedArray := taRecord.TypedArray
	elementSize := TypedArrayElementSize(typedArray)
	offset := typedArray.ByteOffset
	co.value = (accessIndex * elementSize) + offset
	return
}

// 25.4.3.17
func AtomicReadModifyWrite(
	agent *Agent,
	typedArrayValue, index, value Value,
	op AtomicOp,
) Value {
	byteIndexInBuffer := ValidateAtomicAccessOnIntegerTypedArray(
		agent,
		typedArrayValue,
		index,
		false)
	if byteIndexInBuffer.IsError() {
		return byteIndexInBuffer.Error()
	}
	typedArray := MustGetObject(typedArrayValue).(*TypedArrayObject)
	var numericValue Value
	if typedArray.ContentType == TypedArrayContentTypeBigInt {
		numericValue = ToBigInt(agent, value)
	} else {
		numericValue = ToIntegerOrInfinity(agent, value).ToValue()
	}

	RevalidateAtomicAccess(agent, typedArray, byteIndexInBuffer.Data())

	buffer := typedArray.ViewedArrayBuffer
	elementType := TypedArrayElementType(typedArray)

	return GetModifySetValueInBuffer(agent, buffer, byteIndexInBuffer.Data(), elementType, numericValue, op)
}

// 25.1.3.19
func GetModifySetValueInBuffer(
	agent *Agent,
	arrayBuffer *ArrayBufferLike,
	byteIndex JSInt,
	elementType TypedArrayName,
	value Value,
	op AtomicOp,
) Value {
	Assert(!IsDetachedBuffer(arrayBuffer))
	size := getTypedArraySizeFromName(elementType)
	Assert(arrayBuffer.Data().Size() >= byteIndex+size)

	block := arrayBuffer.Data()
	isLittleEndian := agent.IsLittleEndian
	_ = NumericToRawBytes(value, size, isLittleEndian)
	var rawBytesRead []byte
	// TODO: atom calculation
	if IsSharedArrayBuffer(arrayBuffer) {
		rawBytesRead = GetRawBytesFromSharedBlock(block, byteIndex, size, false, Relaxed)
	} else {
		rawBytesRead = block.Slice(byteIndex, byteIndex+size)
	}
	previous := RawBytesToNumeric(size, rawBytesRead, isLittleEndian)
	v := uint64(value.(*NumberValue).Data)

	// TODO: Lock
	var target uint64
	switch op {
	case AtomicOpAdd:
		target = previous + v
	case AtomicOpAnd:
		target = previous & v
	case AtomicOpExchange:
		target = v
	case AtomicOpOr:
		target = previous | v
	case AtomicOpSub:
		target = previous - v
	case AtomicOpXor:
		target = previous ^ v
	default:
		panic("unimplemented")
	}
	block.Set(byteIndex, NumericToRawBytes(JSNumber(target).ToValue(), size, isLittleEndian))
	// return previous
	return JSNumber(previous).ToValue()
}

// 25.4.3.4
func RevalidateAtomicAccess(agent *Agent, typedArray *TypedArrayObject, byteIndexInBuffer JSInt) (co Completion[any]) {
	taRecord := MakeTypedArrayWithBufferWitnessRecord(typedArray, SeqCst)
	if IsTypedArrayOutOfBounds(taRecord) {
		co.err = agent.ThrowRangeError("out of range")
		return
	}
	Assert(byteIndexInBuffer >= typedArray.ByteOffset)
	if byteIndexInBuffer >= taRecord.CachedBufferByteLength.Value {
		co.err = agent.ThrowRangeError("invalid index for typed array")
		return
	}

	return
}

func NewAtomics(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "Atomics")

	atomicsAdd := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		typedArray := argumentsList[0]
		index := argumentsList[1]
		value := argumentsList[2]
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpAdd)
	}
	atomicsAnd := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		typedArray := argumentsList[0]
		index := argumentsList[1]
		value := argumentsList[2]
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpAnd)
	}
	atomicsCompareExchange := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		typedArray := argumentsList[0]
		index := argumentsList[1]
		expectedValue := argumentsList[2]
		replacementValue := argumentsList[3]
		return atomicsCompareExchange(agent, typedArray, index, expectedValue, replacementValue)
	}
	atomicsExchange := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		typedArray := argumentsList[0]
		index := argumentsList[1]
		value := argumentsList[2]
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpExchange)
	}
	atomicsIsLockFree := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		size := argumentsList[0]
		n := ToIntegerOrInfinity(agent, size)
		if n == 1 || n == 2 || n == 4 || n == 8 {
			return TrueValue
		}
		return FalseValue
	}
	atomicsLoad := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		typedArray := argumentsList[0]
		index := argumentsList[1]
		byteIndexInBuffer := ValidateAtomicAccessOnIntegerTypedArray(
			agent,
			typedArray,
			index,
			false)
		ta := MustGetObject(typedArray).(*TypedArrayObject)
		RevalidateAtomicAccess(agent, ta, byteIndexInBuffer.Data())

		buffer := ta.ViewedArrayBuffer
		v := GetValueFromBuffer(
			agent,
			buffer,
			byteIndexInBuffer.Data(),
			getTypedArraySizeFromName(ta.TypedArrayName),
			true,
			SeqCst,
		)
		return JSNumber(v).ToValue()
	}
	atomicsOr := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		typedArray := argumentsList[0]
		index := argumentsList[1]
		value := argumentsList[2]
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpOr)
	}
	atomicsStore := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		typedArray := argumentsList[0]
		index := argumentsList[1]
		value := argumentsList[2]
		return atomicStore(agent, typedArray, index, value)
	}
	atomicsSub := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		typedArray := argumentsList[0]
		index := argumentsList[1]
		value := argumentsList[2]
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpSub)
	}
	atomicsWait := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		panic("not implemented")
	}
	atomicsNotify := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		panic("not implemented")
	}
	atomicsXor := func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		typedArray := argumentsList[0]
		index := argumentsList[1]
		value := argumentsList[2]
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpXor)
	}
	object.defineBuiltinFunction(realm, CMString("add"), atomicsAdd, 3)
	object.defineBuiltinFunction(realm, CMString("and"), atomicsAnd, 3)
	object.defineBuiltinFunction(realm, CMString("compareExchange"), atomicsCompareExchange, 4)
	object.defineBuiltinFunction(realm, CMString("exchange"), atomicsExchange, 3)
	object.defineBuiltinFunction(realm, CMString("isLockFree"), atomicsIsLockFree, 1)
	object.defineBuiltinFunction(realm, CMString("load"), atomicsLoad, 2)
	object.defineBuiltinFunction(realm, CMString("or"), atomicsOr, 3)
	object.defineBuiltinFunction(realm, CMString("store"), atomicsStore, 3)
	object.defineBuiltinFunction(realm, CMString("sub"), atomicsSub, 3)
	object.defineBuiltinFunction(realm, CMString("wait"), atomicsWait, 4)
	object.defineBuiltinFunction(realm, CMString("notify"), atomicsNotify, 3)
	object.defineBuiltinFunction(realm, CMString("xor"), atomicsXor, 3)

	object.defineToStringTag("Atomics")
	return object
}

// MARK: - Internal

func atomicsCompareExchange(
	agent *Agent,
	typedArrayValue, index, expectedValue, replacementValue Value,
) Value {
	byteIndexInBuffer := ValidateAtomicAccessOnIntegerTypedArray(
		agent,
		typedArrayValue,
		index,
		false)
	if byteIndexInBuffer.IsError() {
		return byteIndexInBuffer.Error()
	}
	typedArray := MustGetObject(typedArrayValue).(*TypedArrayObject)
	buffer := typedArray.ViewedArrayBuffer
	block := buffer.Data()
	var expected Value
	var replacement Value
	if typedArray.ContentType == TypedArrayContentTypeBigInt {
		expected = ToBigInt(agent, expectedValue)
		replacement = ToBigInt(agent, replacementValue)
	} else {
		expected = ToIntegerOrInfinity(agent, expectedValue).ToValue()
		replacement = ToIntegerOrInfinity(agent, replacementValue).ToValue()
	}

	RevalidateAtomicAccess(agent, typedArray, byteIndexInBuffer.Data())
	isLittleEndian := agent.IsLittleEndian
	elementType := TypedArrayElementType(typedArray)
	size := getTypedArraySizeFromName(elementType)

	expectedBytes := NumericToRawBytes(expected, size, isLittleEndian)
	replacementBytes := NumericToRawBytes(replacement, size, isLittleEndian)

	var rawBytesRead []byte
	if IsSharedArrayBuffer(buffer) {
		rawBytesRead = GetRawBytesFromSharedBlock(block, byteIndexInBuffer.Data(), size, false, Relaxed)
	} else {
		rawBytesRead = block.Slice(byteIndexInBuffer.Data(), byteIndexInBuffer.Data()+size)
	}
	previous := RawBytesToNumeric(size, rawBytesRead, isLittleEndian)
	expectedUint := RawBytesToNumeric(size, expectedBytes, isLittleEndian)
	if previous == expectedUint {
		block.Set(byteIndexInBuffer.Data(), replacementBytes)
	}

	return JSNumber(previous).ToValue()
}

func atomicStore(
	agent *Agent,
	typedArrayValue, index, value Value,
) Value {
	byteIndexInBuffer := ValidateAtomicAccessOnIntegerTypedArray(
		agent,
		typedArrayValue,
		index,
		false)
	if byteIndexInBuffer.IsError() {
		return byteIndexInBuffer.Error()
	}
	typedArray := MustGetObject(typedArrayValue).(*TypedArrayObject)
	var v Value
	if typedArray.ContentType == TypedArrayContentTypeBigInt {
		v = ToBigInt(agent, value)
	} else {
		v = ToIntegerOrInfinity(agent, value).ToValue()
	}
	RevalidateAtomicAccess(agent, typedArray, byteIndexInBuffer.Data())
	elementType := TypedArrayElementType(typedArray)
	size := getTypedArraySizeFromName(elementType)
	buffer := typedArray.ViewedArrayBuffer

	SetValueInBuffer(
		agent,
		buffer,
		byteIndexInBuffer.Data(),
		v,
		size,
		true,
		SeqCst,
	)
	return v
}
