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
	// Deprecated: get from agent
	// TODO(BM): Remove
	GeneratorContext *ExecutionContext
	// [[GeneratorBrand]]
	GeneratorBrand string
	closure        func() CompletionValue
}

func (g *GeneratorObject) Resume() CompletionValue {
	return g.closure()
}

func NewGeneratorPrototype(realm *Realm) *GeneratorObject {
	agent := realm.Agent
	object := NewObject(agent, realm.Intrinsics.IteratorPrototype, "GeneratorPrototype")
	g := &GeneratorObject{
		Object: object,
	}
	g.ref = g

	var next BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		value := pkg.SliceSafeGet(argumentsList, 0)
		return GeneratorResume(agent, this, value)
	}
	var iteratorReturn BehaviorFn = func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		value := pkg.SliceSafeGet(argumentsList, 0)
		generator := this
		C := Completion[Value]{
			t:     CompletionTypeReturn,
			value: value,
		}
		return GeneratorResumeAbrupt(agent, generator, C)
	}

	g.defineBuiltinFunction(realm, CMString("next"), next, 1)
	g.defineBuiltinFunction(realm, CMString("return"), iteratorReturn, 1)

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
	agent.ExecutionContextMap[generator.GetId()] = genContext
	genVM := genContext.VM
	genContext.Generator = generator
	genContext.generatorCh = make(chan Value)
	closure := func() CompletionValue {
		a := agent
		var result CompletionValue
		if body, ok := generatorBody.(RuntimeSemanticsEvaluation); ok {
			result = body.Evaluation(genVM)
		} else {
			panic("unreachable")
		}
		// TODO: Assert generator status
		a.ExecutionContextStack.Pop()
		generator.GeneratorState = GeneratorStateCompleted

		var completion CompletionValue
		if result.t == CompletionTypeNormal {
			completion = CreateIterResultObject(a, UndefinedValue, true).ToValue().ToCompletion()
		} else if result.t == CompletionTypeReturn {
			completion = CreateIterResultObject(a, result.value, true).ToValue().ToCompletion()
		} else {
			completion = result
		}
		genContext.Result = completion
		a.RunningExecutionContext().Resume()
		return completion
	}
	generator.closure = closure
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
func GeneratorResume(agent *Agent, generator Value, value Value) CompletionValue {
	state := GeneratorValidate(agent, generator)
	if state == GeneratorStateCompleted {
		return CreateIterResultObject(agent, UndefinedValue, true).ToValue().ToCompletion()
	}
	Assert(state == GeneratorStateSuspendedStart || state == GeneratorStateSuspendedYield)
	g := RequireInternalSlot[*GeneratorObject](generator)
	genContext := agent.FindExecutionContextById(g.GetId())
	methodContext := agent.RunningExecutionContext()
	g.GeneratorState = GeneratorStateExecuting
	agent.ExecutionContextStack.Push(genContext)
	if state == GeneratorStateSuspendedStart {
		go g.Resume()
	} else {
		genContext.generatorCh <- value
	}
	methodContext.Suspend()
	Assert(methodContext == agent.RunningExecutionContext())
	return genContext.Result
}

// GeneratorResumeAbrupt
// spec: 27.5.3.4
func GeneratorResumeAbrupt(agent *Agent, generator Value, abruptCompletion CompletionValue) CompletionValue {
	state := GeneratorValidate(agent, generator)
	g := RequireInternalSlot[*GeneratorObject](generator)
	if state == GeneratorStateSuspendedStart {
		g.GeneratorState = GeneratorStateCompleted
		state = GeneratorStateCompleted
	}
	if state == GeneratorStateCompleted {
		if abruptCompletion.t == CompletionTypeReturn {
			return CreateIterResultObject(agent, UndefinedValue, true).ToValue().ToCompletion()
		}
		agent.exception = abruptCompletion.Error()
		return abruptCompletion
	}

	genContext := agent.FindExecutionContextById(g.GetId())
	methodContext := agent.RunningExecutionContext()
	g.GeneratorState = GeneratorStateExecuting
	agent.ExecutionContextStack.Push(genContext)
	go g.closure()
	genContext.Suspend()
	Assert(methodContext == agent.RunningExecutionContext())
	return genContext.Result
}

// 27.5.3.5
type GeneratorKind int

const (
	GeneratorKindNonGenerator GeneratorKind = iota
	GeneratorKindSync
	GeneratorKindAsync
)

// GetGeneratorKind
// spec: 27.5.3.5
func GetGeneratorKind(agent *Agent) GeneratorKind {
	ec := agent.RunningExecutionContext()
	if ec.Generator == nil && ec.AsyncGenerator == nil {
		return GeneratorKindNonGenerator
	} else if ec.AsyncGenerator != nil {
		return GeneratorKindAsync
	}
	return GeneratorKindSync
}

// GeneratorYield
// spec: 27.5.3.6
func GeneratorYield(agent *Agent, iterNextObj ObjectType) (co CompletionValue) {
	genContext := agent.RunningExecutionContext()
	Assert(genContext.Generator != nil)
	generator := genContext.Generator
	Assert(GetGeneratorKind(agent) == GeneratorKindSync)
	generator.GeneratorState = GeneratorStateSuspendedYield
	agent.ExecutionContextStack.Pop()
	callerContext := agent.RunningExecutionContext()
	co.value = iterNextObj.ToValue()

	genContext.Result = co
	callerContext.Resume()
	co.value = <-genContext.generatorCh
	return
}

// Yield
// spec: 27.5.3.7
func Yield(agent *Agent, value Value) (co CompletionValue) {
	generatorKind := GetGeneratorKind(agent)
	switch generatorKind {
	case GeneratorKindNonGenerator:
		panic("unreachable")
	case GeneratorKindSync:
		return GeneratorYield(agent, CreateIterResultObject(agent, value, false))
	case GeneratorKindAsync:
		v, isAbrupt, rt := ReturnIfAbrupt(Await(agent, value), co)
		if isAbrupt {
			return rt
		}
		return AsyncGeneratorYield(agent, v)
	}
	panic("unreachable")
}
