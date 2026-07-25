package coldmoon

import (
	"github.com/Seeingu/coldmoon/pkg"
)

func InternalGetPrototypeOf(object ObjectType) ObjectType {
	return OrdinaryGetPrototypeOf(object)
}

func OrdinaryGetPrototypeOf(object ObjectType) ObjectType {
	return object.Prototype()
}

func InternalSetPrototypeOf(object ObjectType, prototype ObjectType) bool {
	return OrdinarySetPrototypeOf(object, prototype)
}

// 10.1.2.1
func OrdinarySetPrototypeOf(object ObjectType, prototype ObjectType) bool {
	current := object.Prototype()

	if ObjectSameValue(prototype, current) {
		return true
	}

	extensible := object.Extensible()
	if !extensible {
		return false
	}

	p := prototype
	done := false
	for !done {
		if p == nil {
			done = true
		} else if ObjectSameValue(p, object) {
			return false
		} else if !pkg.FuncEqual(p.InternalMethods().GetPrototypeOf, InternalGetPrototypeOf) {
			done = true
		} else {
			p = p.Prototype()
		}
	}

	object.SetPrototype(prototype)

	return true
}

func InternalIsExtensible(object ObjectType) bool {
	return OrdinaryIsExtensible(object.(*Object))
}

func OrdinaryIsExtensible(object *Object) bool {
	return object.Extensible()
}

func InternalPreventExtensions(object ObjectType) bool {
	return OrdinaryPreventExtensions(object.(*Object))
}

func OrdinaryPreventExtensions(object *Object) bool {
	object.data.extensible = false
	return true
}

func InternalGetOwnProperty(object ObjectType, key PropertyKey) *PropertyDescriptor {
	return OrdinaryGetOwnProperty(object, key)
}

func OrdinaryGetOwnProperty(object ObjectType, key PropertyKey) *PropertyDescriptor {
	if !object.PropertyStorage().Has(key) {
		return nil
	}

	d := &PropertyDescriptor{}

	x := object.PropertyStorage().Get(key)
	if x.IsDataDescriptor() {
		d.Value = x.Value
		d.Writable = x.Writable
	} else {
		if !x.IsAccessorDescriptor() {
			return nil
		}

		d.Get = x.Get
		d.Set = x.Set
	}

	d.Enumerable = x.Enumerable
	d.Configurable = x.Configurable

	return d
}

func InternalDefineOwnProperty(object ObjectType, key PropertyKey, desc *PropertyDescriptor) (co Completion[bool]) {
	co.value = OrdinaryDefineOwnProperty(object, key, desc)
	return
}

func OrdinaryDefineOwnProperty(object ObjectType, key PropertyKey, desc *PropertyDescriptor) bool {
	current := object.PropertyStorage().Get(key)

	extensible := object.Extensible()

	return ValidateAndApplyPropertyDescriptor(object, key, extensible, desc, current)
}

// 10.1.6.2
func IsCompatiblePropertyDescriptor(extensible bool, desc, current *PropertyDescriptor) bool {
	return ValidateAndApplyPropertyDescriptor(nil, NewStringPropertyKey(""), extensible, desc, current)
}

// 10.1.6.3
func ValidateAndApplyPropertyDescriptor(
	object ObjectType,
	key PropertyKey,
	extensible bool,
	desc, current *PropertyDescriptor,
) bool {
	// TODO: Assert
	if current == nil {
		if !extensible {
			return false
		}

		if object == nil {
			return true
		}

		if desc.IsAccessorDescriptor() {
			object.PropertyStorage().Set(key, &PropertyDescriptor{
				Get:          desc.Get,
				Set:          desc.Set,
				Enumerable:   desc.Enumerable,
				Configurable: desc.Configurable,
			})
		} else {
			value := desc.Value
			if value == nil {
				value = UndefinedValue
			}
			object.PropertyStorage().Set(key, &PropertyDescriptor{
				Value:        value,
				Writable:     desc.Writable,
				WritableSet:  true,
				Enumerable:   desc.Enumerable,
				Configurable: desc.Configurable,
			})
		}

		return true
	}

	Assert(current.IsFullyPopulated())

	if !desc.HasFields() {
		return true
	}

	if !current.Configurable {
		if desc.Configurable {
			return false
		}

		if desc.Enumerable != current.Enumerable {
			return false
		}

		if (current.IsDataDescriptor() && desc.IsAccessorDescriptor()) ||
			(current.IsAccessorDescriptor() && desc.IsDataDescriptor()) {
			return false
		}

		if current.IsAccessorDescriptor() {
			if desc.Get != nil &&
				current.Get != nil &&
				!pkg.FuncEqual(desc.Get, current.Get) {
				return false
			}

			if desc.Set != nil &&
				current.Set != nil &&
				!pkg.FuncEqual(desc.Set, current.Set) {
				return false
			}
		} else if !current.Writable {
			if desc.Writable {
				return false
			}

			if desc.Value != nil && !SameValue(desc.Value, current.Value) {
				return false
			}
		}
	}

	if object != nil {
		if current.IsDataDescriptor() && desc.IsAccessorDescriptor() {
			// i. If Desc has a [[Configurable]] field, let configurable be Desc.[[Configurable]];
			//    else let configurable be current.[[Configurable]].
			configurable := desc.Configurable || current.Configurable
			// ii. If Desc has a [[Enumerable]] field, let enumerable be Desc.[[Enumerable]]; else
			//     let enumerable be current.[[Enumerable]].
			enumerable := desc.Enumerable || current.Enumerable

			// iii. Replace the property named P of object O with a data property whose
			//      [[Configurable]] and [[Enumerable]] attributes are set to configurable and
			//      enumerable, respectively, and whose [[Value]] and [[Writable]] attributes are
			//      set to the value of the corresponding field in Desc if Desc has that field, or
			//      to the attribute's default value otherwise.
			object.PropertyStorage().Set(key, &PropertyDescriptor{
				Value:        desc.Value,
				Writable:     desc.Writable,
				WritableSet:  true,
				Enumerable:   enumerable,
				Configurable: configurable,
			})
		} else if current.IsAccessorDescriptor() && desc.IsDataDescriptor() {

			// i. If Desc has a [[Configurable]] field, let configurable be Desc.[[Configurable]];
			//    else let configurable be current.[[Configurable]].
			configurable := desc.Configurable || current.Configurable

			// ii. If Desc has a [[Enumerable]] field, let enumerable be Desc.[[Enumerable]]; else
			//     let enumerable be current.[[Enumerable]].
			enumerable := desc.Enumerable || current.Enumerable

			// iii. Replace the property named P of object O with a data property whose
			//      [[Configurable]] and [[Enumerable]] attributes are set to configurable and
			//      enumerable, respectively, and whose [[Value]] and [[Writable]] attributes are
			//      set to the value of the corresponding field in Desc if Desc has that field, or
			//      to the attribute's default value otherwise.
			v := desc.Value
			if v == nil {
				v = UndefinedValue
			}
			object.PropertyStorage().Set(key, &PropertyDescriptor{
				Value:        v,
				Writable:     desc.Writable,
				WritableSet:  true,
				Enumerable:   enumerable,
				Configurable: configurable,
			})
		} else {
			// i. For each field of Desc, set the corresponding attribute of the property named P
			//    of object O to the value of the field.
			v := desc.Value
			if v == nil {
				v = current.Value
			}
			w := current.Writable
			if desc.WritableSet || desc.Writable {
				w = desc.Writable
			}
			e := desc.Enumerable || current.Enumerable
			c := desc.Configurable || current.Configurable
			g := desc.Get
			if g == nil {
				g = current.Get
			}
			s := desc.Set
			if s == nil {
				s = current.Set
			}

			object.PropertyStorage().Set(key, &PropertyDescriptor{
				Value:        v,
				Writable:     w,
				WritableSet:  true,
				Get:          g,
				Set:          s,
				Enumerable:   e,
				Configurable: c,
			})
		}
	}

	return true
}

func InternalHasProperty(object ObjectType, key PropertyKey) bool {
	return OrdinaryHasProperty(object, key)
}

func OrdinaryHasProperty(object ObjectType, key PropertyKey) bool {
	hasOwn := object.InternalMethods().GetOwnProperty(object, key)
	if hasOwn != nil {
		return true
	}

	parent := object.InternalMethods().GetPrototypeOf(object)
	if parent != nil {
		return parent.InternalMethods().HasProperty(parent, key)
	}

	return false
}

func InternalGet(object ObjectType, key PropertyKey, receiver Value) CompletionValue {
	return OrdinaryGet(object, key, receiver)
}

// OrdinaryGet
// spec: 10.1.8.1
func OrdinaryGet(object ObjectType, key PropertyKey, receiver Value) CompletionValue {
	desc := object.InternalMethods().GetOwnProperty(object, key)

	if desc == nil {
		parent := object.InternalMethods().GetPrototypeOf(object)
		if parent == nil {
			return UndefinedValue.ToCompletion()
		}

		return parent.InternalMethods().Get(parent, key, receiver)
	}

	if desc.IsDataDescriptor() {
		return desc.Value.ToCompletion()
	}

	if !desc.IsAccessorDescriptor() {
		panic("")
	}

	getter := desc.Get
	if getter == nil {
		return UndefinedValue.ToCompletion()
	}

	return getter.ToValue().CallNoArgs(receiver)
}

func InternalSet(object ObjectType, key PropertyKey, value Value, receiver Value) Completion[bool] {
	return OrdinarySet(object, key, value, receiver)
}

func OrdinarySet(object ObjectType, key PropertyKey, value Value, receiver Value) Completion[bool] {
	ownDesc := object.InternalMethods().GetOwnProperty(object, key)
	return OrdinarySetWithOwnDescriptor(object, key, value, receiver, ownDesc)
}

// 10.1.9.2
func OrdinarySetWithOwnDescriptor(
	object ObjectType,
	key PropertyKey,
	value Value,
	receiver Value,
	ownDesc *PropertyDescriptor,
) (co Completion[bool]) {
	if ownDesc == nil {
		parent := object.InternalMethods().GetPrototypeOf(object)
		if parent != nil {
			return parent.InternalMethods().Set(parent, key, value, receiver)
		} else {
			ownDesc = &PropertyDescriptor{
				Value:        UndefinedValue,
				Writable:     true,
				Enumerable:   true,
				Configurable: true,
			}
		}
	}

	if ownDesc.IsDataDescriptor() {
		if !ownDesc.Writable {
			return
		}

		r, isObject := receiver.(*ObjectValue)
		if !isObject {
			return
		}
		receiverObject := r.Object

		existingDescriptor := object.InternalMethods().GetOwnProperty(receiverObject, key)

		if existingDescriptor != nil {
			if existingDescriptor.IsAccessorDescriptor() {
				return
			}
			if !existingDescriptor.Writable {
				return
			}

			valueDesc := &PropertyDescriptor{
				Value: value,
			}
			return receiverObject.InternalMethods().DefineOwnProperty(
				receiverObject, key, valueDesc,
			)
		} else {
			Assert(!receiverObject.PropertyStorage().Has(key))

			co.value = receiverObject.CreateDataProperty(key, value)
			return
		}
	}

	Assert(ownDesc.IsAccessorDescriptor())

	setter := ownDesc.Set
	if setter == nil {
		return
	}
	setResult := setter.Call(receiver, []Value{value})
	if setResult.IsAbrupt() {
		return CompletionFrom(co, setResult)
	}
	co.value = true
	return
}

func InternalDelete(object ObjectType, key PropertyKey) (co Completion[bool]) {
	co.value = OrdinaryDelete(object, key)
	return
}

func OrdinaryDelete(object ObjectType, key PropertyKey) bool {
	desc := object.InternalMethods().GetOwnProperty(object, key)
	if desc == nil {
		return true
	}

	if desc.Configurable {
		object.PropertyStorage().Delete(key)
		return true
	}
	return false
}

func InternalOwnPropertyKeys(object ObjectType) []PropertyKey {
	return OrdinaryOwnPropertyKeys(object)
}

func OrdinaryOwnPropertyKeys(object ObjectType) []PropertyKey {
	var keys []PropertyKey

	for key := range object.PropertyStorage().Properties {
		keys = append(keys, key)
	}

	return keys
}

func ObjectSameValue(x, y ObjectType) bool {
	return x == y
}

// OrdinaryObjectCreate
// spec: 10.1.12
func OrdinaryObjectCreate(agent *Agent, proto ObjectType, internalSlotsList []string) *Object {
	obj := NewObjectV2(agent, proto, "OrdinaryObject", internalSlotsList)
	return obj
}

// 10.1.13
func OrdinaryCreateFromConstructor(
	agent *Agent,
	constructor ObjectType,
	intrinsicDefaultProto IntrinsicName,
	internalSlotsList []string,
) *Object {
	proto := GetPrototypeFromConstructor(constructor, intrinsicDefaultProto)

	o := OrdinaryObjectCreate(agent, proto, internalSlotsList)
	return o
}

// 10.1.14
func GetPrototypeFromConstructor(constructor ObjectType, intrinsicDefaultProto IntrinsicName) ObjectType {
	proto := constructor.Get(NewStringPropertyKey("prototype"))

	switch p := proto.(type) {
	case *ObjectValue:
		return p.Object
	default:
		realm := constructor.GetFunctionRealm()
		return realm.Intrinsics.Get(intrinsicDefaultProto)
	}
}
