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
	r.Intrinsics.ObjectPrototype = NewObjectPrototype(r.Agent)
	r.Intrinsics.FunctionPrototype = NewFunctionPrototype(r)
	r.Intrinsics.BooleanPrototype = NewBooleanPrototype(r)
	r.Intrinsics.BooleanConstructor = NewBooleanConstructor(r)
	r.Intrinsics.ThrowTypeError = NewThrowTypeError(r)

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

	properties := []struct {
		key   string
		value Value
	}{
		{
			"Boolean", NewValueFromObject(r.Intrinsics.BooleanConstructor),
		},
	}
	for _, p := range properties {
		name := NewStringPropertyKey(p.key)
		value := p.value
		descriptor := &PropertyDescriptor{
			Value:        value,
			Writable:     true,
			Enumerable:   false,
			Configurable: true,
		}
		global.DefinePropertyOrThrow(name, descriptor)
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
