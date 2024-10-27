package coldmoon

type Data struct {
	prototype       *Object
	extensible      bool
	agent           *Agent
	internalMethods InternalMethods
	propertyStorage PropertyStorage
}

type Object struct {
	ObjectType
	data *Data
}

func NewObject(agent *Agent, prototype *Object) *Object {
	o := &Object{
		data: &Data{
			agent:           agent,
			prototype:       prototype,
			internalMethods: NewInternalMethods(),
			propertyStorage: NewPropertyStorage(),
		},
	}
	return o
}

var EmptyObject = &Object{}

func (o *Object) Prototype() *Object {
	return o.data.prototype
}
func (o *Object) SetPrototype(p *Object) {
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
		if isCallable(method) {
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
func (o *Object) Set(key PropertyKey, value Value, throw bool) {
	success := o.InternalMethods().Set(o, key, value, NewValueFromObject(o))
	if !success && throw {
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

type constructArgs struct {
	newTarget     *Object
	argumentLists []Value
}

// 7.3.14
func (o *Object) Construct(args constructArgs) Value {
	newTarget := args.newTarget
	if newTarget == nil {
		newTarget = o
	}
	return o.InternalMethods().Construct(o, args.argumentLists, newTarget)
}

// 7.3.24
func (o *Object) GetFunctionRealm() *Realm {
	return o.Agent().CurrentRealm()
}
