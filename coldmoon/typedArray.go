package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type TypedArrayContentType int

const (
	TypedArrayContentTypeBigInt TypedArrayContentType = iota
	TypedArrayContentTypeNumber
)

type TypedArrayObject struct {
	*Object
	// [[TypedArrayName]]
	TypedArrayName string
	// [[ContentType]]
	ContentType TypedArrayContentType
	// [[ViewedArrayBuffer]]
	ViewedArrayBuffer *ArrayBufferObject
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
	return typedArray
}

func NewTypedArrayNamePrototype(realm *Realm, name string) ObjectType {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.TypedArrayPrototype, "TypedArrayNamePrototype "+name)

	DefineBuiltinPropertyP(object, "BYTES_PER_ELEMENT", &PropertyDescriptor{
		Value:        NewNumberValue(getTypedArraySizeFromName(name).ToNumber()),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})
	return object
}

func NewTypedArrayConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		panic("TypeError")
	}
	object := CreateBuiltinFunction(agent, behavior, 0, "TypedArray", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})
	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.TypedArrayPrototype),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.TypedArrayPrototype, "constructor", object.ToValue())

	return object
}

func NewTypedArrayNameConstructor(realm *Realm, name string) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return typedArrayBehavior(agent, name, thisArgument, argumentsList, newTarget)
	}
	object := CreateBuiltinFunction(agent, behavior, 3, name, builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.TypedArrayConstructor,
		isConstructor: true,
	})

	intrinsicName := "%" + name + ".prototype%"
	intrinsic := realm.Intrinsics.Get(intrinsicName)
	DefineBuiltinPropertyP(object, "BYTES_PER_ELEMENT", &PropertyDescriptor{
		Value:        NewNumberValue(getTypedArraySizeFromName(name).ToNumber()),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value: NewValueFromObject(intrinsic),
	})
	DefineBuiltinPropertyV(intrinsic, "constructor", NewValueFromObject(object))

	return object
}

// 23.2.5.1
func typedArrayBehavior(agent *Agent, name string, thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
	if newTarget == nil {
		panic("TypeError")
	}
	constructorName := name
	proto := "%TypedArray.prototype%"
	numberOfArgs := len(argumentsList)
	if numberOfArgs == 0 {
		return NewValueFromObject(AllocateTypedArray(agent, constructorName, newTarget, proto, 0))
	} else {
		firstArgument := argumentsList[0]
		if ValueIsObject(firstArgument) {
			O := AllocateTypedArray(agent, constructorName, newTarget, proto, 0)
			firstArgumentObj := MustGetObject(firstArgument)
			if t, ok := firstArgumentObj.(*TypedArrayObject); ok {
				InitializeTypedArrayFromTypedArray(agent, O, t)
			} else if ab, ok := firstArgumentObj.(*ArrayBufferObject); ok {
				byteOffset := pkg.SliceSafeGet(argumentsList, 1)
				length := pkg.SliceSafeGet(argumentsList, 2)
				InitializeTypedArrayFromArrayBuffer(agent, O, ab, byteOffset, length)
			} else {
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
func AllocateTypedArray(agent *Agent, constructorName string, newTarget ObjectType, defaultProto string, length JSInt) *TypedArrayObject {
	proto := GetPrototypeFromConstructor(newTarget, defaultProto)
	obj := TypedArrayCreate(agent, constructorName, proto)
	if length != 0 {
		AllocateTypedArrayBuffer(agent, obj, length)
	}
	return obj
}

func TypedArrayCreate(agent *Agent, name string, proto ObjectType) *TypedArrayObject {
	object := &TypedArrayObject{
		Object:            NewObject(agent, proto, "TypedArray "+name),
		ViewedArrayBuffer: nil,
		TypedArrayName:    name,
	}
	internalMethods := object.InternalMethods()
	internalMethods.PreventExtensions = func(o ObjectType) bool {
		oo := o.(*TypedArrayObject)
		if !IsTypedArrayFixedLength(oo) {
			return false
		}
		return OrdinaryPreventExtensions(oo.Object)
	}
	internalMethods.GetOwnProperty = func(o ObjectType, p PropertyKey) *PropertyDescriptor {
		oo := o.(*TypedArrayObject)
		if i, ok := p.(IntegerIndexPropertyKey); ok {
			value := TypedArrayGetElement(agent, oo, i.Value)
			if value == UndefinedValue {
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
		if i, ok := p.(IntegerIndexPropertyKey); ok {
			return IsValidIntegerIndex(agent, o.(*TypedArrayObject), i.Value)
		}
		return OrdinaryHasProperty(o.(*TypedArrayObject).Object, p)
	}
	internalMethods.DefineOwnProperty = func(o ObjectType, p PropertyKey, desc *PropertyDescriptor) bool {
		if i, ok := p.(IntegerIndexPropertyKey); ok {
			if !IsValidIntegerIndex(agent, o.(*TypedArrayObject), i.Value) {
				return false
			}
			if !desc.Enumerable {
				return false
			}
			if !desc.Configurable {
				return false
			}
			if !desc.Writable {
				return false
			}
			if desc.IsAccessorDescriptor() {
				return false
			}
			if desc.Value != nil {
				TypedArraySetElement(agent, o.(*TypedArrayObject), i.Value, desc.Value)
			}
		}
		return true
	}
	internalMethods.Get = func(o ObjectType, p PropertyKey, receiver Value) Value {
		if i, ok := p.(IntegerIndexPropertyKey); ok {
			return TypedArrayGetElement(agent, o.(*TypedArrayObject), i.Value)
		}
		return OrdinaryGet(o.(*TypedArrayObject).Object, p, receiver)
	}
	internalMethods.Set = func(o ObjectType, p PropertyKey, v Value, receiver Value) bool {
		if i, ok := p.(IntegerIndexPropertyKey); ok {
			TypedArraySetElement(agent, o.(*TypedArrayObject), i.Value, v)
			return true
		}
		return OrdinarySet(o.(*TypedArrayObject).Object, p, v, receiver)
	}
	internalMethods.Delete = func(o ObjectType, p PropertyKey) bool {
		if i, ok := p.(IntegerIndexPropertyKey); ok {
			return !IsValidIntegerIndex(agent, o.(*TypedArrayObject), i.Value)
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

func TypedArraySetElement(agent *Agent, O *TypedArrayObject, index JSInt, value Value) {
	if !IsValidIntegerIndex(agent, O, index) {
		return
	}
	var numValue Value
	if O.ContentType == TypedArrayContentTypeBigInt {
		numValue = ToBigInt(agent, value)
	} else {
		numValue = ToNumber(agent, value)
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
		false,
	)
}

func IsValidIntegerIndex(agent *Agent, O *TypedArrayObject, index JSInt) bool {
	if IsDetachedBuffer(O.ViewedArrayBuffer) {
		return false
	}
	if index < 0 {
		return false
	}
	if index >= O.ArrayLength.Value {
		return false
	}
	return true
}

func TypedArrayGetElement(agent *Agent, O *TypedArrayObject, index JSInt) Value {
	if !IsValidIntegerIndex(agent, O, index) {
		return UndefinedValue
	}
	offset := O.ByteOffset
	elementSize := TypedArrayElementSize(O)
	byteIndexInBuffer := offset + index*elementSize
	return NewNumberValue(GetValueFromBuffer(agent, O.ViewedArrayBuffer, byteIndexInBuffer, elementSize, true, Relaxed, false).ToNumber())
}

func AllocateTypedArrayBuffer(agent *Agent, O *TypedArrayObject, length JSInt) {
	realm := agent.CurrentRealm()
	elementSize := TypedArrayElementSize(O)
	byteLength := length * elementSize
	data := AllocateArrayBuffer(agent, realm.Intrinsics.ArrayBufferConstructor, byteLength, 0)
	O.ViewedArrayBuffer = data.Data().(*ArrayBufferObject)
	O.ByteLength = NewByteLength(byteLength)
	O.ByteOffset = 0
	O.ArrayLength = NewByteLength(length)
}

func TypedArrayElementType(O *TypedArrayObject) string {
	return O.TypedArrayName
}

func getTypedArraySizeFromName(name string) JSInt {
	switch name {
	case "Int8Array":
		return 1
	case "Uint8Array":
		return 1
	case "Uint8ClampedArray":
		return 1
	case "Int16Array":
		return 2
	case "Uint16Array":
		return 2
	case "Int32Array":
		return 4
	case "Uint32Array":
		return 4
	case "BigInt64Array":
		return 8
	case "BigUint64Array":
		return 8
	case "Float32Array":
		return 4
	case "Float64Array":
		return 8
	default:
		panic("unreachable")
	}
}

func TypedArrayElementSize(O *TypedArrayObject) JSInt {
	return getTypedArraySizeFromName(O.TypedArrayName)
}

func InitializeTypedArrayFromTypedArray(agent *Agent, O, srcArray *TypedArrayObject) CompletionObject {
	realm := agent.CurrentRealm()
	srcData := srcArray.ViewedArrayBuffer
	elementSize := TypedArrayElementSize(O)
	elementType := O.TypedArrayName
	srcElementSize := TypedArrayElementSize(srcArray)
	srcType := srcArray.TypedArrayName
	srcByteOffset := srcArray.ByteOffset
	srcRecord := MakeTypedArrayWithBufferWitnessRecord(srcArray, SeqCst)
	if IsTypedArrayOutOfBounds(srcRecord) {
		return NewCompletionObjectError(agent.ThrowException(RangeError, "out of bounds"))
	}
	elementLength := TypedArrayLength(srcRecord)
	byteLength := elementLength * elementSize
	var data CompletionObject
	if elementType == srcType {
		data = CloneArrayBuffer(agent, srcData, srcByteOffset, byteLength)
	} else {
		data = AllocateArrayBuffer(agent, realm.Intrinsics.ArrayBufferConstructor, byteLength, 0)
		if srcArray.ContentType != O.ContentType {
			return NewCompletionObjectError(agent.ThrowException(TypeError, "different content type"))
		}
		srcByteIndex := srcByteOffset
		targetByteIndex := JSInt(0)
		count := elementLength
		for count > 0 {
			value := GetValueFromBuffer(agent, srcData, srcByteIndex, getTypedArraySizeFromName(srcType), true, Relaxed, false)
			SetValueInBuffer(
				agent,
				data.Data().(*ArrayBufferObject),
				targetByteIndex,
				NewNumberValue(value.ToNumber()),
				getTypedArraySizeFromName(srcType),
				true,
				Relaxed,
				false)
			srcByteIndex += srcElementSize
			targetByteIndex += elementSize
			count--
		}
	}

	O.ViewedArrayBuffer = data.Data().(*ArrayBufferObject)
	O.ByteLength = NewByteLength(byteLength)
	O.ByteOffset = 0
	O.ArrayLength = NewByteLength(elementLength)
	// return UNUSED
	return NewCompletionObjectNull()
}

func InitializeTypedArrayFromArrayBuffer(agent *Agent, O *TypedArrayObject, buffer *ArrayBufferObject, byteOffset, length Value) CompletionObject {
	elementSize := TypedArrayElementSize(O)
	offset := ToIndex(agent, byteOffset)
	if offset%elementSize != 0 {
		return NewCompletionObjectError(agent.ThrowException(RangeError, "offset is not a multiple of element size"))
	}
	bufferIsFixedLength := IsFixedLengthArrayBuffer(buffer)
	var newLength JSInt
	if length == UndefinedValue {
		newLength = ToIndex(agent, length)
	}
	if IsDetachedBuffer(buffer) {
		return NewCompletionObjectError(agent.ThrowException(TypeError, "detached buffer"))
	}
	bufferByteLength := ArrayBufferByteLength(buffer, SeqCst)
	if length == UndefinedValue && !bufferIsFixedLength {
		if offset > bufferByteLength {
			return NewCompletionObjectError(agent.ThrowException(RangeError, "offset > bufferByteLength"))
		}
		O.ByteLength.toAuto()
		O.ArrayLength.toAuto()
	} else {
		var byteLength JSInt
		if length == UndefinedValue {
			if bufferByteLength%elementSize != 0 {
				return NewCompletionObjectError(agent.ThrowException(RangeError, "bufferByteLength is not a multiple of element size"))
			}
			byteLength = bufferByteLength - offset
		} else {
			newByteLength := newLength * elementSize
			if offset+newByteLength > bufferByteLength {
				return NewCompletionObjectError(agent.ThrowException(RangeError, "out of bounds"))
			}
			byteLength = newByteLength
		}
		O.ByteLength = NewByteLength(byteLength)
		O.ArrayLength = NewByteLength(byteLength / elementSize)
	}
	O.ViewedArrayBuffer = buffer
	O.ByteOffset = offset
	// return UNUSED
	return NewCompletionObjectNull()
}

// MARK: - TypedArrayWithBufferWitnessRecord

type TypedArrayWithBufferWitnessRecord struct {
	// [[Object]]
	TypedArray *TypedArrayObject
	Order      MemoryOrder
	// [[CachedBufferByteLength]]
	CachedBufferByteLength *CachedBufferByteLength
}

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
		Order:                  order,
		CachedBufferByteLength: cachedBufferByteLength,
	}
}

func IsTypedArrayOutOfBounds(taRecord *TypedArrayWithBufferWitnessRecord) bool {
	// TODO
	return false
}

func IsTypedArrayFixedLength(O *TypedArrayObject) bool {
	if O.ArrayLength.Auto {
		return false
	}
	buffer := O.ViewedArrayBuffer
	if !IsFixedLengthArrayBuffer(buffer) {
		return false
	}
	return true
}

func InitializeTypedArrayFromArrayLike(agent *Agent, O *TypedArrayObject, arrayLike ObjectType) {
	length := arrayLike.LengthOfArrayLike()
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
	// TODO
	return 0
}
