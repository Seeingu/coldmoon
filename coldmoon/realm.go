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

func (r *Realm) ToReferrer() ImportedModuleReferrer {
	return ImportedModuleReferrer{
		Script: nil,
		Module: nil,
		Realm:  r,
	}
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
	r.Intrinsics.ParseInt = NewParseInt(r)
	r.Intrinsics.ParseFloat = NewParseFloat(r)
	r.Intrinsics.DecodeURI = NewDecodeURI(r)
	r.Intrinsics.DecodeURIComponent = NewDecodeURIComponent(r)
	r.Intrinsics.EncodeURI = NewEncodeURI(r)
	r.Intrinsics.EncodeURIComponent = NewEncodeURIComponent(r)
	r.Intrinsics.ObjectConstructor = NewObjectConstructor(r)
	r.Intrinsics.FunctionConstructor = NewFunctionConstructor(r)
	r.Intrinsics.ArrayPrototype = NewArrayPrototype(r)
	r.Intrinsics.ArrayConstructor = NewArrayConstructor(r)
	r.Intrinsics.ArrayIteratorPrototype = NewArrayIteratorPrototype(r)
	r.Intrinsics.ArrayPrototypeValues = MustGetObject(r.Intrinsics.ArrayPrototype.PropertyStorage().Get(NewStringPropertyKey("values")).Value)
	r.Intrinsics.ArrayBufferPrototype = NewArrayBufferPrototype(r)
	r.Intrinsics.ArrayBufferConstructor = NewArrayBufferConstructor(r)
	r.Intrinsics.StringPrototype = NewStringPrototype(r)
	r.Intrinsics.StringConstructor = NewStringConstructor(r)
	r.Intrinsics.StringIteratorPrototype = NewStringIteratorPrototype(r)
	r.Intrinsics.IteratorPrototype = NewIteratorPrototype(r)
	r.Intrinsics.GeneratorFunctionPrototype = NewGeneratorFunctionPrototype(r)
	r.Intrinsics.GeneratorFunctionConstructor = NewGeneratorFunctionConstructor(r)
	r.Intrinsics.GeneratorFunctionPrototypePrototype = NewGeneratorPrototype(r)
	r.Intrinsics.AsyncGeneratorFunction = NewAsyncGeneratorFunctionConstructor(r)
	r.Intrinsics.AsyncGeneratorFunctionPrototype = NewAsyncGeneratorFunctionPrototype(r)
	r.Intrinsics.AsyncGeneratorFunctionPrototypePrototype = NewAsyncGeneratorPrototype(r)
	r.Intrinsics.AsyncIteratorPrototype = NewAsyncIteratorPrototype(r)
	r.Intrinsics.AsyncFunctionPrototype = NewAsyncFunctionPrototype(r)
	r.Intrinsics.AsyncFunctionConstructor = NewAsyncFunctionConstructor(r)
	r.Intrinsics.DataViewPrototype = NewDataViewPrototype(r)
	r.Intrinsics.DataViewConstructor = NewDataViewConstructor(r)
	r.Intrinsics.PromisePrototype = NewPromisePrototype(r)
	r.Intrinsics.Promise = NewPromiseConstructor(r)
	r.Intrinsics.DatePrototype = NewDatePrototype(r)
	r.Intrinsics.DateConstructor = NewDateConstructor(r)
	r.Intrinsics.MapPrototype = NewMapPrototype(r)
	r.Intrinsics.Map = NewMapConstructor(r)
	r.Intrinsics.MapIteratorPrototype = NewMapIteratorPrototype(r)
	r.Intrinsics.SetPrototype = NewSetPrototype(r)
	r.Intrinsics.Set = NewSetConstructor(r)
	r.Intrinsics.SetIteratorPrototype = NewSetIteratorPrototype(r)
	r.Intrinsics.NumberPrototype = NewNumberPrototype(r)
	r.Intrinsics.NumberConstructor = NewNumberConstructor(r)
	r.Intrinsics.Math = NewMathObject(r)
	r.Intrinsics.SymbolPrototype = NewSymbolPrototype(r)
	r.Intrinsics.SymbolConstructor = NewSymbolConstructor(r)
	r.Intrinsics.BigIntPrototype = NewBigIntPrototype(r)
	r.Intrinsics.BigIntConstructor = NewBigIntConstructor(r)
	r.Intrinsics.TypedArrayPrototype = NewTypedArrayPrototype(r)
	r.Intrinsics.TypedArrayConstructor = NewTypedArrayConstructor(r)
	r.Intrinsics.SharedArrayBufferPrototype = NewSharedArrayBufferPrototype(r)
	r.Intrinsics.SharedArrayBufferConstructor = NewSharedArrayBufferConstructor(r)
	r.Intrinsics.BigInt64ArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameBigInt64)
	r.Intrinsics.BigInt64ArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameBigInt64)
	r.Intrinsics.BigUint64ArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameBigUint64)
	r.Intrinsics.BigUint64ArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameBigUint64)
	r.Intrinsics.Float32ArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameFloat32)
	r.Intrinsics.Float32ArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameFloat32)
	r.Intrinsics.Float64ArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameFloat64)
	r.Intrinsics.Float64ArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameFloat64)
	r.Intrinsics.Int8ArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameInt8)
	r.Intrinsics.Int8ArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameInt8)
	r.Intrinsics.Uint8ArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameUint8)
	r.Intrinsics.Uint8ArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameUint8)
	r.Intrinsics.Uint8ClampedArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameUint8Clamped)
	r.Intrinsics.Uint8ClampedArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameUint8Clamped)
	r.Intrinsics.Int16ArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameInt16)
	r.Intrinsics.Int16ArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameInt16)
	r.Intrinsics.Uint16ArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameUint16)
	r.Intrinsics.Uint16ArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameUint16)
	r.Intrinsics.Int32ArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameInt32)
	r.Intrinsics.Int32ArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameInt32)
	r.Intrinsics.Uint32ArrayPrototype = NewTypedArrayNamePrototype(r, TypedArrayNameUint32)
	r.Intrinsics.Uint32ArrayConstructor = NewTypedArrayNameConstructor(r, TypedArrayNameUint32)
	r.Intrinsics.Atomics = NewAtomics(r)
	r.Intrinsics.JSON = NewJSON(r)
	r.Intrinsics.Reflect = NewReflectObject(r)
	r.Intrinsics.Proxy = NewProxyConstructor(r)
	r.Intrinsics.Intl = NewIntlObject(r)
	r.Intrinsics.RegExpPrototype = NewRegExpPrototype(r)
	r.Intrinsics.RegExpConstructor = NewRegExpConstructor(r)
	r.Intrinsics.RegExpStringIteratorPrototype = NewRegExpStringIteratorPrototype(r)
	r.Intrinsics.ForInIteratorPrototype = NewForInIteratorPrototype(r)
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
		ScriptOrModule: nil,
	}

	agent.ExecutionContextStack.Push(newContext)

	global := globalObject
	if global == nil {
		global = OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, []string{})
	}

	realm.SetRealmGlobalObject(global, nil)

	realm.SetDefaultGlobalBindings()
}
