package coldmoon

// 9.1.1.4
type GlobalEnvironment struct {
	EnvironmentRecord
	ObjectRecord      *ObjectEnvironment
	GlobalThisValue   ObjectType
	DeclarativeRecord *DeclarativeEnvironment
	VarNames          []string
	outerEnv          EnvironmentRecord
}

// 9.1.2.5
func NewGlobalEnvironment(globalObj *Object, thisValue ObjectType) *GlobalEnvironment {
	objRec := NewObjectEnvironment(globalObj, false, nil)
	dclRec := NewDeclarativeEnvironment(nil)
	globalEnv := &GlobalEnvironment{
		ObjectRecord:      objRec,
		GlobalThisValue:   thisValue,
		DeclarativeRecord: dclRec,
		VarNames:          []string{},
		outerEnv:          nil,
	}

	return globalEnv
}

// 9.1.1.4.8
func (o *GlobalEnvironment) HasThisBinding() bool {
	return true
}

func (o *GlobalEnvironment) GetThisBinding() ObjectType {
	return o.GlobalThisValue
}

func (o *GlobalEnvironment) OuterEnv() EnvironmentRecord {
	return o.outerEnv
}
