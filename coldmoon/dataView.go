package coldmoon

import (
	"github.com/Seeingu/coldmoon/pkg"
)

// ByteLength Enum
type ByteLength struct {
	Auto  bool
	Value JSInt
}

func NewByteLength(v JSInt) *ByteLength {
	return &ByteLength{
		Value: v,
	}
}

func (b *ByteLength) toAuto() {
	b.Auto = true
	b.Value = 0
}

type DataView struct {
	*Object
	// [[ViewedArrayBuffer]]
	ViewedArrayBuffer *ArrayBufferLike
	// [[ByteLength]]
	ByteLength *ByteLength
	// [[ByteOffset]]
	ByteOffset JSInt
}

func NewDataViewConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	behavior := func(this Value, args []Value, newTarget ObjectType) Value {
		bufferValue := args[0]
		byteOffset := pkg.SliceSafeGet(args, 1)
		byteLength := pkg.SliceSafeGet(args, 2)
		if newTarget == nil {
			panic("TypeError")
		}

		buffer := RequireInternalSlot[*ArrayBufferLike](bufferValue)
		offset := ToIndex(agent, byteOffset)
		if IsDetachedBuffer(buffer) {
			panic("TypeError")
		}
		bufferByteLength := ArrayBufferByteLength(buffer, SeqCst)
		if offset > bufferByteLength {
			panic("RangeError")
		}
		bufferIsFixedLength := IsFixedLengthArrayBuffer(buffer)
		viewByteLength := &ByteLength{}
		if byteLength == nil {
			if bufferIsFixedLength {
				viewByteLength.Auto = true
			} else {
				viewByteLength.Value = bufferByteLength - offset
			}
		} else {
			viewByteLength.Value = ToIndex(agent, byteLength)
			if bufferIsFixedLength {
				if viewByteLength.Value > bufferByteLength-offset {
					panic("RangeError")
				}
			}
		}

		o := OrdinaryCreateFromConstructor(agent, newTarget, "%DataView.prototype%", []string{})
		dataView := &DataView{
			Object: o,
		}
		if IsDetachedBuffer(buffer) {
			panic("TypeError")
		}
		bufferByteLength = ArrayBufferByteLength(buffer, SeqCst)
		if offset > bufferByteLength {
			panic("RangeError")
		}

		dataView.ViewedArrayBuffer = buffer
		dataView.ByteLength = viewByteLength
		dataView.ByteOffset = offset
		return (dataView).ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, "DataView", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        (realm.Intrinsics.DataViewPrototype).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.DataViewPrototype, "constructor", (object).ToValue())

	return object
}

func NewDataViewPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "DataViewPrototype")

	buffer := func(this Value, args []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*DataView](this)
		return (o.ViewedArrayBuffer).ToValue()
	}
	DefineBuiltinAccessor(realm, object, "buffer", buffer, nil)
	byteLength := func(this Value, args []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*DataView](this)
		viewRecord := MakeDataViewWithBufferWitnessRecord(o, SeqCst)
		if IsViewOutOfBounds(viewRecord) {
			panic("RangeError")
		}
		size := GetViewByteLength(viewRecord)
		return NewNumberValue(size.ToNumber())
	}
	DefineBuiltinAccessor(realm, object, "byteLength", byteLength, nil)
	byteOffset := func(this Value, args []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*DataView](this)
		viewRecord := MakeDataViewWithBufferWitnessRecord(o, SeqCst)
		if IsViewOutOfBounds(viewRecord) {
			panic("RangeError")
		}
		offset := o.ByteOffset
		return NewNumberValue(offset.ToNumber())
	}
	DefineBuiltinAccessor(realm, object, "byteOffset", byteOffset, nil)

	getBigInt64 := func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], 8).Data()
	}
	getBigUint64 := func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], 8).Data()
	}
	getFloat32 := func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], 4).Data()
	}
	getFloat64 := func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], 8).Data()
	}
	getInt8 := func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], 1).Data()
	}
	getInt16 := func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], 2).Data()
	}
	getInt32 := func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], 4).Data()
	}
	getUint8 := func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], 1).Data()
	}
	getUint16 := func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], 2).Data()
	}
	getUint32 := func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], 4).Data()
	}
	setBigInt64 := func(this Value, args []Value, newTarget ObjectType) Value {
		return SetViewValue(agent, this, args[0], args[1], 8).Data()
	}
	setBigUint64 := func(this Value, args []Value, newTarget ObjectType) Value {
		return SetViewValue(agent, this, args[0], args[1], 8).Data()
	}
	setFloat32 := func(this Value, args []Value, newTarget ObjectType) Value {
		return SetViewValue(agent, this, args[0], args[1], 4).Data()
	}
	setFloat64 := func(this Value, args []Value, newTarget ObjectType) Value {
		return SetViewValue(agent, this, args[0], args[1], 8).Data()
	}
	setInt8 := func(this Value, args []Value, newTarget ObjectType) Value {
		return SetViewValue(agent, this, args[0], args[1], 1).Data()
	}
	setInt16 := func(this Value, args []Value, newTarget ObjectType) Value {
		return SetViewValue(agent, this, args[0], args[1], 2).Data()
	}
	setInt32 := func(this Value, args []Value, newTarget ObjectType) Value {
		return SetViewValue(agent, this, args[0], args[1], 4).Data()
	}
	setUint8 := func(this Value, args []Value, newTarget ObjectType) Value {
		return SetViewValue(agent, this, args[0], args[1], 1).Data()
	}
	setUint16 := func(this Value, args []Value, newTarget ObjectType) Value {
		return SetViewValue(agent, this, args[0], args[1], 2).Data()
	}
	setUint32 := func(this Value, args []Value, newTarget ObjectType) Value {
		return SetViewValue(agent, this, args[0], args[1], 4).Data()
	}

	DefineBuiltinFunction(object, "getBigInt64", getBigInt64, 2, realm)
	DefineBuiltinFunction(object, "getBigUint64", getBigUint64, 2, realm)
	DefineBuiltinFunction(object, "getFloat32", getFloat32, 2, realm)
	DefineBuiltinFunction(object, "getFloat64", getFloat64, 2, realm)
	DefineBuiltinFunction(object, "getInt8", getInt8, 2, realm)
	DefineBuiltinFunction(object, "getInt16", getInt16, 2, realm)
	DefineBuiltinFunction(object, "getInt32", getInt32, 2, realm)
	DefineBuiltinFunction(object, "getUint8", getUint8, 2, realm)
	DefineBuiltinFunction(object, "getUint16", getUint16, 2, realm)
	DefineBuiltinFunction(object, "getUint32", getUint32, 2, realm)
	DefineBuiltinFunction(object, "setBigInt64", setBigInt64, 3, realm)
	DefineBuiltinFunction(object, "setBigUint64", setBigUint64, 3, realm)
	DefineBuiltinFunction(object, "setFloat32", setFloat32, 3, realm)
	DefineBuiltinFunction(object, "setFloat64", setFloat64, 3, realm)
	DefineBuiltinFunction(object, "setInt8", setInt8, 3, realm)
	DefineBuiltinFunction(object, "setInt16", setInt16, 3, realm)
	DefineBuiltinFunction(object, "setInt32", setInt32, 3, realm)
	DefineBuiltinFunction(object, "setUint8", setUint8, 3, realm)
	DefineBuiltinFunction(object, "setUint16", setUint16, 3, realm)
	DefineBuiltinFunction(object, "setUint32", setUint32, 3, realm)

	DefineBuiltinPropertyP(object, "@@toStringTag", &PropertyDescriptor{
		Value:        NewStringValue("DataView"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

	return object
}

// CachedBufferByteLength Enum
type CachedBufferByteLength struct {
	Detached bool
	Value    JSInt
}
type DataViewWithBufferWitnessRecord struct {
	// [[Object]]
	Object *DataView
	// [[CachedBufferByteLength]]
	CachedBufferByteLength *CachedBufferByteLength
}

func MakeDataViewWithBufferWitnessRecord(object *DataView, order MemoryOrder) *DataViewWithBufferWitnessRecord {
	buffer := object.ViewedArrayBuffer
	byteLength := &CachedBufferByteLength{}
	if IsDetachedBuffer(buffer) {
		byteLength.Detached = true
	} else {
		byteLength.Value = ArrayBufferByteLength(buffer, order)
	}
	return &DataViewWithBufferWitnessRecord{
		Object:                 object,
		CachedBufferByteLength: byteLength,
	}
}

func GetViewByteLength(viewRecord *DataViewWithBufferWitnessRecord) JSInt {
	Assert(!IsViewOutOfBounds(viewRecord))

	view := viewRecord.Object
	if !view.ByteLength.Auto {
		return view.ByteLength.Value
	}

	Assert(IsFixedLengthArrayBuffer(view.ViewedArrayBuffer))

	byteOffset := view.ByteOffset
	byteLength := viewRecord.CachedBufferByteLength
	Assert(!byteLength.Detached)
	return byteLength.Value - byteOffset
}

// 25.3.1.5
func GetViewValue(agent *Agent, viewValue Value, requestIndex Value, size JSInt) CompletionValue {
	view := RequireInternalSlot[*DataView](viewValue)
	getIndex := ToIndex(agent, requestIndex)
	viewOffset := view.ByteOffset
	viewRecord := MakeDataViewWithBufferWitnessRecord(view, SeqCst)
	if IsViewOutOfBounds(viewRecord) {
		return NewCompletionValueError(agent.ThrowException(RangeError, "DataView is out of bounds"))
	}
	viewSize := GetViewByteLength(viewRecord)
	elementSize := size
	if getIndex+elementSize > viewSize {
		return NewCompletionValueError(agent.ThrowException(RangeError, "DataView is out of bounds"))
	}
	bufferIndex := getIndex + viewOffset
	v := GetValueFromBuffer(agent, view.ViewedArrayBuffer, bufferIndex, size, false, SeqCst)
	return NewCompletionValue(NewNumberValue(JSNumber(v)))
}

func SetViewValue(
	agent *Agent,
	viewValue Value,
	requestIndex Value,
	value Value,
	size JSInt,
) CompletionValue {
	view := RequireInternalSlot[*DataView](viewValue)
	setIndex := ToIndex(agent, requestIndex)
	viewOffset := view.ByteOffset
	viewRecord := MakeDataViewWithBufferWitnessRecord(view, SeqCst)
	var numberValue JSNumber
	if IsBigIntElementType(size) {
		numberValue = JSNumber(ToBigInt(agent, value).Data.Uint64())
	} else {
		numberValue = ToNumber(agent, value).Data
	}
	if IsViewOutOfBounds(viewRecord) {
		return NewCompletionValue(agent.ThrowException(RangeError, "DataView is out of bounds"))
	}
	viewSize := GetViewByteLength(viewRecord)
	elementSize := size
	if setIndex+elementSize > viewSize {
		return NewCompletionValue(agent.ThrowException(RangeError, "DataView is out of bounds"))
	}
	bufferIndex := setIndex + viewOffset
	SetValueInBuffer(agent, view.ViewedArrayBuffer, bufferIndex, NewNumberValue(numberValue), size, false, SeqCst)
	return NewCompletionValue(UndefinedValue)
}

func IsViewOutOfBounds(viewRecord *DataViewWithBufferWitnessRecord) bool {
	view := viewRecord.Object
	bufferByteLength := viewRecord.CachedBufferByteLength
	Assert(IsDetachedBuffer(view.ViewedArrayBuffer) == bufferByteLength.Detached)

	byteOffsetStart := view.ByteOffset
	var byteOffsetEnd JSInt
	if view.ByteLength.Auto {
		byteOffsetEnd = bufferByteLength.Value
	} else {
		byteOffsetEnd = view.ByteOffset + view.ByteLength.Value
	}
	if byteOffsetStart > bufferByteLength.Value || byteOffsetEnd > bufferByteLength.Value {
		return true
	}
	return false
}
