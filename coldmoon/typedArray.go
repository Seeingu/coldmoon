package coldmoon

import (
	"github.com/Seeingu/coldmoon/pkg"
)

type TypedArrayContentType int

const (
	TypedArrayContentTypeNumber TypedArrayContentType = iota
	TypedArrayContentTypeBigInt
)

type TypedArrayName int

func (t TypedArrayName) String() string {
	switch t {
	case TypedArrayNameInt8:
		return "Int8Array"
	case TypedArrayNameUint8:
		return "Uint8Array"
	case TypedArrayNameUint8Clamped:
		return "Uint8ClampedArray"
	case TypedArrayNameInt16:
		return "Int16Array"
	case TypedArrayNameUint16:
		return "Uint16Array"
	case TypedArrayNameInt32:
		return "Int32Array"
	case TypedArrayNameUint32:
		return "Uint32Array"
	case TypedArrayNameBigInt64:
		return "BigInt64Array"
	case TypedArrayNameBigUint64:
		return "BigUint64Array"
	case TypedArrayNameFloat32:
		return "Float32Array"
	case TypedArrayNameFloat64:
		return "Float64Array"
	}
	panic("invalid TypedArrayName")
}

func (t TypedArrayName) ToIntrinsicName() IntrinsicName {
	switch t {
	case TypedArrayNameInt8:
		return IntrinsicNameInt8Array
	case TypedArrayNameUint8:
		return IntrinsicNameUint8Array
	case TypedArrayNameUint8Clamped:
		return IntrinsicNameUint8ClampedArray
	case TypedArrayNameInt16:
		return IntrinsicNameInt16Array
	case TypedArrayNameUint16:
		return IntrinsicNameUint16Array
	case TypedArrayNameInt32:
		return IntrinsicNameInt32Array
	case TypedArrayNameUint32:
		return IntrinsicNameUint32Array
	case TypedArrayNameBigInt64:
		return IntrinsicNameBigInt64Array
	case TypedArrayNameBigUint64:
		return IntrinsicNameBigUint64Array
	case TypedArrayNameFloat32:
		return IntrinsicNameFloat32Array
	case TypedArrayNameFloat64:
		return IntrinsicNameFloat64Array
	}
	panic("invalid TypedArrayName")
}

const (
	TypedArrayNameInt8 TypedArrayName = iota
	TypedArrayNameUint8
	TypedArrayNameUint8Clamped
	TypedArrayNameInt16
	TypedArrayNameUint16
	TypedArrayNameInt32
	TypedArrayNameUint32
	TypedArrayNameBigInt64
	TypedArrayNameBigUint64
	TypedArrayNameFloat32
	TypedArrayNameFloat64
)

type TypedArrayObject struct {
	*Object
	// [[TypedArrayName]]
	TypedArrayName TypedArrayName
	// [[ContentType]]
	ContentType TypedArrayContentType
	// [[ViewedArrayBuffer]]
	ViewedArrayBuffer *ArrayBufferLike
	// [[ByteLength]]
	ByteLength *ByteLength
	// [[ByteOffset]]
	ByteOffset JSInt
	// [[ArrayLength]]
	ArrayLength *ByteLength
}

func NewTypedArrayPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.TypedArrayPrototype, "TypedArrayPrototype")
	typedArray := &TypedArrayObject{
		Object: object,
	}
	typedArray.ref = typedArray
	taAt := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		taRecord := ValidateTypedArray(agent, O, SeqCst)
		length := TypedArrayLength(taRecord)
		relativeIndex := ToIntegerOrInfinity(agent, arguments[0])
		var k JSInt
		if relativeIndex >= 0 {
			k = relativeIndex
		} else {
			k = length + relativeIndex
		}
		if k < 0 || k >= length {
			return UndefinedValue
		}
		return MustGetObject(O).Get(NewIntegerIndexPropertyKey(k))
	}
	taBuffer := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		ta := RequireInternalSlot[*TypedArrayObject](O)
		buffer := ta.ViewedArrayBuffer
		return buffer.ToValue()
	}
	taByteLength := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		ta := RequireInternalSlot[*TypedArrayObject](O)
		taRecord := MakeTypedArrayWithBufferWitnessRecord(ta, SeqCst)
		size := TypedArrayByteLength(taRecord)
		return size.ToValue()
	}
	taByteOffset := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		ta := RequireInternalSlot[*TypedArrayObject](O)
		taRecord := MakeTypedArrayWithBufferWitnessRecord(ta, SeqCst)
		if IsTypedArrayOutOfBounds(taRecord) {
			return NewNumberValue(0)
		}
		offset := ta.ByteOffset
		return offset.ToValue()
	}
	taCopyWithin := func(this Value, arguments []Value, newTarget ObjectType) Value {
		target := arguments[0]
		start := arguments[1]
		end := pkg.SliceSafeGet(arguments, 2)
		return typedArrayCopyWith(agent, this, target, start, end)
	}
	taEntries := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		ta := ValidateTypedArray(agent, O, SeqCst)
		return CreateArrayIterator(agent, ta.TypedArray, objectOwnPropertiesKindKeyAndValue).ToValue()
	}
	taEvery := func(this Value, arguments []Value, newTarget ObjectType) Value {
		callback := arguments[0]
		thisArg := pkg.SliceSafeGet(arguments, 1)
		return typedArrayEvery(agent, this, callback, thisArg)
	}
	taFill := func(this Value, arguments []Value, newTarget ObjectType) Value {
		value := arguments[0]
		start := pkg.SliceSafeGet(arguments, 1)
		end := pkg.SliceSafeGet(arguments, 2)
		return typedArrayFill(agent, this, value, start, end)
	}
	taFilter := func(this Value, arguments []Value, newTarget ObjectType) Value {
		callback := arguments[0]
		thisArg := pkg.SliceSafeGet(arguments, 1)
		return typedArrayFilter(agent, this, callback, thisArg)
	}
	taFind := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		predicate := arguments[0]
		thisArg := pkg.SliceSafeGet(arguments, 1)
		taRecord := ValidateTypedArray(agent, this, SeqCst)
		length := TypedArrayLength(taRecord)
		findRec := MustGetObject(O).FindViaPredicate(length, DirectionAscending, predicate, thisArg)
		return findRec.Value
	}
	taFindIndex := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		predicate := arguments[0]
		thisArg := pkg.SliceSafeGet(arguments, 1)
		taRecord := ValidateTypedArray(agent, O, SeqCst)
		length := TypedArrayLength(taRecord)
		findRec := MustGetObject(O).FindViaPredicate(length, DirectionAscending, predicate, thisArg)
		return findRec.Index.ToValue()
	}
	taFindLast := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		predicate := arguments[0]
		thisArg := pkg.SliceSafeGet(arguments, 1)
		taRecord := ValidateTypedArray(agent, O, SeqCst)
		length := TypedArrayLength(taRecord)
		findRec := MustGetObject(O).FindViaPredicate(length, DirectionDescending, predicate, thisArg)
		return findRec.Value
	}
	taFindLastIndex := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		predicate := arguments[0]
		thisArg := pkg.SliceSafeGet(arguments, 1)
		taRecord := ValidateTypedArray(agent, O, SeqCst)
		length := TypedArrayLength(taRecord)
		findRec := MustGetObject(O).FindViaPredicate(length, DirectionDescending, predicate, thisArg)
		return findRec.Index.ToValue()
	}
	taForEach := func(this Value, arguments []Value, newTarget ObjectType) Value {
		callback := arguments[0]
		thisArg := pkg.SliceSafeGet(arguments, 1)
		return typedArrayForEach(agent, this, callback, thisArg)
	}
	taIncludes := func(this Value, arguments []Value, newTarget ObjectType) Value {
		searchElement := arguments[0]
		fromIndex := pkg.SliceSafeGet(arguments, 1)
		return typedArrayIncludes(agent, this, searchElement, fromIndex)
	}
	taIndexOf := func(this Value, arguments []Value, newTarget ObjectType) Value {
		searchElement := arguments[0]
		fromIndex := pkg.SliceSafeGet(arguments, 1)
		return typedArrayIndexOf(agent, this, searchElement, fromIndex)
	}
	taJoin := func(this Value, arguments []Value, newTarget ObjectType) Value {
		separator := pkg.SliceSafeGet(arguments, 0)
		return typedArrayJoin(agent, this, separator)
	}
	taKeys := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		taRecord := ValidateTypedArray(agent, O, SeqCst)
		return CreateArrayIterator(agent, taRecord.TypedArray, objectOwnPropertiesKindKey).ToValue()
	}
	taLastIndexOf := func(this Value, arguments []Value, newTarget ObjectType) Value {
		searchElement := arguments[0]
		fromIndex := pkg.SliceSafeGet(arguments, 1)
		return typedArrayLastIndexOf(agent, this, searchElement, fromIndex)
	}
	taLength := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		taRecord := ValidateTypedArray(agent, O, SeqCst)
		if IsTypedArrayOutOfBounds(taRecord) {
			return NewNumberValue(0)
		}
		length := TypedArrayLength(taRecord)
		return length.ToValue()
	}
	taMap := func(this Value, arguments []Value, newTarget ObjectType) Value {
		callback := arguments[0]
		thisArg := pkg.SliceSafeGet(arguments, 1)
		return typedArrayMap(agent, this, callback, thisArg)
	}
	taReduce := func(this Value, arguments []Value, newTarget ObjectType) Value {
		callback := arguments[0]
		initialValue := pkg.SliceSafeGet(arguments, 1)
		return typedArrayReduce(agent, this, callback, initialValue)
	}
	taReduceRight := func(this Value, arguments []Value, newTarget ObjectType) Value {
		callback := arguments[0]
		initialValue := pkg.SliceSafeGet(arguments, 1)
		return typedArrayReduceRight(agent, this, callback, initialValue)
	}
	taReverse := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := MustGetObject(this)
		taRecord := ValidateTypedArray(agent, O.ToValue(), SeqCst)
		length := TypedArrayLength(taRecord)
		middle := length / 2
		lower := JSInt(0)
		for lower != middle {
			upper := length - lower - 1
			lowerP := NewIntegerIndexPropertyKey(lower)
			upperP := NewIntegerIndexPropertyKey(upper)
			lowerValue := O.Get(lowerP)
			upperValue := O.Get(upperP)
			O.Set(lowerP, upperValue, setThrowTypeThrow)
			O.Set(upperP, lowerValue, setThrowTypeThrow)
			lower++
		}
		return O.ToValue()
	}
	taToReversed := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := MustGetObject(this)
		taRecord := ValidateTypedArray(agent, O.ToValue(), SeqCst)
		length := TypedArrayLength(taRecord)
		A := TypedArrayCreateSameType(agent, taRecord.TypedArray, []Value{NewNumberValue(length.ToNumber())})
		k := JSInt(0)
		for k < length {
			Pk := NewIntegerIndexPropertyKey(k)
			kValue := O.Get(Pk)
			A.Set(Pk, kValue, setThrowTypeThrow)
			k++
		}
		return A.ToValue()
	}
	taSet := func(this Value, arguments []Value, newTarget ObjectType) Value {
		source := arguments[0]
		offset := pkg.SliceSafeGet(arguments, 1)
		return typedArraySet(agent, this, source, offset)
	}
	taSlice := func(this Value, arguments []Value, newTarget ObjectType) Value {
		start := arguments[0]
		end := pkg.SliceSafeGet(arguments, 1)
		return typedArraySlice(agent, this, start, end)
	}
	taSome := func(this Value, arguments []Value, newTarget ObjectType) Value {
		callback := arguments[0]
		thisArg := pkg.SliceSafeGet(arguments, 1)
		return typedArraySome(agent, this, callback, thisArg)
	}
	taSort := func(this Value, arguments []Value, newTarget ObjectType) Value {
		compareFn := pkg.SliceSafeGet(arguments, 0)
		return typedArraySort(agent, this, compareFn)
	}
	taToSorted := func(this Value, arguments []Value, newTarget ObjectType) Value {
		compareFn := arguments[0]
		return typedArrayToSorted(agent, this, compareFn)
	}
	taSubarray := func(this Value, arguments []Value, newTarget ObjectType) Value {
		start := arguments[0]
		end := pkg.SliceSafeGet(arguments, 1)
		return typedArraySubarray(agent, this, start, end)
	}
	taToLocaleString := func(this Value, arguments []Value, newTarget ObjectType) Value {
		panic("not implemented")
	}
	taValues := func(this Value, arguments []Value, newTarget ObjectType) Value {
		O := this
		taRecord := ValidateTypedArray(agent, O, SeqCst)
		return CreateArrayIterator(agent, taRecord.TypedArray, objectOwnPropertiesKindValue).ToValue()
	}
	taWith := func(this Value, arguments []Value, newTarget ObjectType) Value {
		index := arguments[0]
		value := arguments[1]
		return typedArrayWith(agent, this, index, value)
	}
	typedArray.defineBuiltinFunction(realm, CMString("at"), taAt, 1)
	typedArray.defineBuiltinFunction(realm, CMString("buffer"), taBuffer, 0)
	typedArray.defineBuiltinFunction(realm, CMString("byteLength"), taByteLength, 0)
	typedArray.defineBuiltinFunction(realm, CMString("byteOffset"), taByteOffset, 0)
	typedArray.defineBuiltinFunction(realm, CMString("copyWithin"), taCopyWithin, 2)
	typedArray.defineBuiltinFunction(realm, CMString("entries"), taEntries, 0)
	typedArray.defineBuiltinFunction(realm, CMString("every"), taEvery, 1)
	typedArray.defineBuiltinFunction(realm, CMString("fill"), taFill, 1)
	typedArray.defineBuiltinFunction(realm, CMString("filter"), taFilter, 1)
	typedArray.defineBuiltinFunction(realm, CMString("find"), taFind, 1)
	typedArray.defineBuiltinFunction(realm, CMString("findIndex"), taFindIndex, 1)
	typedArray.defineBuiltinFunction(realm, CMString("findLast"), taFindLast, 1)
	typedArray.defineBuiltinFunction(realm, CMString("findLastIndex"), taFindLastIndex, 1)
	typedArray.defineBuiltinFunction(realm, CMString("forEach"), taForEach, 1)
	typedArray.defineBuiltinFunction(realm, CMString("includes"), taIncludes, 1)
	typedArray.defineBuiltinFunction(realm, CMString("indexOf"), taIndexOf, 1)
	typedArray.defineBuiltinFunction(realm, CMString("join"), taJoin, 1)
	typedArray.defineBuiltinFunction(realm, CMString("keys"), taKeys, 0)
	typedArray.defineBuiltinFunction(realm, CMString("lastIndexOf"), taLastIndexOf, 1)
	typedArray.defineBuiltinFunction(realm, CMString("length"), taLength, 0)
	typedArray.defineBuiltinFunction(realm, CMString("map"), taMap, 1)
	typedArray.defineBuiltinFunction(realm, CMString("reduce"), taReduce, 1)
	typedArray.defineBuiltinFunction(realm, CMString("reduceRight"), taReduceRight, 1)
	typedArray.defineBuiltinFunction(realm, CMString("reverse"), taReverse, 0)
	typedArray.defineBuiltinFunction(realm, CMString("toReversed"), taToReversed, 0)
	typedArray.defineBuiltinFunction(realm, CMString("set"), taSet, 1)
	typedArray.defineBuiltinFunction(realm, CMString("slice"), taSlice, 2)
	typedArray.defineBuiltinFunction(realm, CMString("some"), taSome, 1)
	typedArray.defineBuiltinFunction(realm, CMString("sort"), taSort, 1)
	typedArray.defineBuiltinFunction(realm, CMString("toSorted"), taToSorted, 1)
	typedArray.defineBuiltinFunction(realm, CMString("subarray"), taSubarray, 2)
	typedArray.defineBuiltinFunction(realm, CMString("toLocaleString"), taToLocaleString, 0)
	typedArray.defineBuiltinFunction(realm, CMString("values"), taValues, 0)
	typedArray.defineBuiltinFunction(realm, CMString("with"), taWith, 1)

	typedArray.defineBuiltinProperty(CMString("toString"), realm.Intrinsics.ArrayPrototype.ToValue().ToBuiltinPropertyDescriptor())

	toStringTag := func(this Value, arguments []Value, newTarget ObjectType) Value {
		if !this.IsObject() {
			return UndefinedValue
		}
		o := MustGetObject(this)
		if !ObjectIs[*TypedArrayObject](o) {
			return UndefinedValue
		}
		name := o.(*TypedArrayObject).TypedArrayName
		// TODO: should not use internal String() method
		return NewStringValue(name.String())
	}
	object.defineBuiltinAccessor(realm, WellKnownSymbolsToStringTag, builtinAccessorParams{
		Setter: toStringTag,
	})
	return typedArray
}

func NewTypedArrayNamePrototype(realm *Realm, name TypedArrayName) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.TypedArrayPrototype, "TypedArrayNamePrototype "+name.String())

	object.defineBuiltinProperty(CMString("BYTES_PER_ELEMENT"), &PropertyDescriptor{
		Value:        NewNumberValue(getTypedArraySizeFromName(name).ToNumber()),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})
	object.InternalMethods().Get = func(o ObjectType, p PropertyKey, receiver Value) (co CompletionValue) {
		if s, ok := p.(StringPropertyKey); ok {
			if s.Value == "buffer" {
				co.ThrowTypeError(agent, "Method get %TypedArray%.prototype.buffer called on incompatible receiver")
				return
			}
		}
		return InternalGet(o, p, receiver)
	}
	return object
}

func typedArrayFrom(agent *Agent, this Value, source Value, mapper Value, thisArg Value) Value {
	C := this
	if !IsConstructor(C) {
		return agent.ThrowException(TypeError, "is not a constructor")
	}
	var mapping bool
	if mapper == UndefinedValue {
		mapping = false
	} else {
		if !IsCallable(mapper) {
			return agent.ThrowException(TypeError, "is not callable")
		}
		mapping = true
	}

	usingIterator := GetMethod(agent, source, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]))
	if usingIterator != nil {
		values := GetIteratorFromMethod(agent, source, usingIterator).IteratorToList()
		length := JSInt(len(values))
		targetObj := TypedArrayCreateFromConstructor(agent, MustGetObject(C), []Value{NewNumberValue(length.ToNumber())})
		k := JSInt(0)
		for k < length {
			Pk := NewIntegerIndexPropertyKey(k)
			kValue := values[k]
			var mappedValue Value
			if mapping {
				mappedValue = ReturnAssertNormal(
					mapper.Call(thisArg, []Value{kValue, k.ToValue(), source}),
				)
			} else {
				mappedValue = kValue
			}
			targetObj.Set(Pk, mappedValue, setThrowTypeThrow)
			k++
		}
		return targetObj.ToValue()
	}
	arrayLike := MustGetObject(source)
	var co CompletionValue
	length, isAbrupt, rt := ReturnIfAbrupt(arrayLike.LengthOfArrayLike(), co)
	if isAbrupt {
		panic(rt)
	}

	targetObj := TypedArrayCreateFromConstructor(agent, MustGetObject(C), []Value{NewNumberValue(length.ToNumber())})
	k := JSInt(0)
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		kValue := arrayLike.Get(Pk)
		var mappedValue Value
		if mapping {
			mappedValue = ReturnAssertNormal(
				mapper.Call(thisArg, []Value{kValue, k.ToValue(), source}),
			)
		} else {
			mappedValue = kValue
		}
		targetObj.Set(Pk, mappedValue, setThrowTypeThrow)
		k++
	}
	return targetObj.ToValue()
}

// 23.2.3.26.1
func SetTypedArrayFromTypedArray(agent *Agent, target *TypedArrayObject, targetOffset JSInt, source *TypedArrayObject) {
	targetBuffer := target.ViewedArrayBuffer
	targetRecord := MakeTypedArrayWithBufferWitnessRecord(target, SeqCst)
	if IsTypedArrayOutOfBounds(targetRecord) {
		panic("TypeError")
	}
	targetLength := TypedArrayLength(targetRecord)
	srcBuffer := source.ViewedArrayBuffer
	srcRecord := MakeTypedArrayWithBufferWitnessRecord(source, SeqCst)
	if IsTypedArrayOutOfBounds(srcRecord) {
		panic("TypeError")
	}
	srcLength := TypedArrayLength(srcRecord)
	targetType := TypedArrayElementType(target)
	targetElementSize := TypedArrayElementSize(target)
	targetByteOffset := target.ByteOffset
	srcType := source.TypedArrayName
	srcElementSize := TypedArrayElementSize(source)
	srcByteOffset := source.ByteOffset
	if targetOffset.IsPositiveInf() {
		panic("RangeError")
	}
	if srcLength+targetOffset > targetLength {
		panic("RangeError")
	}
	if target.ContentType != source.ContentType {
		panic("TypeError")
	}

	var sameSharedArrayBuffer bool
	if srcBuffer.SharedArrayBuffer != nil && targetBuffer.SharedArrayBuffer != nil {
		if srcBuffer.SharedArrayBuffer == targetBuffer.SharedArrayBuffer {
			sameSharedArrayBuffer = true
		}
	}

	var srcByteIndex JSInt
	if SameObject(srcBuffer, targetBuffer) || sameSharedArrayBuffer {
		srcByteLength := TypedArrayByteLength(srcRecord)
		buffer := CloneArrayBuffer(agent, srcBuffer, srcByteOffset, srcByteLength)
		srcBuffer = NewArrayBufferLike(buffer.Data())
		srcByteIndex = 0
	} else {
		srcByteIndex = srcByteOffset
	}
	targetByteIndex := targetByteOffset + targetOffset*targetElementSize
	limit := targetByteIndex + srcLength*targetElementSize
	if srcType == targetType {
		for targetByteIndex < limit {
			value := GetValueFromBuffer(agent, srcBuffer, srcByteIndex, srcElementSize, true, Relaxed)
			SetValueInBuffer(agent, targetBuffer, targetByteIndex, JSNumber(value).ToValue(), targetElementSize, true, Relaxed)
			srcByteIndex += srcElementSize
			targetByteIndex += targetElementSize
		}
	} else {
		for targetByteIndex < limit {
			value := GetValueFromBuffer(agent, srcBuffer, srcByteIndex, srcElementSize, true, Relaxed)
			SetValueInBuffer(agent, targetBuffer, targetByteIndex, NewNumberValue(JSNumber(value)), targetElementSize, true, Relaxed)
			srcByteIndex += srcElementSize
			targetByteIndex += targetElementSize
		}
	}
}

// 23.2.3.26.2
func SetTypedArrayFromArrayLike(agent *Agent, target *TypedArrayObject, targetOffset JSInt, source Value) {
	targetRecord := MakeTypedArrayWithBufferWitnessRecord(target, SeqCst)
	if IsTypedArrayOutOfBounds(targetRecord) {
		panic("TypeError")
	}
	targetLength := TypedArrayLength(targetRecord)
	src := source.ToObject(agent).value
	var co CompletionValue
	srcLength, isAbrupt, rt := ReturnIfAbrupt(src.LengthOfArrayLike(), co)
	if isAbrupt {
		panic(rt)
	}

	if targetOffset.IsPositiveInf() {
		panic("RangeError")
	}
	if srcLength+targetOffset > targetLength {
		panic("RangeError")
	}
	k := JSInt(0)
	for k < srcLength {
		Pk := NewIntegerIndexPropertyKey(k)
		value := src.Get(Pk)
		targetIndex := targetOffset + k
		TypedArraySetElement(agent, target, targetIndex, value)
		k++
	}
}

// 23.2.4.1
func TypedArraySpeciesCreate(
	agent *Agent,
	exemplar *TypedArrayObject,
	argumentList []Value,
) *TypedArrayObject {
	realm := agent.CurrentRealm()
	defaultConstructor := realm.Intrinsics.Get(exemplar.TypedArrayName.ToIntrinsicName())
	constructor := exemplar.SpeciesConstructor(defaultConstructor)
	result := TypedArrayCreateFromConstructor(agent, constructor.Data(), argumentList)
	if result.(*TypedArrayObject).ContentType != exemplar.ContentType {
		panic("TypeError")
	}
	return result.(*TypedArrayObject)
}

// 23.2.4.2
func TypedArrayCreateFromConstructor(
	agent *Agent,
	constructor ObjectType,
	argumentList []Value,
) ObjectType {
	newTypedArray := constructor.Construct(argumentList, nil).value
	taRecord := ValidateTypedArray(agent, newTypedArray.ToValue(), SeqCst)

	if len(argumentList) == 1 {
		if n, ok := argumentList[0].(*NumberValue); ok {
			if IsTypedArrayOutOfBounds(taRecord) {
				return agent.ThrowTypeExceptionObject("out of bounds")
			}
			length := TypedArrayLength(taRecord)
			if length.ToNumber() < n.Data {
				return agent.ThrowTypeExceptionObject("out of bounds")
			}
		}
	}
	return newTypedArray
}

// 23.2.4.3
func TypedArrayCreateSameType(agent *Agent, exemplar *TypedArrayObject, argumentList []Value) *TypedArrayObject {
	realm := agent.CurrentRealm()
	constructor := realm.Intrinsics.Get(exemplar.TypedArrayName.ToIntrinsicName())
	result := TypedArrayCreateFromConstructor(agent, constructor, argumentList)
	Assert(result.(*TypedArrayObject).ContentType == exemplar.ContentType)
	return result.(*TypedArrayObject)
}

// 23.2.4.4
func ValidateTypedArray(agent *Agent, O Value, order MemoryOrder) *TypedArrayWithBufferWitnessRecord {
	typedArray := RequireInternalSlot[*TypedArrayObject](O)
	taRecord := MakeTypedArrayWithBufferWitnessRecord(typedArray, order)
	if IsTypedArrayOutOfBounds(taRecord) {
		panic("TypeError")
	}
	return taRecord
}

// 23.2.4.7
func CompareTypedArrayElements(agent *Agent, x, y Value, comparator ObjectType) JSNumber {
	xNumber, xIsNumber := x.(*NumberValue)
	yNumber, yIsNumber := y.(*NumberValue)
	xBigInt, xIsBigInt := x.(*BigIntValue)
	yBigInt, yIsBigInt := y.(*BigIntValue)
	Assert((xIsNumber && yIsNumber) || (xIsBigInt && yIsBigInt))
	if comparator != nil {
		v := comparator.Call(UndefinedValue, []Value{x, y}).value.ToNumber(agent)
		if v.IsNaN() {
			return 0
		}
		return v.Data
	}
	if xIsNumber && yIsNumber {
		if xNumber.Data < yNumber.Data {
			return -1
		}
		if xNumber.Data > yNumber.Data {
			return 1
		}
		return 0
	} else {
		xn := xBigInt.Data
		yn := yBigInt.Data
		if xn.Cmp(yn) < 0 {
			return -1
		}
		if xn.Cmp(yn) > 0 {
			return 1
		}
		return 0
	}
}

func NewTypedArrayConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		panic("TypeError")
	}
	object := CreateBuiltinFunction(agent, behavior, 0, CMString("TypedArray"), builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})
	from := func(this Value, arguments []Value, newTarget ObjectType) Value {
		source := arguments[0]
		mapper := pkg.SliceSafeGet(arguments, 1)
		thisArg := pkg.SliceSafeGet(arguments, 2)
		return typedArrayFrom(agent, this, source, mapper, thisArg)
	}
	of := func(this Value, arguments []Value, newTarget ObjectType) Value {
		length := JSInt(len(arguments))
		C := this
		if !IsConstructor(C) {
			return agent.ThrowException(TypeError, "is not a constructor")
		}
		newObj := TypedArrayCreateFromConstructor(agent, MustGetObject(C), arguments)
		k := JSInt(0)
		for k < length {
			Pk := NewIntegerIndexPropertyKey(k)
			kValue := arguments[k]
			newObj.Set(Pk, kValue, setThrowTypeThrow)
			k++
		}
		return newObj.ToValue()
	}

	object.defineBuiltinFunction(realm, CMString("from"), from, 1)
	object.defineBuiltinFunction(realm, CMString("of"), of, 0)
	species := func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return thisArgument
	}
	object.defineBuiltinAccessor(realm, WellKnownSymbolsSpecies, builtinAccessorParams{
		Getter: species,
	})

	BindPrototypeAndConstructor(realm.Intrinsics.TypedArrayPrototype, object)

	return object
}

func NewTypedArrayNameConstructor(realm *Realm, name TypedArrayName) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return typedArrayBehavior(agent, name, thisArgument, argumentsList, newTarget)
	}
	// TODO: should not use String()
	object := CreateBuiltinFunction(agent, behavior, 3, CMString(name.String()), builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.TypedArrayConstructor,
		isConstructor: true,
	})

	intrinsicName := IntrinsicName("%" + name.String() + ".prototype%")
	intrinsic := realm.Intrinsics.Get(intrinsicName)
	object.defineBuiltinProperty(CMString("BYTES_PER_ELEMENT"), &PropertyDescriptor{
		Value:        NewNumberValue(getTypedArraySizeFromName(name).ToNumber()),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	BindPrototypeAndConstructor(intrinsic, object)

	return object
}

// 23.2.5.1
func typedArrayBehavior(agent *Agent, name TypedArrayName, thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
	if newTarget == nil {
		panic("TypeError")
	}
	constructorName := name
	proto := IntrinsicNameTypedArrayPrototype
	numberOfArgs := len(argumentsList)
	if numberOfArgs == 0 {
		return (AllocateTypedArray(agent, constructorName, newTarget, proto, 0)).ToValue()
	} else {
		firstArgument := argumentsList[0]
		if firstArgument.IsObject() {
			O := AllocateTypedArray(agent, constructorName, newTarget, proto, 0)
			firstArgumentObj := MustGetObject(firstArgument)
			switch fao := firstArgumentObj.(type) {
			case *TypedArrayObject:
				InitializeTypedArrayFromTypedArray(agent, O, fao)
			case *ArrayBufferLike, *SharedArrayBufferObject:
				byteOffset := pkg.SliceSafeGet(argumentsList, 1)
				length := pkg.SliceSafeGet(argumentsList, 2)
				bl := NewArrayBufferLike(fao)
				InitializeTypedArrayFromArrayBuffer(agent, O, bl, byteOffset, length)
			default:
				usingIterator := GetMethod(agent, firstArgument, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]))
				if usingIterator != nil {
					values := GetIteratorFromMethod(agent, firstArgument, usingIterator).IteratorToList()
					InitializeTypedArrayFromList(agent, O, values)
				} else {
					InitializeTypedArrayFromArrayLike(agent, O, MustGetObject(firstArgument))
				}
			}
			return O.ToValue()
		} else {
			elementLength := ToIndex(agent, firstArgument)
			return AllocateTypedArray(agent, constructorName, newTarget, proto, elementLength).ToValue()
		}
	}
}

func InitializeTypedArrayFromList(agent *Agent, O *TypedArrayObject, values []Value) {
	length := JSInt(len(values))
	AllocateTypedArrayBuffer(agent, O, length)
	k := JSInt(0)
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		kValue := values[k]
		O.Set(Pk, kValue, setThrowTypeThrow)
		k++
	}
}

// 23.2.5.1.1
func AllocateTypedArray(agent *Agent, constructorName TypedArrayName, newTarget ObjectType, defaultProto IntrinsicName, length JSInt) *TypedArrayObject {
	proto := GetPrototypeFromConstructor(newTarget, defaultProto)
	obj := TypedArrayCreate(agent, constructorName, proto)
	if length != 0 {
		AllocateTypedArrayBuffer(agent, obj, length)
	}
	return obj
}

func TypedArrayCreate(agent *Agent, name TypedArrayName, proto ObjectType) *TypedArrayObject {
	object := &TypedArrayObject{
		Object:            NewObject(agent, proto, "TypedArray "+name.String()),
		ViewedArrayBuffer: nil,
		TypedArrayName:    name,
	}
	object.ref = object
	internalMethods := object.InternalMethods()
	internalMethods.PreventExtensions = func(o ObjectType) bool {
		oo := o.(*TypedArrayObject)
		if !oo.IsTypedArrayFixedLength() {
			return false
		}
		return OrdinaryPreventExtensions(oo.Object)
	}
	internalMethods.GetOwnProperty = func(o ObjectType, p PropertyKey) *PropertyDescriptor {
		oo := o.Ref().(*TypedArrayObject)
		if i, err := p.GetIndex(); err == nil {
			value := TypedArrayGetElement(agent, oo, i)
			if IsUndefinedOrNil(value) {
				return nil
			}
			return &PropertyDescriptor{
				Value:        value,
				Writable:     true,
				Enumerable:   true,
				Configurable: true,
			}
		}
		return OrdinaryGetOwnProperty(oo.Object, p)
	}
	internalMethods.HasProperty = func(o ObjectType, p PropertyKey) bool {
		if i, err := p.GetIndex(); err == nil {
			return o.(*TypedArrayObject).IsValidIntegerIndex(agent, i)
		}
		return OrdinaryHasProperty(o.(*TypedArrayObject).Object, p)
	}
	internalMethods.DefineOwnProperty = func(o ObjectType, p PropertyKey, desc *PropertyDescriptor) bool {
		if i, err := p.GetIndex(); err == nil {
			if !o.(*TypedArrayObject).IsValidIntegerIndex(agent, i) {
				return false
			}
			// TODO: Only check for undefined
			//if !desc.Enumerable {
			//	return false
			//}
			//if !desc.Configurable {
			//	return false
			//}
			//if !desc.Writable {
			//	return false
			//}
			//if desc.IsAccessorDescriptor() {
			//	return false
			//}
			if desc.Value != nil {
				TypedArraySetElement(agent, o.(*TypedArrayObject), i, desc.Value)
			}
		}
		return true
	}
	internalMethods.Get = func(o ObjectType, p PropertyKey, receiver Value) CompletionValue {
		if i, err := p.GetIndex(); err == nil {
			return TypedArrayGetElement(agent, o.(*TypedArrayObject), i).ToCompletion()
		}
		return OrdinaryGet(o.(*TypedArrayObject).Object, p, receiver)
	}
	internalMethods.Set = func(o ObjectType, p PropertyKey, v Value, receiver Value) bool {
		if i, err := p.GetIndex(); err == nil {
			TypedArraySetElement(agent, o.(*TypedArrayObject), i, v)
			return true
		}
		return OrdinarySet(o.(*TypedArrayObject).Object, p, v, receiver)
	}
	internalMethods.Delete = func(o ObjectType, p PropertyKey) bool {
		if i, err := p.GetIndex(); err == nil {
			return !o.(*TypedArrayObject).IsValidIntegerIndex(agent, i)
		}
		return OrdinaryDelete(o.(*TypedArrayObject).Object, p)
	}
	internalMethods.OwnPropertyKeys = func(o ObjectType) []PropertyKey {
		oo := o.(*TypedArrayObject)
		keys := OrdinaryOwnPropertyKeys(oo.Object)
		length := oo.ArrayLength.Value
		for k := JSInt(0); k < length; k++ {
			keys = append(keys, NewIntegerIndexPropertyKey(k))
		}
		return keys
	}

	return object
}

// 10.4.5.18
func TypedArraySetElement(agent *Agent, O *TypedArrayObject, index JSInt, value Value) {
	if !O.IsValidIntegerIndex(agent, index) {
		return
	}
	var numValue Value
	if O.ContentType == TypedArrayContentTypeBigInt {
		numValue = ToBigInt(agent, value)
	} else {
		numValue = value.ToNumber(agent)
	}

	offset := O.ByteOffset
	elementSize := TypedArrayElementSize(O)
	byteIndexInBuffer := offset + index*elementSize
	SetValueInBuffer(
		agent,
		O.ViewedArrayBuffer,
		byteIndexInBuffer,
		numValue,
		elementSize,
		true,
		Relaxed,
	)
}

// 10.4.5.14
func (t *TypedArrayObject) IsValidIntegerIndex(agent *Agent, index JSInt) bool {
	if IsDetachedBuffer(t.ViewedArrayBuffer) {
		return false
	}
	if index < 0 {
		return false
	}
	taRecord := MakeTypedArrayWithBufferWitnessRecord(t, Relaxed)
	if IsTypedArrayOutOfBounds(taRecord) {
		return false
	}
	length := TypedArrayLength(taRecord)
	if index >= length {
		return false
	}
	return true
}

// 10.4.5.17
func TypedArrayGetElement(agent *Agent, O *TypedArrayObject, index JSInt) Value {
	if !O.IsValidIntegerIndex(agent, index) {
		return UndefinedValue
	}
	offset := O.ByteOffset
	elementSize := TypedArrayElementSize(O)
	byteIndexInBuffer := offset + index*elementSize
	u := GetValueFromBuffer(agent, O.ViewedArrayBuffer, byteIndexInBuffer, elementSize, true, Relaxed)
	return NewNumberValue(JSNumber(u))
}

func AllocateTypedArrayBuffer(agent *Agent, O *TypedArrayObject, length JSInt) {
	realm := agent.CurrentRealm()
	elementSize := TypedArrayElementSize(O)
	byteLength := length * elementSize
	data := AllocateArrayBuffer(agent, realm.Intrinsics.ArrayBufferConstructor, byteLength, 0)
	O.ViewedArrayBuffer = NewArrayBufferLike(data.Data())
	O.ByteLength = NewByteLength(byteLength)
	O.ByteOffset = 0
	O.ArrayLength = NewByteLength(length)
}

func TypedArrayElementType(O *TypedArrayObject) TypedArrayName {
	return O.TypedArrayName
}

func getTypedArraySizeFromName(name TypedArrayName) JSInt {
	switch name {
	case TypedArrayNameInt8:
		return 1
	case TypedArrayNameUint8:
		return 1
	case TypedArrayNameUint8Clamped:
		return 1
	case TypedArrayNameInt16:
		return 2
	case TypedArrayNameUint16:
		return 2
	case TypedArrayNameInt32:
		return 4
	case TypedArrayNameUint32:
		return 4
	case TypedArrayNameFloat32:
		return 4
	case TypedArrayNameFloat64:
		return 8
	case TypedArrayNameBigInt64:
		return 8
	case TypedArrayNameBigUint64:
		return 8
	}
	panic("unreachable")
}

func TypedArrayElementSize(O *TypedArrayObject) JSInt {
	return getTypedArraySizeFromName(O.TypedArrayName)
}

func InitializeTypedArrayFromTypedArray(agent *Agent, O, srcArray *TypedArrayObject) (co Completion[ObjectType]) {
	realm := agent.CurrentRealm()
	srcData := srcArray.ViewedArrayBuffer
	elementSize := TypedArrayElementSize(O)
	elementType := O.TypedArrayName
	srcElementSize := TypedArrayElementSize(srcArray)
	srcType := srcArray.TypedArrayName
	srcByteOffset := srcArray.ByteOffset
	srcRecord := MakeTypedArrayWithBufferWitnessRecord(srcArray, SeqCst)
	if IsTypedArrayOutOfBounds(srcRecord) {
		co.err = agent.ThrowException(RangeError, "out of bounds")
		return
	}
	elementLength := TypedArrayLength(srcRecord)
	byteLength := elementLength * elementSize
	var data Completion[ObjectType]
	if elementType == srcType {
		data = CloneArrayBuffer(agent, srcData, srcByteOffset, byteLength)
	} else {
		data = AllocateArrayBuffer(agent, realm.Intrinsics.ArrayBufferConstructor, byteLength, 0)
		if srcArray.ContentType != O.ContentType {
			co.err = agent.ThrowException(TypeError, "different content type")
			return
		}
		srcByteIndex := srcByteOffset
		targetByteIndex := JSInt(0)
		count := elementLength
		for count > 0 {
			value := GetValueFromBuffer(agent, srcData, srcByteIndex, getTypedArraySizeFromName(srcType), true, Relaxed)
			SetValueInBuffer(
				agent,
				NewArrayBufferLike(data.Data()),
				targetByteIndex,
				NewNumberValue(JSNumber(value)),
				getTypedArraySizeFromName(srcType),
				true,
				Relaxed)
			srcByteIndex += srcElementSize
			targetByteIndex += elementSize
			count--
		}
	}

	O.ViewedArrayBuffer = NewArrayBufferLike(data.Data())
	O.ByteLength = NewByteLength(byteLength)
	O.ByteOffset = 0
	O.ArrayLength = NewByteLength(elementLength)
	// return UNUSED
	return
}

func InitializeTypedArrayFromArrayBuffer(
	agent *Agent,
	O *TypedArrayObject,
	buffer *ArrayBufferLike, byteOffset, length Value,
) (co Completion[ObjectType]) {
	elementSize := TypedArrayElementSize(O)
	offset := ToIndex(agent, byteOffset)
	if offset%elementSize != 0 {
		co.err = agent.ThrowException(RangeError, "offset is not a multiple of element size")
		return
	}
	bufferIsFixedLength := IsFixedLengthArrayBuffer(buffer)
	var newLength JSInt
	if length == UndefinedValue {
		newLength = ToIndex(agent, length)
	}
	if IsDetachedBuffer(buffer) {
		co.err = agent.ThrowException(TypeError, "detached buffer")
		return
	}
	bufferByteLength := ArrayBufferByteLength(buffer, SeqCst)
	if IsUndefinedOrNil(length) && !bufferIsFixedLength {
		if offset > bufferByteLength {
			co.err = agent.ThrowException(RangeError, "offset > bufferByteLength")
			return
		}
		O.ByteLength.toAuto()
		O.ArrayLength.toAuto()
	} else {
		var byteLength JSInt
		if IsUndefinedOrNil(length) {
			if bufferByteLength%elementSize != 0 {
				co.err = agent.ThrowException(RangeError, "bufferByteLength is not a multiple of element size")
				return
			}
			byteLength = bufferByteLength - offset
		} else {
			newByteLength := newLength * elementSize
			if offset+newByteLength > bufferByteLength {
				co.err = agent.ThrowException(RangeError, "out of bounds")
				return
			}
			byteLength = newByteLength
		}
		O.ByteLength = NewByteLength(byteLength)
		O.ArrayLength = NewByteLength(byteLength / elementSize)
	}
	O.ViewedArrayBuffer = buffer
	O.ByteOffset = offset
	// return UNUSED
	return
}

// MARK: - TypedArrayWithBufferWitnessRecord

type TypedArrayWithBufferWitnessRecord struct {
	// [[Object]]
	TypedArray *TypedArrayObject
	// [[CachedBufferByteLength]]
	CachedBufferByteLength *CachedBufferByteLength
}

// 10.4.5.9
func MakeTypedArrayWithBufferWitnessRecord(O *TypedArrayObject, order MemoryOrder) *TypedArrayWithBufferWitnessRecord {
	buffer := O.ViewedArrayBuffer
	cachedBufferByteLength := &CachedBufferByteLength{}
	if IsDetachedBuffer(buffer) {
		cachedBufferByteLength.Detached = true
	} else {
		cachedBufferByteLength.Value = ArrayBufferByteLength(buffer, order)
	}
	return &TypedArrayWithBufferWitnessRecord{
		TypedArray:             O,
		CachedBufferByteLength: cachedBufferByteLength,
	}
}

// 10.4.5.12
func TypedArrayByteLength(taRecord *TypedArrayWithBufferWitnessRecord) JSInt {
	if IsTypedArrayOutOfBounds(taRecord) {
		return 0
	}
	length := TypedArrayLength(taRecord)
	if length == 0 {
		return 0
	}
	O := taRecord.TypedArray
	if !O.ByteLength.Auto {
		return O.ByteLength.Value
	}
	elementSize := TypedArrayElementSize(O)
	return length * elementSize
}

// 10.4.5.14
func IsTypedArrayOutOfBounds(taRecord *TypedArrayWithBufferWitnessRecord) bool {
	O := taRecord.TypedArray
	bufferByteLength := taRecord.CachedBufferByteLength
	Assert(IsDetachedBuffer(O.ViewedArrayBuffer) == bufferByteLength.Detached)
	if bufferByteLength.Detached {
		return true
	}
	byteOffsetStart := O.ByteOffset
	var byteOffsetEnd JSInt
	if O.ArrayLength.Auto {
		byteOffsetEnd = bufferByteLength.Value
	} else {
		elementSize := TypedArrayElementSize(O)
		byteOffsetEnd = byteOffsetStart + O.ArrayLength.Value*elementSize
	}
	if byteOffsetStart > byteOffsetEnd ||
		byteOffsetEnd > bufferByteLength.Value {
		return true
	}
	return false
}

// 10.4.5.15
func (t *TypedArrayObject) IsTypedArrayFixedLength() bool {
	if t.ArrayLength.Auto {
		return false
	}
	buffer := t.ViewedArrayBuffer
	if !IsFixedLengthArrayBuffer(buffer) {
		return false
	}
	return true
}

// 10.4.5.16

func InitializeTypedArrayFromArrayLike(agent *Agent, O *TypedArrayObject, arrayLike ObjectType) {
	var co CompletionValue
	length, isAbrupt, rt := ReturnIfAbrupt(arrayLike.LengthOfArrayLike(), co)
	if isAbrupt {
		panic(rt)
	}

	AllocateTypedArrayBuffer(agent, O, length)
	k := JSInt(0)
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		kValue := arrayLike.Get(Pk)
		O.Set(Pk, kValue, setThrowTypeThrow)
		k++
	}
}

func TypedArrayLength(taRecord *TypedArrayWithBufferWitnessRecord) JSInt {
	Assert(!IsTypedArrayOutOfBounds(taRecord))
	O := taRecord.TypedArray
	if !O.ArrayLength.Auto {
		return O.ArrayLength.Value
	}
	Assert(!IsFixedLengthArrayBuffer(O.ViewedArrayBuffer))
	byteOffset := O.ByteOffset
	elementSize := TypedArrayElementSize(O)
	byteLength := taRecord.CachedBufferByteLength
	Assert(!byteLength.Detached)
	return (byteLength.Value - byteOffset) / elementSize
}

// MARK: - Internal

func typedArrayCopyWith(agent *Agent, this Value, target, start, end Value) Value {
	O := this

	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	relativeTarget := ToIntegerOrInfinity(agent, target)

	var targetIndex JSInt
	if relativeTarget.IsNegInf() {
		targetIndex = 0
	} else if relativeTarget < 0 {
		targetIndex = (length + relativeTarget).Max(0)
	} else {
		targetIndex = relativeTarget.Min(length)
	}

	relativeStart := ToIntegerOrInfinity(agent, start)
	var startIndex JSInt
	if relativeStart.IsNegInf() {
		startIndex = 0
	} else if relativeStart < 0 {
		startIndex = (length + relativeStart).Max(0)
	} else {
		startIndex = relativeStart.Min(length)
	}

	var relativeEnd JSInt
	if IsUndefinedOrNil(end) {
		relativeEnd = length
	} else {
		relativeEnd = ToIntegerOrInfinity(agent, end)
	}
	var endIndex JSInt
	if relativeEnd.IsNegInf() {
		endIndex = 0
	} else if relativeEnd < 0 {
		endIndex = (length + relativeEnd).Max(0)
	} else {
		endIndex = relativeEnd.Min(length)
	}

	count := (endIndex - startIndex).Min(length - targetIndex)
	if count > 0 {
		buffer := ta.ViewedArrayBuffer
		taRecord = MakeTypedArrayWithBufferWitnessRecord(ta, SeqCst)
		if IsTypedArrayOutOfBounds(taRecord) {
			return agent.ThrowException(RangeError, "out of bounds")
		}
		length = TypedArrayLength(taRecord)
		elementSize := TypedArrayElementSize(ta)
		byteOffset := ta.ByteOffset
		bufferByteLimit := (length * elementSize) + byteOffset
		toByteIndex := (targetIndex * elementSize) + byteOffset
		fromByteIndex := (startIndex * elementSize) + byteOffset
		countBytes := count * elementSize
		var dir JSInt
		if fromByteIndex < toByteIndex &&
			toByteIndex < fromByteIndex+countBytes {
			fromByteIndex += countBytes - 1
			toByteIndex += countBytes - 1
			dir = -1
		} else {
			dir = 1
		}
		for countBytes > 0 {
			if fromByteIndex < bufferByteLimit &&
				toByteIndex < bufferByteLimit {
				value := GetValueFromBuffer(agent, buffer, fromByteIndex, elementSize, true, Relaxed)
				SetValueInBuffer(agent, buffer, toByteIndex, JSNumber(value).ToValue(), elementSize, true, Relaxed)
				fromByteIndex += dir
				toByteIndex += dir
				countBytes--
			} else {
				countBytes = 0
			}
		}
	}
	return ta.ToValue()
}

func typedArrayEvery(agent *Agent, this Value, callback Value, thisArg Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if !IsCallable(callback) {
		return agent.ThrowException(TypeError, "is not callable")
	}
	if length == 0 {
		return TrueValue
	}
	k := JSInt(0)
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		kValue := ta.Get(Pk)
		testResult := ReturnAssertNormal(
			callback.Call(thisArg, []Value{kValue, k.ToValue(), ta.ToValue()}),
		)
		if !testResult.ToBoolean() {
			return FalseValue
		}
		k++
	}
	return TrueValue
}

func typedArrayFill(agent *Agent, this Value, _value, start, end Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)

	var value Value
	if ta.ContentType == TypedArrayContentTypeBigInt {
		value = ToBigInt(agent, _value)
	} else {
		value = _value.ToNumber(agent)
	}

	relativeStart := ToIntegerOrInfinity(agent, start)
	var startIndex JSInt
	if relativeStart.IsNegInf() {
		startIndex = 0
	} else if relativeStart < 0 {
		startIndex = (length + relativeStart).Max(0)
	} else {
		startIndex = relativeStart.Min(length)
	}

	var relativeEnd JSInt
	if IsUndefinedOrNil(end) {
		relativeEnd = length
	} else {
		relativeEnd = ToIntegerOrInfinity(agent, end)
	}
	var endIndex JSInt
	if relativeEnd.IsNegInf() {
		endIndex = 0
	} else if relativeEnd < 0 {
		endIndex = (length + relativeEnd).Max(0)
	} else {
		endIndex = relativeEnd.Min(length)
	}

	taRecord = MakeTypedArrayWithBufferWitnessRecord(ta, SeqCst)
	if IsTypedArrayOutOfBounds(taRecord) {
		return agent.ThrowException(RangeError, "out of bounds")
	}
	length = TypedArrayLength(taRecord)
	endIndex = endIndex.Min(length)
	k := startIndex
	for k < endIndex {
		Pk := NewIntegerIndexPropertyKey(k)
		ta.Set(Pk, value, setThrowTypeThrow)
		k++
	}

	return ta.ToValue()
}

func typedArrayFilter(agent *Agent, this Value, callback Value, thisArg Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if !IsCallable(callback) {
		return agent.ThrowException(TypeError, "is not callable")
	}
	kept := make([]Value, 0)
	captured := JSInt(0)
	k := JSInt(0)
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		kValue := ta.Get(Pk)
		selected := ReturnAssertNormal(
			callback.Call(thisArg, []Value{kValue, k.ToValue(), ta.ToValue()}),
		)
		if selected.ToBoolean() {
			kept = append(kept, kValue)
			captured++
		}
		k++
	}

	A := TypedArraySpeciesCreate(agent, ta, []Value{captured.ToValue()})
	for i, value := range kept {
		Pk := NewIntegerIndexPropertyKey(JSInt(i))
		A.Set(Pk, value, setThrowTypeThrow)
	}

	return A.ToValue()
}

func typedArrayForEach(agent *Agent, this Value, callback Value, thisArg Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if !IsCallable(callback) {
		return agent.ThrowException(TypeError, "is not callable")
	}
	k := JSInt(0)
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		kValue := ta.Get(Pk)
		callback.Call(thisArg, []Value{kValue, k.ToValue(), ta.ToValue()})
		k++
	}
	return UndefinedValue
}

func typedArrayIncludes(agent *Agent, this Value, searchElement, fromIndex Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if length == 0 {
		return FalseValue
	}
	n := ToIntegerOrInfinity(agent, fromIndex)
	if IsUndefinedOrNil(fromIndex) {
		Assert(n == 0)
	}
	if n.IsPositiveInf() {
		return FalseValue
	}
	if n.IsNegInf() {
		n = 0
	}

	var k JSInt
	if n >= 0 {
		k = n
	} else {
		k = (length + n).Max(0)
	}
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		elementK := ta.Get(Pk)
		if SameValueZero(searchElement, elementK) {
			return TrueValue
		}
		k++
	}
	return FalseValue
}

func typedArrayIndexOf(agent *Agent, this Value, searchElement, fromIndex Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if length == 0 {
		return NewNumberValue(-1)
	}
	n := ToIntegerOrInfinity(agent, fromIndex)
	if IsUndefinedOrNil(fromIndex) {
		Assert(n == 0)
	}
	if n.IsPositiveInf() {
		return NewNumberValue(-1)
	}
	if n.IsNegInf() {
		n = 0
	}

	var k JSInt
	if n >= 0 {
		k = n
	} else {
		k = (length + n).Max(0)
	}
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		kPresent := ta.HasProperty(Pk)
		if kPresent {
			elementK := ta.Get(Pk)
			if IsStrictlyEqual(searchElement, elementK) {
				return k.ToValue()
			}
		}
		k++
	}
	return NewNumberValue(-1)
}

func typedArrayJoin(agent *Agent, this Value, separator Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if length == 0 {
		return NewStringValue("")
	}
	sep := ","
	if !IsUndefinedOrNil(separator) {
		sep = ToString(agent, separator).Data
	}
	var R string
	k := JSInt(0)
	for k < length {
		if k > 0 {
			R += sep
		}
		Pk := NewIntegerIndexPropertyKey(k)
		element := ta.Get(Pk)
		if !IsUndefinedOrNil(element) {
			R += ToString(agent, element).Data
		}
		k++
	}
	return NewStringValue(R)
}

func typedArrayLastIndexOf(agent *Agent, this Value, searchElement, fromIndex Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if length == 0 {
		return NewNumberValue(-1)
	}
	var n JSInt
	if IsUndefinedOrNil(fromIndex) {
		n = length - 1
	} else {
		n = ToIntegerOrInfinity(agent, fromIndex)
	}
	var k JSInt
	if n >= 0 {
		k = n.Min(length - 1)
	} else {
		k = length + n
	}
	for k >= 0 {
		Pk := NewIntegerIndexPropertyKey(k)
		kPresent := ta.HasProperty(Pk)
		if kPresent {
			elementK := ta.Get(Pk)
			if IsStrictlyEqual(searchElement, elementK) {
				return k.ToValue()
			}
		}
		k--
	}
	return NewNumberValue(-1)
}

func typedArrayMap(agent *Agent, this Value, callback Value, thisArg Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if !IsCallable(callback) {
		return agent.ThrowException(TypeError, "is not callable")
	}
	A := TypedArraySpeciesCreate(agent, ta, []Value{length.ToValue()})
	k := JSInt(0)
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		kValue := ta.Get(Pk)
		mappedValue := ReturnAssertNormal(
			callback.Call(thisArg, []Value{kValue, k.ToValue(), ta.ToValue()}),
		)
		A.Set(Pk, mappedValue, setThrowTypeThrow)
		k++
	}
	return A.ToValue()
}

func typedArrayReduce(agent *Agent, this Value, callback Value, initialValue Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if !IsCallable(callback) {
		return agent.ThrowException(TypeError, "is not callable")
	}
	if length == 0 && IsUndefinedOrNil(initialValue) {
		return agent.ThrowException(TypeError, "empty")
	}
	k := JSInt(0)
	var accumulator Value
	if IsUndefinedOrNil(initialValue) {
		Pk := NewIntegerIndexPropertyKey(k)
		accumulator = ta.Get(Pk)
		k++
	} else {
		accumulator = initialValue
	}
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		kValue := ta.Get(Pk)
		accumulator = ReturnAssertNormal(
			callback.Call(UndefinedValue, []Value{accumulator, kValue, k.ToValue(), ta.ToValue()}),
		)
		k++
	}
	return accumulator
}

func typedArrayReduceRight(agent *Agent, this Value, callback Value, initialValue Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if !IsCallable(callback) {
		return agent.ThrowException(TypeError, "is not callable")
	}
	if length == 0 && IsUndefinedOrNil(initialValue) {
		return agent.ThrowException(TypeError, "empty")
	}
	k := length - 1
	var accumulator Value
	if IsUndefinedOrNil(initialValue) {
		Pk := NewIntegerIndexPropertyKey(k)
		accumulator = ta.Get(Pk)
		k--
	} else {
		accumulator = initialValue
	}
	for k >= 0 {
		Pk := NewIntegerIndexPropertyKey(k)
		kValue := ta.Get(Pk)
		accumulator = ReturnAssertNormal(
			callback.Call(UndefinedValue, []Value{accumulator, kValue, k.ToValue(), ta.ToValue()}),
		)
		k--
	}
	return accumulator
}

func typedArraySet(agent *Agent, this Value, source, offset Value) Value {
	target := RequireInternalSlot[*TypedArrayObject](this)
	targetOffset := ToIntegerOrInfinity(agent, offset)
	if targetOffset < 0 {
		return agent.ThrowException(RangeError, "negative offset")
	}
	var sourceIsTypedArray bool
	if s, ok := source.GetObject(); ok {
		if ta, ok := s.(*TypedArrayObject); ok {
			SetTypedArrayFromTypedArray(agent, target, targetOffset, ta)
			sourceIsTypedArray = true
		}
	}
	if !sourceIsTypedArray {
		SetTypedArrayFromArrayLike(agent, target, targetOffset, source)
	}

	return UndefinedValue
}

func typedArraySlice(agent *Agent, this Value, start, end Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray

	srcArrayLength := TypedArrayLength(taRecord)
	relativeStart := ToIntegerOrInfinity(agent, start)
	var startIndex JSInt
	if relativeStart.IsNegInf() {
		startIndex = 0
	} else if relativeStart < 0 {
		startIndex = (srcArrayLength + relativeStart).Max(0)
	} else {
		startIndex = relativeStart.Min(srcArrayLength)
	}
	var relativeEnd JSInt
	if IsUndefinedOrNil(end) {
		relativeEnd = srcArrayLength
	} else {
		relativeEnd = ToIntegerOrInfinity(agent, end)
	}
	var endIndex JSInt
	if relativeEnd.IsNegInf() {
		endIndex = 0
	} else if relativeEnd < 0 {
		endIndex = (srcArrayLength + relativeEnd).Max(0)
	} else {
		endIndex = relativeEnd.Min(srcArrayLength)
	}

	countBytes := (endIndex - startIndex).Max(0)

	A := TypedArraySpeciesCreate(agent, ta, []Value{countBytes.ToValue()})
	if countBytes > 0 {
		taRecord = MakeTypedArrayWithBufferWitnessRecord(ta, SeqCst)
		if IsTypedArrayOutOfBounds(taRecord) {
			return agent.ThrowException(RangeError, "out of bounds")
		}

		endIndex = endIndex.Min(TypedArrayLength(taRecord))
		countBytes = (endIndex - startIndex).Max(0)
		srcType := ta.TypedArrayName
		targetType := A.TypedArrayName
		if srcType == targetType {
			srcBuffer := ta.ViewedArrayBuffer
			targetBuffer := A.ViewedArrayBuffer
			elementSize := TypedArrayElementSize(ta)
			srcByteOffset := ta.ByteOffset
			srcByteIndex := startIndex*elementSize + srcByteOffset
			targetByteIndex := A.ByteOffset
			endByteIndex := targetByteIndex + countBytes*elementSize
			for targetByteIndex < endByteIndex {
				value := GetValueFromBuffer(agent, srcBuffer, srcByteIndex, elementSize, true, Relaxed)
				SetValueInBuffer(agent, targetBuffer, targetByteIndex, JSNumber(value).ToValue(), elementSize, true, Relaxed)
				srcByteIndex += elementSize
				targetByteIndex += elementSize
			}
		} else {
			n := JSInt(0)
			k := startIndex
			for k < endIndex {
				Pk := NewIntegerIndexPropertyKey(k)
				kValue := ta.Get(Pk)
				A.Set(NewIntegerIndexPropertyKey(n), kValue, setThrowTypeThrow)
				k++
				n++
			}
		}
	}
	return A.ToValue()
}

func typedArraySome(agent *Agent, this Value, callback Value, thisArg Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	if !IsCallable(callback) {
		return agent.ThrowException(TypeError, "is not callable")
	}
	if length == 0 {
		return FalseValue
	}
	k := JSInt(0)
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		kValue := ta.Get(Pk)
		testResult := ReturnAssertNormal(
			callback.Call(thisArg, []Value{kValue, k.ToValue(), ta.ToValue()}),
		)
		if testResult.ToBoolean() {
			return TrueValue
		}
		k++
	}
	return FalseValue
}

func typedArraySort(agent *Agent, this Value, compareFn Value) Value {
	if !IsUndefinedOrNil(compareFn) && !IsCallable(compareFn) {
		return agent.ThrowException(TypeError, "is not callable")
	}
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)

	sortCompare := SortCompare{
		impl:      CompareTypedArrayElements,
		compareFn: MustGetObject(compareFn),
	}
	sortedList := SortIndexedProperties(agent, ta, length, sortCompare, sortHolesTypeReadThroughHoles)

	j := JSInt(0)
	for j < length {
		Pk := NewIntegerIndexPropertyKey(j)
		ta.Set(Pk, sortedList[j], setThrowTypeThrow)
		j++
	}
	return ta.ToValue()
}

func typedArrayToSorted(agent *Agent, this Value, compareFn Value) Value {
	if !IsUndefinedOrNil(compareFn) && !IsCallable(compareFn) {
		return agent.ThrowException(TypeError, "is not callable")
	}
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)

	sortCompare := SortCompare{
		impl:      CompareTypedArrayElements,
		compareFn: MustGetObject(compareFn),
	}
	sortedList := SortIndexedProperties(agent, ta, length, sortCompare, sortHolesTypeReadThroughHoles)

	A := TypedArrayCreateSameType(agent, ta, []Value{length.ToValue()})
	for j, value := range sortedList {
		Pk := NewIntegerIndexPropertyKey(JSInt(j))
		A.Set(Pk, value, setThrowTypeThrow)
	}

	return A.ToValue()
}

func typedArraySubarray(agent *Agent, this Value, begin, end Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	buffer := ta.ViewedArrayBuffer
	srcRecord := MakeTypedArrayWithBufferWitnessRecord(ta, SeqCst)
	var srcLength JSInt
	if IsTypedArrayOutOfBounds(srcRecord) {
		srcLength = 0
	} else {
		srcLength = TypedArrayLength(srcRecord)
	}
	relativeStart := ToIntegerOrInfinity(agent, begin)
	var startIndex JSInt
	if relativeStart.IsNegInf() {
		startIndex = 0
	} else if relativeStart < 0 {
		startIndex = (srcLength + relativeStart).Max(0)
	} else {
		startIndex = relativeStart.Min(srcLength)
	}

	elementSize := TypedArrayElementSize(ta)
	srcByteOffset := ta.ByteOffset
	beginByteOffset := startIndex*elementSize + srcByteOffset
	var argumentsList []Value
	if ta.ArrayLength.Auto && IsUndefinedOrNil(end) {
		argumentsList = []Value{buffer.ToValue(), beginByteOffset.ToValue()}
	} else {
		var relativeEnd JSInt
		if IsUndefinedOrNil(end) {
			relativeEnd = srcLength
		} else {
			relativeEnd = ToIntegerOrInfinity(agent, end)
		}
		var endIndex JSInt
		if relativeEnd.IsNegInf() {
			endIndex = 0
		} else if relativeEnd < 0 {
			endIndex = (srcLength + relativeEnd).Max(0)
		} else {
			endIndex = relativeEnd.Min(srcLength)
		}

		newLength := (endIndex - startIndex).Max(0)
		argumentsList = []Value{buffer.ToValue(), beginByteOffset.ToValue(), newLength.ToValue()}
	}
	return TypedArraySpeciesCreate(agent, ta, argumentsList).ToValue()
}

func typedArrayWith(agent *Agent, this Value, index, value Value) Value {
	O := this
	taRecord := ValidateTypedArray(agent, O, SeqCst)
	ta := taRecord.TypedArray
	length := TypedArrayLength(taRecord)
	relativeIndex := ToIntegerOrInfinity(agent, index)
	var actualIndex JSInt
	if relativeIndex >= 0 {
		actualIndex = relativeIndex
	} else {
		actualIndex = length + relativeIndex
	}

	var numericValue Value
	if ta.ContentType == TypedArrayContentTypeBigInt {
		numericValue = ToBigInt(agent, value)
	} else {
		numericValue = value.ToNumber(agent)
	}
	if !ta.IsValidIntegerIndex(agent, actualIndex) {
		return agent.ThrowException(RangeError, "invalid index")
	}
	A := TypedArrayCreateSameType(agent, ta, []Value{length.ToValue()})
	k := JSInt(0)
	for k < length {
		Pk := NewIntegerIndexPropertyKey(k)
		if k == actualIndex {
			A.Set(Pk, numericValue, setThrowTypeThrow)
		} else {
			A.Set(Pk, ta.Get(Pk), setThrowTypeThrow)
		}
		k++
	}
	return A.ToValue()
}
