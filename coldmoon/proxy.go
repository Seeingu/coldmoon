package coldmoon

import (
	"github.com/samber/lo"
)

type ProxyObject struct {
	*Object
	Target  ObjectType
	Handler ObjectType
}

func NewProxyObject(agent *Agent, target, handler Value) *ProxyObject {
	if !ValueIsObject(target) {
		panic("TypeError")
	}
	if !ValueIsObject(handler) {
		panic("TypeError")
	}
	p := &ProxyObject{
		Object:  NewObject(agent, nil, "Proxy"),
		Target:  MustGetObject(target),
		Handler: MustGetObject(handler),
	}

	getPrototypeOf := func(o ObjectType) ObjectType {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()

		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("getPrototypeOf"))
		if trap == nil {
			return t.InternalMethods().GetPrototypeOf(t)
		}

		handlerPrototype := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t)},
		)
		handlerProtoObject, handlerProtoIsObject := handlerPrototype.(*ObjectValue)

		if !handlerProtoIsObject && handlerPrototype != NullValue {
			panic("TypeError")
		}

		extensibleTarget := t.IsExtensible()
		if extensibleTarget {
			if handlerProtoIsObject {
				return handlerProtoObject.Object
			} else {
				return nil
			}
		}

		targetProto := t.InternalMethods().GetPrototypeOf(t)
		if !SameValue(handlerPrototype, NewValueFromObject(targetProto)) {
			panic("TypeError")
		}

		return handlerProtoObject.Object
	}
	setPrototypeOf := func(object ObjectType, prototype ObjectType) bool {
		proxy := object.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("setPrototypeOf"))
		if trap == nil {
			return t.InternalMethods().SetPrototypeOf(t, prototype)
		}

		booleanTrapResult := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t), NewValueFromObject(prototype)},
		).ToBoolean()
		if !booleanTrapResult {
			return false
		}
		extensibleTarget := t.IsExtensible()
		if extensibleTarget {
			return true
		}
		targetProto := t.InternalMethods().GetPrototypeOf(t)
		if !SameValue(NewValueFromObject(prototype), NewValueFromObject(targetProto)) {
			panic("TypeError")
		}
		return true
	}
	isExtensible := func(o ObjectType) bool {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("isExtensible"))
		if trap == nil {
			return t.InternalMethods().IsExtensible(t)
		}

		booleanTrapResult := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t)},
		).ToBoolean()
		targetExtensible := t.IsExtensible()
		if booleanTrapResult != targetExtensible {
			panic("TypeError")
		}
		return booleanTrapResult
	}
	preventExtensions := func(o ObjectType) bool {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("preventExtensions"))
		if trap == nil {
			return t.InternalMethods().PreventExtensions(t)
		}

		booleanTrapResult := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t)},
		).ToBoolean()
		if booleanTrapResult {
			targetExtensible := t.IsExtensible()
			if targetExtensible {
				panic("TypeError")
			}
		}
		return booleanTrapResult
	}
	getOwnProperty := func(o ObjectType, pk PropertyKey) *PropertyDescriptor {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("getOwnPropertyDescriptor"))
		if trap == nil {
			return t.InternalMethods().GetOwnProperty(t, pk)
		}

		trapResultObjValue := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t), pk.ToValue()},
		)
		_, trapResultIsObject := trapResultObjValue.(*ObjectValue)
		if !trapResultIsObject {
			panic("TypeError")
		}
		targetDesc := t.InternalMethods().GetOwnProperty(t, pk)
		if trapResultObjValue == UndefinedValue {
			if targetDesc == nil {
				return nil
			}
			if !targetDesc.Configurable {
				panic("TypeError")
			}

			extensibleTarget := t.IsExtensible()
			if !extensibleTarget {
				panic("TypeError")
			}
		}
		return nil
	}
	defineOwnProperty := func(o ObjectType, pk PropertyKey, desc *PropertyDescriptor) bool {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("defineProperty"))
		if trap == nil {
			return t.InternalMethods().DefineOwnProperty(t, pk, desc)
		}

		descObj := desc.FromPropertyDescriptor(agent, desc)
		booleanTrapResult := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t), pk.ToValue(), NewValueFromObject(descObj)},
		).ToBoolean()

		if !booleanTrapResult {
			return false
		}
		targetDesc := t.InternalMethods().GetOwnProperty(t, pk)
		extensibleTarget := t.IsExtensible()
		settingConfigFalse := desc.Configurable == false
		if targetDesc == nil {
			if !extensibleTarget {
				panic("TypeError")
			} else {
				if settingConfigFalse {
					panic("TypeError")
				}
			}
		} else {
			if !IsCompatiblePropertyDescriptor(extensibleTarget, desc, targetDesc) {
				panic("TypeError")
			}
			if settingConfigFalse && targetDesc.Configurable {
				panic("TypeError")
			}

			if desc.IsDataDescriptor() && !targetDesc.Configurable && targetDesc.Writable &&
				!desc.Writable {
				panic("TypeError")
			}
		}
		return true
	}
	hasProperty := func(o ObjectType, pk PropertyKey) bool {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("has"))
		if trap == nil {
			return t.InternalMethods().HasProperty(t, pk)
		}

		booleanTrapResult := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t), pk.ToValue()},
		).ToBoolean()
		if !booleanTrapResult {
			targetDesc := t.InternalMethods().GetOwnProperty(t, pk)
			if targetDesc != nil && !targetDesc.Configurable {
				panic("TypeError")
			}

			extensibleTarget := t.IsExtensible()
			if !extensibleTarget {
				panic("TypeError")
			}
		}
		return booleanTrapResult
	}
	get := func(o ObjectType, pk PropertyKey, receiver Value) Value {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("get"))
		if trap == nil {
			return t.InternalMethods().Get(t, pk, receiver)
		}

		trapResult := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t), pk.ToValue(), receiver},
		)
		targetDesc := t.InternalMethods().GetOwnProperty(t, pk)
		if targetDesc != nil && !targetDesc.Configurable {
			if targetDesc.IsDataDescriptor() && !targetDesc.Writable {
				if !SameValue(trapResult, targetDesc.Value) {
					panic("TypeError")
				}
			}
			if targetDesc.IsAccessorDescriptor() && targetDesc.Get == nil {
				if trapResult != UndefinedValue {
					panic("TypeError")
				}
			}
		}
		return trapResult
	}
	set := func(o ObjectType, pk PropertyKey, v Value, receiver Value) bool {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("set"))
		if trap == nil {
			return t.InternalMethods().Set(t, pk, v, receiver)
		}

		booleanTrapResult := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t), pk.ToValue(), v, receiver},
		).ToBoolean()
		if !booleanTrapResult {
			return false
		}
		targetDesc := t.InternalMethods().GetOwnProperty(t, pk)
		if targetDesc != nil && !targetDesc.Configurable {
			if targetDesc.IsDataDescriptor() && !targetDesc.Writable {
				if !SameValue(v, targetDesc.Value) {
					panic("TypeError")
				}
			}
			if targetDesc.IsAccessorDescriptor() && targetDesc.Set == nil {
				if v != UndefinedValue {
					panic("TypeError")
				}
			}
		}
		return true
	}
	proxyDelete := func(o ObjectType, pk PropertyKey) bool {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("deleteProperty"))
		if trap == nil {
			return t.InternalMethods().Delete(t, pk)
		}

		booleanTrapResult := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t), pk.ToValue()},
		).ToBoolean()
		if !booleanTrapResult {
			return false
		}
		targetDesc := t.InternalMethods().GetOwnProperty(t, pk)
		if targetDesc == nil {
			return true
		}
		if !targetDesc.Configurable {
			panic("TypeError")
		}
		extensibleTarget := t.IsExtensible()
		if !extensibleTarget {
			panic("TypeError")
		}
		return true
	}
	ownPropertyKeys := func(o ObjectType) []PropertyKey {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("ownKeys"))
		if trap == nil {
			return t.InternalMethods().OwnPropertyKeys(t)
		}

		trapResultArray := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t)},
		)

		elements := CreateListFromArrayLike(agent, trapResultArray)
		var trapResult []PropertyKey
		uniquePropertyKeys := make(map[PropertyKey]struct{})
		for _, element := range elements {
			var pk PropertyKey
			switch el := element.(type) {
			case *StringValue:
				pk = NewStringPropertyKey(el.Data)
			case *SymbolValue:
				pk = NewSymbolPropertyKey(el)
			default:
				panic("unreachable")
			}
			trapResult = append(trapResult, pk)
			uniquePropertyKeys[pk] = struct{}{}
		}
		if len(trapResult) != len(uniquePropertyKeys) {
			panic("TypeError")
		}

		extensibleTarget := t.IsExtensible()
		targetKeys := t.InternalMethods().OwnPropertyKeys(t)
		var targetConfigurableKeys []PropertyKey
		var targetNonConfigurableKeys []PropertyKey
		for _, key := range targetKeys {
			desc := t.InternalMethods().GetOwnProperty(t, key)
			if desc != nil && !desc.Configurable {
				targetNonConfigurableKeys = append(targetNonConfigurableKeys, key)
			} else {
				targetConfigurableKeys = append(targetConfigurableKeys, key)
			}
		}

		if extensibleTarget && len(targetNonConfigurableKeys) == 0 {
			return trapResult
		}

		var uncheckedResultKeys []PropertyKey
		for _, key := range trapResult {
			uncheckedResultKeys = append(uncheckedResultKeys, key)
		}
		for _, key := range targetNonConfigurableKeys {
			index := lo.IndexOf(uncheckedResultKeys, key)
			if index == -1 {
				panic("TypeError")
			}
			uncheckedResultKeys = append(uncheckedResultKeys[:index], uncheckedResultKeys[index+1:]...)
		}

		if extensibleTarget {
			return trapResult
		}
		for _, key := range targetConfigurableKeys {
			index := lo.IndexOf(uncheckedResultKeys, key)
			if index == -1 {
				panic("TypeError")
			}
			uncheckedResultKeys = append(uncheckedResultKeys[:index], uncheckedResultKeys[index+1:]...)
		}
		if len(uncheckedResultKeys) != 0 {
			panic("TypeError")
		}
		return trapResult
	}
	call := func(o ObjectType, this Value, arguments []Value) Value {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("apply"))
		if trap == nil {
			return t.InternalMethods().Call(t, this, arguments)
		}

		argArray := CreateArrayFromList(agent, arguments)
		return NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t), this, NewValueFromObject(argArray)},
		)
	}
	construct := func(o ObjectType, arguments []Value, newTarget ObjectType) ObjectType {
		proxy := o.(*ProxyObject)
		proxy.validateNonRevokedProxy()
		t := proxy.Target
		Assert(IsConstructor(NewValueFromObject(t)))
		h := proxy.Handler
		trap := GetMethod(agent, NewValueFromObject(h), NewStringPropertyKey("construct"))
		if trap == nil {
			return t.InternalMethods().Construct(t, arguments, newTarget)
		}

		argArray := CreateArrayFromList(agent, arguments)
		newObj := NewValueFromObject(trap).CallAssumeCallable(
			NewValueFromObject(h),
			[]Value{NewValueFromObject(t), NewValueFromObject(argArray), NewValueFromObject(newTarget)},
		)
		if !ValueIsObject(newObj) {
			panic("TypeError")
		}
		return MustGetObject(newObj)
	}

	p.InternalMethods().GetPrototypeOf = getPrototypeOf
	p.InternalMethods().SetPrototypeOf = setPrototypeOf
	p.InternalMethods().IsExtensible = isExtensible
	p.InternalMethods().PreventExtensions = preventExtensions
	p.InternalMethods().GetOwnProperty = getOwnProperty
	p.InternalMethods().DefineOwnProperty = defineOwnProperty
	p.InternalMethods().HasProperty = hasProperty
	p.InternalMethods().Get = get
	p.InternalMethods().Set = set
	p.InternalMethods().Delete = proxyDelete
	p.InternalMethods().OwnPropertyKeys = ownPropertyKeys

	if IsCallable(target) {
		p.InternalMethods().Call = call
		if IsConstructor(target) {
			p.InternalMethods().Construct = construct
		}
	}
	return p
}

// 10.5.14
func (p *ProxyObject) validateNonRevokedProxy() {
	if p.Target == nil {
		panic("TypeError")
	}
	Assert(p.Handler != nil)
}

var ProxyCreate = NewProxyObject

func NewProxyConstructor(realm *Realm) ObjectType {
	agent := realm.Agent

	var behavior BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		target := arguments[0]
		handler := arguments[1]
		if newTarget == nil {
			panic("TypeError")
		}
		return NewValueFromObject(ProxyCreate(agent, target, handler))
	}

	obj := CreateBuiltinFunction(agent, behavior, 2, "Proxy", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})

	var revocable BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		target := argumentsList[0]
		handler := argumentsList[1]

		proxy := ProxyCreate(agent, target, handler)

		var revokerClosure BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
			f := agent.ActiveFunctionObject().(*BuiltinFunction)
			p := f.RevocableProxy.(*ProxyObject)
			if p == nil {
				return UndefinedValue
			}
			f.RevocableProxy = nil
			p.Target = nil
			p.Handler = nil
			return UndefinedValue
		}
		revoker := CreateBuiltinFunction(agent, revokerClosure, 0, "", builtinFunctionArgs{
			revocableProxy: proxy,
		})

		result := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
		result.CreateDataPropertyOrThrow(NewStringPropertyKey("proxy"), NewValueFromObject(proxy))
		result.CreateDataPropertyOrThrow(NewStringPropertyKey("revoke"), NewValueFromObject(revoker))
		return NewValueFromObject(result)
	}
	DefineBuiltinFunction(obj, "revocable", revocable, 2, realm)

	return obj
}

func (p *ProxyObject) GetFunctionRealm() *Realm {
	p.validateNonRevokedProxy()
	target := p.Target
	return target.GetFunctionRealm()
}
