package coldmoon

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

var nextRealmRandomSeed = time.Now().UnixNano()

type realmState uint8

const (
	realmStateBuilding realmState = iota
	realmStateReady
)

type Realm struct {
	AgentSignifier any
	Intrinsics     *Intrinsics
	GlobalObject   *Object
	GlobalEnv      *GlobalEnvironment
	TemplateMap    map[*TemplateLiteral]ObjectType
	HostDefined    any
	Agent          *Agent
	Rng            rand.Rand
	state          realmState
}

// IsReady reports whether intrinsic construction and validation completed.
func (r *Realm) IsReady() bool {
	return r != nil && r.state == realmStateReady
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
		// Seed each Realm independently while retaining the public rand.Rand
		// value seam that embedders can replace for deterministic execution.
		Rng:   *rand.New(rand.NewSource(atomic.AddInt64(&nextRealmRandomSeed, 1))),
		state: realmStateBuilding,
	}
	r.createIntrinsics()
	if missing := missingIntrinsicFields(r.Intrinsics); len(missing) > 0 {
		panic(fmt.Sprintf("incomplete intrinsic bootstrap: %s", strings.Join(missing, ", ")))
	}
	AddRestrictedFunctionProperties(r.Intrinsics.FunctionPrototype, r)
	r.state = realmStateReady

	return r
}

// 9.3.2
func (r *Realm) createIntrinsics() {
	Assert(r.state == realmStateBuilding)
	r.createFoundationIntrinsics()
	r.createIterationAndCallableIntrinsics()
	r.createStandardLibraryIntrinsics()
	r.createTypedArrayIntrinsics()
	r.createUtilityIntrinsics()
	r.createErrorIntrinsics()
}

type intrinsicDependency struct {
	name  string
	value ObjectType
}

func (r *Realm) requireIntrinsicDependencies(phase string, dependencies ...intrinsicDependency) {
	for _, dependency := range dependencies {
		if dependency.value == nil {
			panic(fmt.Sprintf("intrinsic bootstrap phase %q requires %s", phase, dependency.name))
		}
	}
}

// createFoundationIntrinsics builds the Object/Function cycle first, then the
// primitive constructors and global functions that every later phase uses.
func (r *Realm) createFoundationIntrinsics() {
	r.Intrinsics.ObjectPrototype = NewObjectPrototypeSkeleton(r)
	r.requireIntrinsicDependencies("function prototype", intrinsicDependency{"ObjectPrototype", r.Intrinsics.ObjectPrototype})
	NewFunctionPrototypeWithIntrinsicsBinding(r)
	r.requireIntrinsicDependencies("object prototype", intrinsicDependency{"FunctionPrototype", r.Intrinsics.FunctionPrototype})
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
}

// createIterationAndCallableIntrinsics establishes the iterator family before
// generator and async prototypes, and ArrayBuffer before DataView/typed arrays.
func (r *Realm) createIterationAndCallableIntrinsics() {
	r.requireIntrinsicDependencies(
		"iteration and callable",
		intrinsicDependency{"ObjectPrototype", r.Intrinsics.ObjectPrototype},
		intrinsicDependency{"FunctionPrototype", r.Intrinsics.FunctionPrototype},
		intrinsicDependency{"ObjectConstructor", r.Intrinsics.ObjectConstructor},
		intrinsicDependency{"FunctionConstructor", r.Intrinsics.FunctionConstructor},
	)
	r.Intrinsics.ArrayPrototype = NewArrayPrototype(r)
	r.Intrinsics.ArrayConstructor = NewArrayConstructor(r)
	r.Intrinsics.IteratorPrototype = NewIteratorPrototype(r)
	r.Intrinsics.ArrayIteratorPrototype = NewArrayIteratorPrototype(r)
	r.Intrinsics.ArrayPrototypeValues = MustGetObject(r.Intrinsics.ArrayPrototype.propertyStorage().Get(NewStringPropertyKey("values")).Value)
	r.Intrinsics.ArrayBufferPrototype = NewArrayBufferPrototype(r)
	r.Intrinsics.ArrayBufferConstructor = NewArrayBufferConstructor(r)
	r.Intrinsics.StringPrototype = NewStringPrototype(r)
	r.Intrinsics.StringConstructor = NewStringConstructor(r)
	r.Intrinsics.StringIteratorPrototype = NewStringIteratorPrototype(r)
	r.Intrinsics.GeneratorFunctionPrototype = NewGeneratorFunctionPrototype(r)
	r.Intrinsics.GeneratorFunctionConstructor = NewGeneratorFunctionConstructor(r)
	r.Intrinsics.GeneratorFunctionPrototypePrototype = NewGeneratorPrototype(r)
	r.Intrinsics.AsyncIteratorPrototype = NewAsyncIteratorPrototype(r)
	r.Intrinsics.AsyncGeneratorFunction = NewAsyncGeneratorFunctionConstructor(r)
	r.Intrinsics.AsyncGeneratorFunctionPrototype = NewAsyncGeneratorFunctionPrototype(r)
	r.Intrinsics.AsyncGeneratorFunctionPrototypePrototype = NewAsyncGeneratorPrototype(r)
	r.Intrinsics.AsyncFunctionPrototype = NewAsyncFunctionPrototype(r)
	r.Intrinsics.AsyncFunctionConstructor = NewAsyncFunctionConstructor(r)
}

// createStandardLibraryIntrinsics builds consumers of the iterator, callable,
// string, and buffer foundations without depending on typed-array constructors.
func (r *Realm) createStandardLibraryIntrinsics() {
	r.requireIntrinsicDependencies(
		"standard library",
		intrinsicDependency{"ArrayConstructor", r.Intrinsics.ArrayConstructor},
		intrinsicDependency{"ArrayBufferConstructor", r.Intrinsics.ArrayBufferConstructor},
		intrinsicDependency{"StringConstructor", r.Intrinsics.StringConstructor},
		intrinsicDependency{"IteratorPrototype", r.Intrinsics.IteratorPrototype},
		intrinsicDependency{"AsyncIteratorPrototype", r.Intrinsics.AsyncIteratorPrototype},
		intrinsicDependency{"AsyncFunctionConstructor", r.Intrinsics.AsyncFunctionConstructor},
	)
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
}

// createTypedArrayIntrinsics constructs the shared TypedArray base before each
// concrete numeric view; Atomics is intentionally deferred until all exist.
func (r *Realm) createTypedArrayIntrinsics() {
	r.requireIntrinsicDependencies(
		"typed arrays",
		intrinsicDependency{"ArrayBufferConstructor", r.Intrinsics.ArrayBufferConstructor},
		intrinsicDependency{"IteratorPrototype", r.Intrinsics.IteratorPrototype},
		intrinsicDependency{"NumberConstructor", r.Intrinsics.NumberConstructor},
		intrinsicDependency{"BigIntConstructor", r.Intrinsics.BigIntConstructor},
	)
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
}

// createUtilityIntrinsics installs namespace-style objects and reflection only
// after their collection, buffer, and typed-array dependencies are complete.
func (r *Realm) createUtilityIntrinsics() {
	r.requireIntrinsicDependencies(
		"utility objects",
		intrinsicDependency{"TypedArrayConstructor", r.Intrinsics.TypedArrayConstructor},
		intrinsicDependency{"SharedArrayBufferConstructor", r.Intrinsics.SharedArrayBufferConstructor},
		intrinsicDependency{"Uint32ArrayConstructor", r.Intrinsics.Uint32ArrayConstructor},
	)
	r.Intrinsics.Atomics = NewAtomics(r)
	r.Intrinsics.JSON = NewJSON(r)
	r.Intrinsics.Reflect = NewReflectObject(r)
	r.Intrinsics.Proxy = NewProxyConstructor(r)
	r.Intrinsics.Intl = NewIntlObject(r)
	r.Intrinsics.RegExpPrototype = NewRegExpPrototype(r)
	r.Intrinsics.RegExpConstructor = NewRegExpConstructor(r)
	r.Intrinsics.RegExpStringIteratorPrototype = NewRegExpStringIteratorPrototype(r)
	r.Intrinsics.ForInIteratorPrototype = NewForInIteratorPrototype(r)
}

// createErrorIntrinsics is last so every constructor reachable while creating
// an error object has already been published into the draft Intrinsics table.
func (r *Realm) createErrorIntrinsics() {
	r.requireIntrinsicDependencies(
		"errors",
		intrinsicDependency{"ObjectPrototype", r.Intrinsics.ObjectPrototype},
		intrinsicDependency{"FunctionPrototype", r.Intrinsics.FunctionPrototype},
		intrinsicDependency{"RegExpConstructor", r.Intrinsics.RegExpConstructor},
	)
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

func missingIntrinsicFields(intrinsics *Intrinsics) []string {
	if intrinsics == nil {
		return []string{"Intrinsics"}
	}
	value := reflect.ValueOf(intrinsics).Elem()
	kind := value.Type()
	var missing []string
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		switch field.Kind() {
		case reflect.Interface, reflect.Pointer, reflect.Map, reflect.Slice, reflect.Func:
			if field.IsNil() {
				missing = append(missing, kind.Field(index).Name)
			}
		}
	}
	return missing
}

// 9.3.3
func (r *Realm) SetRealmGlobalObject(globalObj *Object, thisValue ObjectType) {
	Assert(r.IsReady())
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
	Assert(r.IsReady())
	Assert(r.GlobalObject != nil)
	Assert(r.GlobalEnv != nil)
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
		ch:             make(chan struct{}),
		ScriptOrModule: nil,
	}

	agent.resumeExecutionContext(newContext)

	global := globalObject
	if global == nil {
		global = OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, []string{})
	}

	realm.SetRealmGlobalObject(global, nil)

	realm.SetDefaultGlobalBindings()
}
