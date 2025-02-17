package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type GeneratorState int

const (
	GeneratorStateSuspendedStart GeneratorState = iota
	GeneratorStateSuspendedYield
	GeneratorStateExecuting
	GeneratorStateCompleted
)

type GeneratorObject struct {
	*Object
	// [[GeneratorState]]
	GeneratorState GeneratorState
	// [[GeneratorContext]]
	GeneratorContext *ExecutionContext
	// [[GeneratorBrand]]
	GeneratorBrand string
	closure        func(bytecode *BytecodeContext) ObjectType
	bytecode       *BytecodeContext
	result         Value
}

func NewGeneratorPrototype(realm *Realm) *GeneratorObject {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.IteratorPrototype, "GeneratorPrototype")
	g := &GeneratorObject{
		Object: object,
	}
	g.ref = g

	var next BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		value := pkg.SliceSafeGet(argumentsList, 0)
		return GeneratorResume(agent, this, value).ToValue()
	}
	var iteratorReturn BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		value := pkg.SliceSafeGet(argumentsList, 0)
		generator := this
		C := NewCompletionReturnValue(value)
		return GeneratorResumeAbrupt(agent, generator, C).ToValue()
	}

	DefineBuiltinFunction(realm, CMString("next"), g, next, 1)
	DefineBuiltinFunction(realm, CMString("return"), g, iteratorReturn, 1)

	g.defineBuiltinProperty(CMString("constructor"), &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.GeneratorFunctionPrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	g.defineToStringTag("Generator")

	return g
}

// GeneratorStart
// spec: 27.5.3.1
func GeneratorStart(agent *Agent, generator *GeneratorObject, generatorBody *ECMAScriptFunction) {
	Assert(generator.GeneratorState == GeneratorStateSuspendedStart)
	genContext := agent.RunningExecutionContext()
	genContext.Generator = generator

	// TODO(BM): use new vm arch
	closure := func(bytecode *BytecodeContext) ObjectType {
		a := bytecode.agent
		acGenContext := a.RunningExecutionContext()
		acGenerator := acGenContext.Generator
		if bytecode.IsFinished() {
			a.ExecutionContextStack.Pop()
			acGenerator.GeneratorState = GeneratorStateCompleted
			return CreateIterResultObject(a, UndefinedValue, true)
		}
		result := bytecode.Run()
		if result.IsError() {
			return MustGetObject(result.Error())
		}
		acGenerator.GeneratorState = GeneratorStateSuspendedYield
		return CreateIterResultObject(a, result.Data(), false)
	}
	generator.closure = closure
	generator.bytecode = GenerateBytecode(agent, generatorBody.ECMAScriptCode)
	generator.GeneratorContext = genContext
}

// 27.5.3.2
func GeneratorValidate(agent *Agent, generator Value) GeneratorState {
	g := RequireInternalSlot[*GeneratorObject](generator)
	if g.GeneratorState == GeneratorStateExecuting {
		// TODO:
		panic("TypeError")
	}
	return g.GeneratorState
}

// 27.5.3.3
func GeneratorResume(agent *Agent, generator Value, value Value) ObjectType {
	state := GeneratorValidate(agent, generator)
	if state == GeneratorStateCompleted {
		return CreateIterResultObject(agent, UndefinedValue, true)
	}
	Assert(state == GeneratorStateSuspendedStart || state == GeneratorStateSuspendedYield)
	g := RequireInternalSlot[*GeneratorObject](generator)
	genContext := g.GeneratorContext
	// methodContext := agent.RunningExecutionContext()
	g.GeneratorState = GeneratorStateExecuting
	agent.ExecutionContextStack.Push(genContext)
	result := g.closure(g.bytecode)
	// Assert(methodContext == agent.RunningExecutionContext())
	return result
}

func GeneratorResumeAbrupt(agent *Agent, generator Value, abruptCompletion CompletionValue) ObjectType {
	state := GeneratorValidate(agent, generator)
	g := RequireInternalSlot[*GeneratorObject](generator)
	if state == GeneratorStateSuspendedStart {
		g.GeneratorState = GeneratorStateCompleted
		state = GeneratorStateCompleted
	}
	if state == GeneratorStateCompleted {
		if abruptCompletion.Type == CompletionTypeReturn {
			return CreateIterResultObject(agent, UndefinedValue, true)
		}
		agent.exception = abruptCompletion.Error()
		return MustGetObject(abruptCompletion.Error())
	}

	genContext := g.GeneratorContext
	// methodContext:= agent.RunningExecutionContext()
	g.GeneratorState = GeneratorStateExecuting
	agent.ExecutionContextStack.Push(genContext)
	result := g.closure(g.bytecode)
	return result
}

// 27.5.3.5
type GeneratorKind int

const (
	GeneratorKindNonGenerator GeneratorKind = iota
	GeneratorKindSync
	GeneratorKindAsync
)

func GetGeneratorKind(agent *Agent) GeneratorKind {
	ec := agent.RunningExecutionContext()
	if ec.Generator == nil {
		return GeneratorKindNonGenerator
	}
	// TODO: Async
	return GeneratorKindSync
}

// 27.5.3.6
func GeneratorYield(agent *Agent, iteratorResult ObjectType) CompletionValue {
	genContext := agent.RunningExecutionContext()
	Assert(genContext.Generator != nil)
	generator := genContext.Generator
	Assert(GetGeneratorKind(agent) == GeneratorKindSync)
	generator.GeneratorState = GeneratorStateSuspendedYield
	agent.ExecutionContextStack.Pop()
	generator.result = iteratorResult.ToValue()
	return NewCompletionValue(iteratorResult.Get(CMString("value").ToPropertyKey()))
}

// 27.5.3.7
func Yield(agent *Agent, value Value) CompletionValue {
	generatorKind := GetGeneratorKind(agent)
	switch generatorKind {
	case GeneratorKindNonGenerator:
		panic("unreachable")
	case GeneratorKindSync:
		return GeneratorYield(agent, CreateIterResultObject(agent, value, false))
	case GeneratorKindAsync:
		panic("unimplemented")
	}
	panic("unreachable")
}
