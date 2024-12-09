package coldmoon

import (
	"math"
	"strconv"
	"strings"

	"github.com/Seeingu/coldmoon/pkg"
	"github.com/samber/lo"
)

type ArrayObject struct {
	*Object
}

func getArrayLength(array ObjectType) JSInt {
	lengthDesc := OrdinaryGetOwnProperty(array, NewStringPropertyKey("length"))
	Assert(lengthDesc.IsDataDescriptor())
	return lengthDesc.Value.(*NumberValue).Data.ToInt()
}

// 10.4.2.2
func ArrayCreate(agent *Agent, length JSInt, proto ObjectType) ObjectType {
	realm := agent.CurrentRealm()
	// 10.4.2.1
	defineOwnProperty := func(array ObjectType, p PropertyKey, desc *PropertyDescriptor) bool {
		propertyKeyString, ok := p.(StringPropertyKey)
		if ok && propertyKeyString.Value == "length" {
			return ArraySetLength(agent, array, desc)
		}
		var index JSInt
		if propertyKeyIndex, err := strconv.ParseFloat(propertyKeyString.Value, 64); err != nil {
			intValue, ok := p.(IntegerIndexPropertyKey)
			if !ok {
				panic("unexpected")
			}
			index = intValue.Value
		} else {
			index = JSInt(propertyKeyIndex)
		}
		lengthDesc := OrdinaryGetOwnProperty(array, NewStringPropertyKey("length"))
		Assert(lengthDesc.IsDataDescriptor())
		Assert(!lengthDesc.Configurable)

		lengthValue := lengthDesc.Value
		length := lengthValue.(*NumberValue).Data.ToInt()
		Assert(length.IsInf() || length >= 0)

		if index >= length && !lengthDesc.Writable {
			return false
		}

		succeeded := OrdinaryDefineOwnProperty(array, p, desc)

		if !succeeded {
			return false
		}

		if index >= length {
			lengthDesc.Value = NewNumberValue((index + 1).ToNumber())

			succeeded = OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), lengthDesc)
			Assert(succeeded)
		}
		return true
	}

	if float64(length) > POW_2_32-1 {
		panic("RangeError")
	}

	if proto == nil {
		proto = realm.Intrinsics.ArrayPrototype
	}

	arr := &ArrayObject{
		Object: NewObject(agent, proto, "Array"),
	}
	arr.Object.InternalMethods().DefineOwnProperty = defineOwnProperty
	OrdinaryDefineOwnProperty(arr, NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(length.ToNumber()),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	return arr
}

// 10.4.2.3
func ArraySpeciesCreate(agent *Agent, originalArray ObjectType, length JSInt) ObjectType {
	isArray := IsArray(NewValueFromObject(originalArray))
	if !isArray {
		return ArrayCreate(agent, length, nil)
	}

	c := originalArray.Get(NewStringPropertyKey("constructor"))
	constructorObject, isObject := c.(*ObjectValue)
	if IsConstructor(c) {
		thisRealm := agent.CurrentRealm()
		realmC := constructorObject.Object.GetFunctionRealm()
		if thisRealm != realmC {
			if pkg.FuncEqual(constructorObject.Object, realmC.Intrinsics.ArrayConstructor) {
				c = UndefinedValue
			}
		}
	}

	if isObject {
		c = constructorObject.Object.Get(NewStringPropertyKey("Symbol.species"))
		if c == NullValue {
			c = UndefinedValue
		}
	}
	if c == UndefinedValue {
		return ArrayCreate(agent, length, nil)
	}
	if !IsConstructor(c) {
		panic("TypeError")
	}

	return constructorObject.Object.Construct([]Value{NewNumberValue(length.ToNumber())}, nil)
}

// 10.4.2.4
func ArraySetLength(agent *Agent, array ObjectType, desc *PropertyDescriptor) bool {
	if desc.Value == nil {
		return OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), desc)
	}
	newLenDesc := desc

	newLenValue := desc.Value
	newLen := JSInt(newLenValue.(*NumberValue).Data)
	numberLen := ToNumber(agent, newLenValue)

	if JSInt(numberLen.Data) != newLen {
		panic("RangeError")
	}

	newLenDesc.Value = NewNumberValue(newLen.ToNumber())

	oldLenDesc := OrdinaryGetOwnProperty(array, NewStringPropertyKey("length"))
	Assert(oldLenDesc.IsDataDescriptor())
	Assert(!oldLenDesc.Configurable)
	oldLen := JSInt(oldLenDesc.Value.(*NumberValue).Data)

	if newLen >= oldLen {
		return OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), newLenDesc)
	}

	if !desc.Writable {
		return false
	}

	newWritable := false
	// TODO: nil
	if newLenDesc.Writable {
		newWritable = true
	} else {
		newWritable = false
		newLenDesc.Writable = true
	}

	succeeded := OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), newLenDesc)
	if !succeeded {
		return false
	}

	for k := oldLen - 1; k >= newLen; k-- {
		deleteSucceeded := array.InternalMethods().Delete(array, NewIntegerIndexPropertyKey(k))
		if !deleteSucceeded {
			newLenDesc.Value = NewNumberValue(JSNumber(k) + 1)
			if !newWritable {
				succeeded = OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), &PropertyDescriptor{
					Writable: false,
				})
				Assert(succeeded)
			}
			return false
		}
	}

	if !newWritable {
		succeeded = OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), &PropertyDescriptor{
			Writable: false,
		})
		Assert(succeeded)
	}

	return true
}

func NewArrayConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		if newTarget == nil {
			newTarget = agent.ActiveFunctionObject()
		}

		proto := GetPrototypeFromConstructor(newTarget, "%Array.prototype%")

		numberOfArgs := JSInt(len(args))
		if numberOfArgs == 0 {
			return NewValueFromObject(ArrayCreate(agent, 0, proto))
		} else if numberOfArgs == 1 {
			length := args[0]
			array := ArrayCreate(agent, 0, proto)

			var intLen JSNumber
			if _, ok := length.(*NumberValue); ok {
				array.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(0), length)
				intLen = 1
			} else {
				intLen = ToUint32(agent, length).ToNumber()
			}

			array.Set(
				NewStringPropertyKey("length"),
				NewNumberValue(intLen),
				setThrowTypeThrow,
			)

			return NewValueFromObject(array)
		} else {
			Assert(numberOfArgs >= 2)
			array := ArrayCreate(agent, numberOfArgs, proto)

			for k := range numberOfArgs {
				propertyKey := NewIntegerIndexPropertyKey(k)
				array.CreateDataPropertyOrThrow(propertyKey, args[k])
			}

			Assert(getArrayLength(array) == numberOfArgs)
			return NewValueFromObject(array)
		}
	}

	var isArray BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		arg := args[0]
		return NewBooleanValue(IsArray(arg))
	}
	var of BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		length := JSInt(len(args))
		lenNumber := NewNumberValue(length.ToNumber())

		constructor := this
		var array ObjectType
		if IsConstructor(constructor) {
			array = MustGetObject(constructor).Construct(args, nil)
		} else {
			array = ArrayCreate(realm.Agent, length, nil)
		}

		for k := range args {
			propertyKey := NewIntegerIndexPropertyKey(JSInt(k))
			array.CreateDataPropertyOrThrow(propertyKey, args[k])
		}

		array.Set(NewStringPropertyKey("length"), lenNumber, setThrowTypeThrow)
		return NewValueFromObject(array)
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 1, "Array", builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	DefineBuiltinFunction(object, "isArray", isArray, 1, realm)
	DefineBuiltinFunction(object, "of", of, 0, realm)

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.ArrayPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// 23.1.2.5
	var getter BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		return this
	}
	DefineBuiltinAccessor(realm, object, "@@species", getter, nil)
	DefineBuiltinPropertyV(realm.Intrinsics.ArrayPrototype, "constructor", NewValueFromObject(object))

	return object
}

func NewArrayPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := &ArrayObject{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "ArrayPrototype"),
	}

	DefineBuiltinPropertyP(object.Object, "length", &PropertyDescriptor{
		Value:        NewNumberValue(0),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})

	var arrayMap BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		callbackFn := args[0]
		thisArg := args[1]

		array := MustGetObject(this)
		length := array.LengthOfArrayLike()

		if !IsCallable(callbackFn) {
			panic("TypeError")
		}

		A := ArraySpeciesCreate(agent, array, length)
		for k := range length {
			pk := NewIntegerIndexPropertyKey(k)
			mappedValue := callbackFn.CallAssumeCallable(
				thisArg,
				[]Value{array.Get(pk), NewNumberValue(k.ToNumber()), this},
			)

			A.CreateDataPropertyOrThrow(pk, mappedValue)
		}
		return NewValueFromObject(A)
	}

	var join BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		array := MustGetObject(this)
		length := array.LengthOfArrayLike()
		sep := ","
		if len(args) > 0 {
			sep = args[0].String()
		}

		var elements []string
		for k := range length {
			element := array.Get(NewIntegerIndexPropertyKey(k))

			var next string
			if element == nil || element == UndefinedValue || element == NullValue {
			} else {
				next = element.String()
			}
			elements = append(elements, next)
		}
		return NewStringValue(strings.Join(elements, sep))
	}

	var toString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		array := ValueToObject(agent, this)
		fun := array.Get(NewStringPropertyKey("join"))
		if !IsCallable(fun) {
			fun = realm.Intrinsics.ObjectPrototype.Get(NewStringPropertyKey("toString"))
		}
		return CallAssumeCallableNoArgs(fun, this)
	}

	var forEach BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		callbackFn := args[0]
		thisArg := args[1]

		array := MustGetObject(this)
		length := array.LengthOfArrayLike()

		if !IsCallable(callbackFn) {
			panic("TypeError")
		}

		for k := range length {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := array.HasProperty(pk)
			if kPresent {
				kValue := array.Get(pk)
				callbackFn.CallAssumeCallable(
					thisArg,
					[]Value{kValue, NewNumberValue(k.ToNumber()), this},
				)
			}
		}
		return UndefinedValue
	}
	var push BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		array := MustGetObject(this)
		length := array.LengthOfArrayLike()
		argCount := len(args)
		for i := 0; i < argCount; i++ {
			array.Set(NewIntegerIndexPropertyKey(length), args[i], setThrowTypeThrow)
			length++
		}

		array.Set(NewStringPropertyKey("length"), NewNumberValue(length.ToNumber()), setThrowTypeThrow)
		return NewNumberValue(length.ToNumber())
	}
	var pop BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		array := MustGetObject(this)
		length := array.LengthOfArrayLike()
		if length == 0 {
			array.Set(NewStringPropertyKey("length"), NewNumberValue(0), setThrowTypeThrow)
			return UndefinedValue
		}
		length--
		element := array.Get(NewIntegerIndexPropertyKey(length))
		deleteSucceeded := array.DeletePropertyOrThrow(NewIntegerIndexPropertyKey(length))
		if !deleteSucceeded {
			panic("TypeError")
		}
		array.Set(NewStringPropertyKey("length"), NewNumberValue(length.ToNumber()), setThrowTypeThrow)
		return element
	}
	var toLocaleString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		array := MustGetObject(this)
		length := array.LengthOfArrayLike()
		separator := ", "
		var elements []string
		for k := range length {
			nextElement := array.Get(NewIntegerIndexPropertyKey(k))
			if nextElement == nil || nextElement == UndefinedValue {
				elements = append(elements, "")
			} else {
				s := ValueInvoke(agent, nextElement, NewStringPropertyKey("toLocaleString"), nil).String()
				elements = append(elements, s)
			}
		}
		return NewStringValue(strings.Join(elements, separator))
	}
	var includes BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		searchElement := args[0]
		fromIndex := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		if length == 0 {
			return FalseValue
		}
		n := ToIntegerOrInfinity(agent, fromIndex)
		if fromIndex == UndefinedValue {
			Assert(n == 0)
		}
		if n.IsPositiveInf() {
			return FalseValue
		} else if n.IsNegInf() {
			n = 0
		}

		k := n
		if k < 0 {
			k = (length + k).Max(0)
		}

		for k < length {
			elementK := o.Get(NewIntegerIndexPropertyKey(k))
			if SameValueZero(searchElement, elementK) {
				return TrueValue
			}
			k++
		}
		return FalseValue
	}
	var indexOf BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		searchElement := args[0]
		fromIndex := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		if length == 0 {
			return NewNumberValue(-1)
		}
		n := ToIntegerOrInfinity(agent, fromIndex)
		if fromIndex == UndefinedValue {
			Assert(n == 0)
		}
		if n.IsPositiveInf() {
			return NewNumberValue(-1)
		} else if n.IsNegInf() {
			n = 0
		}

		k := n.Max(0)
		if k < 0 {
			k = (length + k).Max(0)
		}

		for k < length {
			kPresent := o.HasProperty(NewIntegerIndexPropertyKey(k))
			if kPresent {
				elementK := o.Get(NewIntegerIndexPropertyKey(k))
				if IsStrictlyEqual(searchElement, elementK) {
					return NewNumberValue(k.ToNumber())
				}
			}
			k++
		}
		return NewNumberValue(-1)
	}
	var find BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		predicate := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		findRec := o.(*ArrayObject).findViaPredicate(length, DirectionAscending, predicate, thisArg)
		return findRec.value
	}
	var findIndex BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		predicate := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		findRec := o.(*ArrayObject).findViaPredicate(length, DirectionAscending, predicate, thisArg)
		return NewNumberValue(findRec.index.ToNumber())
	}
	var findLast BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		predicate := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		findRec := o.(*ArrayObject).findViaPredicate(length, DirectionDescending, predicate, thisArg)
		return findRec.value
	}
	var findLastIndex BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		predicate := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		findRec := o.(*ArrayObject).findViaPredicate(length, DirectionDescending, predicate, thisArg)
		return NewNumberValue(findRec.index.ToNumber())
	}
	var lastIndexOf BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		searchElement := args[0]
		fromIndex := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		if length == 0 {
			return NewNumberValue(-1)
		}
		var n JSInt
		if len(args) > 1 {
			n = ToIntegerOrInfinity(agent, fromIndex)
		} else {
			n = length - 1
		}

		if n.IsNegInf() {
			return NewNumberValue(-1)
		}

		k := n.Max(0).Min(length - 1)
		for k >= 0 {
			kPresent := o.HasProperty(NewIntegerIndexPropertyKey(k))
			if kPresent {
				elementK := o.Get(NewIntegerIndexPropertyKey(k))
				if IsStrictlyEqual(searchElement, elementK) {
					return NewNumberValue(k.ToNumber())
				}
			}
			k--
		}
		return NewNumberValue(-1)
	}
	var at BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		index := args[0]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		relativeIndex := ToIntegerOrInfinity(agent, index)
		k := relativeIndex
		if k < 0 {
			k += length
		}
		return o.Get(NewIntegerIndexPropertyKey(k))
	}
	var every BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		callbackFn := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()

		if !IsCallable(callbackFn) {
			panic("TypeError")
		}

		for k := JSInt(0); k < length; k++ {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				testResult := callbackFn.CallAssumeCallable(thisArg, []Value{kValue, NewNumberValue(k.ToNumber()), this})
				if !testResult.ToBoolean() {
					return FalseValue
				}
			}
		}
		return TrueValue
	}
	var some BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		callbackFn := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()

		if !IsCallable(callbackFn) {
			panic("TypeError")
		}

		for k := JSInt(0); k < length; k++ {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				testResult := callbackFn.CallAssumeCallable(thisArg, []Value{kValue, NewNumberValue(k.ToNumber()), this})
				if testResult.ToBoolean() {
					return TrueValue
				}
			}
		}
		return FalseValue
	}
	var with BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		index := args[0]
		value := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		relativeIndex := ToIntegerOrInfinity(agent, index)

		actualIndex := relativeIndex
		if actualIndex < 0 {
			actualIndex += length
		}

		array := ArrayCreate(agent, length, nil)
		for k := JSInt(0); k < length; k++ {
			pk := NewIntegerIndexPropertyKey(k)
			if k == actualIndex {
				array.CreateDataPropertyOrThrow(pk, value)
			} else {
				array.CreateDataPropertyOrThrow(pk, o.Get(pk))
			}
		}
		return NewValueFromObject(array)
	}
	var from BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		items := args[0]
		mapFn := args[1]
		thisArg := args[2]

		c := this
		var mapping bool
		if mapFn == nil || mapFn == UndefinedValue {
			mapping = false
		} else {
			if !IsCallable(mapFn) {
				panic("TypeError")
			}
			mapping = true
		}
		usingIterator := GetMethod(agent, items, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]))
		if usingIterator != nil {
			var a ObjectType
			if IsConstructor(c) {
				a = MustGetObject(c).Construct([]Value{NewNumberValue(0)}, nil)
			} else {
				a = ArrayCreate(agent, 0, nil)
			}

			iteratorRecord := GetIteratorFromMethod(agent, items, usingIterator)

			for k := JSInt(0); ; k++ {
				pk := NewIntegerIndexPropertyKey(k)
				next := iteratorRecord.IteratorStep()
				if next == nil {
					a.Set(NewStringPropertyKey("length"), NewNumberValue(k.ToNumber()), setThrowTypeThrow)
					return NewValueFromObject(a)
				}

				nextValue := IteratorValue(next)
				var mappedValue Value
				if mapping {
					mappedValue = mapFn.CallAssumeCallable(thisArg, []Value{nextValue, NewNumberValue(k.ToNumber())})
				} else {
					mappedValue = nextValue
				}
				a.CreateDataPropertyOrThrow(pk, mappedValue)
			}

		}
		arrayLike := ValueToObject(agent, items)
		length := arrayLike.LengthOfArrayLike()
		var a ObjectType
		if IsConstructor(c) {
			a = MustGetObject(c).Construct([]Value{NewNumberValue(length.ToNumber())}, nil)
		} else {
			a = ArrayCreate(agent, length, nil)
		}

		for k := JSInt(0); k < length; k++ {
			pk := NewIntegerIndexPropertyKey(k)
			kValue := arrayLike.Get(pk)
			var mappedValue Value
			if mapping {
				mappedValue = mapFn.CallAssumeCallable(thisArg, []Value{kValue, NewNumberValue(k.ToNumber())})
			} else {
				mappedValue = kValue
			}
			a.CreateDataPropertyOrThrow(pk, mappedValue)
		}
		a.Set(NewStringPropertyKey("length"), NewNumberValue(length.ToNumber()), setThrowTypeThrow)
		return NewValueFromObject(a)
	}
	var entries BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		return NewValueFromObject(CreateArrayIterator(agent, o.(*ArrayObject), objectOwnPropertiesKindKeyAndValue))
	}
	var keys BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		return NewValueFromObject(CreateArrayIterator(agent, o.(*ArrayObject), objectOwnPropertiesKindKey))
	}
	var values BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		return NewValueFromObject(CreateArrayIterator(agent, o.(*ArrayObject), objectOwnPropertiesKindValue))
	}
	var shift BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		if length == 0 {
			o.Set(NewStringPropertyKey("length"), NewNumberValue(0), setThrowTypeThrow)
			return UndefinedValue
		}
		first := o.Get(NewIntegerIndexPropertyKey(0))
		for k := JSInt(1); k < length; k++ {
			from := NewIntegerIndexPropertyKey(k)
			to := NewIntegerIndexPropertyKey(k - 1)
			fromPresent := o.HasProperty(from)
			if fromPresent {
				fromValue := o.Get(from)
				o.Set(to, fromValue, setThrowTypeThrow)
			} else {
				o.DeletePropertyOrThrow(to)
			}
		}
		deleteSucceeded := o.DeletePropertyOrThrow(NewIntegerIndexPropertyKey(length - 1))
		if !deleteSucceeded {
			panic("TypeError")
		}
		o.Set(NewStringPropertyKey("length"), NewNumberValue((length - 1).ToNumber()), setThrowTypeThrow)
		return first
	}
	var unshift BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		argCount := JSInt(len(args))
		if argCount == 0 {
			return NewNumberValue(length.ToNumber())
		}

		k := length
		for k > 0 {
			k--
			from := NewIntegerIndexPropertyKey(k - 1)
			to := NewIntegerIndexPropertyKey(k + argCount - 1)
			fromPresent := o.HasProperty(from)
			if fromPresent {
				fromValue := o.Get(from)
				o.Set(to, fromValue, setThrowTypeThrow)
			} else {
				o.DeletePropertyOrThrow(to)
			}
		}
		for j, arg := range args {
			key := NewIntegerIndexPropertyKey(JSInt(j))
			o.Set(key, arg, setThrowTypeThrow)
		}
		newLength := (length + argCount).ToNumber()
		o.Set(NewStringPropertyKey("length"), NewNumberValue(newLength), setThrowTypeThrow)
		return NewNumberValue(newLength)
	}
	var filter BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		callbackFn := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		if !IsCallable(callbackFn) {
			panic("TypeError")
		}
		A := ArraySpeciesCreate(agent, o, 0)
		k := JSInt(0)
		to := JSInt(0)
		for k < length {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				selected := callbackFn.CallAssumeCallable(thisArg, []Value{kValue, NewNumberValue(k.ToNumber()), this})
				if selected.ToBoolean() {
					A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(to), kValue)
					to++
				}
			}
			k++
		}
		return NewValueFromObject(A)
	}
	var reduce BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		callbackFn := args[0]
		initialValue := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		if !IsCallable(callbackFn) {
			panic("TypeError")
		}
		if length == 0 && initialValue == UndefinedValue {
			panic("TypeError")
		}
		k := JSInt(0)
		var accumulator Value
		if initialValue == UndefinedValue {
			kPresent := false
			for {
				pk := NewIntegerIndexPropertyKey(k)
				kPresent = o.HasProperty(pk)
				if kPresent {
					accumulator = o.Get(pk)
					k++
					break
				}
				k++
				if k >= length {
					panic("TypeError")
				}
			}
		} else {
			accumulator = initialValue
		}
		for k < length {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				accumulator = callbackFn.CallAssumeCallable(UndefinedValue, []Value{accumulator, kValue, NewNumberValue(k.ToNumber()), this})
			}
			k++
		}
		return accumulator
	}
	var reduceRight BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		callbackFn := args[0]
		initialValue := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		if !IsCallable(callbackFn) {
			panic("TypeError")
		}
		if length == 0 && initialValue == UndefinedValue {
			panic("TypeError")
		}
		k := JSInt(length) - 1
		var accumulator Value
		if initialValue == UndefinedValue {
			kPresent := false
			for {
				pk := NewIntegerIndexPropertyKey(k)
				kPresent = o.HasProperty(pk)
				if kPresent {
					accumulator = o.Get(pk)
					k--
					break
				}
				k--
				if k < 0 {
					panic("TypeError")
				}
			}
		} else {
			accumulator = initialValue
		}
		for k >= 0 {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				accumulator = callbackFn.CallAssumeCallable(UndefinedValue, []Value{accumulator, kValue, NewNumberValue(k.ToNumber()), this})
			}
			k--
		}
		return accumulator
	}
	var concat BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		A := ArraySpeciesCreate(agent, o, 0)
		n := JSInt(0)

		for index := range len(args) + 1 {
			var element Value
			if index == 0 {
				element = NewValueFromObject(o)
			} else {
				element = args[index-1]
			}

			spreadable := IsConcatSpreadable(agent, element)
			if spreadable {
				length := MustGetObject(element).LengthOfArrayLike()
				if float64(n)+float64(length) > POW_2_53-1 {
					panic("TypeError")
				}

				k := JSInt(0)
				for k < length {
					pk := NewIntegerIndexPropertyKey(k)
					kPresent := MustGetObject(element).HasProperty(pk)
					if kPresent {
						kValue := MustGetObject(element).Get(pk)
						A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(n), kValue)
					}
					k++
					n++
				}
			} else {
				A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(n), element)
				n++
			}
		}

		A.Set(NewStringPropertyKey("length"), NewNumberValue(n.ToNumber()), setThrowTypeThrow)
		return NewValueFromObject(A)
	}
	var slice BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		start := args[0]
		end := args[1]

		relativeStart := ToIntegerOrInfinity(agent, start)

		var k JSInt
		if relativeStart.IsNegInf() {
			k = 0
		} else if relativeStart < 0 {
			k = (length + relativeStart).Max(0)
		} else {
			k = relativeStart.Min(length)
		}

		var relativeEnd JSInt
		if end == UndefinedValue {
			relativeEnd = length
		} else {
			relativeEnd = ToIntegerOrInfinity(agent, end)
		}

		var final JSInt
		if relativeEnd.IsNegInf() {
			final = 0
		} else if relativeEnd < 0 {
			final = JSInt(math.Max(float64(length+relativeEnd), 0))
		} else {
			final = JSInt(math.Min(float64(relativeEnd), float64(length)))
		}
		count := (final - k).Max(0)

		n := JSInt(0)
		A := ArraySpeciesCreate(agent, o, count)
		for k < final {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(n), kValue)
			}
			k++
			n++
		}
		A.Set(NewStringPropertyKey("length"), NewNumberValue(n.ToNumber()), setThrowTypeThrow)
		return NewValueFromObject(A)
	}
	var fill BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		value := args[0]
		start := args[1]
		end := args[2]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()

		relativeStart := ToIntegerOrInfinity(agent, start)
		k := relativeStart.Max(0)
		if end == UndefinedValue {
			end = NewNumberValue(length.ToNumber())
		}
		var relativeEnd JSInt
		if end == UndefinedValue {
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
		for k < final {
			pk := NewIntegerIndexPropertyKey(k)
			o.Set(pk, value, setThrowTypeThrow)
			k++
		}
		return NewValueFromObject(o)
	}
	var copyWithin BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		target := args[0]
		start := args[1]
		end := args[2]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()

		relativeTarget := ToIntegerOrInfinity(agent, target)
		var to JSInt
		if relativeTarget.IsNegInf() {
			to = 0
		} else if relativeTarget < 0 {
			to = (length + relativeTarget).Max(0)
		} else {
			to = relativeTarget.Min(length)
		}

		relativeStart := ToIntegerOrInfinity(agent, start)
		var from JSInt
		if relativeStart.IsNegInf() {
			from = 0
		} else if relativeStart < 0 {
			from = (length + relativeStart).Max(0)
		} else {
			from = relativeStart.Min(length)
		}

		var relativeEnd JSInt
		if end == UndefinedValue {
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

		count := (final - from).Min(length - to)

		var direction int
		if from < to && to < from+count {
			direction = -1
			from += count - 1
			to += count - 1
		} else {
			direction = 1
		}

		for count > 0 {
			fromKey := NewIntegerIndexPropertyKey(from)
			toKey := NewIntegerIndexPropertyKey(to)
			fromPresent := o.HasProperty(fromKey)
			if fromPresent {
				fromValue := o.Get(fromKey)
				o.Set(toKey, fromValue, setThrowTypeThrow)
			} else {
				o.DeletePropertyOrThrow(toKey)
			}
			from += JSInt(direction)
			to += JSInt(direction)
			count--
		}
		return NewValueFromObject(o)
	}
	var reverse BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		middle := length / 2
		lower := JSInt(0)
		for lower < middle {
			upper := length - lower - 1
			lowerP := NewIntegerIndexPropertyKey(lower)
			upperP := NewIntegerIndexPropertyKey(upper)
			lowerExists := o.HasProperty(lowerP)
			upperExists := o.HasProperty(upperP)

			var lowerValue Value = UndefinedValue
			if lowerExists {
				lowerValue = o.Get(lowerP)
			}
			var upperValue Value = UndefinedValue
			if upperExists {
				upperValue = o.Get(upperP)
			}
			if lowerExists && upperExists {
				o.Set(lowerP, upperValue, setThrowTypeThrow)
				o.Set(upperP, lowerValue, setThrowTypeThrow)
			} else if !lowerExists && upperExists {
				o.Set(lowerP, o.Get(upperP), setThrowTypeThrow)
				o.DeletePropertyOrThrow(upperP)
			} else if lowerExists && !upperExists {
				o.Set(upperP, o.Get(lowerP), setThrowTypeThrow)
				o.DeletePropertyOrThrow(lowerP)
			}
			lower++
		}
		return NewValueFromObject(o)
	}
	var toReversed BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		A := ArrayCreate(agent, length, nil)
		for k := JSInt(0); k < length; k++ {
			from := NewIntegerIndexPropertyKey(length - k - 1)
			fromValue := o.Get(from)
			pk := NewIntegerIndexPropertyKey(k)
			A.CreateDataPropertyOrThrow(pk, fromValue)
		}
		return NewValueFromObject(A)
	}
	var sort BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		compareFn := args[0]
		if compareFn != UndefinedValue && !IsCallable(compareFn) {
			panic("TypeError")
		}
		obj := ValueToObject(agent, this)
		length := obj.LengthOfArrayLike()

		sortCompare := SortCompare{
			compareFn: MustGetObject(compareFn),
			impl: func(agent *Agent, x Value, y Value, objectType ObjectType) int {
				return CompareArrayElements(agent, x, y, objectType)
			},
		}

		sortedList := SortIndexedProperties(agent, obj, length, sortCompare, sortHolesTypeSkipHoles)
		itemCount := JSInt(len(sortedList))

		j := JSInt(0)
		for ; j < itemCount; j++ {
			obj.Set(NewIntegerIndexPropertyKey(j), sortedList[j], setThrowTypeThrow)
		}
		for ; j < length; j++ {
			obj.DeletePropertyOrThrow(NewIntegerIndexPropertyKey(j))
		}
		return NewValueFromObject(obj)
	}
	var toSorted BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		compareFn := args[0]
		if compareFn != UndefinedValue && !IsCallable(compareFn) {
			panic("TypeError")
		}
		obj := ValueToObject(agent, this)
		length := obj.LengthOfArrayLike()

		sortCompare := SortCompare{
			compareFn: MustGetObject(compareFn),
			impl: func(agent *Agent, x Value, y Value, objectType ObjectType) int {
				return CompareArrayElements(agent, x, y, objectType)
			},
		}

		sortedList := SortIndexedProperties(agent, obj, length, sortCompare, sortHolesTypeReadThroughHoles)
		A := ArrayCreate(agent, length, nil)
		for k, v := range sortedList {
			A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(JSInt(k)), v)
		}
		return NewValueFromObject(A)
	}
	var flat BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		depth := args[0]
		o := ValueToObject(agent, this)
		sourceLen := o.LengthOfArrayLike()
		var depthNum JSInt = 1
		if depth != UndefinedValue {
			depthNum = ToIntegerOrInfinity(agent, depth)
			if depthNum < 0 {
				depthNum = 0
			}
		}
		A := ArraySpeciesCreate(agent, o, 0)
		FlattenIntoArray(agent, A, o, sourceLen, 0, depthNum, nil, nil)
		return NewValueFromObject(A)
	}
	var flatMap BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		mapperFunction := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		sourceLen := o.LengthOfArrayLike()
		if !IsCallable(mapperFunction) {
			panic("TypeError")
		}
		A := ArraySpeciesCreate(agent, o, 0)
		FlattenIntoArray(agent, A, o, sourceLen, 0, 1, MustGetObject(mapperFunction), thisArg)
		return NewValueFromObject(A)
	}
	var splice BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		start := args[0]
		deleteCount := args[1]
		var items []Value
		if len(args) > 2 {
			items = args[2:]
		}
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()

		var relativeStart JSInt = 0
		if start != nil {
			relativeStart = ToIntegerOrInfinity(agent, start)
		}
		var actualStart JSInt
		if relativeStart.IsNegInf() {
			actualStart = 0
		} else if relativeStart < 0 {
			actualStart = (length + relativeStart).Max(0)
		} else {
			actualStart = relativeStart.Min(length)
		}
		itemCount := JSInt(len(items))
		var actualDeleteCount JSInt
		if start == nil {
			actualDeleteCount = 0
		} else if deleteCount == nil {
			actualDeleteCount = length - actualStart
		} else {
			actualDeleteCount = ToIntegerOrInfinity(agent, deleteCount)
		}

		if float64(length+itemCount-actualDeleteCount) > POW_2_53-1 {
			panic("TypeError")
		}

		A := ArraySpeciesCreate(agent, o, actualDeleteCount)
		for k := JSInt(0); k < actualDeleteCount; k++ {
			from := NewIntegerIndexPropertyKey(actualStart + k)
			fromPresent := o.HasProperty(from)
			if fromPresent {
				fromValue := o.Get(from)
				A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(k), fromValue)
			}
		}
		A.Set(NewStringPropertyKey("length"), NewNumberValue(actualDeleteCount.ToNumber()), setThrowTypeThrow)
		if itemCount < actualDeleteCount {
			k := actualStart
			for k < length-actualDeleteCount {
				from := NewIntegerIndexPropertyKey(k + actualDeleteCount)
				to := NewIntegerIndexPropertyKey(k + itemCount)
				fromPresent := o.HasProperty(from)
				if fromPresent {
					fromValue := o.Get(from)
					o.Set(to, fromValue, setThrowTypeThrow)
				} else {
					o.DeletePropertyOrThrow(to)
				}
				k++
			}
			k = length
			for k > length-actualDeleteCount+itemCount {
				k--
				o.DeletePropertyOrThrow(NewIntegerIndexPropertyKey(k))
			}
		} else if itemCount > actualDeleteCount {
			k := length - actualDeleteCount
			for k > actualStart {
				from := NewIntegerIndexPropertyKey(k + actualDeleteCount - 1)
				to := NewIntegerIndexPropertyKey(k + itemCount - 1)
				fromPresent := o.HasProperty(from)
				if fromPresent {
					fromValue := o.Get(from)
					o.Set(to, fromValue, setThrowTypeThrow)
				} else {
					o.DeletePropertyOrThrow(to)
				}
				k--
			}
		}
		k := actualStart
		for _, E := range items {
			o.Set(NewIntegerIndexPropertyKey(k), E, setThrowTypeThrow)
			k++
		}
		o.Set(NewStringPropertyKey("length"), NewNumberValue((length - actualDeleteCount + itemCount).ToNumber()), setThrowTypeThrow)
		return NewValueFromObject(A)
	}
	var toSpliced BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		start := args[0]
		skipCount := args[1]
		var items []Value
		if len(args) > 2 {
			items = args[2:]
		}
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()

		var relativeStart JSInt = 0
		if start != nil {
			relativeStart = ToIntegerOrInfinity(agent, start)
		}
		var actualStart JSInt
		if relativeStart.IsNegInf() {
			actualStart = 0
		} else if relativeStart < 0 {
			actualStart = (length + relativeStart).Max(0)
		} else {
			actualStart = relativeStart.Min(length)
		}
		insertCount := JSInt(len(items))
		var actualSkipCount JSInt
		if start == nil {
			actualSkipCount = 0
		} else if skipCount == nil {
			actualSkipCount = length - actualStart
		} else {
			sc := ToIntegerOrInfinity(agent, skipCount)
			actualSkipCount = lo.Clamp(sc, 0, length-actualStart)
		}

		newLen := length + insertCount - actualSkipCount

		A := ArrayCreate(agent, newLen, nil)
		i := JSInt(0)
		r := actualStart + actualSkipCount
		for ; i < actualStart; i++ {
			from := NewIntegerIndexPropertyKey(i)
			fromValue := o.Get(from)
			A.CreateDataPropertyOrThrow(from, fromValue)
		}
		for _, E := range items {
			A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(i), E)
			i++
		}
		for ; i < newLen; i++ {
			from := NewIntegerIndexPropertyKey(r)
			fromValue := o.Get(from)
			A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(i), fromValue)
			r++
		}

		return NewValueFromObject(A)
	}

	DefineBuiltinFunction(object, "join", join, 1, realm)
	DefineBuiltinFunction(object, "toString", toString, 0, realm)
	DefineBuiltinFunction(object, "forEach", forEach, 1, realm)
	DefineBuiltinFunction(object, "push", push, 1, realm)
	DefineBuiltinFunction(object, "pop", pop, 0, realm)
	DefineBuiltinFunction(object, "map", arrayMap, 1, realm)
	DefineBuiltinFunction(object, "toLocaleString", toLocaleString, 0, realm)
	DefineBuiltinFunction(object, "includes", includes, 1, realm)
	DefineBuiltinFunction(object, "indexOf", indexOf, 1, realm)
	DefineBuiltinFunction(object, "find", find, 1, realm)
	DefineBuiltinFunction(object, "findIndex", findIndex, 1, realm)
	DefineBuiltinFunction(object, "findLast", findLast, 1, realm)
	DefineBuiltinFunction(object, "findLastIndex", findLastIndex, 1, realm)
	DefineBuiltinFunction(object, "lastIndexOf", lastIndexOf, 1, realm)
	DefineBuiltinFunction(object, "at", at, 1, realm)
	DefineBuiltinFunction(object, "every", every, 1, realm)
	DefineBuiltinFunction(object, "some", some, 1, realm)
	DefineBuiltinFunction(object, "with", with, 2, realm)
	DefineBuiltinFunction(object, "from", from, 1, realm)
	DefineBuiltinFunction(object, "entries", entries, 0, realm)
	DefineBuiltinFunction(object, "keys", keys, 0, realm)
	DefineBuiltinFunction(object, "values", values, 0, realm)
	DefineBuiltinFunction(object, "shift", shift, 0, realm)
	DefineBuiltinFunction(object, "unshift", unshift, 1, realm)
	DefineBuiltinFunction(object, "filter", filter, 1, realm)
	DefineBuiltinFunction(object, "reduce", reduce, 1, realm)
	DefineBuiltinFunction(object, "reduceRight", reduceRight, 1, realm)
	DefineBuiltinFunction(object, "concat", concat, 1, realm)
	DefineBuiltinFunction(object, "slice", slice, 2, realm)
	DefineBuiltinFunction(object, "fill", fill, 1, realm)
	DefineBuiltinFunction(object, "copyWithin", copyWithin, 2, realm)
	DefineBuiltinFunction(object, "reverse", reverse, 0, realm)
	DefineBuiltinFunction(object, "toReversed", toReversed, 0, realm)
	DefineBuiltinFunction(object, "sort", sort, 1, realm)
	DefineBuiltinFunction(object, "toSorted", toSorted, 1, realm)
	DefineBuiltinFunction(object, "flat", flat, 0, realm)
	DefineBuiltinFunction(object, "flatMap", flatMap, 1, realm)
	DefineBuiltinFunction(object, "splice", splice, 2, realm)
	DefineBuiltinFunction(object, "toSpliced", toSpliced, 2, realm)

	var unscopablesValue Value
	unscopablesList := OrdinaryObjectCreate(agent, nil, nil)
	unscopablesProps := []string{
		"at",
		"copyWithin",
		"entries",
		"fill",
		"find",
		"findIndex",
		"findLast",
		"findLastIndex",
		"flat",
		"flatMap",
		"includes",
		"keys",
		"values",
		"toReversed",
		"toSorted",
		"toSpliced",
	}
	for _, prop := range unscopablesProps {
		unscopablesList.CreateDataPropertyOrThrow(NewStringPropertyKey(prop), NewBooleanValue(true))
	}

	DefineBuiltinPropertyP(object, "@@unscopables", &PropertyDescriptor{
		Value:        unscopablesValue,
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	DefineBuiltinPropertyP(object, "@@iterator", object.PropertyStorage().Get(NewStringPropertyKey("values")))

	return object
}

type direction int

const (
	DirectionAscending direction = iota
	DirectionDescending
)

type FoundResult struct {
	index JSInt
	value Value
}

func (a *ArrayObject) findViaPredicate(
	len JSInt,
	direction direction,
	predicate Value,
	thisArg Value,
) FoundResult {
	if !IsCallable(predicate) {
		panic("TypeError")
	}

	var k JSInt
	if direction == DirectionAscending {
		k = 0
	} else {
		k = len - 1
	}

	for {
		if direction == DirectionAscending && k >= len {
			break
		}
		if direction == DirectionDescending && k < 0 {
			break
		}
		pk := NewIntegerIndexPropertyKey(k)
		kValue := a.Get(pk)
		testResult := predicate.CallAssumeCallable(thisArg, []Value{kValue, NewNumberValue(k.ToNumber()), NewValueFromObject(a)})

		if testResult.ToBoolean() {
			return FoundResult{
				index: k,
				value: kValue,
			}
		}
	}

	return FoundResult{
		index: -1,
		value: UndefinedValue,
	}
}

// 23.1.3.2.1
func IsConcatSpreadable(agent *Agent, value Value) bool {
	if !ValueIsObject(value) {
		return false
	}
	spreadable := MustGetObject(value).Get(NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIsConcatSpreadable]))
	if spreadable != UndefinedValue {
		return spreadable.ToBoolean()
	}
	return IsArray(value)
}

type SortCompare struct {
	compareFn ObjectType
	impl      func(*Agent, Value, Value, ObjectType) int
}

type sortHolesType int

const (
	sortHolesTypeSkipHoles sortHolesType = iota
	sortHolesTypeReadThroughHoles
)

func InsertionSort(agent *Agent, items []Value, sortCompare SortCompare) {
	if len(items) == 0 {
		return
	}
	for i := 1; i < len(items); i++ {
		x := items[i]
		j := i
		for j > 0 {
			y := items[j-1]
			if sortCompare.impl(agent, x, y, sortCompare.compareFn) >= 0 {
				break
			}
			items[j] = y
			j--
		}
		items[j] = x
	}
}

// 23.1.3.30.1
func SortIndexedProperties(agent *Agent, obj ObjectType, length JSInt, sortCompare SortCompare, holes sortHolesType) (items []Value) {
	k := JSInt(0)
	for k < length {
		pk := NewIntegerIndexPropertyKey(k)
		var kRead bool
		if holes == sortHolesTypeSkipHoles {
			kRead = obj.HasProperty(pk)
		} else {
			kRead = true
		}
		if kRead {
			kValue := obj.Get(pk)
			items = append(items, kValue)
		}
		k++
	}
	InsertionSort(agent, items, sortCompare)

	return
}

// 23.1.3.30.2
func CompareArrayElements(agent *Agent, x, y Value, compareFn ObjectType) int {
	if x == UndefinedValue && y == UndefinedValue {
		return 0
	}
	if x == UndefinedValue {
		return 1
	}
	if y == UndefinedValue {
		return -1
	}
	if compareFn != nil {
		v := ToNumber(agent,
			NewValueFromObject(compareFn).CallAssumeCallable(
				UndefinedValue,
				[]Value{x, y}),
		)
		if v.IsNaN() {
			return 0
		}
		return int(v.Data)
	}

	xString := x.String()
	yString := y.String()
	xSmaller := IsLessThan(agent, NewStringValue(xString), NewStringValue(yString), IsLessThanOrderLeftFirst)
	if xSmaller {
		return -1
	}

	ySmaller := IsLessThan(agent, NewStringValue(yString), NewStringValue(xString), IsLessThanOrderLeftFirst)
	if ySmaller {
		return 1
	}
	return 0
}

func FlattenIntoArray(agent *Agent, target, source ObjectType, sourceLen JSInt, start, depth JSInt, mapperFunction ObjectType, thisArg Value) JSInt {
	if mapperFunction != nil {
		Assert(IsCallable(NewValueFromObject(mapperFunction)))
		Assert(thisArg != nil)
		Assert(depth == 1)
	}

	targetIndex := start
	sourceIndex := JSInt(0)
	for sourceIndex < sourceLen {
		p := NewIntegerIndexPropertyKey(sourceIndex)
		exists := source.HasProperty(p)
		if exists {
			element := source.Get(p)
			if mapperFunction != nil {
				element = NewValueFromObject(mapperFunction).CallAssumeCallable(
					thisArg,
					[]Value{element, NewNumberValue(sourceIndex.ToNumber()), NewValueFromObject(source)},
				)
			}

			shouldFlatten := false
			if depth > 0 {
				shouldFlatten = IsArray(element)
			}
			if shouldFlatten {
				var newDepth JSInt
				if depth.IsPositiveInf() {
					newDepth = JSInt(math.Inf(1))
				} else {
					newDepth = depth - 1
				}
				elementLen := MustGetObject(element).LengthOfArrayLike()
				targetIndex = FlattenIntoArray(agent, target, MustGetObject(element), elementLen, targetIndex, newDepth, mapperFunction, thisArg)
			} else {
				if float64(targetIndex) >= POW_2_53-1 {
					panic("TypeError")
				}
				target.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(targetIndex), element)
				targetIndex++
			}
		}
	}
	return targetIndex
}

// MARK: - Internal

func (a *ArrayObject) String() string {
	s := "ArrayObject ["
	for i := JSInt(0); i < a.LengthOfArrayLike(); i++ {
		if i > 0 {
			s += ", "
		}
		s += a.Get(NewIntegerIndexPropertyKey(i)).String()
	}
	s += "]"
	return s
}
