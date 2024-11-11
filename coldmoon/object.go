package coldmoon

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
}

type Object struct {
	ObjectType
	data *Data
}

func (o *Object) ToObject() *Object {
	return o
}

func NewObject(agent *Agent, prototype ObjectType) *Object {
	o := &Object{
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
			result := CallAssumeCallableNoArgs(method, NewValueFromObject(o))
			if _, isObject := result.(*ObjectValue); !isObject {
				return result
			}
		}
	}

	message := "Could not convert object to primitive"
	o.Agent().ThrowException(TypeError, message)
	panic("")
}

// 7.2.5
func (o *Object) IsExtensible() bool {
	return o.InternalMethods().IsExtensible(o)
}

// 7.3.2
func (o *Object) Get(key PropertyKey) Value {
	return o.InternalMethods().Get(o, key, NewValueFromObject(o))
}

// 7.3.4
func (o *Object) Set(key PropertyKey, value Value, throw setThrowType) {
	success := o.InternalMethods().Set(o, key, value, NewValueFromObject(o))
	if !success && throw == setThrowTypeThrow {
		o.Agent().ThrowException(TypeError, "Set failed")
	}
}

// 7.3.5
func (o *Object) CreateDataProperty(key PropertyKey, value Value) bool {
	newDesc := o.InternalMethods().DefineOwnProperty(o, key, &PropertyDescriptor{
		Value:        value,
		Writable:     true,
		Enumerable:   true,
		Configurable: true,
	})
	return newDesc
}

// 7.3.6
func (o *Object) CreateDataPropertyOrThrow(key PropertyKey, value Value) bool {
	success := o.CreateDataProperty(key, value)
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
	success := o.InternalMethods().DefineOwnProperty(o, key, desc)
	if !success {
		o.Agent().ThrowException(TypeError, "DefinePropertyOrThrow failed")
	}
	return success
}

// 7.3.9
func (o *Object) DeletePropertyOrThrow(key PropertyKey) bool {
	success := o.InternalMethods().Delete(o, key)
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

// 7.3.19
func (o *Object) LengthOfArrayLike() uint64 {
	return ToLength(o.Agent(), o.Get(NewStringPropertyKey("length")))
}

// 7.3.24
func (o *Object) GetFunctionRealm() *Realm {
	return o.Agent().CurrentRealm()
}

// MARK: - Object Constructor

// 20.1.1
type ObjectConstructor struct {
	*Object
}

func NewObjectConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		value := argumentsList[0]
		if newTarget != nil && newTarget != agent.ActiveFunctionObject() {
			return NewValueFromObject(OrdinaryCreateFromConstructor(
				agent, newTarget, "%Object.prototype%", []string{},
			))
		}

		if value == nil {
			return NewValueFromObject(
				OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, []string{}),
			)
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
	DefineBuiltinFunction(object, "freeze", freeze, 1, realm)

	// 20.1.2.15
	var ObjectIs BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		arg1 := args[0]
		arg2 := args[1]
		return NewBooleanValue(SameValue(arg1, arg2))
	}
	DefineBuiltinFunction(object, "is", ObjectIs, 2, realm)

	// 20.1.2.16
	var isExtensible BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj, ok := objectValue.(*ObjectValue)
		if !ok {
			return NewBooleanValue(false)
		}
		return NewBooleanValue(obj.Object.IsExtensible())
	}
	DefineBuiltinFunction(object, "isExtensible", isExtensible, 1, realm)

	// 20.1.2.17
	var isFrozen BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj, ok := objectValue.(*ObjectValue)
		if !ok {
			return NewBooleanValue(true)
		}
		return NewBooleanValue(TestIntegrityLevel(obj.Object, IntegrityLevelFrozen))
	}
	DefineBuiltinFunction(object, "isFrozen", isFrozen, 1, realm)

	// 20.1.2.18
	var isSealed BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		objectValue := args[0]
		obj, ok := objectValue.(*ObjectValue)
		if !ok {
			return NewBooleanValue(true)
		}
		return NewBooleanValue(TestIntegrityLevel(obj.Object, IntegrityLevelSealed))
	}
	DefineBuiltinFunction(object, "isSealed", isSealed, 1, realm)

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
	DefineBuiltinFunction(object, "preventExtensions", preventExtensions, 1, realm)

	// 20.1.2.21
	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.ObjectPrototype),
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
	DefineBuiltinFunction(object, "seal", seal, 1, realm)

	// 20.1.3.1
	DefineBuiltinProperty(realm.Intrinsics.ObjectPrototype, "constructor", NewValueFromObject(object))

	// 20.1.3.6 toString
	var toString = func(this Value, args []Value, newTarget ObjectType) Value {
		if this == UndefinedValue {
			return NewStringValue("[object Undefined]")
		}
		if this == NullValue {
			return NewStringValue("[object Null]")
		}
		o := ValueToObject(agent, this)
		_isArray := isArray(this)
		var builtInTag string
		if _isArray {
			builtInTag = "Array"
		} else if o.InternalMethods().Call != nil {
			builtInTag = "Function"
		} else if _, ok := o.(*BooleanObject); ok {
			builtInTag = "Boolean"
		} else {
			builtInTag = "Object"
		}

		symbol := WellKnownSymbols[WellKnownSymbolsToStringTag]
		tagValue := o.Get(NewSymbolPropertyKey(&symbol))

		var tag string
		if stringTag, ok := tagValue.(*StringValue); ok {
			tag = stringTag.Data
		} else {
			tag = builtInTag
		}
		return NewStringValue("[object " + tag + "]")

	}
	DefineBuiltinFunction(object, "toString", toString, 0, realm)

	var valueOf = func(this Value, args []Value, newTarget ObjectType) Value {
		return NewValueFromObject(ValueToObject(agent, this))
	}
	DefineBuiltinFunction(object, "valueOf", valueOf, 0, realm)

	return object

}
