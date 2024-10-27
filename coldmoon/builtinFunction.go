package coldmoon

type BehaviorFn func(thisArgument Value, argumentsList []Value, newTarget *Object) Value
type BuiltinFunction struct {
	*Object
	Realm       *Realm
	InitialName string
	Behavior    BehaviorFn
}

// 10.3.1
func BuiltinCall(o ObjectType, thisArgument Value, argumentsList []Value) Value {
	return o.(*BuiltinFunction).BuiltinCallOrConstruct(thisArgument, argumentsList, nil)
}

// 10.3.2
func BuiltinConstruct(b ObjectType, argumentsList []Value, newTarget *Object) Value {
	return b.(*BuiltinFunction).BuiltinCallOrConstruct(&NullValue, argumentsList, newTarget)
}

// 10.3.3
func (b *BuiltinFunction) BuiltinCallOrConstruct(thisArgument Value, argumentsList []Value, newTarget *Object) Value {
	a := b.Agent()
	callerContext := a.runningExecutionContext()
	_ = callerContext

	calleeContext := &ExecutionContext{
		Function:       b.Object,
		Realm:          b.Realm,
		ScriptOrModule: ScriptOrModuleNull,
	}

	a.executionContextStack = append(a.executionContextStack, calleeContext)

	result := b.Behavior(thisArgument, argumentsList, newTarget)

	a.executionContextStack = a.executionContextStack[:len(a.executionContextStack)-1]
	return result
}

type builtinFunctionArgs struct {
	realm     *Realm
	prototype *Object
	prefix    string
}

// 10.3.4
func CreateBuiltinFunction(
	agent *Agent,
	behavior BehaviorFn,
	length float64,
	name string,
	args builtinFunctionArgs,
) *BuiltinFunction {
	realm := args.realm
	if realm == nil {
		realm = agent.CurrentRealm()
	}

	prototype := args.prototype
	if prototype == nil {
		//prototype = realm.Intrinsics
		panic("unimplemented")
	}

	object := NewObject(agent, prototype)
	object.SetExtensible(true)
	function := &BuiltinFunction{
		Object:      object,
		Realm:       realm,
		Behavior:    behavior,
		InitialName: "",
	}
	function.InternalMethods().Call = BuiltinCall
	function.InternalMethods().Construct = BuiltinConstruct

	SetFunctionLength(function.Object, length)
	SetFunctionName(function.Object, NewStringPropertyKey(name), args.prefix)

	return function
}
