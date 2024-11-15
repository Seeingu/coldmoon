package coldmoon

import "github.com/samber/lo"

// 9.1.1.4
type GlobalEnvironment struct {
	EnvironmentRecord
	ObjectRecord      *ObjectEnvironment
	GlobalThisValue   ObjectType
	DeclarativeRecord *DeclarativeEnvironment
	VarNames          []string
	outerEnv          EnvironmentRecord
}

var _ EnvironmentRecord = (*GlobalEnvironment)(nil)

func (g *GlobalEnvironment) CreateGlobalVarBinding(name string, deletable bool) {
	objRec := g.ObjectRecord
	globalObject := objRec.BindingObject
	hasProperty := ObjectHasOwnProperty(globalObject, NewStringPropertyKey(name))
	extensible := globalObject.IsExtensible()

	if !hasProperty && extensible {
		objRec.CreateMutableBinding(name, deletable)
		objRec.InitializeBinding(name, UndefinedValue)
	}

	if !lo.Contains(g.VarNames, name) {
		g.VarNames = append(g.VarNames, name)
	}
}

func (g *GlobalEnvironment) InitializeBinding(name string, value Value) {
	if g.DeclarativeRecord.HasBinding(name) {
		g.DeclarativeRecord.InitializeBinding(name, value)
		return
	}
	g.ObjectRecord.InitializeBinding(name, value)
}

func (g *GlobalEnvironment) CreateMutableBinding(name string, deletable bool) {
	if g.DeclarativeRecord.HasBinding(name) {
		panic("TypeError: Binding already exists")
	}
	g.DeclarativeRecord.CreateMutableBinding(name, deletable)
}

func (g *GlobalEnvironment) CreateImmutableBinding(name string, deletable bool) {
	if g.DeclarativeRecord.HasBinding(name) {
		panic("TypeError: Binding already exists")
	}

	g.DeclarativeRecord.CreateImmutableBinding(name, deletable)
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

// 9.1.1.4.1
func (g *GlobalEnvironment) HasBinding(name string) bool {
	DclRec := g.DeclarativeRecord
	if DclRec.HasBinding(name) {
		return true
	}
	ObjRec := g.ObjectRecord
	return ObjRec.HasBinding(name)
}

func (g *GlobalEnvironment) SetMutableBinding(name string, value Value, strict bool) {
	DclRec := g.DeclarativeRecord
	if DclRec.HasBinding(name) {
		DclRec.SetMutableBinding(name, value, strict)
		return
	}
	ObjRec := g.ObjectRecord
	ObjRec.SetMutableBinding(name, value, strict)
}

// 9.1.1.4.6
func (g *GlobalEnvironment) GetBindingValue(name string, strict bool) Value {
	DclRec := g.DeclarativeRecord
	if DclRec.HasBinding(name) {
		return DclRec.GetBindingValue(name, strict)
	}
	ObjRec := g.ObjectRecord
	return ObjRec.GetBindingValue(name, strict)
}

// 9.1.1.4.8
func (g *GlobalEnvironment) HasThisBinding() bool {
	return true
}

// 9.1.1.4.10
func (g *GlobalEnvironment) WithBaseObject() ObjectType {
	return nil
}

func (g *GlobalEnvironment) GetThisBinding() Value {
	return NewValueFromObject(g.GlobalThisValue)
}

func (g *GlobalEnvironment) OuterEnv() EnvironmentRecord {
	return g.outerEnv
}
