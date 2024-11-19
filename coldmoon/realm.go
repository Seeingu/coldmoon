package coldmoon

import (
	"math/rand"
)

type Realm struct {
	AgentSignifier interface{}
	Intrinsics     *Intrinsics
	GlobalObject   *Object
	GlobalEnv      *GlobalEnvironment
	TemplateMap    interface{}
	LoadedModules  interface{}
	HostDefined    interface{}
	Agent          *Agent
	Rng            rand.Rand
}

// CreateRealm creates a new realm.
// 9.3.1
func CreateRealm(agent *Agent) *Realm {
	r := &Realm{
		Agent:      agent,
		Intrinsics: &Intrinsics{},
	}
	r.CreateIntrinsics()
	AddRestrictedFunctionProperties(r.Intrinsics.FunctionPrototype, r)

	return r
}

// 9.3.2
func (r *Realm) CreateIntrinsics() {
	r.Intrinsics.ObjectPrototype = NewObjectPrototypeSkeleton(r)
	NewFunctionPrototypeWithIntrinsicsBinding(r)
	r.Intrinsics.ObjectPrototype = NewObjectPrototypeWithObject(r, r.Intrinsics.ObjectPrototype)
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
	r.Intrinsics.ArrayIteratorPrototype = NewArrayIteratorPrototype(r)
	r.Intrinsics.StringPrototype = NewStringPrototype(r)
	r.Intrinsics.StringConstructor = NewStringConstructor(r)
	r.Intrinsics.StringIteratorPrototype = NewStringIteratorPrototype(r)
	r.Intrinsics.DatePrototype = NewDatePrototype(r)
	r.Intrinsics.DateConstructor = NewDateConstructor(r)
	r.Intrinsics.MapPrototype = NewMapPrototype(r)
	r.Intrinsics.Map = NewMapConstructor(r)
	r.Intrinsics.NumberPrototype = NewNumberPrototype(r)
	r.Intrinsics.NumberConstructor = NewNumberConstructor(r)
	r.Intrinsics.Math = NewMathObject(r)
	r.Intrinsics.SymbolPrototype = NewSymbolPrototype(r)
	r.Intrinsics.SymbolConstructor = NewSymbolConstructor(r)
	r.Intrinsics.BigIntPrototype = NewBigIntPrototype(r)
	r.Intrinsics.BigIntConstructor = NewBigIntConstructor(r)
	r.Intrinsics.Reflect = NewReflectObject(r)
	r.Intrinsics.Proxy = NewProxyConstructor(r)
	r.Intrinsics.IteratorPrototype = NewIteratorPrototype(r)
	r.Intrinsics.ErrorPrototype = NewErrorPrototype(r)
	r.Intrinsics.ErrorConstructor = NewErrorConstructor(r)
	r.Intrinsics.SyntaxErrorPrototype = NewNativeErrorPrototype(r, "SyntaxError")
	r.Intrinsics.SyntaxErrorConstructor = NewNativeErrorConstructor(r, "SyntaxError")
	r.Intrinsics.TypeErrorPrototype = NewNativeErrorPrototype(r, "TypeError")
	r.Intrinsics.TypeErrorConstructor = NewNativeErrorConstructor(r, "TypeError")
	r.Intrinsics.RangeErrorPrototype = NewNativeErrorPrototype(r, "RangeError")
	r.Intrinsics.RangeErrorConstructor = NewNativeErrorConstructor(r, "RangeError")
	r.Intrinsics.ReferenceErrorPrototype = NewNativeErrorPrototype(r, "ReferenceError")
	r.Intrinsics.ReferenceErrorConstructor = NewNativeErrorConstructor(r, "ReferenceError")
	r.Intrinsics.URIErrorPrototype = NewNativeErrorPrototype(r, "URIError")
	r.Intrinsics.URIErrorConstructor = NewNativeErrorConstructor(r, "URIError")
	r.Intrinsics.EvalErrorPrototype = NewNativeErrorPrototype(r, "EvalError")
	r.Intrinsics.EvalErrorConstructor = NewNativeErrorConstructor(r, "EvalError")
	r.Intrinsics.AggregateErrorPrototype = NewAggregateErrorPrototype(r)
	r.Intrinsics.AggregateErrorConstructor = NewAggregateErrorConstructor(r)

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
	r.GlobalEnv = NewGlobalEnvironment(obj, this)
	Assert(r.GlobalEnv != nil)
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
