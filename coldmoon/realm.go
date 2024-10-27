package coldmoon

type (
	Environment struct{}
	Intrinsics  struct{}
)

func (i *Intrinsics) Get(key string) interface{} {
	return nil
}

type Realm struct {
	AgentSignifier interface{}
	Intrinsics     Intrinsics
	GlobalObject   Object
	GlobalEnv      Environment
	TemplateMap    interface{}
	LoadedModules  interface{}
	HostDefined    interface{}
}

// CreateRealm creates a new realm.
// 9.3.1
func CreateRealm() *Realm {
	r := &Realm{}
	r.CreateIntrinsics()

	return r
}

// 9.3.2
func (r *Realm) CreateIntrinsics() {
	r.Intrinsics = Intrinsics{}
}

// 9.6
func InitializeHostDefinedRealm(agent *Agent) {
	realm := CreateRealm()
	newContext := &ExecutionContext{
		Function:       nil,
		Realm:          realm,
		ScriptOrModule: ScriptOrModuleNull,
	}

	agent.executionContextStack = append(agent.executionContextStack, newContext)
}
