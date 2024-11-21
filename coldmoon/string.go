package coldmoon

import (
	"github.com/samber/lo"
	"math"
	"sort"
	"strings"
)

type StringObject struct {
	*Object
	Data string
}

func StringGetOwnProperty(s *StringObject, p PropertyKey) *PropertyDescriptor {
	intIndex, ok := p.(*IntegerIndexPropertyKey)
	if !ok {
		return nil
	}
	index := intIndex.Value

	str := s.Data
	l := len(str)
	if l <= index {
		return nil
	}

	resultStr := str[index : index+1]
	return &PropertyDescriptor{
		Value:        NewStringValue(resultStr),
		Writable:     false,
		Enumerable:   true,
		Configurable: false,
	}
}

func NewStringObject(agent *Agent, s string, prototype ObjectType) *StringObject {
	// 10.4.3.1
	var getOwnProperty GetOwnPropertyFn = func(o ObjectType, p PropertyKey) *PropertyDescriptor {
		desc := OrdinaryGetOwnProperty(o, p)

		if desc != nil {
			return desc
		}
		return StringGetOwnProperty(o.(*StringObject), p)
	}
	var defineOwnProperty DefineOwnPropertyFn = func(o ObjectType, p PropertyKey, desc *PropertyDescriptor) bool {
		s := o.(*StringObject)
		stringDesc := StringGetOwnProperty(s, p)
		if stringDesc != nil {
			extensible := s.Extensible()
			return IsCompatiblePropertyDescriptor(extensible, desc, stringDesc)
		}
		return OrdinaryDefineOwnProperty(o, p, desc)
	}
	// 10.4.3.3
	var ownPropertyKeys = func(o ObjectType) []PropertyKey {
		propertiesMap := o.PropertyStorage().Properties
		str := o.(*StringObject).Data
		length := len(str)
		keys := make([]PropertyKey, length+len(propertiesMap))
		for i := range length {
			keys[i] = NewIntegerIndexPropertyKey(i)
		}

		keysGreaterThanLength :=
			lo.Filter(lo.Keys(propertiesMap), func(pk PropertyKey, index int) bool {
				if index, ok := pk.(IntegerIndexPropertyKey); ok {
					return index.Value >= length
				}
				return false
			})
		sort.Slice(keysGreaterThanLength, func(i, j int) bool {
			return keysGreaterThanLength[i].(IntegerIndexPropertyKey).Value < keysGreaterThanLength[j].(IntegerIndexPropertyKey).Value
		})
		copy(keys[length:], keysGreaterThanLength)

		for pk, _ := range propertiesMap {
			if _, ok := pk.(StringPropertyKey); ok {
				keys = append(keys, pk)
			}
		}
		for pk, _ := range propertiesMap {
			if _, ok := pk.(SymbolPropertyKey); ok {
				keys = append(keys, pk)
			}
		}
		return keys
	}

	stringObject := &StringObject{
		Object: NewObject(agent, prototype),
		Data:   s,
	}
	stringObject.InternalMethods().GetOwnProperty = getOwnProperty
	stringObject.InternalMethods().DefineOwnProperty = defineOwnProperty
	stringObject.InternalMethods().OwnPropertyKeys = ownPropertyKeys

	length := uint64(len(s))
	stringObject.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(float64(length)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return stringObject
}

// 10.4.3.4
var StringCreate = NewStringObject

// MARK: - StringConstructor

func NewStringConstructor(realm *Realm) ObjectType {
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		value := argumentsList[0]
		var s string
		if len(argumentsList) > 0 {
			symbolValue, isSymbol := value.(*SymbolValue)
			if newTarget == nil && isSymbol {
				return NewStringValue(symbolValue.SymbolDescriptiveString())
			}
			s = value.String()
		}

		if newTarget == nil {
			return NewStringValue(s)
		}
		return NewValueFromObject(StringCreate(
			realm.Agent, s, GetPrototypeFromConstructor(newTarget, "%String.prototype%")),
		)
	}

	object := CreateBuiltinFunction(realm.Agent, behavior, 1, "String", builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		isConstructor: true,
	})

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.StringPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.StringPrototype, "constructor", NewValueFromObject(object))

	return object
}

func NewStringPrototype(realm *Realm) *StringObject {
	stringPrototype := &StringObject{
		Object: NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype),
		Data:   "",
	}

	agent := realm.Agent
	var toString BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		s := thisStringValue(agent, thisArgument)
		return NewStringValue(s)
	}
	var valueOf BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return NewStringValue(thisStringValue(agent, thisArgument))
	}
	var charAt BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		position := int(ToIntegerOrInfinity(agent, argumentsList[0]))
		size := len(s)
		if position < 0 || position >= size {
			return NewStringValue("")
		}
		return NewStringValue(string(s[position]))
	}
	var charCodeAt BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		position := int(ToIntegerOrInfinity(agent, argumentsList[0]))
		size := len(s)
		if position < 0 || position >= size {
			return NewNumberValue(math.NaN())
		}
		return NewNumberValue(float64(s[position]))
	}
	var iterator BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		return NewValueFromObject(&StringIteratorObject{
			Object: NewObject(agent, realm.Intrinsics.StringIteratorPrototype),
			Data:   s,
		})
	}
	var at BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		index := argumentsList[0]
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		length := len(s)
		relativeIndex := int(ToIntegerOrInfinity(agent, index))
		if relativeIndex < 0 || relativeIndex >= length {
			return UndefinedValue
		}
		k := relativeIndex
		return NewStringValue(string(s[k]))
	}
	var slice BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		intStart := ToIntegerOrInfinity(agent, argumentsList[0])
		intEnd := ToIntegerOrInfinity(agent, argumentsList[1])

		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		length := len(s)

		var from float64
		if intStart == math.Inf(-1) {
			from = 0
		} else if intStart < 0 {
			from = math.Max(float64(length)+intStart, 0)
		} else {
			from = math.Min(intStart, float64(length))
		}

		var to float64
		if intEnd == math.Inf(-1) {
			to = 0
		} else if intEnd < 0 {
			to = math.Max(float64(length)+intEnd, 0)
		} else {
			to = math.Min(intEnd, float64(length))
		}
		if from >= to {
			return NewStringValue("")
		}
		return NewStringValue(s[int(from):int(to)])
	}
	var repeat BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		n := ToIntegerOrInfinity(agent, argumentsList[0])
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		if n < 0 || n == math.Inf(1) {
			panic("RangeError")
		}
		if n == 0 {
			return NewStringValue("")
		}
		return NewStringValue(strings.Repeat(s, int(n)))
	}
	var concat BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		o := RequireObjectCoercible(agent, thisArgument)
		s := o.String()
		for _, arg := range argumentsList {
			s += arg.String()
		}
		return NewStringValue(s)
	}

	DefineBuiltinFunction(stringPrototype, "toString", toString, 0, realm)
	DefineBuiltinFunction(stringPrototype, "valueOf", valueOf, 0, realm)
	DefineBuiltinFunction(stringPrototype, "charAt", charAt, 1, realm)
	DefineBuiltinFunction(stringPrototype, "charCodeAt", charCodeAt, 1, realm)
	DefineBuiltinFunction(stringPrototype, "iterator", iterator, 0, realm)
	DefineBuiltinFunction(stringPrototype, "at", at, 1, realm)
	DefineBuiltinFunction(stringPrototype, "slice", slice, 2, realm)
	DefineBuiltinFunction(stringPrototype, "repeat", repeat, 1, realm)
	DefineBuiltinFunction(stringPrototype, "concat", concat, 1, realm)

	return stringPrototype
}

func thisStringValue(agent *Agent, v Value) string {
	switch v := v.(type) {
	case *StringValue:
		return v.Data
	case *ObjectValue:
		s, ok := v.Object.(*StringObject)
		if ok {
			return s.Data
		}

	}
	panic("TypeError")
}

// 22.1.3.32.1
func (s *StringValue) TrimString() string {
	return strings.TrimSpace(s.Data)
}
