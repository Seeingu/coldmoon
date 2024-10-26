package coldmoon

type (
	Environment struct{}
	Intrinsics  struct{}
)

type Realm struct {
	AgentSignifier interface{}
	Instringics    Intrinsics
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

func (r *Realm) CreateIntrinsics() {
	r.Instringics = Intrinsics{}
}
