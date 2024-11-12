package coldmoon

type Realm struct {
	AgentSignifier interface{}
	Intrinsics     *Intrinsics
	GlobalObject   *Object
	GlobalEnv      *GlobalEnvironment
	TemplateMap    interface{}
	LoadedModules  interface{}
	HostDefined    interface{}
	Agent          *Agent
}

// CreateRealm creates a new realm.
// 9.3.1
func CreateRealm(agent *Agent) *Realm {
	r := &Realm{
		Agent:      agent,
		Intrinsics: &Intrinsics{},
	}
	r.CreateIntrinsics()

	return r
}

// 9.3.2
func (r *Realm) CreateIntrinsics() {
	r.Intrinsics.ObjectPrototype = NewObjectPrototype(r)
	f := NewFunctionPrototype(r)
	r.Intrinsics.FunctionPrototype = f
	InitFunctionMethods(f, r)
	r.Intrinsics.BooleanPrototype = NewBooleanPrototype(r)
	r.Intrinsics.BooleanConstructor = NewBooleanConstructor(r)
	r.Intrinsics.ThrowTypeError = NewThrowTypeError(r)
	r.Intrinsics.IsFinite = NewIsFinite(r)
	r.Intrinsics.IsNaN = NewIsNaN(r)
	r.Intrinsics.Eval = NewEval(r)
	r.Intrinsics.ObjectConstructor = NewObjectConstructor(r)
	r.Intrinsics.FunctionConstructor = NewFunctionConstructor(r)
	r.Intrinsics.ArrayPrototype = NewArrayPrototype(r)
	r.Intrinsics.ArrayConstructor = NewArrayConstructor(r)
	r.Intrinsics.StringPrototype = NewStringPrototype(r)
	r.Intrinsics.StringConstructor = NewStringConstructor(r)
	r.Intrinsics.NumberPrototype = NewNumberPrototype(r)
	r.Intrinsics.NumberConstructor = NewNumberConstructor(r)
	r.Intrinsics.SymbolPrototype = NewSymbolPrototype(r)
	r.Intrinsics.SymbolConstructor = NewSymbolConstructor(r)

	AddRestrictedFunctionProperties(r.Intrinsics.FunctionPrototype, r)
}

// 9.3.3
func (r *Realm) SetRealmGlobalObject(globalObj *Object, thisValue ObjectType) {
	obj := globalObj
	if obj == nil {
		obj = OrdinaryObjectCreate(r.Agent, r.Intrinsics.ObjectPrototype, []string{})
	}

	this := thisValue
	if this == nil {
		this = globalObj
	}

	r.GlobalObject = obj
	r.GlobalEnv = NewGlobalEnvironment(globalObj, this)
}

// 9.3.4
func (r *Realm) SetDefaultGlobalBindings() *Object {
	global := r.GlobalObject

	properties := GlobalObjectProperties(r)
	for _, p := range properties {
		name := NewStringPropertyKey(p.Name)
		desc := p.PropertyDescriptor
		global.DefinePropertyOrThrow(name, desc)
	}

	return global
}

// 9.6
func InitializeHostDefinedRealm(
	agent *Agent,
	globalObject *Object,
) {
	realm := CreateRealm(agent)
	newContext := &ExecutionContext{
		Function:       nil,
		Realm:          realm,
		ScriptOrModule: ScriptOrModuleNull,
	}

	agent.ExecutionContextStack.Push(newContext)

	global := globalObject

	realm.SetRealmGlobalObject(global, nil)

	realm.SetDefaultGlobalBindings()
}
