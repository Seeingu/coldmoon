package coldmoon

import "fmt"

type ConstructorKind int

const (
	ConstructorKindBase ConstructorKind = iota
	ConstructorKindDerived
)

type ThisMode int

const (
	ThisModeLexical ThisMode = iota
	ThisModeStrict
	ThisModeGlobal
)

type ECMAScriptFunction struct {
	*Object
	Realm              *Realm
	Environment        EnvironmentRecord
	PrivateEnvironment *PrivateEnvironment
	FormalParameters   *FormalParameters
	ECMAScriptCode     *FunctionBody
	ConstructorKind    ConstructorKind
	ScriptOrModule     ScriptOrModule
	ThisMode           ThisMode
	Strict             bool
	HomeObject         ObjectType
	SourceText         string
	IsClassConstructor bool
}

// 7.3.24
func (e *ECMAScriptFunction) GetFunctionRealm() *Realm {
	return e.Realm
}

// 10.2.1
func Call(object ObjectType, thisArgument Value, argumentsList []Value) Value {
	agent := object.Agent()
	function := object.(*ECMAScriptFunction)

	calleeContext := PrepareForOrdinaryCall(agent, function, nil)
	Assert(calleeContext == agent.runningExecutionContext())

	if function.IsClassConstructor {
		panic("TypeError")
	}

	OrdinaryCallBindThis(agent, function, calleeContext, thisArgument)

	result := OrdinaryCallEvaluateBody(agent, function, argumentsList)

	agent.ExecutionContextStack.Pop()

	if result.Type == Return {
		return result.Value
	}
	return nil
}

// 10.2.1.1
func PrepareForOrdinaryCall(agent *Agent, function *ECMAScriptFunction, newTarget ObjectType) *ExecutionContext {
	localEnv := NewFunctionEnvironment(function, newTarget)

	calleeContext := &ExecutionContext{
		Function:       function.Object,
		Realm:          function.Realm,
		ScriptOrModule: function.ScriptOrModule,
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  localEnv,
			VariableEnvironment: localEnv,
			PrivateEnvironment:  function.PrivateEnvironment,
		},
	}

	agent.ExecutionContextStack.Push(calleeContext)

	return calleeContext
}

// 10.2.1.2
func OrdinaryCallBindThis(agent *Agent, function *ECMAScriptFunction, calleeContext *ExecutionContext, thisArgument Value) {
	thisMode := function.ThisMode

	if thisMode == ThisModeLexical {
		return
	}

	calleeRealm := function.Realm

	localEnv := calleeContext.ECMAScriptCode.LexicalEnvironment

	var thisValue Value
	if thisMode == ThisModeStrict {
		thisValue = thisArgument
	} else {
		if thisArgument == nil || thisArgument == UndefinedValue || thisArgument == NullValue {
			globalEnv := calleeRealm.GlobalEnv
			thisValue = NewValueFromObject(globalEnv.GlobalThisValue)
		} else {
			thisValue = NewValueFromObject(ValueToObject(agent, thisArgument))
		}
	}

	funEnv, ok := localEnv.(*FunctionEnvironment)
	Assert(ok)

	_ = funEnv.BindThisValue(thisValue)

	return
}

// 10.2.1.4
func OrdinaryCallEvaluateBody(agent *Agent, function *ECMAScriptFunction, argumentsList []Value) *CompletionRecord {
	calleeContext := agent.runningExecutionContext()
	calleeEnv := calleeContext.ECMAScriptCode.LexicalEnvironment
	env := NewDeclarativeEnvironment(calleeEnv)
	calleeContext.ECMAScriptCode.LexicalEnvironment = env
	for i, item := range function.FormalParameters.Items {
		identifier := item.(*FormalParameter).BindingElement.Identifier
		value := argumentsList[i]
		env.CreateMutableBinding(string(identifier), false)
		env.InitializeBinding(string(identifier), value)
	}
	return GenerateAndRunBytecode(agent, function.ECMAScriptCode)
}

type functionCreateThisMode int

const (
	functionCreateThisModeLexical functionCreateThisMode = iota
	functionCreateThisModeNonLexical
)

// 10.2.2
func ECMAScriptFunctionConstruct(
	object ObjectType,
	argumentsList []Value,
	newTarget ObjectType,
) ObjectType {
	agent := object.Agent()
	function := object.(*ECMAScriptFunction)

	kind := function.ConstructorKind

	var thisArgument Value
	if kind == ConstructorKindBase {
		thisArgument = NewValueFromObject(
			OrdinaryCreateFromConstructor(
				agent,
				newTarget,
				"%Object.prototype%",
				nil,
			))
	}

	calleeContext := PrepareForOrdinaryCall(agent, function, newTarget)
	Assert(calleeContext == agent.runningExecutionContext())

	if kind == ConstructorKindBase {
		OrdinaryCallBindThis(agent, function, calleeContext, thisArgument)
	}

	constructorEnv := calleeContext.ECMAScriptCode.LexicalEnvironment

	result := OrdinaryCallEvaluateBody(agent, function, argumentsList)

	agent.ExecutionContextStack.Pop()

	if result.Type == Return {
		if o, ok := result.Value.(*ObjectValue); ok {
			return o.Object
		}
		if kind == ConstructorKindBase {
			return thisArgument.(*ObjectValue).Object
		}
		if result.Value != UndefinedValue {
			panic("TypeError")
		}
	} else {
		panic("ReturnIfAbrupt")
	}

	thisBinding := constructorEnv.GetThisBinding()
	thisBindingObject, ok := thisBinding.(*ObjectValue)
	Assert(ok)

	return thisBindingObject.Object
}

// 10.2.3
func OrdinaryFunctionCreate(
	agent *Agent,
	functionPrototype ObjectType,
	sourceText string,
	parameterList *FormalParameters,
	body *FunctionBody,
	functionCreateThisMode functionCreateThisMode,
	env EnvironmentRecord,
	privateEnv *PrivateEnvironment,
) *ECMAScriptFunction {
	var thisMode ThisMode = ThisModeLexical
	strict := body.FunctionBodyContainsUseStrict()

	function := &ECMAScriptFunction{
		Object:             NewObject(agent, functionPrototype),
		Realm:              agent.CurrentRealm(),
		SourceText:         sourceText,
		FormalParameters:   parameterList,
		ECMAScriptCode:     body,
		Strict:             strict,
		ThisMode:           thisMode,
		IsClassConstructor: false,
		Environment:        env,
		PrivateEnvironment: privateEnv,
		ScriptOrModule:     agent.GetActiveScriptOrModule(),
		HomeObject:         nil,
		ConstructorKind:    ConstructorKindBase,
	}
	function.InternalMethods().Call = Call

	length := parameterList.ExpectedArgumentCount()
	SetFunctionLength(function.Object, float64(length))

	return function
}

// 10.2.4
func AddRestrictedFunctionProperties(F ObjectType, realm *Realm) {
	// TODO: Assert
	thrower := realm.Intrinsics.ThrowTypeError
	F.DefinePropertyOrThrow(NewStringPropertyKey("caller"), &PropertyDescriptor{
		Get:          thrower,
		Set:          thrower,
		Enumerable:   false,
		Configurable: true,
	})

	F.DefinePropertyOrThrow(NewStringPropertyKey("arguments"), &PropertyDescriptor{
		Get:          thrower,
		Set:          thrower,
		Enumerable:   false,
		Configurable: true,
	})

}

// 10.2.5
func MakeConstructor(F ObjectType, writable bool, prototype ObjectType) {
	agent := F.Agent()
	realm := agent.CurrentRealm()
	fun, isECMAScriptFunction := F.(*ECMAScriptFunction)
	if isECMAScriptFunction {
		Assert(!isConstructor(NewValueFromObject(fun)))

		Assert(fun.IsExtensible() &&
			!F.PropertyStorage().Has(NewStringPropertyKey("prototype")),
		)
		F.InternalMethods().Construct = ECMAScriptFunctionConstruct
	} else {
		F.InternalMethods().Construct = BuiltinConstruct
	}

	if isECMAScriptFunction {
		fun.ConstructorKind = ConstructorKindBase
	}

	proto := prototype
	if proto == nil {
		proto = OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)

		proto.DefinePropertyOrThrow(NewStringPropertyKey("constructor"), &PropertyDescriptor{
			Value:        NewValueFromObject(F),
			Writable:     writable,
			Enumerable:   false,
			Configurable: true,
		})
	}

	F.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
		Value:        NewValueFromObject(proto),
		Writable:     writable,
		Enumerable:   false,
		Configurable: false,
	})
}

// 10.2.9
func SetFunctionName(function ObjectType, key PropertyKey, prefix string) {
	Assert(function.IsExtensible())
	Assert(!function.PropertyStorage().Has(NewStringPropertyKey("name")))

	var name string
	switch k := key.(type) {
	case SymbolPropertyKey:
		description := k.Value.Description

		if description == "" {
			name = description
		} else {
			name = fmt.Sprintf("[%s]", description)
		}
	case StringPropertyKey:
		name = k.Value
	default:
		panic("unimplemented")
	}

	builtinFunction, isBuiltinFunction := function.(*BuiltinFunction)
	if isBuiltinFunction {
		builtinFunction.InitialName = name
	}
	if prefix != "" {
		name = prefix + " " + name
		builtinFunction.InitialName = name
	}

	function.DefinePropertyOrThrow(NewStringPropertyKey("name"), &PropertyDescriptor{
		Value:        NewStringValue(name),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})

}

// 10.2.10
func SetFunctionLength(function ObjectType, length float64) {
	Assert(function.IsExtensible())
	Assert(!function.PropertyStorage().Has(NewStringPropertyKey("length")))

	function.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(length),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
}
