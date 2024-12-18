package coldmoon

import (
	"fmt"

	"github.com/samber/lo"
)

type IntegrityLevel int

const (
	IntegrityLevelSealed IntegrityLevel = iota
	IntegrityLevelFrozen
)

type Data struct {
	prototype       ObjectType
	extensible      bool
	agent           *Agent
	internalMethods InternalMethods
	propertyStorage PropertyStorage
	privateElements map[PrivateName]*PrivateElement
}

type Object struct {
	ObjectType
	typeName string
	data     *Data
	// ref is used to store the reference of the object
	ref ObjectType
}

func (o *Object) ToObject() *Object {
	return o
}

func (o *Object) ToValue() Value {
	return NewValueFromObject(o.Ref())
}

// Ref returns the reference of the object
// If the object does not have a reference, it returns itself
func (o *Object) Ref() ObjectType {
	if o.ref == nil {
		return o
	}
	return o.ref
}

func NewObject(agent *Agent, prototype ObjectType, typeName string) *Object {
	o := &Object{
		typeName: typeName,
		data: &Data{
			agent:           agent,
			prototype:       prototype,
			extensible:      true,
			internalMethods: NewInternalMethods(),
			propertyStorage: NewPropertyStorage(),
		},
	}
	return o
}

var EmptyObject = &Object{}

func (o *Object) Prototype() ObjectType {
	return o.data.prototype
}

func (o *Object) SetPrototype(p ObjectType) {
	o.data.prototype = p
}

func (o *Object) Extensible() bool {
	return o.data.extensible
}

func (o *Object) SetExtensible(v bool) {
	o.data.extensible = v
}

func (o *Object) Agent() *Agent {
	return o.data.agent
}

func (o *Object) InternalMethods() *InternalMethods {
	return &o.data.internalMethods
}

func (o *Object) PropertyStorage() *PropertyStorage {
	return &o.data.propertyStorage
}

// 7.1.1.1
func (o *Object) OrdinaryToPrimitive(hint PreferredType) Value {
	var methodNames []string
	switch hint {
	case PreferredTypeString:
		methodNames = []string{"toString", "valueOf"}
	default:
		methodNames = []string{"valueOf", "toString"}
	}

	for _, name := range methodNames {
		method := o.Get(NewStringPropertyKey(name))
		if IsCallable(method) {
			result := method.CallNoArgs(o.ToValue())
			if _, isObject := result.(*ObjectValue); !isObject {
				return result
			}
		}
	}

	message := "Could not convert object to primitive"
	return o.Agent().ThrowException(TypeError, message)
}

// 7.2.5
func (o *Object) IsExtensible() bool {
	return o.InternalMethods().IsExtensible(o)
}

// 7.3.2
func (o *Object) Get(key PropertyKey) Value {
	return o.InternalMethods().Get(o.Ref(), key, o.ToValue())
}

// 7.3.4
func (o *Object) Set(key PropertyKey, value Value, throw setThrowType) {
	success := o.InternalMethods().Set(o.Ref(), key, value, (o).ToValue())
	if !success && throw == setThrowTypeThrow {
		o.Agent().ThrowException(TypeError, "SetObject failed")
	}
}

// 7.3.5
func (o *Object) CreateDataProperty(key PropertyKey, value Value) bool {
	newDesc := o.InternalMethods().DefineOwnProperty(o.Ref(), key, &PropertyDescriptor{
		Value:        value,
		Writable:     true,
		Enumerable:   true,
		Configurable: true,
	})
	return newDesc
}

// 7.3.6

func (o *Object) CreateDataPropertyOrThrow(key PropertyKey, value Value) bool {
	success := o.Ref().CreateDataProperty(key, value)
	if !success {
		o.Agent().ThrowException(TypeError, "CreateDataPropertyOrThrow failed")
	}
	return success
}

// 7.3.7
func (o *Object) CreateNonEnumerableDataProperty(key PropertyKey, value Value) bool {
	for _, p := range o.PropertyStorage().Properties {
		Assert(p.Configurable)
	}

	newDesc := o.DefinePropertyOrThrow(key, &PropertyDescriptor{
		Value:        value,
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
	return newDesc
}

// 7.3.8
func (o *Object) DefinePropertyOrThrow(key PropertyKey, desc *PropertyDescriptor) bool {
	success := o.InternalMethods().DefineOwnProperty(o.Ref(), key, desc)
	if !success {
		o.Agent().ThrowException(TypeError, "DefinePropertyOrThrow failed")
	}
	return success
}

// 7.3.9
func (o *Object) DeletePropertyOrThrow(key PropertyKey) bool {
	success := o.InternalMethods().Delete(o.Ref(), key)
	if !success {
		o.Agent().ThrowException(TypeError, "DeletePropertyOrThrow failed")
	}
	return success
}

// 7.3.11
func (o *Object) HasProperty(key PropertyKey) bool {
	return o.InternalMethods().HasProperty(o, key)
}

// 7.3.12
func ObjectHasOwnProperty(o ObjectType, key PropertyKey) bool {
	desc := o.InternalMethods().GetOwnProperty(o, key)
	return desc != nil
}

// 7.3.14
func ObjectConstruct(
	o ObjectType,
	_argumentLists []Value,
	_newTarget ObjectType,
) ObjectType {
	newTarget := _newTarget
	if newTarget == nil {
		newTarget = o
	}
	return o.InternalMethods().Construct(o, _argumentLists, newTarget)
}

func (o *Object) Construct(
	argumentLists []Value,
	newTarget ObjectType,
) ObjectType {
	if newTarget == nil {
		newTarget = o
	}
	return o.InternalMethods().Construct(o, argumentLists, newTarget)
}

// 7.3.15
func SetIntegrityLevel(o ObjectType, level IntegrityLevel) bool {
	status := o.InternalMethods().PreventExtensions(o)

	if !status {
		return false
	}

	keys := o.InternalMethods().OwnPropertyKeys(o)
	switch level {
	case IntegrityLevelSealed:
		for _, k := range keys {
			o.DefinePropertyOrThrow(k, &PropertyDescriptor{
				Configurable: false,
			})
		}
	case IntegrityLevelFrozen:
		for _, k := range keys {
			currentDesc := o.InternalMethods().GetOwnProperty(o, k)
			var desc *PropertyDescriptor

			if currentDesc != nil {
				if currentDesc.IsAccessorDescriptor() {
					desc = &PropertyDescriptor{
						Configurable: false,
					}
				} else {
					desc = &PropertyDescriptor{
						Configurable: false,
						Writable:     false,
					}
				}
				o.DefinePropertyOrThrow(k, desc)
			}
		}
	}
	return true
}

// 7.3.16
func TestIntegrityLevel(o ObjectType, level IntegrityLevel) bool {
	extensible := o.IsExtensible()
	if extensible {
		return false
	}

	keys := o.InternalMethods().OwnPropertyKeys(o)

	for _, k := range keys {
		currentDesc := o.InternalMethods().GetOwnProperty(o, k)
		if currentDesc == nil {
			continue
		}

		if !currentDesc.Configurable {
			return false
		}

		if level == IntegrityLevelFrozen && currentDesc.IsDataDescriptor() {
			if currentDesc.Writable {
				return false
			}
		}
	}
	return true
}

// 7.3.18
func (o *Object) LengthOfArrayLike() JSInt {
	return ToLength(o.Agent(), o.Get(NewStringPropertyKey("length")))
}

// 7.3.22
func (o *Object) SpeciesConstructor(defaultConstructor ObjectType) CompletionObject {
	objectRef := o.Ref()
	c := objectRef.Get(NewStringPropertyKey("constructor"))
	if c == UndefinedValue {
		return NewCompletionObject(defaultConstructor)
	}
	if !ValueIsObject(c) {
		return NewCompletionObjectError(objectRef.Agent().ThrowException(TypeError, c.String()+" is not an object"))
	}
	cObject := MustGetObject(c)
	s := cObject.Get(NewSymbolPropertyKey(WellKnownSymbols[WellKnownSymbolsSpecies]))
	if s == UndefinedValue || s == NullValue {
		return NewCompletionObject(defaultConstructor)
	}
	if IsConstructor(s) {
		return NewCompletionObject(MustGetObject(s))
	}
	return NewCompletionObject(defaultConstructor)
}

func (o *Object) ToCompletion() CompletionValue {
	return (o).ToValue().ToCompletion()
}

// MARK: - 7.3.23
type objectOwnPropertiesKind int

const (
	objectOwnPropertiesKindKey objectOwnPropertiesKind = iota
	objectOwnPropertiesKindValue
	objectOwnPropertiesKindKeyAndValue
)

func (o *Object) EnumerableOwnProperties(kind objectOwnPropertiesKind) (results []Value) {
	ownKeys := o.InternalMethods().OwnPropertyKeys(o)

	for _, key := range ownKeys {
		if _, err := key.GetIndex(); err == nil {
			desc := o.InternalMethods().GetOwnProperty(o, key)
			if desc != nil && desc.Enumerable {
				switch kind {
				case objectOwnPropertiesKindKey:
					results = append(results, key.ToValue())
				case objectOwnPropertiesKindValue:
					results = append(results, o.Get(key))
				case objectOwnPropertiesKindKeyAndValue:
					keyValue := key.ToValue()
					entry := CreateArrayFromList(o.Agent(), []Value{
						keyValue,
						o.Get(key),
					})
					results = append(results, (entry).ToValue())
				}
			}
		}
	}
	return
}

// 7.3.24
func (o *Object) GetFunctionRealm() *Realm {
	return o.Agent().CurrentRealm()
}

// MARK: - 7.3.25

func (o *Object) CopyDataProperties(source Value, excludedItems []PropertyKey) {
	if source == UndefinedValue || source == NullValue {
		return
	}
	from := ValueToObject(o.Agent(), source)
	keys := from.InternalMethods().OwnPropertyKeys(from)

	for _, key := range keys {
		excluded := false
		if lo.Contains(excludedItems, key) {
			excluded = true
		}
		if !excluded {
			desc := from.InternalMethods().GetOwnProperty(from, key)
			if desc != nil && desc.Enumerable {
				propValue := from.Get(key)
				o.CreateDataPropertyOrThrow(key, propValue)
			}
		}
	}
}

func (o *Object) PrivateFieldAdd(privateName *PrivateName, value Value) {
	// TODO
}

// MARK: - DefineField

func (o *Object) DefineField(field *ClassFieldDefinition) {
	fieldName := field.Name
	var initializer Value = UndefinedValue
	if field.Initializer != nil {
		initializer = field.Initializer.ToValue()
	}

	switch name := fieldName.(type) {
	case PropertyKey:
		o.CreateDataPropertyOrThrow(name, initializer)
	case *PrivateName:
		o.PrivateFieldAdd(name, initializer)
	default:
		panic("unreachable")
	}
}

func (o *Object) PrivateElementFind(privateName PrivateName) *PrivateElement {
	return o.data.privateElements[privateName]
}

// 7.3.26
func (o *Object) PrivateMethodOrAccessorAdd(
	privateName PrivateName,
	method *PrivateElement,
) {
	Assert(method.Kind == PrivateElementKindMethod || method.Kind == PrivateElementKindAccessor)
	o.Agent().HostHooks.HostEnsureCanAddPrivateElement()
	entry := o.PrivateElementFind(privateName)
	if entry != nil {
		panic("TypeError")
	}
	o.data.privateElements[privateName] = method
}

// 7.3.30
func (o *Object) PrivateGet(privateName PrivateName) Value {
	entry := o.PrivateElementFind(privateName)
	if entry == nil {
		return o.Agent().ThrowTypeError("PrivateGet failed")
	}
	switch entry.Kind {
	case PrivateElementKindField:
		return entry.Value
	case PrivateElementKindMethod:
		return entry.Value
	case PrivateElementKindAccessor:
		getter := entry.Get
		if getter == nil {
			return o.Agent().ThrowTypeError("PrivateGet failed: getter is nil")
		}
		return getter.ToValue().Call(o.ToValue(), []Value{})
	}
	panic("unreachable")
}

// MARK: - InitializeInstanceElements

func (o *Object) InitializeInstanceElements(constructor ObjectType) {
	methods := constructor.(InternalSlotPrivateMethods).PrivateMethods()
	for _, method := range methods {
		o.PrivateMethodOrAccessorAdd(method.PrivateName, method.PrivateElement)
	}

	fields := constructor.(InternalSlotFields).Fields()
	for _, field := range fields {
		o.DefineField(field)
	}
}

// MARK: - Object Constructor

// 20.1.1
func NewObjectConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		value := argumentsList[0]
		if newTarget != nil && newTarget != agent.ActiveFunctionObject() {
			return (OrdinaryCreateFromConstructor(
				agent, newTarget, "%Object.prototype%", []string{},
			)).ToValue()
		}

		if value == nil {
			return OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, []string{}).ToValue()
		}

		return value
	}

	object := CreateBuiltinFunction(
		agent,
		behavior,
		1, "Object",
		builtinFunctionArgs{
			realm:         realm,
			prototype:     realm.Intrinsics.FunctionPrototype,
			prefix:        "",
			isConstructor: true,
		},
	)

	var create BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		o := arguments[0]
		properties := arguments[1]
		if !ValueIsObject(o) {
			panic("TypeError")
		}

		obj := OrdinaryObjectCreate(agent, MustGetObject(o), []string{})

		if properties != nil {
			return (objectDefineProperties(agent, obj, properties)).ToValue()
		}

		return (obj).ToValue()
	}

	var defineProperties BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
		o := arguments[0]
		properties := arguments[1]
		if !ValueIsObject(o) {
			panic("TypeError")
		}
		return (objectDefineProperties(agent, MustGetObject(o), properties)).ToValue()
	}

	defineProperty := func(this Value, arguments []Value, newTarget ObjectType) Value {
		o := arguments[0]
		property := arguments[1]
		attributes := arguments[2]
		if !ValueIsObject(o) {
			panic("TypeError")
		}

		key := ToPropertyKey(agent, property)
		desc := ToPropertyDescriptor(agent, attributes)
		MustGetObject(o).DefinePropertyOrThrow(key, desc)

		return o
	}

	// 20.1.2.6
	var freeze BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj, ok := objectValue.(*ObjectValue)
		if !ok {
			return obj
		}

		status := SetIntegrityLevel(obj.Object, IntegrityLevelFrozen)
		if !status {
			realm.Agent.ThrowException(TypeError, "SetIntegrityLevel failed")
		}
		return obj
	}

	// 20.1.2.8
	var getOwnPropertyDescriptor BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := args[0]
		p := args[1]
		obj := ValueToObject(agent, o)
		key := ToPropertyKey(agent, p)
		desc := obj.InternalMethods().GetOwnProperty(obj, key)

		if desc == nil {
			return UndefinedValue
		}
		return (desc.FromPropertyDescriptor(agent, desc)).ToValue()
	}
	// 20.1.2.9
	var getOwnPropertyDescriptors BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := args[0]
		obj := ValueToObject(agent, o)

		ownKeys := obj.InternalMethods().OwnPropertyKeys(obj)

		descriptors := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, []string{})
		for _, key := range ownKeys {
			desc := obj.InternalMethods().GetOwnProperty(obj, key)
			if desc != nil {
				descValue := (desc.FromPropertyDescriptor(agent, desc)).ToValue()
				descriptors.CreateDataPropertyOrThrow(key, descValue)
			}
		}
		return (descriptors).ToValue()
	}

	// 20.1.2.12
	var getPrototypeOf BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := args[0]

		obj := ValueToObject(agent, o)
		proto := obj.InternalMethods().GetPrototypeOf(obj)
		return (proto).ToValue()
	}

	// 20.1.2.15
	var objectIs BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		arg1 := args[0]
		arg2 := args[1]
		return NewBooleanValue(SameValue(arg1, arg2))
	}

	// 20.1.2.16
	var isExtensible BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj, ok := objectValue.(*ObjectValue)
		if !ok {
			return NewBooleanValue(false)
		}
		return NewBooleanValue(obj.Object.IsExtensible())
	}

	// 20.1.2.17
	var isFrozen BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj, ok := objectValue.(*ObjectValue)
		if !ok {
			return NewBooleanValue(true)
		}
		return NewBooleanValue(TestIntegrityLevel(obj.Object, IntegrityLevelFrozen))
	}

	// 20.1.2.18
	var isSealed BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj, ok := objectValue.(*ObjectValue)
		if !ok {
			return NewBooleanValue(true)
		}
		return NewBooleanValue(TestIntegrityLevel(obj.Object, IntegrityLevelSealed))
	}

	// 20.1.2.20
	var preventExtensions BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj, ok := objectValue.(*ObjectValue)
		if !ok {
			return obj
		}

		status := obj.Object.InternalMethods().PreventExtensions(obj.Object)
		if !status {
			realm.Agent.ThrowException(TypeError, "PreventExtensions failed")
		}
		return obj
	}

	// 20.1.2.21
	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        (realm.Intrinsics.ObjectPrototype).ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// 20.1.2.22
	var seal BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj, ok := objectValue.(*ObjectValue)
		if !ok {
			return obj
		}

		status := SetIntegrityLevel(obj.Object, IntegrityLevelSealed)
		if !status {
			realm.Agent.ThrowException(TypeError, "SetIntegrityLevel failed")
		}
		return obj
	}

	// 20.1.2.23
	var setPrototypeOf BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		proto := args[1]

		RequireObjectCoercible(agent, objectValue)

		obj, ok := objectValue.(*ObjectValue)
		if !ok && proto != nil {
			panic("TypeError")
		}
		if !ok {
			return obj
		}

		var protoObj ObjectType
		if po, ok := proto.(*ObjectValue); ok {
			protoObj = po.Object
		}
		status := obj.Object.InternalMethods().SetPrototypeOf(obj.Object, protoObj)
		if !status {
			panic("TypeError")
		}

		return obj
	}

	var hasOwn BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		key := args[1]

		obj := ValueToObject(agent, objectValue)
		p := ToPropertyKey(agent, key)
		return NewBooleanValue(ObjectHasOwnProperty(obj, p))
	}

	var entries BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj := ValueToObject(agent, objectValue)
		entryList := obj.EnumerableOwnProperties(objectOwnPropertiesKindKeyAndValue)
		return (CreateArrayFromList(agent, entryList)).ToValue()
	}
	var keys BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj := ValueToObject(agent, objectValue)
		keyList := obj.EnumerableOwnProperties(objectOwnPropertiesKindKey)
		return (CreateArrayFromList(agent, keyList)).ToValue()
	}
	var values BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj := ValueToObject(agent, objectValue)
		valueList := obj.EnumerableOwnProperties(objectOwnPropertiesKindValue)
		return (CreateArrayFromList(agent, valueList)).ToValue()
	}
	var assign BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		target := args[0]
		to := ValueToObject(agent, target)
		sources := args[1:]
		if len(sources) == 0 {
			return (to).ToValue()
		}
		for _, nextSource := range sources {
			if nextSource != UndefinedValue && nextSource != NullValue {
				from := ValueToObject(agent, nextSource)
				pKeys := from.InternalMethods().OwnPropertyKeys(from)
				for _, nextKey := range pKeys {
					desc := from.InternalMethods().GetOwnProperty(from, nextKey)
					if desc != nil && desc.Enumerable {
						propValue := from.Get(nextKey)
						to.Set(nextKey, propValue, setThrowTypeThrow)
					}
				}
			}
		}
		return (to).ToValue()
	}
	var getOwnPropertyNames BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj := ValueToObject(agent, objectValue)
		keys := obj.InternalMethods().OwnPropertyKeys(obj)
		keyNames := lo.Filter(keys, func(key PropertyKey, _ int) bool {
			_, ok := key.(SymbolPropertyKey)
			return !ok
		})
		keyValues := lo.Map(keyNames, func(key PropertyKey, _ int) Value {
			return key.ToValue()
		})
		return (CreateArrayFromList(agent, keyValues)).ToValue()
	}
	var getOwnPropertySymbols BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj := ValueToObject(agent, objectValue)
		keys := obj.InternalMethods().OwnPropertyKeys(obj)
		symbols := lo.Filter(keys, func(key PropertyKey, _ int) bool {
			_, ok := key.(SymbolPropertyKey)
			return ok
		})
		symbolValues := lo.Map(symbols, func(key PropertyKey, _ int) Value {
			return key.ToValue()
		})
		return (CreateArrayFromList(agent, symbolValues)).ToValue()
	}
	fromEntries := func(this Value, args []Value, newTarget ObjectType) Value {
		iterable := args[0]
		RequireObjectCoercible(agent, iterable)
		obj := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, []string{})
		type Captures struct {
			object ObjectType
		}
		var closure BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
			f := agent.ActiveFunctionObject()
			captures := f.(*BuiltinFunction).AdditionalFieldsV2.(*Captures)
			k := args[0]
			v := args[1]
			propertyKey := ToPropertyKey(agent, k)
			captures.object.CreateDataPropertyOrThrow(propertyKey, v)
			return UndefinedValue
		}

		adder := CreateBuiltinFunction(agent, closure, 2, "", builtinFunctionArgs{
			additionalFieldsV2: &Captures{object: obj},
		})
		return AddEntriesFromIterable(agent, obj, iterable, adder).ToValue()
	}

	DefineBuiltinFunction(object, "hasOwn", hasOwn, 2, realm)
	DefineBuiltinFunction(object, "getPrototypeOf", getPrototypeOf, 1, realm)
	DefineBuiltinFunction(object, "create", create, 2, realm)
	DefineBuiltinFunction(object, "defineProperties", defineProperties, 2, realm)
	DefineBuiltinFunction(object, "defineProperty", defineProperty, 3, realm)
	DefineBuiltinFunction(object, "getOwnPropertyDescriptor", getOwnPropertyDescriptor, 2, realm)
	DefineBuiltinFunction(object, "getOwnPropertyDescriptors", getOwnPropertyDescriptors, 1, realm)
	DefineBuiltinFunction(object, "getOwnPropertyNames", getOwnPropertyNames, 1, realm)
	DefineBuiltinFunction(object, "getOwnPropertySymbols", getOwnPropertySymbols, 1, realm)
	DefineBuiltinFunction(object, "freeze", freeze, 1, realm)
	DefineBuiltinFunction(object, "is", objectIs, 2, realm)
	DefineBuiltinFunction(object, "isExtensible", isExtensible, 1, realm)
	DefineBuiltinFunction(object, "isFrozen", isFrozen, 1, realm)
	DefineBuiltinFunction(object, "isSealed", isSealed, 1, realm)
	DefineBuiltinFunction(object, "preventExtensions", preventExtensions, 1, realm)
	DefineBuiltinFunction(object, "seal", seal, 1, realm)
	DefineBuiltinFunction(object, "setPrototypeOf", setPrototypeOf, 2, realm)
	DefineBuiltinFunction(object, "entries", entries, 1, realm)
	DefineBuiltinFunction(object, "keys", keys, 1, realm)
	DefineBuiltinFunction(object, "values", values, 1, realm)
	DefineBuiltinFunction(object, "assign", assign, 2, realm)
	DefineBuiltinFunction(object, "fromEntries", fromEntries, 1, realm)

	// 20.1.3.1
	DefineBuiltinPropertyV(realm.Intrinsics.ObjectPrototype, "constructor", (object).ToValue())

	return object
}

// NewObjectPrototypeSkeleton init %Object.prototype% at first
func NewObjectPrototypeSkeleton(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, nil, "Object")
	return object
}

func NewObjectPrototypeWithObject(realm *Realm, object ObjectType) ObjectType {
	agent := realm.Agent
	object.InternalMethods().SetPrototypeOf = ImmutableSetPrototypeOf
	realm.Intrinsics.ObjectPrototype = object

	valueOf := func(this Value, args []Value, newTarget ObjectType) Value {
		return (ValueToObject(agent, this)).ToValue()
	}

	// 20.1.3.6 toString
	toString := func(this Value, args []Value, newTarget ObjectType) Value {
		if this == UndefinedValue {
			return NewStringValue("[object Undefined]")
		}
		if this == NullValue {
			return NewStringValue("[object Null]")
		}
		o := ValueToObject(agent, this)
		_isArray := IsArray(this)
		var builtInTag string
		if _isArray {
			builtInTag = "Array"
		} else if o.InternalMethods().Call != nil {
			builtInTag = "Function"
		} else if ObjectIs[*BooleanObject](o) {
			builtInTag = "Boolean"
		} else if ObjectIs[*ErrorObject](o) {
			builtInTag = "Error"
		} else if ObjectIs[*NumberObject](o) {
			builtInTag = "Number"
		} else if ObjectIs[*StringObject](o) {
			builtInTag = "String"
		} else if ObjectIs[*DateObject](o) {
			builtInTag = "Date"
		} else if ObjectIs[*RegExpObject](o) {
			builtInTag = "RegExp"
		} else {
			builtInTag = "Object"
		}

		symbol := WellKnownSymbols[WellKnownSymbolsToStringTag]
		tagValue := o.Get(NewSymbolPropertyKey(symbol))

		var tag string
		if stringTag, ok := tagValue.(*StringValue); ok {
			tag = stringTag.Data
		} else {
			tag = builtInTag
		}
		return NewStringValue("[object " + tag + "]")
	}
	hasOwnProperty := func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		p := ToPropertyKey(agent, args[0])
		return NewBooleanValue(ObjectHasOwnProperty(o, p))
	}
	isPrototypeOf := func(this Value, args []Value, newTarget ObjectType) Value {
		v := args[0]
		if !ValueIsObject(v) {
			return NewBooleanValue(false)
		}
		o := ValueToObject(agent, this)
		target := ValueToObject(agent, v)
		for {
			target = target.InternalMethods().GetPrototypeOf(target)
			if target == nil {
				return NewBooleanValue(false)
			}
			if ObjectSameValue(target, o) {
				return NewBooleanValue(true)
			}
		}
	}
	propertyIsEnumerable := func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		p := ToPropertyKey(agent, args[0])
		desc := o.InternalMethods().GetOwnProperty(o, p)
		if desc == nil {
			return NewBooleanValue(false)
		}
		return NewBooleanValue(desc.Enumerable)
	}
	toLocaleString := func(this Value, args []Value, newTarget ObjectType) Value {
		o := ValueToObject(agent, this)
		return ValueInvoke(agent, o.ToValue(), NewStringPropertyKey("toString"), []Value{})
	}

	DefineBuiltinFunction(object, "toString", toString, 0, realm)
	DefineBuiltinFunction(object, "valueOf", valueOf, 0, realm)
	DefineBuiltinFunction(object, "hasOwnProperty", hasOwnProperty, 1, realm)
	DefineBuiltinFunction(object, "isPrototypeOf", isPrototypeOf, 1, realm)
	DefineBuiltinFunction(object, "propertyIsEnumerable", propertyIsEnumerable, 1, realm)
	DefineBuiltinFunction(object, "toLocaleString", toLocaleString, 0, realm)

	return object
}

// 9.2.12
func CoerceOptionsToObject(agent *Agent, options Value) ObjectType {
	if options == UndefinedValue {
		return nil
	}
	return ValueToObject(agent, options)
}

// 9.2.13

// 20.1.2.3.1
func objectDefineProperties(agent *Agent, object ObjectType, properties Value) ObjectType {
	props := ValueToObject(agent, properties)

	keys := props.InternalMethods().OwnPropertyKeys(props)

	for _, key := range keys {
		descValue := props.Get(key)
		desc := ToPropertyDescriptor(agent, descValue)
		if desc != nil {
			object.DefinePropertyOrThrow(key, desc)
		}
	}

	return object
}

// MARK: - FindViaPredicate

type direction int

const (
	DirectionAscending direction = iota
	DirectionDescending
)

type FoundResult struct {
	Index JSInt
	Value Value
}

// 23.1.3.12.1
func (o *Object) FindViaPredicate(
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
		kValue := o.Ref().Get(pk)
		testResult := predicate.Call(thisArg, []Value{kValue, NewNumberValue(k.ToNumber()), o.ToValue()})

		if testResult.ToBoolean() {
			return FoundResult{
				Index: k,
				Value: kValue,
			}
		}
	}

	return FoundResult{
		Index: -1,
		Value: UndefinedValue,
	}
}

func SameObject(o1, o2 ObjectType) bool {
	return o1 == o2
}

// MARK: - Internal

func (o *Object) String() string {
	return fmt.Sprintf("Object[%s]", o.typeName)
}
