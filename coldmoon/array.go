package coldmoon

import (
	"github.com/Seeingu/coldmoon/pkg"
	"math"
	"strings"
)

type ArrayObject struct {
	*Object
}

func getArrayLength(array ObjectType) float64 {
	lengthDesc := OrdinaryGetOwnProperty(array, NewStringPropertyKey("length"))
	Assert(lengthDesc.IsDataDescriptor())
	return lengthDesc.Value.(*NumberValue).Data
}

// 10.4.2.2
func ArrayCreate(agent *Agent, length float64, proto ObjectType) ObjectType {
	realm := agent.CurrentRealm()
	// 10.4.2.1
	var defineOwnProperty = func(array ObjectType, p PropertyKey, desc *PropertyDescriptor) bool {
		propertyKeyString, ok := p.(*StringPropertyKey)
		if ok && propertyKeyString.Value == "length" {
			return ArraySetLength(agent, array, desc)
		}
		propertyKeyIndex, ok := p.(*IntegerIndexPropertyKey)
		if ok {
			lengthDesc := OrdinaryGetOwnProperty(array, NewStringPropertyKey("length"))
			Assert(lengthDesc.IsDataDescriptor())
			Assert(!lengthDesc.Configurable)

			lengthValue := lengthDesc.Value
			length := lengthValue.(*NumberValue).Data
			Assert(math.IsInf(length, 0) || length >= 0)

			index := propertyKeyIndex.Value

			if index >= int(length) && !lengthDesc.Writable {
				return false
			}

			succeeded := OrdinaryDefineOwnProperty(array, p, desc)

			if !succeeded {
				return false
			}

			if index >= int(length) {
				lengthDesc.Value = NewNumberValue(float64(index) + 1)

				succeeded = OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), lengthDesc)
				Assert(succeeded)
			}
		}
		return true
	}

	if length > POW_2_32-1 {
		panic("RangeError")
	}

	if proto == nil {
		proto = realm.Intrinsics.ArrayPrototype
	}

	arr := &ArrayObject{
		Object: NewObject(agent, proto),
	}
	arr.Object.InternalMethods().DefineOwnProperty = defineOwnProperty
	OrdinaryDefineOwnProperty(arr, NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(length),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	return arr
}

// 10.4.2.3
func ArraySpeciesCreate(agent *Agent, originalArray ObjectType, length float64) ObjectType {
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

	return constructorObject.Object.Construct([]Value{NewNumberValue(length)}, nil)
}

// 10.4.2.4
func ArraySetLength(agent *Agent, array ObjectType, desc *PropertyDescriptor) bool {
	if desc.Value == nil {
		return OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), desc)
	}
	newLenDesc := desc

	newLenValue := desc.Value
	newLen := newLenValue.(*NumberValue).Data
	numberLen := ToNumber(agent, newLenValue)

	if numberLen.Data != newLen {
		panic("RangeError")
	}

	newLenDesc.Value = NewNumberValue(newLen)

	oldLenDesc := OrdinaryGetOwnProperty(array, NewStringPropertyKey("length"))
	Assert(oldLenDesc.IsDataDescriptor())
	Assert(!oldLenDesc.Configurable)
	oldLen := oldLenDesc.Value.(*NumberValue).Data

	if newLen >= oldLen {
		return OrdinaryDefineOwnProperty(array, NewStringPropertyKey("length"), newLenDesc)
	}

	if !desc.Writable {
		return false
	}

	var newWritable = false
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
		deleteSucceeded := array.InternalMethods().Delete(array, NewIntegerIndexPropertyKey(int(k)))
		if !deleteSucceeded {
			newLenDesc.Value = NewNumberValue(float64(k) + 1)
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

		numberOfArgs := len(args)
		if numberOfArgs == 0 {
			return NewValueFromObject(ArrayCreate(agent, 0, proto))
		} else if numberOfArgs == 1 {
			length := args[0]
			array := ArrayCreate(agent, 0, proto)

			var intLen uint32
			if _, ok := length.(*NumberValue); ok {
				array.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(0), length)
				intLen = 1
			} else {
				intLen = ToUint32(agent, length)
			}

			array.Set(
				NewStringPropertyKey("length"),
				NewNumberValue(float64(intLen)),
				setThrowTypeThrow,
			)

			return NewValueFromObject(array)
		} else {
			Assert(numberOfArgs >= 2)
			array := ArrayCreate(agent, float64(numberOfArgs), proto)

			for k := range numberOfArgs {
				propertyKey := NewIntegerIndexPropertyKey(k)
				array.CreateDataPropertyOrThrow(propertyKey, args[k])
			}

			Assert(getArrayLength(array) == float64(numberOfArgs))
			return NewValueFromObject(array)
		}
	}

	var isArray BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		arg := args[0]
		return NewBooleanValue(IsArray(arg))
	}
	var of BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		length := len(args)
		lenNumber := NewNumberValue(float64(length))

		constructor := this
		var array ObjectType
		if IsConstructor(constructor) {
			array = MustGetObject(constructor).Construct(args, nil)
		} else {
			array = ArrayCreate(realm.Agent, float64(length), nil)
		}

		for k := range args {
			propertyKey := NewIntegerIndexPropertyKey(k)
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

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
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
	DefineBuiltinProperty(realm.Intrinsics.ArrayPrototype, "constructor", NewValueFromObject(object))

	return object
}

func NewArrayPrototype(realm *Realm) ObjectType {
	agent := realm.Agent
	object := &ArrayObject{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype),
	}

	DefineBuiltinProperty(object.Object, "length", &PropertyDescriptor{
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

		A := ArraySpeciesCreate(agent, array, float64(length))
		for k := range length {
			pk := NewIntegerIndexPropertyKey(int(k))
			mappedValue := callbackFn.CallAssumeCallable(
				thisArg,
				[]Value{array.Get(pk), NewNumberValue(float64(k)), this},
			)

			A.CreateDataPropertyOrThrow(pk, mappedValue)
		}
		return NewValueFromObject(A)
	}

	var join BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		array := MustGetObject(this)
		length := array.LengthOfArrayLike()
		var sep = ","
		if args[0] != nil {
			sep = args[0].String()
		}

		var elements []string
		for k := range length {
			element := array.Get(NewIntegerIndexPropertyKey(int(k)))

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
			pk := NewIntegerIndexPropertyKey(int(k))
			kPresent := array.HasProperty(pk)
			if kPresent {
				kValue := array.Get(pk)
				callbackFn.CallAssumeCallable(
					thisArg,
					[]Value{kValue, NewNumberValue(float64(k)), this},
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
			array.Set(NewIntegerIndexPropertyKey(int(length)), args[i], setThrowTypeThrow)
			length++
		}

		array.Set(NewStringPropertyKey("length"), NewNumberValue(float64(length)), setThrowTypeThrow)
		return NewNumberValue(float64(length))
	}
	var pop BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		array := MustGetObject(this)
		length := array.LengthOfArrayLike()
		if length == 0 {
			array.Set(NewStringPropertyKey("length"), NewNumberValue(0), setThrowTypeThrow)
			return UndefinedValue
		}
		length--
		element := array.Get(NewIntegerIndexPropertyKey(int(length)))
		deleteSucceeded := array.DeletePropertyOrThrow(NewIntegerIndexPropertyKey(int(length)))
		if !deleteSucceeded {
			panic("TypeError")
		}
		array.Set(NewStringPropertyKey("length"), NewNumberValue(float64(length)), setThrowTypeThrow)
		return element
	}
	var toLocaleString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		array := MustGetObject(this)
		length := array.LengthOfArrayLike()
		separator := ", "
		var elements []string
		for k := range length {
			nextElement := array.Get(NewIntegerIndexPropertyKey(int(k)))
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
		if n == math.Inf(1) {
			return FalseValue
		} else if n == math.Inf(-1) {
			n = 0
		}

		k := int(n)
		if k < 0 {
			k = int(math.Max(float64(length)+float64(k), 0))
		}

		for k < int(length) {
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
		if n == math.Inf(1) {
			return NewNumberValue(-1)
		} else if n == math.Inf(-1) {
			n = 0
		}

		k := int(math.Max(n, 0))
		if k < 0 {
			k = int(math.Max(float64(length)+float64(k), 0))
		}

		for k < int(length) {
			kPresent := o.HasProperty(NewIntegerIndexPropertyKey(k))
			if kPresent {
				elementK := o.Get(NewIntegerIndexPropertyKey(k))
				if IsStrictlyEqual(searchElement, elementK) {
					return NewNumberValue(float64(k))
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
		findRec := o.(*ArrayObject).findViaPredicate(int(length), DirectionAscending, predicate, thisArg)
		return findRec.value
	}
	var findIndex BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		predicate := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		findRec := o.(*ArrayObject).findViaPredicate(int(length), DirectionAscending, predicate, thisArg)
		return NewNumberValue(float64(findRec.index))
	}
	var findLast BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		predicate := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		findRec := o.(*ArrayObject).findViaPredicate(int(length), DirectionDescending, predicate, thisArg)
		return findRec.value
	}
	var findLastIndex BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		predicate := args[0]
		thisArg := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		findRec := o.(*ArrayObject).findViaPredicate(int(length), DirectionDescending, predicate, thisArg)
		return NewNumberValue(float64(findRec.index))
	}
	var lastIndexOf BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		searchElement := args[0]
		fromIndex := args[1]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		if length == 0 {
			return NewNumberValue(-1)
		}
		var n float64
		if len(args) > 1 {
			n = ToIntegerOrInfinity(agent, fromIndex)
		} else {
			n = float64(length - 1)
		}

		if n == math.Inf(-1) {
			return NewNumberValue(-1)
		}

		k := int(math.Min(math.Max(n, 0), float64(length-1)))
		for k >= 0 {
			kPresent := o.HasProperty(NewIntegerIndexPropertyKey(k))
			if kPresent {
				elementK := o.Get(NewIntegerIndexPropertyKey(k))
				if IsStrictlyEqual(searchElement, elementK) {
					return NewNumberValue(float64(k))
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
		k := int(relativeIndex)
		if k < 0 {
			k += int(length)
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

		for k := 0; k < int(length); k++ {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				testResult := callbackFn.CallAssumeCallable(thisArg, []Value{kValue, NewNumberValue(float64(k)), this})
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

		for k := 0; k < int(length); k++ {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				testResult := callbackFn.CallAssumeCallable(thisArg, []Value{kValue, NewNumberValue(float64(k)), this})
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

		actualIndex := int(relativeIndex)
		if actualIndex < 0 {
			actualIndex += int(length)
		}

		array := ArrayCreate(agent, float64(length), nil)
		for k := 0; k < int(length); k++ {
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

			for k := 0; ; k++ {
				pk := NewIntegerIndexPropertyKey(k)
				next := iteratorRecord.IteratorStep()
				if next == nil {
					a.Set(NewStringPropertyKey("length"), NewNumberValue(float64(k)), setThrowTypeThrow)
					return NewValueFromObject(a)
				}

				nextValue := IteratorValue(next)
				var mappedValue Value
				if mapping {
					mappedValue = mapFn.CallAssumeCallable(thisArg, []Value{nextValue, NewNumberValue(float64(k))})
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
			a = MustGetObject(c).Construct([]Value{NewNumberValue(float64(length))}, nil)
		} else {
			a = ArrayCreate(agent, float64(length), nil)
		}

		for k := 0; k < int(length); k++ {
			pk := NewIntegerIndexPropertyKey(k)
			kValue := arrayLike.Get(pk)
			var mappedValue Value
			if mapping {
				mappedValue = mapFn.CallAssumeCallable(thisArg, []Value{kValue, NewNumberValue(float64(k))})
			} else {
				mappedValue = kValue
			}
			a.CreateDataPropertyOrThrow(pk, mappedValue)
		}
		a.Set(NewStringPropertyKey("length"), NewNumberValue(float64(length)), setThrowTypeThrow)
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
		for k := 1; k < int(length); k++ {
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
		deleteSucceeded := o.DeletePropertyOrThrow(NewIntegerIndexPropertyKey(int(length - 1)))
		if !deleteSucceeded {
			panic("TypeError")
		}
		o.Set(NewStringPropertyKey("length"), NewNumberValue(float64(length-1)), setThrowTypeThrow)
		return first
	}
	var unshift BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		argCount := len(args)
		if argCount == 0 {
			return NewNumberValue(float64(length))
		}

		k := int(length)
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
			key := NewIntegerIndexPropertyKey(j)
			o.Set(key, arg, setThrowTypeThrow)
		}
		o.Set(NewStringPropertyKey("length"), NewNumberValue(float64(int(length)+argCount)), setThrowTypeThrow)
		return NewNumberValue(float64(int(length) + argCount))
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
		k := 0
		to := 0
		for k < int(length) {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				selected := callbackFn.CallAssumeCallable(thisArg, []Value{kValue, NewNumberValue(float64(k)), this})
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
		k := 0
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
				if k >= int(length) {
					panic("TypeError")
				}
			}
		} else {
			accumulator = initialValue
		}
		for k < int(length) {
			pk := NewIntegerIndexPropertyKey(k)
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				accumulator = callbackFn.CallAssumeCallable(UndefinedValue, []Value{accumulator, kValue, NewNumberValue(float64(k)), this})
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
		k := int(length) - 1
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
				accumulator = callbackFn.CallAssumeCallable(UndefinedValue, []Value{accumulator, kValue, NewNumberValue(float64(k)), this})
			}
			k--
		}
		return accumulator
	}
	var concat BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		A := ArraySpeciesCreate(agent, o, 0)
		n := 0

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

				k := 0
				for k < int(length) {
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

		A.Set(NewStringPropertyKey("length"), NewNumberValue(float64(n)), setThrowTypeThrow)
		return NewValueFromObject(A)
	}
	var slice BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		start := args[0]
		end := args[1]

		relativeStart := ToIntegerOrInfinity(agent, start)

		var k float64
		if relativeStart == math.Inf(-1) {
			k = 0
		} else if relativeStart < 0 {
			k = math.Max(float64(length)+relativeStart, 0)
		} else {
			k = math.Min(relativeStart, float64(length))
		}

		var relativeEnd float64
		if end == UndefinedValue {
			relativeEnd = float64(length)
		} else {
			relativeEnd = ToIntegerOrInfinity(agent, end)
		}

		var final float64
		if relativeEnd == math.Inf(-1) {
			final = 0
		} else if relativeEnd < 0 {
			final = math.Max(float64(length)+relativeEnd, 0)
		} else {
			final = math.Min(relativeEnd, float64(length))
		}
		count := math.Max(final-k, 0)

		n := 0
		A := ArraySpeciesCreate(agent, o, count)
		for k < final {
			pk := NewIntegerIndexPropertyKey(int(k))
			kPresent := o.HasProperty(pk)
			if kPresent {
				kValue := o.Get(pk)
				A.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(n), kValue)
			}
			k++
			n++
		}
		A.Set(NewStringPropertyKey("length"), NewNumberValue(float64(n)), setThrowTypeThrow)
		return NewValueFromObject(A)
	}
	var fill BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		value := args[0]
		start := args[1]
		end := args[2]
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()

		relativeStart := ToIntegerOrInfinity(agent, start)
		k := math.Max(relativeStart, 0)
		if end == UndefinedValue {
			end = NewNumberValue(float64(length))
		}
		var relativeEnd float64
		if end == UndefinedValue {
			relativeEnd = float64(length)
		} else {
			relativeEnd = ToIntegerOrInfinity(agent, end)
		}

		var final float64
		if relativeEnd == math.Inf(-1) {
			final = 0
		} else if relativeEnd < 0 {
			final = math.Max(float64(length)+relativeEnd, 0)
		} else {
			final = math.Min(relativeEnd, float64(length))
		}
		for k < final {
			pk := NewIntegerIndexPropertyKey(int(k))
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
		var to float64
		if relativeTarget == math.Inf(-1) {
			to = 0
		} else if relativeTarget < 0 {
			to = math.Max(float64(length)+relativeTarget, 0)
		} else {
			to = math.Min(relativeTarget, float64(length))
		}

		relativeStart := ToIntegerOrInfinity(agent, start)
		var from float64
		if relativeStart == math.Inf(-1) {
			from = 0
		} else if relativeStart < 0 {
			from = math.Max(float64(length)+relativeStart, 0)
		} else {
			from = math.Min(relativeStart, float64(length))
		}

		var relativeEnd float64
		if end == UndefinedValue {
			relativeEnd = float64(length)
		} else {
			relativeEnd = ToIntegerOrInfinity(agent, end)
		}

		var final float64
		if relativeEnd == math.Inf(-1) {
			final = 0
		} else if relativeEnd < 0 {
			final = math.Max(float64(length)+relativeEnd, 0)
		} else {
			final = math.Min(relativeEnd, float64(length))
		}

		count := math.Min(final-from, float64(length)-to)

		var direction int
		if from < to && to < from+count {
			direction = -1
			from += count - 1
			to += count - 1
		} else {
			direction = 1
		}

		for count > 0 {
			fromKey := NewIntegerIndexPropertyKey(int(from))
			toKey := NewIntegerIndexPropertyKey(int(to))
			fromPresent := o.HasProperty(fromKey)
			if fromPresent {
				fromValue := o.Get(fromKey)
				o.Set(toKey, fromValue, setThrowTypeThrow)
			} else {
				o.DeletePropertyOrThrow(toKey)
			}
			from += float64(direction)
			to += float64(direction)
			count--
		}
		return NewValueFromObject(o)
	}
	var reverse BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		length := o.LengthOfArrayLike()
		middle := int(length / 2)
		lower := 0
		for lower < middle {
			upper := int(length) - lower - 1
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
		A := ArrayCreate(agent, float64(length), nil)
		for k := 0; k < int(length); k++ {
			from := NewIntegerIndexPropertyKey(int(length) - k - 1)
			fromValue := o.Get(from)
			pk := NewIntegerIndexPropertyKey(k)
			A.CreateDataPropertyOrThrow(pk, fromValue)
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

	DefineBuiltinProperty(object, "@@unscopables", &PropertyDescriptor{
		Value:        unscopablesValue,
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	DefineBuiltinProperty(object, "@@iterator", object.PropertyStorage().Get(NewStringPropertyKey("values")))

	return object
}

type direction int

const (
	DirectionAscending direction = iota
	DirectionDescending
)

type FoundResult struct {
	index int
	value Value
}

func (a *ArrayObject) findViaPredicate(
	len int,
	direction direction,
	predicate Value,
	thisArg Value,
) FoundResult {
	if !IsCallable(predicate) {
		panic("TypeError")
	}

	var k int
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
		testResult := predicate.CallAssumeCallable(thisArg, []Value{kValue, NewNumberValue(float64(k)), NewValueFromObject(a)})

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
