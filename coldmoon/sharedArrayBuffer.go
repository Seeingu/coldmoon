package coldmoon

import (
	"github.com/Seeingu/coldmoon/pkg"
)

type SharedArrayBufferObject struct {
	*Object
	ArrayBufferData          *DataBlock
	ArrayBufferByteLength    JSInt
	ArrayBufferMaxByteLength JSInt
}

// 25.2.2.1
func AllocateSharedArrayBuffer(
	agent *Agent,
	constructor ObjectType,
	byteLength JSInt,
	maxByteLength JSInt,
) ObjectType {
	allocatingGrowableBuffer := maxByteLength != 0
	if allocatingGrowableBuffer {
		if byteLength > maxByteLength {
			return agent.ThrowRangeExceptionObject("byteLength > maxByteLength")
		}
	}
	obj := OrdinaryCreateFromConstructor(
		agent,
		constructor,
		IntrinsicNameSharedArrayBufferPrototype,
		nil,
	)

	var allocLength JSInt
	if allocatingGrowableBuffer {
		allocLength = maxByteLength
	} else {
		allocLength = byteLength
	}
	block := CreateSharedByteDataBlock(agent, allocLength)

	sharedArrayBuffer := &SharedArrayBufferObject{
		Object:          obj,
		ArrayBufferData: block,
	}
	if allocatingGrowableBuffer {
		Assert(maxByteLength >= byteLength)
		sharedArrayBuffer.ArrayBufferMaxByteLength = maxByteLength
	}
	sharedArrayBuffer.ref = sharedArrayBuffer

	return sharedArrayBuffer
}

func NewSharedArrayBufferConstructor(realm *Realm) ObjectType {
	agent := realm.Agent

	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		length := argumentsList[0]
		options := pkg.SliceSafeGet(argumentsList, 1)
		if options == nil {
			options = UndefinedValue
		}
		if newTarget == nil {
			return agent.ThrowTypeError("newTarget is nil in SharedArrayBuffer")
		}
		byteLength := ToIndex(agent, length)
		requestedMaxByteLength := GetArrayBufferMaxByteLengthOption(agent, options)
		return AllocateSharedArrayBuffer(agent, newTarget, byteLength, requestedMaxByteLength).ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, CMString("SharedArrayBuffer"), builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	object.defineBuiltinAccessor(realm, WellKnownSymbolsSpecies, builtinAccessorParams{
		Getter: func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
			return this
		},
	})
	BindPrototypeAndConstructor(realm.Intrinsics.SharedArrayBufferPrototype, object)
	return object
}

func NewSharedArrayBufferPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "SharedArrayBuffer")

	var byteLength BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		O := RequireInternalSlot[*SharedArrayBufferObject](this)
		length := ArrayBufferByteLength(NewArrayBufferLike(O), SeqCst)
		return length.ToValue()
	}
	grow := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		newLength := argumentsList[0]
		return sharedArrayBufferGrow(agent, this, newLength)
	}
	growable := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		O := RequireInternalSlot[*SharedArrayBufferObject](this)
		if IsFixedLengthArrayBuffer(NewArrayBufferLike(O)) {
			return FalseValue
		}
		return TrueValue
	}
	maxByteLength := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		O := RequireInternalSlot[*SharedArrayBufferObject](this)
		var length JSInt
		if IsFixedLengthArrayBuffer(NewArrayBufferLike(O)) {
			length = ArrayBufferByteLength(NewArrayBufferLike(O), SeqCst)
		} else {
			length = O.ArrayBufferMaxByteLength
		}
		return length.ToValue()
	}
	slice := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		start := argumentsList[0]
		end := argumentsList[1]
		return sharedArrayBufferSlice(agent, this, start, end)
	}
	object.defineBuiltinAccessor(realm, CMString("byteLength"), builtinAccessorParams{
		Getter: byteLength,
	})
	object.defineBuiltinFunction(realm, CMString("grow"), grow, 1)
	object.defineBuiltinFunction(realm, CMString("slice"), slice, 2)
	object.defineBuiltinAccessor(realm, CMString("growable"), builtinAccessorParams{
		Getter: growable,
	})
	object.defineBuiltinAccessor(realm, CMString("maxByteLength"), builtinAccessorParams{
		Getter: maxByteLength,
	})

	object.defineToStringTag("SharedArrayBuffer")
	return object
}

// MARK: - Internal

func sharedArrayBufferGrow(agent *Agent, this Value, newLength Value) Value {
	O := RequireInternalSlot[*SharedArrayBufferObject](this)
	if O.ArrayBufferMaxByteLength == 0 {
		return agent.ThrowTypeError("SharedArrayBuffer.prototype.grow called on a non-growable SharedArrayBuffer")
	}
	currentLength := ArrayBufferByteLength(NewArrayBufferLike(O), SeqCst)
	newByteLength := ToIndex(agent, newLength)
	if newByteLength <= currentLength {
		return agent.ThrowRangeError("newByteLength <= currentLength")
	}
	if newByteLength > O.ArrayBufferMaxByteLength {
		return agent.ThrowRangeError("newByteLength > O.ArrayBufferMaxByteLength")
	}
	// TODO: resize

	return UndefinedValue
}

func sharedArrayBufferSlice(agent *Agent, this Value, start Value, end Value) Value {
	realm := agent.CurrentRealm()
	O := RequireInternalSlot[*SharedArrayBufferObject](this)
	length := ArrayBufferByteLength(NewArrayBufferLike(O), SeqCst)
	relativeStart := ToIntegerOrInfinity(agent, start)
	var first JSInt
	if relativeStart.IsNegInf() {
		first = 0
	} else if relativeStart < 0 {
		first = (length + relativeStart).Max(0)
	} else {
		first = relativeStart.Min(length)
	}

	var relativeEnd JSInt
	if IsUndefinedOrNil(end) {
		relativeEnd = length
	} else {
		relativeEnd = ToIntegerOrInfinity(agent, end)
	}
	var final JSInt
	if relativeEnd.IsNegInf() {
		final = 0
	} else if relativeEnd < 0 {
		final = (length + relativeEnd).Max(0)
	} else {
		final = relativeEnd.Min(length)
	}

	newLen := (final - first).Max(0)
	ctor := O.SpeciesConstructor(realm.Intrinsics.SharedArrayBufferConstructor)
	newObject := ctor.Data().Construct([]Value{newLen.ToValue()}, nil).value
	sharedArrayBuffer := RequireInternalSlot[*SharedArrayBufferObject](newObject.ToValue())
	if sharedArrayBuffer.ArrayBufferData.Equal(O.ArrayBufferData) {
		return agent.ThrowTypeError("should return a new ArrayBuffer instance")
	}

	if ArrayBufferByteLength(NewArrayBufferLike(sharedArrayBuffer), SeqCst) < newLen {
		return agent.ThrowTypeError("size of new ArrayBuffer is less than newLen")
	}

	fromBuf := O.ArrayBufferData
	toBuf := sharedArrayBuffer.ArrayBufferData
	CopyDataBlockBytes(toBuf, 0, fromBuf, first, newLen)
	return newObject.ToValue()
}
