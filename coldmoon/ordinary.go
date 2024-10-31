package coldmoon

import (
	"fmt"
	"github.com/Seeingu/coldmoon/pkg"
)

func InternalGetPrototypeOf(object ObjectType) ObjectType {
	return OrdinaryGetPrototypeOf(object)
}

func OrdinaryGetPrototypeOf(object ObjectType) ObjectType {
	return object.Prototype()
}

func InternalSetPrototypeOf(object ObjectType, prototype ObjectType) bool {
	return OrdinarySetPrototypeOf(object.(*Object), prototype)
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

	object.(*Object).SetPrototype(prototype)

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
			panic("")
		}

		d.Get = x.Get
		d.Set = x.Set
	}

	d.Enumerable = x.Enumerable
	d.Configurable = x.Configurable

	return d
}

func InternalDefineOwnProperty(object ObjectType, key PropertyKey, desc *PropertyDescriptor) bool {
	return OrdinaryDefineOwnProperty(object.(*Object), key, desc)
}

func OrdinaryDefineOwnProperty(object *Object, key PropertyKey, desc *PropertyDescriptor) bool {
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
			object.PropertyStorage().Set(key, &PropertyDescriptor{
				Value:        desc.Value,
				Writable:     desc.Writable,
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

			if !SameValue(desc.Value, current.Value) {
				return false
			}
		}
	}

	if object != nil {
		if current.IsDataDescriptor() && desc.IsAccessorDescriptor() {
			// TODO: i. If Desc has a [[Configurable]] field, let configurable be Desc.[[Configurable]];
			//    else let configurable be current.[[Configurable]].
			configurable := desc.Configurable
			// TODO: ii. If Desc has a [[Enumerable]] field, let enumerable be Desc.[[Enumerable]]; else
			//     let enumerable be current.[[Enumerable]].
			enumerable := desc.Enumerable

			// TODO: iii. Replace the property named P of object O with a data property whose
			//      [[Configurable]] and [[Enumerable]] attributes are set to configurable and
			//      enumerable, respectively, and whose [[Value]] and [[Writable]] attributes are
			//      set to the value of the corresponding field in Desc if Desc has that field, or
			//      to the attribute's default value otherwise.
			object.PropertyStorage().Set(key, &PropertyDescriptor{
				Value:        desc.Value,
				Writable:     desc.Writable,
				Enumerable:   enumerable,
				Configurable: configurable,
			})
		} else {
			// TODO: i. For each field of Desc, set the corresponding attribute of the property named P
			//    of object O to the value of the field.
			object.PropertyStorage().Set(key, &PropertyDescriptor{
				Value:        desc.Value,
				Writable:     desc.Writable,
				Get:          desc.Get,
				Set:          desc.Set,
				Enumerable:   desc.Enumerable,
				Configurable: desc.Configurable,
			})
		}
	}

	return true
}

func InternalHasProperty(object ObjectType, key PropertyKey) bool {
	return OrdinaryHasProperty(object.(*Object), key)
}

func OrdinaryHasProperty(object *Object, key PropertyKey) bool {
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

func InternalGet(object ObjectType, key PropertyKey, receiver Value) Value {
	return OrdinaryGet(object, key, receiver)
}

func OrdinaryGet(object ObjectType, key PropertyKey, receiver Value) Value {
	desc := object.InternalMethods().GetOwnProperty(object, key)

	if desc == nil {
		parent := object.InternalMethods().GetPrototypeOf(object)
		if parent == nil {
			return UndefinedValue
		}

		return parent.InternalMethods().Get(parent, key, receiver)
	}

	if desc.IsDataDescriptor() {
		return desc.Value
	}

	if !desc.IsAccessorDescriptor() {
		panic("")
	}

	getter := desc.Get
	if getter == nil {
		fmt.Println("DEBUG: getter is nil")
		return UndefinedValue
	}

	return CallAssumeCallableNoArgs(NewValueFromObject(getter), receiver)
}

func InternalSet(object ObjectType, key PropertyKey, value Value, receiver Value) bool {
	return OrdinarySet(object.(*Object), key, value, receiver)
}

func OrdinarySet(object *Object, key PropertyKey, value Value, receiver Value) bool {
	ownDesc := object.InternalMethods().GetOwnProperty(object, key)
	return OrdinarySetWithOwnDescriptor(object, key, value, receiver, ownDesc)

}

// 10.1.9.2
func OrdinarySetWithOwnDescriptor(
	object *Object,
	key PropertyKey,
	value Value,
	receiver Value,
	ownDesc *PropertyDescriptor,
) bool {
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
			return false
		}

		r, isObject := receiver.(*ObjectValue)
		if !isObject {
			return false
		}
		receiverObject := r.Object.ToObject()

		existingDescriptor := object.InternalMethods().GetOwnProperty(receiverObject, key)

		if existingDescriptor != nil {
			if existingDescriptor.IsAccessorDescriptor() {
				return false
			}
			if !existingDescriptor.Writable {
				return false
			}

			valueDesc := &PropertyDescriptor{
				Value: value,
			}
			return receiverObject.InternalMethods().DefineOwnProperty(
				receiverObject, key, valueDesc,
			)
		} else {
			Assert(!receiverObject.PropertyStorage().Has(key))

			return receiverObject.CreateDataProperty(key, value)
		}
	}

	Assert(ownDesc.IsAccessorDescriptor())

	setter := ownDesc.Set
	if setter == nil {
		return false
	}
	_ = CallAssumeCallable(NewValueFromObject(setter), receiver, []Value{value})
	return true
}

func InternalDelete(object ObjectType, key PropertyKey) bool {
	return OrdinaryDelete(object.(*Object), key)
}

func OrdinaryDelete(object *Object, key PropertyKey) bool {
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
	return OrdinaryOwnPropertyKeys(object.(*Object))
}

func OrdinaryOwnPropertyKeys(object *Object) []PropertyKey {
	var keys []PropertyKey

	for key := range object.PropertyStorage().Properties {
		keys = append(keys, key)
	}

	return keys
}

func ObjectSameValue(x, y ObjectType) bool {
	return x == y
}

// 10.1.12
func OrdinaryObjectCreate(agent *Agent, proto ObjectType, internalSlotsList []string) *Object {
	obj := NewObject(agent, proto)

	return obj
}

// 10.1.13
func OrdinaryCreateFromConstructor(agent *Agent, constructor ObjectType, intrinsicDefaultProto string, internalSlotsList []string) *Object {
	// TODO: Assert
	proto := GetPrototypeFromConstructor(constructor, intrinsicDefaultProto)

	return OrdinaryObjectCreate(agent, proto, internalSlotsList)
}

// 10.1.14
func GetPrototypeFromConstructor(constructor ObjectType, intrinsicDefaultProto string) ObjectType {
	// TODO: Assert
	proto := constructor.ToObject().Get(NewStringPropertyKey("prototype"))

	switch p := proto.(type) {
	case *ObjectValue:
		return p.Object
	default:
		realm := constructor.ToObject().GetFunctionRealm()
		return realm.Intrinsics.Get(intrinsicDefaultProto).(*Object)
	}

}
