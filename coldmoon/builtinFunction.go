package coldmoon

type BehaviorFn func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value
type BuiltinFunction struct {
	*Object
	Realm            *Realm
	InitialName      string
	Behavior         BehaviorFn
	RevocableProxy   ObjectType
	AdditionalFields *AdditionalFields
}

func (b *BuiltinFunction) ToObject() *Object {
	return b.Object
}

func (b *BuiltinFunction) ToValue() Value {
	return NewValueFromObject(b)
}

// 7.3.24
func (b *BuiltinFunction) GetFunctionRealm() *Realm {
	return b.Realm
}

// 10.3.1
func BuiltinCall(o ObjectType, thisArgument Value, argumentsList []Value) Value {
	return o.(*BuiltinFunction).BuiltinCallOrConstruct(thisArgument, argumentsList, nil)
}

// 10.3.2
func BuiltinConstruct(b ObjectType, argumentsList []Value, newTarget ObjectType) ObjectType {
	r := b.(*BuiltinFunction).BuiltinCallOrConstruct(NullValue, argumentsList, newTarget)
	return r.(*ObjectValue).Object
}

// 10.3.3
func (b *BuiltinFunction) BuiltinCallOrConstruct(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
	a := b.Agent()
	callerContext := a.runningExecutionContext()
	_ = callerContext

	calleeContext := &ExecutionContext{
		Function:       b.Object,
		Realm:          b.Realm,
		ScriptOrModule: ScriptOrModuleNull,
	}

	a.ExecutionContextStack.Push(calleeContext)

	result := b.Behavior(thisArgument, argumentsList, newTarget)

	a.ExecutionContextStack.Pop()
	return result
}

type builtinFunctionArgs struct {
	realm            *Realm
	prototype        ObjectType
	prefix           string
	isConstructor    bool
	revocableProxy   ObjectType
	additionalFields *AdditionalFields
}

// 10.3.4
func CreateBuiltinFunction(
	agent *Agent,
	behavior BehaviorFn,
	length float64,
	name string,
	args builtinFunctionArgs,
) ObjectType {
	realm := args.realm
	if realm == nil {
		realm = agent.CurrentRealm()
	}

	prototype := args.prototype
	if prototype == nil {
		prototype = realm.Intrinsics.FunctionPrototype
	}
	if prototype == nil {
		panic("FunctionPrototype is nil")
	}

	object := NewObject(agent, prototype.ToObject())
	object.SetExtensible(true)
	function := &BuiltinFunction{
		Object:           object,
		Realm:            realm,
		Behavior:         behavior,
		InitialName:      "",
		RevocableProxy:   args.revocableProxy,
		AdditionalFields: args.additionalFields,
	}
	function.InternalMethods().Call = BuiltinCall
	if args.isConstructor {
		function.InternalMethods().Construct = BuiltinConstruct
	}

	SetFunctionLength(function.Object, length)
	SetFunctionName(function, NewStringPropertyKey(name), args.prefix)

	return function
}
