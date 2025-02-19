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
	closure        func() Value
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
		return GeneratorResume(agent, this, value)
	}
	var iteratorReturn BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		value := pkg.SliceSafeGet(argumentsList, 0)
		generator := this
		C := NewCompletionReturnValue(value)
		return GeneratorResumeAbrupt(agent, generator, C)
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

type AbstractClosure func() Value

// GeneratorBody is Expression or Abstract closure
type GeneratorBody interface{}

// GeneratorStart
// spec: 27.5.3.1
func GeneratorStart(agent *Agent, generator *GeneratorObject, generatorBody GeneratorBody) {
	Assert(generator.GeneratorState == GeneratorStateSuspendedStart)
	genContext := agent.RunningExecutionContext()
	genContext.Generator = generator

	genVM := NewVM2(agent)
	genVM.suspendedGeneratorBody = generatorBody
	closure := func() Value {
		a := agent
		acGenContext := a.RunningExecutionContext()
		acGenerator := acGenContext.Generator
		// TODO: result should be a completion
		var result Value
		if body, ok := genVM.suspendedGeneratorBody.(RuntimeSemanticsEvaluation); ok {
			result = body.Evaluation(genVM)
		} else {
			result = generatorBody.(AbstractClosure)()
		}
		// TODO: not standard
		if acGenerator.GeneratorState == GeneratorStateSuspendedYield {
			return CreateIterResultObject(a, result, false).ToValue()
		}
		// TODO: Assert generator status
		a.ExecutionContextStack.Pop()
		acGenerator.GeneratorState = GeneratorStateCompleted
		var resultValue Value = result
		return CreateIterResultObject(a, resultValue, true).ToValue()
	}
	generator.closure = closure
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

// GeneratorResume
// spec: 27.5.3.3
func GeneratorResume(agent *Agent, generator Value, value Value) Value {
	state := GeneratorValidate(agent, generator)
	if state == GeneratorStateCompleted {
		return CreateIterResultObject(agent, UndefinedValue, true).ToValue()
	}
	Assert(state == GeneratorStateSuspendedStart || state == GeneratorStateSuspendedYield)
	g := RequireInternalSlot[*GeneratorObject](generator)
	genContext := g.GeneratorContext
	methodContext := agent.RunningExecutionContext()
	g.GeneratorState = GeneratorStateExecuting
	agent.ExecutionContextStack.Push(genContext)
	result := g.closure()
	Assert(methodContext == agent.RunningExecutionContext())
	return result
}

func GeneratorResumeAbrupt(agent *Agent, generator Value, abruptCompletion CompletionValue) Value {
	state := GeneratorValidate(agent, generator)
	g := RequireInternalSlot[*GeneratorObject](generator)
	if state == GeneratorStateSuspendedStart {
		g.GeneratorState = GeneratorStateCompleted
		state = GeneratorStateCompleted
	}
	if state == GeneratorStateCompleted {
		if abruptCompletion.Type == CompletionTypeReturn {
			return CreateIterResultObject(agent, UndefinedValue, true).ToValue()
		}
		agent.exception = abruptCompletion.Error()
		return abruptCompletion.Error()
	}

	genContext := g.GeneratorContext
	// methodContext:= agent.RunningExecutionContext()
	g.GeneratorState = GeneratorStateExecuting
	agent.ExecutionContextStack.Push(genContext)
	result := g.closure()
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

// GeneratorYield
// spec: 27.5.3.6
func GeneratorYield(agent *Agent, iterNextObj ObjectType) CompletionValue {
	genContext := agent.RunningExecutionContext()
	Assert(genContext.Generator != nil)
	generator := genContext.Generator
	Assert(GetGeneratorKind(agent) == GeneratorKindSync)
	generator.GeneratorState = GeneratorStateSuspendedYield
	agent.ExecutionContextStack.Pop()
	generator.result = iterNextObj.ToValue()
	return NewCompletionValue(iterNextObj.Get(CMString("value").ToPropertyKey()))
}

// Yield
// spec: 27.5.3.7
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
