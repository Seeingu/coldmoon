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
	object := CreateBuiltinFunction(agent, behavior, 1, CMString("DataView"), builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})

	BindPrototypeAndConstructor(realm.Intrinsics.DataViewPrototype, object)

	return object
}

func NewDataViewPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype, "DataViewPrototype")

	buffer := func(this Value, args []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*DataView](this)
		return (o.ViewedArrayBuffer).ToValue()
	}
	object.defineBuiltinAccessor(realm, CMString("buffer"), builtinAccessorParams{
		Getter: buffer,
	})
	byteLength := func(this Value, args []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*DataView](this)
		viewRecord := MakeDataViewWithBufferWitnessRecord(o, SeqCst)
		if IsViewOutOfBounds(viewRecord) {
			panic("RangeError")
		}
		size := GetViewByteLength(viewRecord)
		return NewNumberValue(size.ToNumber())
	}
	object.defineBuiltinAccessor(realm, CMString("byteLength"), builtinAccessorParams{
		Getter: byteLength,
	})
	byteOffset := func(this Value, args []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*DataView](this)
		viewRecord := MakeDataViewWithBufferWitnessRecord(o, SeqCst)
		if IsViewOutOfBounds(viewRecord) {
			panic("RangeError")
		}
		offset := o.ByteOffset
		return NewNumberValue(offset.ToNumber())
	}
	object.defineBuiltinAccessor(realm, CMString("byteOffset"), builtinAccessorParams{
		Getter: byteOffset,
	})

	getBigInt64 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := GetViewValue(agent, this, args[0], 8)
		return value.Data()
	}
	getBigUint64 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := GetViewValue(agent, this, args[0], 8)
		return value.Data()
	}
	getFloat32 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := GetViewValue(agent, this, args[0], 4)
		return value.Data()
	}
	getFloat64 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := GetViewValue(agent, this, args[0], 8)
		return value.Data()
	}
	getInt8 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := GetViewValue(agent, this, args[0], 1)
		return value.Data()
	}
	getInt16 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := GetViewValue(agent, this, args[0], 2)
		return value.Data()
	}
	getInt32 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := GetViewValue(agent, this, args[0], 4)
		return value.Data()
	}
	getUint8 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := GetViewValue(agent, this, args[0], 1)
		return value.Data()
	}
	getUint16 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := GetViewValue(agent, this, args[0], 2)
		return value.Data()
	}
	getUint32 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := GetViewValue(agent, this, args[0], 4)
		return value.Data()
	}
	setBigInt64 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := SetViewValue(agent, this, args[0], args[1], 8)
		return value.Data()
	}
	setBigUint64 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := SetViewValue(agent, this, args[0], args[1], 8)
		return value.Data()
	}
	setFloat32 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := SetViewValue(agent, this, args[0], args[1], 4)
		return value.Data()
	}
	setFloat64 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := SetViewValue(agent, this, args[0], args[1], 8)
		return value.Data()
	}
	setInt8 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := SetViewValue(agent, this, args[0], args[1], 1)
		return value.Data()
	}
	setInt16 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := SetViewValue(agent, this, args[0], args[1], 2)
		return value.Data()
	}
	setInt32 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := SetViewValue(agent, this, args[0], args[1], 4)
		return value.Data()
	}
	setUint8 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := SetViewValue(agent, this, args[0], args[1], 1)
		return value.Data()
	}
	setUint16 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := SetViewValue(agent, this, args[0], args[1], 2)
		return value.Data()
	}
	setUint32 := func(this Value, args []Value, newTarget ObjectType) Value {
		value := SetViewValue(agent, this, args[0], args[1], 4)
		return value.Data()
	}

	object.defineBuiltinFunction(realm, CMString("getBigInt64"), getBigInt64, 2)
	object.defineBuiltinFunction(realm, CMString("getBigUint64"), getBigUint64, 2)
	object.defineBuiltinFunction(realm, CMString("getFloat32"), getFloat32, 2)
	object.defineBuiltinFunction(realm, CMString("getFloat64"), getFloat64, 2)
	object.defineBuiltinFunction(realm, CMString("getInt8"), getInt8, 2)
	object.defineBuiltinFunction(realm, CMString("getInt16"), getInt16, 2)
	object.defineBuiltinFunction(realm, CMString("getInt32"), getInt32, 2)
	object.defineBuiltinFunction(realm, CMString("getUint8"), getUint8, 2)
	object.defineBuiltinFunction(realm, CMString("getUint16"), getUint16, 2)
	object.defineBuiltinFunction(realm, CMString("getUint32"), getUint32, 2)
	object.defineBuiltinFunction(realm, CMString("setBigInt64"), setBigInt64, 3)
	object.defineBuiltinFunction(realm, CMString("setBigUint64"), setBigUint64, 3)
	object.defineBuiltinFunction(realm, CMString("setFloat32"), setFloat32, 3)
	object.defineBuiltinFunction(realm, CMString("setFloat64"), setFloat64, 3)
	object.defineBuiltinFunction(realm, CMString("setInt8"), setInt8, 3)
	object.defineBuiltinFunction(realm, CMString("setInt16"), setInt16, 3)
	object.defineBuiltinFunction(realm, CMString("setInt32"), setInt32, 3)
	object.defineBuiltinFunction(realm, CMString("setUint8"), setUint8, 3)
	object.defineBuiltinFunction(realm, CMString("setUint16"), setUint16, 3)
	object.defineBuiltinFunction(realm, CMString("setUint32"), setUint32, 3)

	object.defineToStringTag("DataView")
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
func GetViewValue(agent *Agent, viewValue Value, requestIndex Value, size JSInt) (co CompletionValue) {
	view := RequireInternalSlot[*DataView](viewValue)
	getIndex := ToIndex(agent, requestIndex)
	viewOffset := view.ByteOffset
	viewRecord := MakeDataViewWithBufferWitnessRecord(view, SeqCst)
	if IsViewOutOfBounds(viewRecord) {
		co.err = agent.ThrowException(RangeError, "DataView is out of bounds")
		return
	}
	viewSize := GetViewByteLength(viewRecord)
	elementSize := size
	if getIndex+elementSize > viewSize {
		co.err = agent.ThrowException(RangeError, "DataView is out of bounds")
		return
	}
	bufferIndex := getIndex + viewOffset
	v := GetValueFromBuffer(agent, view.ViewedArrayBuffer, bufferIndex, size, false, SeqCst)
	co.value = NewNumberValue(JSNumber(v))
	return
}

func SetViewValue(
	agent *Agent,
	viewValue Value,
	requestIndex Value,
	value Value,
	size JSInt,
) (co CompletionValue) {
	view := RequireInternalSlot[*DataView](viewValue)
	setIndex := ToIndex(agent, requestIndex)
	viewOffset := view.ByteOffset
	viewRecord := MakeDataViewWithBufferWitnessRecord(view, SeqCst)
	var numberValue JSNumber
	if IsBigIntElementType(size) {
		numberValue = JSNumber(ToBigInt(agent, value).Data.Uint64())
	} else {
		numberValue = value.ToNumber(agent).Data
	}
	if IsViewOutOfBounds(viewRecord) {
		co.err = agent.ThrowException(RangeError, "DataView is out of bounds")
		return
	}
	viewSize := GetViewByteLength(viewRecord)
	elementSize := size
	if setIndex+elementSize > viewSize {
		co.err = agent.ThrowException(RangeError, "DataView is out of bounds")
		return
	}
	bufferIndex := setIndex + viewOffset
	SetValueInBuffer(agent, view.ViewedArrayBuffer, bufferIndex, NewNumberValue(numberValue), size, false, SeqCst)
	co.value = UndefinedValue
	return
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
