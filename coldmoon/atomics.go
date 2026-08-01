package coldmoon

import (
	"math"
	"math/big"
	"sync"
	"time"
)

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
		switch name {
		case TypedArrayNameInt8, TypedArrayNameUint8,
			TypedArrayNameInt16, TypedArrayNameUint16,
			TypedArrayNameInt32, TypedArrayNameUint32,
			TypedArrayNameBigInt64, TypedArrayNameBigUint64:
		default:
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

type atomicWaitLocation struct {
	block     *DataBlock
	byteIndex JSInt
}

type atomicWaiter struct {
	wake chan struct{}
}

var atomicsWaiters = struct {
	sync.Mutex
	queues map[atomicWaitLocation][]*atomicWaiter
}{queues: make(map[atomicWaitLocation][]*atomicWaiter)}

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
	accessIndex, isAbrupt, rt := ReturnIfAbrupt(ToIndex(agent, requestIndex), co)
	if isAbrupt {
		return rt
	}
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

// AtomicReadModifyWrite
// spec: 25.4.3.17
func AtomicReadModifyWrite(
	agent *Agent,
	typedArrayValue, index, value Value,
	op AtomicOp,
) (co CompletionValue) {
	byteIndexInBuffer, isAbrupt, rt := ReturnIfAbrupt(ValidateAtomicAccessOnIntegerTypedArray(
		agent,
		typedArrayValue,
		index,
		false), co)
	if isAbrupt {
		return rt
	}
	typedArray := MustGetObject(typedArrayValue).(*TypedArrayObject)
	numericValue, isAbrupt, rt := ReturnIfAbrupt(atomicNumericValue(agent, typedArray.ContentType, value), co)
	if isAbrupt {
		return rt
	}

	_, isAbrupt, rt = ReturnIfAbrupt(RevalidateAtomicAccess(agent, typedArray, byteIndexInBuffer), co)
	if isAbrupt {
		return rt
	}

	buffer := typedArray.ViewedArrayBuffer
	elementType := TypedArrayElementType(typedArray)

	return GetModifySetValueInBuffer(agent, buffer, byteIndexInBuffer, elementType, numericValue, op)
}

// 25.1.3.19
func GetModifySetValueInBuffer(
	agent *Agent,
	arrayBuffer *ArrayBufferLike,
	byteIndex JSInt,
	elementType TypedArrayName,
	value Value,
	op AtomicOp,
) (co CompletionValue) {
	Assert(!IsDetachedBuffer(arrayBuffer))
	size := getTypedArraySizeFromName(elementType)
	Assert(arrayBuffer.Data().Size() >= byteIndex+size)

	block := arrayBuffer.Data()
	isLittleEndian := agent.IsLittleEndian
	valueRaw := RawBytesToNumeric(size, NumericToRawBytes(value, elementType, isLittleEndian), isLittleEndian)
	rawBytesRead := block.AtomicModify(byteIndex, size, func(previousBytes []byte) []byte {
		previous := RawBytesToNumeric(size, previousBytes, isLittleEndian)
		var target uint64
		switch op {
		case AtomicOpAdd:
			target = previous + valueRaw
		case AtomicOpAnd:
			target = previous & valueRaw
		case AtomicOpExchange:
			target = valueRaw
		case AtomicOpOr:
			target = previous | valueRaw
		case AtomicOpSub:
			target = previous - valueRaw
		case AtomicOpXor:
			target = previous ^ valueRaw
		default:
			panic("GetModifySetValueInBuffer: unsupported atomic operation")
		}
		if size < 8 {
			target &= (uint64(1) << uint(size*8)) - 1
		}
		return rawUint64ToBytes(target, size, isLittleEndian)
	})
	previous := RawBytesToNumeric(size, rawBytesRead, isLittleEndian)
	co.value = atomicRawValue(elementType, previous)
	return
}

func rawUint64ToBytes(raw uint64, size JSInt, isLittleEndian bool) []byte {
	return rawUint64Bytes(raw, size, isLittleEndian)
}

func atomicRawValue(elementType TypedArrayName, raw uint64) Value {
	switch elementType {
	case TypedArrayNameInt8:
		return NewNumberValue(JSNumber(int8(raw)))
	case TypedArrayNameUint8:
		return NewNumberValue(JSNumber(uint8(raw)))
	case TypedArrayNameInt16:
		return NewNumberValue(JSNumber(int16(raw)))
	case TypedArrayNameUint16:
		return NewNumberValue(JSNumber(uint16(raw)))
	case TypedArrayNameInt32:
		return NewNumberValue(JSNumber(int32(raw)))
	case TypedArrayNameUint32:
		return NewNumberValue(JSNumber(uint32(raw)))
	case TypedArrayNameBigInt64:
		return NewBigIntValue(big.NewInt(int64(raw)))
	case TypedArrayNameBigUint64:
		return NewBigIntValue(new(big.Int).SetUint64(raw))
	default:
		panic("atomicRawValue: non-integer typed array")
	}
}

func atomicNumericValue(agent *Agent, contentType TypedArrayContentType, value Value) (co CompletionValue) {
	if contentType == TypedArrayContentTypeBigInt {
		bigint, isAbrupt, rt := ReturnIfAbrupt(ToBigInt(agent, value), co)
		if isAbrupt {
			return rt
		}
		co.value = bigint
		return
	}
	number, isAbrupt, rt := ReturnIfAbrupt(value.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	if number.IsNaN() {
		co.value = NewNumberValue(0)
	} else if number.IsFinite() {
		co.value = NewNumberValue(number.Truncate())
	} else {
		co.value = number
	}
	return
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

	atomicsAdd := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		typedArray := argumentAt(argumentsList, 0)
		index := argumentAt(argumentsList, 1)
		value := argumentAt(argumentsList, 2)
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpAdd)
	}
	atomicsAnd := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		typedArray := argumentAt(argumentsList, 0)
		index := argumentAt(argumentsList, 1)
		value := argumentAt(argumentsList, 2)
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpAnd)
	}
	atomicsCompareExchange := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		typedArray := argumentAt(argumentsList, 0)
		index := argumentAt(argumentsList, 1)
		expectedValue := argumentAt(argumentsList, 2)
		replacementValue := argumentAt(argumentsList, 3)
		return atomicsCompareExchange(agent, typedArray, index, expectedValue, replacementValue)
	}
	atomicsExchange := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		typedArray := argumentAt(argumentsList, 0)
		index := argumentAt(argumentsList, 1)
		value := argumentAt(argumentsList, 2)
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpExchange)
	}
	atomicsIsLockFree := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		size := argumentAt(argumentsList, 0)
		n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, size), co)
		if isAbrupt {
			return rt
		}
		if n == 1 || n == 2 || n == 4 || n == 8 {
			return TrueValue
		}
		return FalseValue
	}
	atomicsLoad := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		typedArray := argumentAt(argumentsList, 0)
		index := argumentAt(argumentsList, 1)
		byteIndexInBuffer, isAbrupt, rt := ReturnIfAbrupt(
			ValidateAtomicAccessOnIntegerTypedArray(agent, typedArray, index, false),
			co,
		)
		if isAbrupt {
			return rt
		}
		ta := MustGetObject(typedArray).(*TypedArrayObject)
		_, isAbrupt, rt = ReturnIfAbrupt(RevalidateAtomicAccess(agent, ta, byteIndexInBuffer), co)
		if isAbrupt {
			return rt
		}

		buffer := ta.ViewedArrayBuffer
		raw := GetValueFromBuffer(
			agent,
			buffer,
			byteIndexInBuffer,
			getTypedArraySizeFromName(ta.TypedArrayName),
			true,
			SeqCst,
		)
		return atomicRawValue(ta.TypedArrayName, raw)
	}
	atomicsOr := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		typedArray := argumentAt(argumentsList, 0)
		index := argumentAt(argumentsList, 1)
		value := argumentAt(argumentsList, 2)
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpOr)
	}
	atomicsStore := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		typedArray := argumentAt(argumentsList, 0)
		index := argumentAt(argumentsList, 1)
		value := argumentAt(argumentsList, 2)
		return atomicStore(agent, typedArray, index, value)
	}
	atomicsSub := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		typedArray := argumentAt(argumentsList, 0)
		index := argumentAt(argumentsList, 1)
		value := argumentAt(argumentsList, 2)
		return AtomicReadModifyWrite(agent, typedArray, index, value, AtomicOpSub)
	}
	atomicsWait := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return atomicWaitOperation(agent, argumentAt(argumentsList, 0), argumentAt(argumentsList, 1), argumentAt(argumentsList, 2), argumentAt(argumentsList, 3))
	}
	atomicsNotify := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return atomicNotifyOperation(agent, argumentAt(argumentsList, 0), argumentAt(argumentsList, 1), argumentAt(argumentsList, 2))
	}
	atomicsXor := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		typedArray := argumentAt(argumentsList, 0)
		index := argumentAt(argumentsList, 1)
		value := argumentAt(argumentsList, 2)
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

func atomicWaitOperation(agent *Agent, typedArrayValue, index, expectedValue, timeoutValue Value) (co CompletionValue) {
	taRecord, isAbrupt, rt := ReturnIfAbrupt(ValidateIntegerTypedArray(agent, typedArrayValue, true), co)
	if isAbrupt {
		return rt
	}
	typedArray := taRecord.TypedArray
	buffer := typedArray.ViewedArrayBuffer
	if !IsSharedArrayBuffer(buffer) {
		return co.ThrowTypeError(agent, "Atomics.wait requires a SharedArrayBuffer")
	}
	byteIndex, isAbrupt, rt := ReturnIfAbrupt(ValidateAtomicAccess(agent, taRecord, index), co)
	if isAbrupt {
		return rt
	}

	var expected Value
	if typedArray.TypedArrayName == TypedArrayNameBigInt64 {
		value, isAbrupt, rt := ReturnIfAbrupt(ToBigInt64(expectedValue, agent), co)
		if isAbrupt {
			return rt
		}
		expected = NewBigIntValue(big.NewInt(value))
	} else {
		value, isAbrupt, rt := ReturnIfAbrupt(ToInt32(agent, expectedValue), co)
		if isAbrupt {
			return rt
		}
		expected = NewNumberValue(JSNumber(value))
	}
	timeout, isAbrupt, rt := ReturnIfAbrupt(timeoutValue.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}
	timeoutMilliseconds := timeout.Data.ToFloat()
	if math.IsNaN(timeoutMilliseconds) || math.IsInf(timeoutMilliseconds, 1) {
		timeoutMilliseconds = math.Inf(1)
	} else if timeoutMilliseconds < 0 || math.IsInf(timeoutMilliseconds, -1) {
		timeoutMilliseconds = 0
	}

	location := atomicWaitLocation{block: buffer.Data(), byteIndex: byteIndex}
	expectedBytes := NumericToRawBytes(expected, typedArray.TypedArrayName, agent.IsLittleEndian)
	atomicsWaiters.Lock()
	currentBytes := buffer.Data().Slice(byteIndex, byteIndex+JSInt(len(expectedBytes)))
	if !byteListsEqual(currentBytes, expectedBytes) {
		atomicsWaiters.Unlock()
		return NewStringValue("not-equal").ToCompletion()
	}
	if timeoutMilliseconds == 0 {
		atomicsWaiters.Unlock()
		return NewStringValue("timed-out").ToCompletion()
	}
	waiter := &atomicWaiter{wake: make(chan struct{})}
	atomicsWaiters.queues[location] = append(atomicsWaiters.queues[location], waiter)
	atomicsWaiters.Unlock()

	if math.IsInf(timeoutMilliseconds, 1) || timeoutMilliseconds > float64(math.MaxInt64)/float64(time.Millisecond) {
		<-waiter.wake
		return NewStringValue("ok").ToCompletion()
	}
	timer := time.NewTimer(time.Duration(timeoutMilliseconds * float64(time.Millisecond)))
	defer timer.Stop()
	select {
	case <-waiter.wake:
		return NewStringValue("ok").ToCompletion()
	case <-timer.C:
		atomicsWaiters.Lock()
		removed := removeAtomicWaiterLocked(location, waiter)
		atomicsWaiters.Unlock()
		if removed {
			return NewStringValue("timed-out").ToCompletion()
		}
		// A notifier removed this waiter at the same instant as the timer.
		<-waiter.wake
		return NewStringValue("ok").ToCompletion()
	}
}

func atomicNotifyOperation(agent *Agent, typedArrayValue, index, countValue Value) (co CompletionValue) {
	taRecord, isAbrupt, rt := ReturnIfAbrupt(ValidateIntegerTypedArray(agent, typedArrayValue, true), co)
	if isAbrupt {
		return rt
	}
	byteIndex, isAbrupt, rt := ReturnIfAbrupt(ValidateAtomicAccess(agent, taRecord, index), co)
	if isAbrupt {
		return rt
	}
	limit := math.MaxInt
	if !IsUndefinedOrNil(countValue) {
		count, isAbrupt, rt := ReturnIfAbrupt(countValue.ToNumber(agent), co)
		if isAbrupt {
			return rt
		}
		countNumber := count.Data.ToFloat()
		if math.IsNaN(countNumber) || countNumber <= 0 || math.IsInf(countNumber, -1) {
			limit = 0
		} else if !math.IsInf(countNumber, 1) {
			countNumber = math.Trunc(countNumber)
			if countNumber < float64(limit) {
				limit = int(countNumber)
			}
		}
	}
	buffer := taRecord.TypedArray.ViewedArrayBuffer
	if !IsSharedArrayBuffer(buffer) || limit == 0 {
		return NewNumberValue(0).ToCompletion()
	}

	location := atomicWaitLocation{block: buffer.Data(), byteIndex: byteIndex}
	atomicsWaiters.Lock()
	queue := atomicsWaiters.queues[location]
	count := len(queue)
	if count > limit {
		count = limit
	}
	selected := append([]*atomicWaiter(nil), queue[:count]...)
	if count == len(queue) {
		delete(atomicsWaiters.queues, location)
	} else {
		atomicsWaiters.queues[location] = queue[count:]
	}
	for _, waiter := range selected {
		close(waiter.wake)
	}
	atomicsWaiters.Unlock()
	return NewNumberValue(JSNumber(count)).ToCompletion()
}

func removeAtomicWaiterLocked(location atomicWaitLocation, target *atomicWaiter) bool {
	queue := atomicsWaiters.queues[location]
	for index, waiter := range queue {
		if waiter != target {
			continue
		}
		queue = append(queue[:index], queue[index+1:]...)
		if len(queue) == 0 {
			delete(atomicsWaiters.queues, location)
		} else {
			atomicsWaiters.queues[location] = queue
		}
		return true
	}
	return false
}

func byteListsEqual(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

// MARK: - Internal

func atomicsCompareExchange(
	agent *Agent,
	typedArrayValue, index, expectedValue, replacementValue Value,
) (co CompletionValue) {
	byteIndexInBuffer, isAbrupt, rt := ReturnIfAbrupt(ValidateAtomicAccessOnIntegerTypedArray(
		agent,
		typedArrayValue,
		index,
		false), co)
	if isAbrupt {
		return rt
	}
	typedArray := MustGetObject(typedArrayValue).(*TypedArrayObject)
	buffer := typedArray.ViewedArrayBuffer
	block := buffer.Data()
	expected, isAbrupt, rt := ReturnIfAbrupt(atomicNumericValue(agent, typedArray.ContentType, expectedValue), co)
	if isAbrupt {
		return rt
	}
	replacement, isAbrupt, rt := ReturnIfAbrupt(atomicNumericValue(agent, typedArray.ContentType, replacementValue), co)
	if isAbrupt {
		return rt
	}

	_, isAbrupt, rt = ReturnIfAbrupt(RevalidateAtomicAccess(agent, typedArray, byteIndexInBuffer), co)
	if isAbrupt {
		return rt
	}
	isLittleEndian := agent.IsLittleEndian
	elementType := TypedArrayElementType(typedArray)
	size := getTypedArraySizeFromName(elementType)

	expectedBytes := NumericToRawBytes(expected, elementType, isLittleEndian)
	replacementBytes := NumericToRawBytes(replacement, elementType, isLittleEndian)

	rawBytesRead := block.AtomicCompareExchange(byteIndexInBuffer, expectedBytes, replacementBytes)
	previous := RawBytesToNumeric(size, rawBytesRead, isLittleEndian)
	co.value = atomicRawValue(elementType, previous)
	return
}

func atomicStore(
	agent *Agent,
	typedArrayValue, index, value Value,
) (co CompletionValue) {
	byteIndexInBuffer, isAbrupt, rt := ReturnIfAbrupt(ValidateAtomicAccessOnIntegerTypedArray(
		agent,
		typedArrayValue,
		index,
		false), co)
	if isAbrupt {
		return rt
	}
	typedArray := MustGetObject(typedArrayValue).(*TypedArrayObject)
	v, isAbrupt, rt := ReturnIfAbrupt(atomicNumericValue(agent, typedArray.ContentType, value), co)
	if isAbrupt {
		return rt
	}
	_, isAbrupt, rt = ReturnIfAbrupt(RevalidateAtomicAccess(agent, typedArray, byteIndexInBuffer), co)
	if isAbrupt {
		return rt
	}
	elementType := TypedArrayElementType(typedArray)
	buffer := typedArray.ViewedArrayBuffer

	SetTypedArrayValueInBuffer(
		agent,
		buffer,
		byteIndexInBuffer,
		v,
		elementType,
		SeqCst,
	)
	co.value = v
	return
}
