package coldmoon

import (
	"github.com/Seeingu/coldmoon/pkg"
)

// ByteLength Enum
type ByteLength struct {
	Auto  bool
	Value int
}

type DataView struct {
	*Object
	// [[ViewedArrayBuffer]]
	ViewedArrayBuffer *ArrayBufferObject
	// [[ByteLength]]
	ByteLength *ByteLength
	// [[ByteOffset]]
	ByteOffset int
}

func NewDataViewConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior = func(this Value, args []Value, newTarget ObjectType) Value {
		bufferValue := args[0]
		byteOffset := pkg.SliceSafeGet(args, 1)
		byteLength := pkg.SliceSafeGet(args, 2)
		if newTarget == nil {
			panic("TypeError")
		}

		buffer := RequireInternalSlot[*ArrayBufferObject](bufferValue)
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
				viewByteLength.Value = int(bufferByteLength - offset)
			}
		} else {
			viewByteLength.Value = int(ToIndex(agent, byteLength))
			if bufferIsFixedLength {
				if viewByteLength.Value > int(bufferByteLength-offset) {
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
		dataView.ByteOffset = int(offset)
		return dataView.ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 1, "DataView", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        realm.Intrinsics.DataViewPrototype.ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.DataViewPrototype, "constructor", object.ToValue())

	return object
}

func NewDataViewPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.ObjectPrototype)

	var buffer = func(this Value, args []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*DataView](this)
		return o.ViewedArrayBuffer.ToValue()
	}
	DefineBuiltinAccessor(realm, object, "buffer", buffer, nil)
	var byteLength = func(this Value, args []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*DataView](this)
		viewRecord := MakeDataViewWithBufferWitnessRecord(o, SeqCst)
		if IsViewOutOfBounds(viewRecord) {
			panic("RangeError")
		}
		size := GetViewByteLength(viewRecord)
		return NewNumberValue(float64(size))
	}
	DefineBuiltinAccessor(realm, object, "byteLength", byteLength, nil)
	var byteOffset = func(this Value, args []Value, newTarget ObjectType) Value {
		o := RequireInternalSlot[*DataView](this)
		viewRecord := MakeDataViewWithBufferWitnessRecord(o, SeqCst)
		if IsViewOutOfBounds(viewRecord) {
			panic("RangeError")
		}
		offset := o.ByteOffset
		return NewNumberValue(float64(offset))
	}
	DefineBuiltinAccessor(realm, object, "byteOffset", byteOffset, nil)

	var getBigInt64 = func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], args[1], 8).Value
	}
	var getBigUint64 = func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], args[1], 8).Value
	}
	var getFloat32 = func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], args[1], 4).Value
	}
	var getFloat64 = func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], args[1], 8).Value
	}
	var getInt8 = func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], args[1], 1).Value
	}
	var getInt16 = func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], args[1], 2).Value
	}
	var getInt32 = func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], args[1], 4).Value
	}
	var getUint8 = func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], args[1], 1).Value
	}
	var getUint16 = func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], args[1], 2).Value
	}
	var getUint32 = func(this Value, args []Value, newTarget ObjectType) Value {
		return GetViewValue(agent, this, args[0], args[1], 4).Value
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
	Value    int
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
		byteLength.Value = int(ArrayBufferByteLength(buffer, order))
	}
	return &DataViewWithBufferWitnessRecord{
		Object:                 object,
		CachedBufferByteLength: byteLength,
	}
}

func GetViewByteLength(viewRecord *DataViewWithBufferWitnessRecord) int {
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
func GetViewValue(agent *Agent, viewValue Value, requestIndex Value, isLittleEndian Value, size uint64) *CompletionValue {
	view := RequireInternalSlot[*DataView](viewValue)
	getIndex := ToIndex(agent, requestIndex)
	isLittleEndianBool := isLittleEndian.ToBoolean()
	viewOffset := view.ByteOffset
	viewRecord := MakeDataViewWithBufferWitnessRecord(view, SeqCst)
	if IsViewOutOfBounds(viewRecord) {
		return NewCompletionValueError(agent.ThrowException(RangeError, "DataView is out of bounds"))
	}
	viewSize := GetViewByteLength(viewRecord)
	elementSize := size
	if getIndex+elementSize > uint64(viewSize) {
		return NewCompletionValueError(agent.ThrowException(RangeError, "DataView is out of bounds"))
	}
	bufferIndex := getIndex + uint64(viewOffset)
	v := GetValueFromBuffer(agent, view.ViewedArrayBuffer, bufferIndex, size, false, SeqCst, isLittleEndianBool)
	return NewCompletionValue(NewNumberValue(float64(v)))
}

func IsViewOutOfBounds(viewRecord *DataViewWithBufferWitnessRecord) bool {
	view := viewRecord.Object
	bufferByteLength := viewRecord.CachedBufferByteLength
	Assert(IsDetachedBuffer(view.ViewedArrayBuffer) == bufferByteLength.Detached)

	byteOffsetStart := view.ByteOffset
	var byteOffsetEnd int
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
