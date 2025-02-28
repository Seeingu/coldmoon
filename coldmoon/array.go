package coldmoon

import (
	"math"
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

// ArrayCreate
// spec: 10.4.2.2
// proto is optional
func ArrayCreate(agent *Agent, length JSInt, proto ObjectType) *ArrayObject {
	realm := agent.CurrentRealm()
	// 10.4.2.1
	defineOwnProperty := func(array ObjectType, p PropertyKey, desc *PropertyDescriptor) bool {
		propertyKeyString, ok := p.(StringPropertyKey)
		if ok && propertyKeyString.Value == "length" {
			return ReturnAssertNormal(ArraySetLength(agent, array, desc))
		}
		index, err := p.GetIndex()
		if err != nil {
			return OrdinaryDefineOwnProperty(array, p, desc)
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
	arr.ref = arr
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
	isArray := IsArray(originalArray.ToValue())
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

	return constructorObject.Object.Construct([]Value{NewNumberValue(length.ToNumber())}, nil).value
}

// ArraySetLength
// spec: 10.4.2.4
func ArraySetLength(agent *Agent, array ObjectType, desc *PropertyDescriptor) (co Completion[bool]) {
	if desc.Value == nil {
		co.value = OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), desc)
		return
	}
	newLenDesc := desc

	newLenValue := desc.Value
	newLen := JSInt(newLenValue.(*NumberValue).Data)
	numberLen, isAbrupt, rt := ReturnIfAbrupt(newLenValue.ToNumber(agent), co)
	if isAbrupt {
		return rt
	}

	if JSInt(numberLen.Data) != newLen {
		return co.ThrowError(agent, RangeError, "Invalid array length")
	}

	newLenDesc.Value = NewNumberValue(newLen.ToNumber())

	oldLenDesc := OrdinaryGetOwnProperty(array, NewStringPropertyKey("length"))
	Assert(oldLenDesc.IsDataDescriptor())
	Assert(!oldLenDesc.Configurable)
	oldLen := JSInt(oldLenDesc.Value.(*NumberValue).Data)

	if newLen >= oldLen {
		co.value = OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), newLenDesc)
		return
	}

	if !desc.Writable {
		co.value = false
		return
	}

	newWritable := false
	if newLenDesc.Writable {
		newWritable = true
	} else {
		newWritable = false
		newLenDesc.Writable = true
	}

	succeeded := OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), newLenDesc)
	if !succeeded {
		co.value = false
		return
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
			co.value = false
			return
		}
	}

	if !newWritable {
		succeeded = OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), &PropertyDescriptor{
			Writable: false,
		})
		Assert(succeeded)
	}

	co.value = true
	return
}

func NewArrayConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		if newTarget == nil {
			newTarget = agent.ActiveFunctionObject()
		}

		proto := GetPrototypeFromConstructor(newTarget, "%Array.prototype%")

		numberOfArgs := JSInt(len(args))
		if numberOfArgs == 0 {
			return ArrayCreate(agent, 0, proto).ToValue()
		} else if numberOfArgs == 1 {
			length := args[0]
			array := ArrayCreate(agent, 0, proto)

			var intLen JSNumber
			if _, ok := length.(*NumberValue); ok {
				array.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(0), length)
				intLen = 1
			} else {
				n, isAbrupt, rt := ReturnIfAbrupt(ToUint32(agent, length), co)
				if isAbrupt {
					return rt
				}
				intLen = n.ToNumber()
			}

			array.Set(
				NewStringPropertyKey("length"),
				NewNumberValue(intLen),
				setThrowTypeThrow,
			)

			return array.ToValue()
		} else {
			Assert(numberOfArgs >= 2)
			array := ArrayCreate(agent, numberOfArgs, proto)

			for k := range numberOfArgs {
				propertyKey := NewIntegerIndexPropertyKey(k)
				array.CreateDataPropertyOrThrow(propertyKey, args[k])
			}

			Assert(getArrayLength(array) == numberOfArgs)
			return array.ToValue()
		}
	}

	var isArray BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		arg := args[0]
		return NewBooleanValue(IsArray(arg))
	}
	var of BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		length := JSInt(len(args))
		lenNumber := NewNumberValue(length.ToNumber())

		constructor := this
		var array ObjectType
		if IsConstructor(constructor) {
			array = MustGetObject(constructor).Construct(args, nil).value
		} else {
			array = ArrayCreate(realm.Agent, length, nil)
		}

		for k := range args {
			propertyKey := NewIntegerIndexPropertyKey(JSInt(k))
			array.CreateDataPropertyOrThrow(propertyKey, args[k])
		}

		array.Set(NewStringPropertyKey("length"), lenNumber, setThrowTypeThrow)
		return array.ToValue()
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 1, CMString("Array"), builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	object.defineBuiltinFunction(realm, CMString("isArray"), isArray, 1)
	object.defineBuiltinFunction(realm, CMString("of"), of, 0)

	// 23.1.2.5
	var getter BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return this
	}
	object.defineBuiltinAccessor(realm, WellKnownSymbolsSpecies, builtinAccessorParams{
		Getter: getter,
	})
	// 23.1.2.1
	var from BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		items := args[0]
		mapFn := pkg.SliceSafeGet(args, 1)
		thisArg := pkg.SliceSafeGet(args, 2)

		c := this
		var mapping bool
		if mapFn == nil || mapFn == UndefinedValue {
			mapping = false
		} else {
			if !IsCallable(mapFn) {
				return co.ThrowTypeError(agent, "mapFn is not callable")
			}
			mapping = true
		}
		usingIterator := GetMethod(agent, items, NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsIterator]))
		if usingIterator != nil {
			var a ObjectType
			if IsConstructor(c) {
				a = MustGetObject(c).Construct([]Value{NewNumberValue(0)}, nil).value
			} else {
				a = ArrayCreate(agent, 0, nil)
			}

			iteratorRecord := GetIteratorFromMethod(agent, items, usingIterator)

			for k := JSInt(0); ; k++ {
				pk := NewIntegerIndexPropertyKey(k)
				next := iteratorRecord.IteratorStep()
				if next == nil {
					a.Set(NewStringPropertyKey("length"), NewNumberValue(k.ToNumber()), setThrowTypeThrow)
					return a.ToValue()
				}

				nextValue := IteratorValue(next)
				var mappedValue Value
				if mapping {
					mappedValue = ReturnAssertNormal(
						mapFn.Call(agent, thisArg, []Value{nextValue, NewNumberValue(k.ToNumber())}),
					)
				} else {
					mappedValue = nextValue
				}
				a.CreateDataPropertyOrThrow(pk, mappedValue)
			}

		}
		arrayLike := ReturnAssertNormal(items.ToObject(agent))
		length, isAbrupt, rt := ReturnIfAbrupt(arrayLike.LengthOfArrayLike(), co)
		if isAbrupt {
			return rt
		}
		var a ObjectType
		if IsConstructor(c) {
			_a, isAbrupt, rt := ReturnIfAbrupt(
				MustGetObject(c).Construct([]Value{NewNumberValue(length.ToNumber())}, nil),
				co,
			)
			if isAbrupt {
				return rt
			}
			a = _a
		} else {
			a = ArrayCreate(agent, length, nil)
		}

		for k := JSInt(0); k < length; k++ {
			pk := NewIntegerIndexPropertyKey(k)
			kValue := arrayLike.Get(pk)
			var mappedValue Value
			if mapping {
				mappedValue = ReturnAssertNormal(
					mapFn.Call(agent, thisArg, []Value{kValue, NewNumberValue(k.ToNumber())}),
				)
			} else {
				mappedValue = kValue
			}
			a.CreateDataPropertyOrThrow(pk, mappedValue)
		}
		a.Set(NewStringPropertyKey("length"), NewNumberValue(length.ToNumber()), setThrowTypeThrow)
		return a.ToValue()
	}
	object.defineBuiltinFunction(realm, CMString("from"), from, 1)
	BindPrototypeAndConstructor(realm.Intrinsics.ArrayPrototype, object)

	return object
}

func NewArrayPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := &ArrayObject{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "ArrayPrototype"),
	}
	object.ref = object

	object.defineBuiltinProperty(CMString("length"), &PropertyDescriptor{
		Value:        NewNumberValue(0),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})

	var arrayMap BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		callbackFn := args[0]
		thisArg := args[1]

		array := MustGetObject(this)
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(array.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		if !IsCallable(callbackFn) {
			panic("TypeError")
		}

		A := ArraySpeciesCreate(agent, array, length)
		for k := range length {
			pk := NewIntegerIndexPropertyKey(k)
			mappedValue := ReturnAssertNormal(callbackFn.Call(
				agent,
				thisArg,
				[]Value{array.Get(pk), NewNumberValue(k.ToNumber()), this},
			))

			A.CreateDataPropertyOrThrow(pk, mappedValue)
		}
		return A.ToValue()
	}

	var join BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		array := MustGetObject(this)
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(array.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
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

	var toString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		array := ReturnAssertNormal(this.ToObject(agent))
		fun := array.Get(NewStringPropertyKey("join"))
		if !IsCallable(fun) {
			fun = realm.Intrinsics.ObjectPrototype.Get(NewStringPropertyKey("toString"))
		}
		return fun.Call(agent, this, nil)
	}

	var forEach BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		callbackFn := args[0]
		thisArg := args[1]

		array := MustGetObject(this)
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(array.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		if !IsCallable(callbackFn) {
			panic("TypeError")
		}

		for k := range length {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := array.HasProperty(pk)
			if kPresent {
				kValue := array.Get(pk)
				callbackFn.Call(
					agent,
					thisArg,
					[]Value{kValue, NewNumberValue(k.ToNumber()), this},
				)
			}
		}
		return UndefinedValue
	}
	var push BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		array := MustGetObject(this)
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(array.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		argCount := len(args)
		for i := 0; i < argCount; i++ {
			array.Set(NewIntegerIndexPropertyKey(length), args[i], setThrowTypeThrow)
			length++
		}

		array.Set(NewStringPropertyKey("length"), NewNumberValue(length.ToNumber()), setThrowTypeThrow)
		return NewNumberValue(length.ToNumber())
	}
	var pop BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		array := MustGetObject(this)
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(array.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		if length == 0 {
			array.Set(NewStringPropertyKey("length"), NewNumberValue(0), setThrowTypeThrow)
			return UndefinedValue.ToCompletion()
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
	var toLocaleString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		array := MustGetObject(this)
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(array.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		separator := ", "
		var elements []string
		for k := range length {
			nextElement := array.Get(NewIntegerIndexPropertyKey(k))
			if nextElement == nil || nextElement == UndefinedValue {
				elements = append(elements, "")
			} else {
				s := ReturnAssertNormal(ValueInvoke(agent, nextElement, NewStringPropertyKey("toLocaleString"), nil)).String()
				elements = append(elements, s)
			}
		}
		return NewStringValue(strings.Join(elements, separator))
	}
	var includes BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		searchElement := args[0]
		fromIndex := args[1]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		if length == 0 {
			return FalseValue
		}
		n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, fromIndex), co)
		if isAbrupt {
			return rt
		}
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
	var indexOf BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		searchElement := args[0]
		fromIndex := args[1]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			return rt
		}
		if length == 0 {
			return NewNumberValue(-1)
		}
		n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, fromIndex), co)
		if isAbrupt {
			return rt
		}
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
	var find BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		predicate := args[0]
		thisArg := args[1]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		findRec := o.FindViaPredicate(length, DirectionAscending, predicate, thisArg)
		return findRec.Value
	}
	var findIndex BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		predicate := args[0]
		thisArg := args[1]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		findRec := o.FindViaPredicate(length, DirectionAscending, predicate, thisArg)
		return NewNumberValue(findRec.Index.ToNumber())
	}
	var findLast BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		predicate := args[0]
		thisArg := args[1]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		findRec := o.FindViaPredicate(length, DirectionDescending, predicate, thisArg)
		return findRec.Value
	}
	var findLastIndex BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		predicate := args[0]
		thisArg := args[1]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		findRec := o.FindViaPredicate(length, DirectionDescending, predicate, thisArg)
		return NewNumberValue(findRec.Index.ToNumber())
	}
	var lastIndexOf BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		searchElement := args[0]
		fromIndex := args[1]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		if length == 0 {
			return NewNumberValue(-1)
		}
		var n JSInt
		if len(args) > 1 {
			_n, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, fromIndex), co)
			if isAbrupt {
				return rt
			}
			n = _n
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
	var at BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		index := args[0]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		relativeIndex, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, index), co)
		if isAbrupt {
			return rt
		}
		k := relativeIndex
		if k < 0 {
			k += length
		}
		return o.Get(NewIntegerIndexPropertyKey(k))
	}
	var every BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		callbackFn := args[0]
		thisArg := args[1]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		if !IsCallable(callbackFn) {
			panic("TypeError")
		}

		for k := JSInt(0); k < length; k++ {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				testResult := ReturnAssertNormal(
					callbackFn.Call(agent, thisArg, []Value{kValue, NewNumberValue(k.ToNumber()), this}),
				)
				if !testResult.ToBoolean() {
					return FalseValue
				}
			}
		}
		return TrueValue
	}
	var some BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		callbackFn := args[0]
		thisArg := args[1]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		if !IsCallable(callbackFn) {
			panic("TypeError")
		}

		for k := JSInt(0); k < length; k++ {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				testResult := ReturnAssertNormal(
					callbackFn.Call(agent, thisArg, []Value{kValue, NewNumberValue(k.ToNumber()), this}),
				)
				if testResult.ToBoolean() {
					return TrueValue
				}
			}
		}
		return FalseValue
	}
	var with BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		index := args[0]
		value := args[1]
		o := ReturnAssertNormal(this.ToObject(agent))
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		relativeIndex, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, index), co)
		if isAbrupt {
			return rt
		}

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
		return array.ToValue()
	}
	var entries BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := this.ToObject(agent).value
		return CreateArrayIterator(agent, o, objectOwnPropertiesKindKeyAndValue).ToValue()
	}
	var keys BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := this.ToObject(agent).value
		return CreateArrayIterator(agent, o, objectOwnPropertiesKindKey).ToValue()
	}
	var values BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := this.ToObject(agent).value
		iterator := CreateArrayIterator(agent, o, objectOwnPropertiesKindValue).ToValue()
		return iterator
	}
	var shift BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
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
	var unshift BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		o := this.ToObject(agent).value
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			return rt
		}
		argCount := JSInt(len(args))
		if argCount == 0 {
			return NewNumberValue(length.ToNumber())
		}
		if float64(length)+float64(argCount) > POW_2_32-1 {
			return co.ThrowTypeError(agent, "size is too large")
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
	var filter BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		callbackFn := args[0]
		thisArg := args[1]
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
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
				selected := ReturnAssertNormal(
					callbackFn.Call(agent, thisArg, []Value{kValue, NewNumberValue(k.ToNumber()), this}),
				)
				if selected.ToBoolean() {
					A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(to), kValue)
					to++
				}
			}
			k++
		}
		return A.ToValue()
	}
	var reduce BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		callbackFn := args[0]
		initialValue := args[1]
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
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
				accumulator = ReturnAssertNormal(
					callbackFn.Call(agent, UndefinedValue, []Value{accumulator, kValue, NewNumberValue(k.ToNumber()), this}),
				)
			}
			k++
		}
		return accumulator
	}
	var reduceRight BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		callbackFn := args[0]
		initialValue := args[1]
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
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
				accumulator = ReturnAssertNormal(
					callbackFn.Call(agent, UndefinedValue, []Value{accumulator, kValue, NewNumberValue(k.ToNumber()), this}),
				)
			}
			k--
		}
		return accumulator
	}
	var concat BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := this.ToObject(agent).value
		A := ArraySpeciesCreate(agent, o, 0)
		n := JSInt(0)

		for index := range len(args) + 1 {
			var element Value
			if index == 0 {
				element = o.ToValue()
			} else {
				element = args[index-1]
			}

			spreadable := IsConcatSpreadable(agent, element)
			if spreadable {
				var co CompletionValue
				length, isAbrupt, rt := ReturnIfAbrupt(MustGetObject(element).LengthOfArrayLike(), co)
				if isAbrupt {
					panic(rt)
				}

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
		return A.ToValue()
	}
	var slice BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		start := args[0]
		end := args[1]

		relativeStart, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, start), co)
		if isAbrupt {
			return rt
		}

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
			_relativeEnd, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, end), co)
			if isAbrupt {
				return rt
			}
			relativeEnd = _relativeEnd
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
		return A.ToValue()
	}
	var fill BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		value := args[0]
		start := args[1]
		end := args[2]
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		relativeStart, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, start), co)
		if isAbrupt {
			return rt
		}
		k := relativeStart.Max(0)
		if end == UndefinedValue {
			end = NewNumberValue(length.ToNumber())
		}
		var relativeEnd JSInt
		if end == UndefinedValue {
			relativeEnd = length
		} else {
			_relativeEnd, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, end), co)
			if isAbrupt {
				return rt
			}
			relativeEnd = _relativeEnd

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
		return o.ToValue()
	}
	var copyWithin BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		target := args[0]
		start := args[1]
		end := args[2]
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		relativeTarget, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, target), co)
		if isAbrupt {
			return rt
		}
		var to JSInt
		if relativeTarget.IsNegInf() {
			to = 0
		} else if relativeTarget < 0 {
			to = (length + relativeTarget).Max(0)
		} else {
			to = relativeTarget.Min(length)
		}

		relativeStart, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, start), co)
		if isAbrupt {
			return rt
		}
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
			_relativeEnd, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, end), co)
			if isAbrupt {
				return rt
			}
			relativeEnd = _relativeEnd
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
		return o.ToValue()
	}
	var reverse BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
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
		return o.ToValue()
	}
	var toReversed BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}
		A := ArrayCreate(agent, length, nil)
		for k := JSInt(0); k < length; k++ {
			from := NewIntegerIndexPropertyKey(length - k - 1)
			fromValue := o.Get(from)
			pk := NewIntegerIndexPropertyKey(k)
			A.CreateDataPropertyOrThrow(pk, fromValue)
		}
		return A.ToValue()
	}
	var sort BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		compareFn := args[0]
		if compareFn != UndefinedValue && !IsCallable(compareFn) {
			panic("TypeError")
		}
		obj := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(obj.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		sortCompare := SortCompare{
			compareFn: MustGetObject(compareFn),
			impl:      CompareArrayElements,
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
		return obj.ToValue()
	}
	var toSorted BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		compareFn := args[0]
		if compareFn != UndefinedValue && !IsCallable(compareFn) {
			panic("TypeError")
		}
		obj := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(obj.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		sortCompare := SortCompare{
			compareFn: MustGetObject(compareFn),
			impl:      CompareArrayElements,
		}

		sortedList := SortIndexedProperties(agent, obj, length, sortCompare, sortHolesTypeReadThroughHoles)
		A := ArrayCreate(agent, length, nil)
		for k, v := range sortedList {
			A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(JSInt(k)), v)
		}
		return A.ToValue()
	}
	var flat BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		depth := args[0]
		o := this.ToObject(agent).value
		var co CompletionValue
		sourceLen, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		var depthNum JSInt = 1
		if depth != UndefinedValue {
			_depthNum, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, depth), co)
			if isAbrupt {
				return rt
			}
			depthNum = _depthNum
			if depthNum < 0 {
				depthNum = 0
			}
		}
		A := ArraySpeciesCreate(agent, o, 0)
		FlattenIntoArray(agent, A, o, sourceLen, 0, depthNum, nil, nil)
		return A.ToValue()
	}
	var flatMap BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		mapperFunction := args[0]
		thisArg := args[1]
		o := this.ToObject(agent).value
		var co CompletionValue
		sourceLen, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		if !IsCallable(mapperFunction) {
			panic("TypeError")
		}
		A := ArraySpeciesCreate(agent, o, 0)
		FlattenIntoArray(agent, A, o, sourceLen, 0, 1, MustGetObject(mapperFunction), thisArg)
		return A.ToValue()
	}
	var splice BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		start := args[0]
		deleteCount := args[1]
		var items []Value
		if len(args) > 2 {
			items = args[2:]
		}
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		var relativeStart JSInt = 0
		if start != nil {
			_relativeStart, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, start), co)
			if isAbrupt {
				return rt
			}
			relativeStart = _relativeStart
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
			_actualDeleteCount, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, deleteCount), co)
			if isAbrupt {
				return rt
			}
			actualDeleteCount = _actualDeleteCount
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
		return A.ToValue()
	}
	var toSpliced BehaviorFn = func(this Value, args []Value, newTarget ObjectType) CompletionConvertable[Value] {
		start := args[0]
		skipCount := args[1]
		var items []Value
		if len(args) > 2 {
			items = args[2:]
		}
		o := this.ToObject(agent).value
		var co CompletionValue
		length, isAbrupt, rt := ReturnIfAbrupt(o.LengthOfArrayLike(), co)
		if isAbrupt {
			panic(rt)
		}

		var relativeStart JSInt = 0
		if start != nil {
			_relativeStart, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, start), co)
			if isAbrupt {
				return rt
			}
			relativeStart = _relativeStart
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
			sc, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, skipCount), co)
			if isAbrupt {
				return rt
			}
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

		return A.ToValue()
	}

	object.defineBuiltinFunction(realm, CMString("join"), join, 1)
	object.defineBuiltinFunction(realm, CMString("toString"), toString, 0)
	object.defineBuiltinFunction(realm, CMString("forEach"), forEach, 1)
	object.defineBuiltinFunction(realm, CMString("push"), push, 1)
	object.defineBuiltinFunction(realm, CMString("pop"), pop, 0)
	object.defineBuiltinFunction(realm, CMString("map"), arrayMap, 1)
	object.defineBuiltinFunction(realm, CMString("toLocaleString"), toLocaleString, 0)
	object.defineBuiltinFunction(realm, CMString("includes"), includes, 1)
	object.defineBuiltinFunction(realm, CMString("indexOf"), indexOf, 1)
	object.defineBuiltinFunction(realm, CMString("find"), find, 1)
	object.defineBuiltinFunction(realm, CMString("findIndex"), findIndex, 1)
	object.defineBuiltinFunction(realm, CMString("findLast"), findLast, 1)
	object.defineBuiltinFunction(realm, CMString("findLastIndex"), findLastIndex, 1)
	object.defineBuiltinFunction(realm, CMString("lastIndexOf"), lastIndexOf, 1)
	object.defineBuiltinFunction(realm, CMString("at"), at, 1)
	object.defineBuiltinFunction(realm, CMString("every"), every, 1)
	object.defineBuiltinFunction(realm, CMString("some"), some, 1)
	object.defineBuiltinFunction(realm, CMString("with"), with, 2)
	object.defineBuiltinFunction(realm, CMString("entries"), entries, 0)
	object.defineBuiltinFunction(realm, CMString("keys"), keys, 0)
	object.defineBuiltinFunction(realm, CMString("values"), values, 0)
	object.defineBuiltinFunction(realm, CMString("shift"), shift, 0)
	object.defineBuiltinFunction(realm, CMString("unshift"), unshift, 1)
	object.defineBuiltinFunction(realm, CMString("filter"), filter, 1)
	object.defineBuiltinFunction(realm, CMString("reduce"), reduce, 1)
	object.defineBuiltinFunction(realm, CMString("reduceRight"), reduceRight, 1)
	object.defineBuiltinFunction(realm, CMString("concat"), concat, 1)
	object.defineBuiltinFunction(realm, CMString("slice"), slice, 2)
	object.defineBuiltinFunction(realm, CMString("fill"), fill, 1)
	object.defineBuiltinFunction(realm, CMString("copyWithin"), copyWithin, 2)
	object.defineBuiltinFunction(realm, CMString("reverse"), reverse, 0)
	object.defineBuiltinFunction(realm, CMString("toReversed"), toReversed, 0)
	object.defineBuiltinFunction(realm, CMString("sort"), sort, 1)
	object.defineBuiltinFunction(realm, CMString("toSorted"), toSorted, 1)
	object.defineBuiltinFunction(realm, CMString("flat"), flat, 0)
	object.defineBuiltinFunction(realm, CMString("flatMap"), flatMap, 1)
	object.defineBuiltinFunction(realm, CMString("splice"), splice, 2)
	object.defineBuiltinFunction(realm, CMString("toSpliced"), toSpliced, 2)

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

	object.defineUnscopables(unscopablesList.ToValue())
	value := object.PropertyStorage().Get(NewStringPropertyKey("values"))
	object.defineBuiltinProperty(WellKnownSymbolsIterator, value)

	return object
}

// 23.1.3.2.1
func IsConcatSpreadable(agent *Agent, value Value) bool {
	if !value.IsObject() {
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
	impl      func(*Agent, Value, Value, ObjectType) Completion[JSNumber]
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
			c := ReturnAssertNormal(sortCompare.impl(agent, x, y, sortCompare.compareFn))
			if c >= 0 {
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

// CompareArrayElements
// spec: 23.1.3.30.2
func CompareArrayElements(agent *Agent, x, y Value, compareFn ObjectType) (co Completion[JSNumber]) {
	if x == UndefinedValue && y == UndefinedValue {
		co.value = 0
		return
	}
	if x == UndefinedValue {
		co.value = 1
		return
	}
	if y == UndefinedValue {
		co.value = -1
		return
	}
	if compareFn != nil {
		v, isAbrupt, rt := ReturnIfAbrupt(compareFn.
			ToValue().
			Call(agent, UndefinedValue, []Value{x, y}).
			value.
			ToNumber(agent), co)
		if isAbrupt {
			return rt
		}
		if v.IsNaN() {
			co.value = 0
			return
		}
		co.value = v.Data
		return
	}

	xString := x.String()
	yString := y.String()
	xSmaller, isAbrupt, rt := ReturnIfAbrupt(
		IsLessThan(agent, NewStringValue(xString), NewStringValue(yString), IsLessThanOrderLeftFirst),
		co)
	if isAbrupt {
		return rt
	}
	if xSmaller.ToBoolean() {
		co.value = -1
		return
	}

	ySmaller, isAbrupt, rt := ReturnIfAbrupt(
		IsLessThan(agent, NewStringValue(yString), NewStringValue(xString), IsLessThanOrderLeftFirst),
		co,
	)
	if isAbrupt {
		return rt
	}
	if ySmaller.ToBoolean() {
		co.value = 1
		return
	}
	co.value = 0
	return
}

// FlattenIntoArray
// spec: 23.1.3.13.1
func FlattenIntoArray(
	agent *Agent,
	target, source ObjectType,
	sourceLen JSInt,
	start, depth JSInt,
	mapperFunction ObjectType,
	thisArg Value,
) JSInt {
	if mapperFunction != nil {
		Assert(IsCallable((mapperFunction).ToValue()))
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
				element = mapperFunction.Call(
					thisArg,
					[]Value{element, NewNumberValue(sourceIndex.ToNumber()), (source).ToValue()},
				).value
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
				var co CompletionValue
				elementLen, isAbrupt, rt := ReturnIfAbrupt(MustGetObject(element).LengthOfArrayLike(), co)
				if isAbrupt {
					panic(rt)
				}

				targetIndex = FlattenIntoArray(agent, target, MustGetObject(element), elementLen, targetIndex, newDepth, mapperFunction, thisArg)
			} else {
				if float64(targetIndex) >= POW_2_53-1 {
					panic("TypeError")
				}
				target.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(targetIndex), element)
				targetIndex++
			}
		}
		sourceIndex++
	}
	return targetIndex
}

// MARK: - Internal

func (a *ArrayObject) String() string {
	s := "ArrayObject ["
	for i := JSInt(0); i < a.LengthOfArrayLike().value; i++ {
		if i > 0 {
			s += ", "
		}
		s += a.Get(NewIntegerIndexPropertyKey(i)).String()
	}
	s += "]"
	return s
}
