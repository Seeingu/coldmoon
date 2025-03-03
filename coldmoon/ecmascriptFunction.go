package coldmoon

import (
	"fmt"

	"github.com/Seeingu/coldmoon/pkg"
	"github.com/samber/lo"
)

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
	InternalSlotPrivateMethods
	InternalSlotFields
	InternalSlotClassFieldInitializerName
	privateMethods            []*PrivateElement
	fields                    []*ClassFieldDefinition
	classFieldInitializerName ClassFieldInitializerName
	Realm                     *Realm
	Environment               EnvironmentRecord
	PrivateEnvironment        *PrivateEnvironment
	FormalParameters          *FormalParameters
	ECMAScriptCode            *FunctionBody
	ConstructorKind           ConstructorKind
	ScriptOrModule            ScriptOrModule
	ThisMode                  ThisMode
	Strict                    bool
	HomeObject                ObjectType
	SourceText                string
	IsClassConstructor        bool
}

var (
	_ InternalSlotPrivateMethods            = (*ECMAScriptFunction)(nil)
	_ InternalSlotFields                    = (*ECMAScriptFunction)(nil)
	_ InternalSlotClassFieldInitializerName = (*ECMAScriptFunction)(nil)
)

func (e *ECMAScriptFunction) PrivateMethods() []*PrivateElement {
	return e.privateMethods
}

func (e *ECMAScriptFunction) Fields() []*ClassFieldDefinition {
	return e.fields
}

func (e *ECMAScriptFunction) SetFields(f []*ClassFieldDefinition) {
	e.fields = f
}

func (e *ECMAScriptFunction) ClassFieldInitializerName() ClassFieldInitializerName {
	return e.classFieldInitializerName
}

func (e *ECMAScriptFunction) SetClassFieldInitializerName(n ClassFieldInitializerName) {
	e.classFieldInitializerName = n
}

func (e *ECMAScriptFunction) EvaluateBody() Completion[Value] {
	return RunNode(e.Agent(), e.ECMAScriptCode)
}

// 7.3.24
func (e *ECMAScriptFunction) GetFunctionRealm() *Realm {
	return e.Realm
}

// Call
// spec: 10.2.1
// returns either normal completion or throw
func (e *ECMAScriptFunction) Call(thisArgument Value, argumentsList []Value) (co CompletionValue) {
	agent := e.Agent()
	function := e

	calleeContext := PrepareForOrdinaryCall(agent, function, nil)
	Assert(calleeContext == agent.RunningExecutionContext())

	if function.IsClassConstructor {
		return co.ThrowTypeError(e.Agent(), "IsClassConstructor")
	}

	OrdinaryCallBindThis(agent, function, calleeContext, thisArgument)

	result := OrdinaryCallEvaluateBody(agent, function, argumentsList)

	agent.ExecutionContextStack.Pop()

	if result.IsAbrupt() {
		if result.t != CompletionTypeThrow {
			result.t = CompletionTypeNormal
		}
		return result
	}
	return UndefinedValue.ToCompletion()
}

// PrepareForOrdinaryCall
// spec: 10.2.1.1
func PrepareForOrdinaryCall(agent *Agent, function *ECMAScriptFunction, newTarget ObjectType) *ExecutionContext {
	localEnv := NewFunctionEnvironment(function, newTarget)
	calleeContext := &ExecutionContext{
		Function:       function,
		Realm:          function.Realm,
		VM:             NewVM2(agent),
		ch:             make(chan struct{}),
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
			thisValue = (globalEnv.GlobalThisValue).ToValue()
		} else {
			thisValue = thisArgument.ToObject(agent).value.ToValue()
		}
	}

	funEnv, ok := localEnv.(*FunctionEnvironment)
	Assert(ok)

	_ = funEnv.BindThisValue(thisValue)

	return
}

// OrdinaryCallEvaluateBody
// spec: 10.2.1.4
func OrdinaryCallEvaluateBody(agent *Agent, function *ECMAScriptFunction, argumentsList []Value) (value CompletionValue) {
	return function.ECMAScriptCode.EvaluateBody(agent, function, argumentsList)
}

func EvaluateAsyncFunctionBody(agent *Agent, function *ECMAScriptFunction, argumentsList []Value) (co Completion[Value]) {
	realm := agent.CurrentRealm()
	promiseCapability := NewPromiseCapability(agent, (realm.Intrinsics.Promise).ToValue())
	completion := FunctionDeclarationInstantiation(agent, function, argumentsList)
	if completion.IsError() {
		ex := agent.exception
		promiseCapability.Reject.Call(
			UndefinedValue,
			[]Value{ex},
		)
		agent.exception = nil
	} else {
		AsyncFunctionStart(agent, promiseCapability, function)
	}
	co.t = CompletionTypeReturn
	co.value = promiseCapability.Promise.ToValue()
	return
}

func EvaluateFunctionBody(agent *Agent, function *ECMAScriptFunction, argumentsList []Value) CompletionValue {
	functionBody := function.ECMAScriptCode
	FunctionDeclarationInstantiation(agent, function, argumentsList)
	return RunNode(agent, functionBody)
}

func EvaluateAsyncGeneratorBody(agent *Agent, function *ECMAScriptFunction, argumentsList []Value) (co CompletionValue) {
	FunctionDeclarationInstantiation(agent, function, argumentsList)
	G := OrdinaryCreateFromConstructor(agent, function, "%AsyncGeneratorFunction.prototype.prototype%", nil)
	co.t = CompletionTypeReturn
	co.value = G.ToValue()
	return
}

// EvaluateGeneratorBody
// spec: 15.5.2
func EvaluateGeneratorBody(agent *Agent, function *ECMAScriptFunction, argumentsList []Value) (co CompletionValue) {
	FunctionDeclarationInstantiation(agent, function, argumentsList)
	o := OrdinaryCreateFromConstructor(agent, function, "%GeneratorFunction.prototype.prototype%", nil)
	G := &GeneratorObject{
		Object: o,
	}
	G.ref = G

	GeneratorStart(agent, G, function.ECMAScriptCode)
	co.t = CompletionTypeReturn
	co.value = G.ToValue()
	return
}

// FunctionDeclarationInstantiation
// spec: 10.2.11
func FunctionDeclarationInstantiation(agent *Agent, function *ECMAScriptFunction, argumentsList ArgumentsList) (co Completion[Value]) {
	calleeContext := agent.RunningExecutionContext()
	code := function.ECMAScriptCode
	strict := function.Strict
	formals := function.FormalParameters
	parameterNames := formals.BoundNames()
	var hasDuplicates bool
	uniqueNames := make(map[IdentifierName]bool)
loop:
	for _, name := range parameterNames {
		if _, exists := uniqueNames[name]; exists {
			hasDuplicates = true
			break loop
		}
		uniqueNames[name] = true
	}

	simpleParameterList := formals.IsSimpleParameterList()
	hasParameterExpressions := formals.ContainsExpression()

	// varNames := code.VarDeclaredNames()
	varDeclarations := code.VarScopedDeclarations()
	lexicalNames := code.LexicallyDeclaredNames()
	var functionNames []IdentifierName

	argumentsObjectNeeded := true
	if function.ThisMode == ThisModeLexical {
		argumentsObjectNeeded = false
	} else if lo.Contains(parameterNames, "arguments") {
		argumentsObjectNeeded = false
	} else if !hasParameterExpressions {
		if lo.Contains(functionNames, "arguments") || lo.Contains(lexicalNames, "arguments") {
			argumentsObjectNeeded = false
		}
	}
	var env EnvironmentRecord
	if strict || !hasParameterExpressions {
		env = calleeContext.ECMAScriptCode.LexicalEnvironment
	} else {
		calleeEnv := calleeContext.ECMAScriptCode.LexicalEnvironment
		Assert(calleeContext.ECMAScriptCode.LexicalEnvironment == calleeEnv)
		env = NewDeclarativeEnvironment(calleeEnv)
	}

	for _, paramName := range parameterNames {
		alreadyDeclared := env.HasBinding(string(paramName))
		if !alreadyDeclared {
			env.CreateMutableBinding(string(paramName), false)
			if hasDuplicates {
				env.InitializeBinding(string(paramName), UndefinedValue)
			}
		}
	}
	var parameterBindings []string
	if argumentsObjectNeeded {
		var argumentsObject ObjectType
		if strict || !simpleParameterList {
			argumentsObject = CreateUnmappedArgumentsObject(agent, argumentsList)
		} else {
			argumentsObject = CreateMappedArgumentsObject(agent, function, formals, argumentsList, env)
		}
		if strict {
			env.CreateImmutableBinding("arguments", false)
		} else {
			env.CreateMutableBinding("arguments", false)
		}
		env.InitializeBinding("arguments", (argumentsObject).ToValue())

		for _, parameterName := range parameterNames {
			parameterBindings = append(parameterBindings, string(parameterName))
		}
		parameterBindings = append(parameterBindings, "arguments")
	} else {
		for _, parameterName := range parameterNames {
			parameterBindings = append(parameterBindings, string(parameterName))
		}
	}

	for i, item := range formals.Items {
		e := env
		switch param := item.(type) {
		case *FormalParameter:
			name := param.BindingElement.SingleNameBinding.BindingIdentifier
			initializer := param.BindingElement.SingleNameBinding.Initializer
			value := pkg.SliceSafeGet(argumentsList, i)
			if value == nil {
				value = UndefinedValue
			}
			ref := agent.ResolveBinding(string(name), e, strict)
			if initializer != nil {
				node := RunNode(agent, &ExpressionStatement{
					Expression: initializer,
				})
				value = node.Data()
			}
			if e == nil {
				_, isAbrupt, rt := ReturnIfAbrupt(ref.PutValue(agent, value), co)
				if isAbrupt {
					return rt
				}
			} else {
				ref.InitializeReferencedBinding(value)
			}
		case *FormalParameterFunctionRestParameter:
			name := param.BindingRestElement.(*BindingRestElementIdentifier).Identifier
			ref := agent.ResolveBinding(string(name), e, strict)
			array := ArrayCreate(agent, 0, nil)
			rest := argumentsList[i:]
			for j, value := range rest {
				array.CreateDataPropertyOrThrow(NewIntegerIndexPropertyKey(JSInt(j)), value)
			}
			if e == nil {
				_, isAbrupt, rt := ReturnIfAbrupt(ref.PutValue(agent, array.ToValue()), co)
				if isAbrupt {
					return rt
				}
			} else {
				ref.InitializeReferencedBinding(array.ToValue())
			}
		default:
			panic("unreachable")
		}
	}

	var varEnv EnvironmentRecord
	if !hasParameterExpressions {
		instantiatedVarNames := make(map[IdentifierName]bool)
		for _, paramBinding := range parameterBindings {
			instantiatedVarNames[IdentifierName(paramBinding)] = true
		}
		for _, declaration := range varDeclarations {
			varName := declaration.BindingIdentifier
			if _, exists := instantiatedVarNames[varName]; !exists {
				instantiatedVarNames[varName] = true
				env.CreateMutableBinding(string(varName), false)
				env.InitializeBinding(string(varName), UndefinedValue)
			}
		}
		varEnv = env
	} else {
		varEnv = NewDeclarativeEnvironment(env)
		calleeContext.ECMAScriptCode.VariableEnvironment = varEnv
		instantiatedVarNames := make(map[IdentifierName]bool)
		for _, declaration := range varDeclarations {
			varName := declaration.BindingIdentifier
			if _, exists := instantiatedVarNames[varName]; !exists {
				instantiatedVarNames[varName] = true
				varEnv.CreateMutableBinding(string(varName), false)
				var initialValue Value
				if !lo.Contains(parameterNames, varName) || lo.Contains(functionNames, varName) {
				} else {
					value := env.GetBindingValue(agent, string(varName), false)
					initialValue = value.Data()
				}
				varEnv.InitializeBinding(string(varName), initialValue)
			}
		}
	}

	lexEnv := varEnv
	if !strict {
		lexEnv = NewDeclarativeEnvironment(varEnv)
	}
	calleeContext.ECMAScriptCode.LexicalEnvironment = lexEnv
	co.value = UndefinedValue
	return
}

type functionCreateThisMode int

const (
	functionCreateThisModeLexical functionCreateThisMode = iota
	functionCreateThisModeNonLexical
)

// Construct
// spec: 10.2.2
func (e *ECMAScriptFunction) Construct(
	argumentsList []Value,
	_newTarget ObjectType,
) (co Completion[ObjectType]) {
	newTarget := _newTarget
	if newTarget == nil {
		newTarget = e
	}
	agent := e.Agent()
	function := e

	kind := function.ConstructorKind

	var thisArgument Value
	if kind == ConstructorKindBase {
		thisArgument = OrdinaryCreateFromConstructor(
			agent,
			newTarget,
			"%Object.prototype%",
			nil,
		).ToValue()
	}

	calleeContext := PrepareForOrdinaryCall(agent, function, newTarget)
	Assert(calleeContext == agent.RunningExecutionContext())

	if kind == ConstructorKindBase {
		OrdinaryCallBindThis(agent, function, calleeContext, thisArgument)

		MustGetObject(thisArgument).InitializeInstanceElements(function)
	}

	constructorEnv := calleeContext.ECMAScriptCode.LexicalEnvironment

	result, isAbrupt, rt := ReturnIfAbrupt(OrdinaryCallEvaluateBody(agent, function, argumentsList), co)

	agent.ExecutionContextStack.Pop()

	if isAbrupt {
		if o, ok := result.(*ObjectValue); ok {
			return o.Object.ToCompletion()
		}
		if kind == ConstructorKindBase {
			return MustGetObject(thisArgument).ToCompletion()
		}
		if result != UndefinedValue {
			return co.ThrowTypeError(agent, "Construct TypeError")
		}
		return rt
	}

	thisBinding := constructorEnv.GetThisBinding()
	thisBindingObject, ok := thisBinding.(*ObjectValue)
	Assert(ok)

	return thisBindingObject.Object.ToCompletion()
}

func ECMAScriptFunctionConstruct(
	object ObjectType,
	argumentsList []Value,
	newTarget ObjectType,
) Completion[ObjectType] {
	return object.(*ECMAScriptFunction).Construct(argumentsList, newTarget)
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
	var thisMode ThisMode
	strict := body.Strict
	if functionCreateThisMode == functionCreateThisModeNonLexical {
		if strict {
			thisMode = ThisModeStrict
		} else {
			thisMode = ThisModeGlobal
		}
	} else {
		thisMode = ThisModeLexical
	}

	function := &ECMAScriptFunction{
		Object:             NewObject(agent, functionPrototype, "ECMAScriptFunction"),
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
	function.ref = function
	function.InternalMethods().Call = func(o ObjectType, this Value, arguments []Value) CompletionValue {
		return o.(*ECMAScriptFunction).Call(this, arguments)
	}

	length := parameterList.ExpectedArgumentCount()
	SetFunctionLength(function.Object, length)

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

func MakeClassConstructor(function *ECMAScriptFunction) {
	Assert(!function.IsClassConstructor)
	function.IsClassConstructor = true
}

// 10.2.5
func MakeConstructor(F ObjectType, writable bool, prototype ObjectType) {
	agent := F.Agent()
	realm := agent.CurrentRealm()
	fun, isECMAScriptFunction := F.(*ECMAScriptFunction)

	if isECMAScriptFunction {
		Assert(!IsConstructor((fun).ToValue()))

		Assert(fun.IsExtensible() &&
			!F.PropertyStorage().Has(NewStringPropertyKey("prototype")),
		)
		F.InternalMethods().Construct = ECMAScriptFunctionConstruct
	} else {
		F.InternalMethods().Construct = BuiltinConstruct
	}

	if isECMAScriptFunction {
		fun.ConstructorKind = ConstructorKindBase
	} else if b, isBuiltinFunction := F.(*BuiltinFunction); isBuiltinFunction {
		b.AdditionalFields.ClassConstructorFields.ConstructorKind = ConstructorKindBase
	}

	proto := prototype
	if proto == nil {
		proto = OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)

		proto.DefinePropertyOrThrow(NewStringPropertyKey("constructor"), &PropertyDescriptor{
			Value:        F.ToValue(),
			Writable:     writable,
			Enumerable:   false,
			Configurable: true,
		})
	}

	F.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
		Value:        proto.ToValue(),
		Writable:     writable,
		Enumerable:   false,
		Configurable: false,
	})
}

// 10.2.7
func MakeMethod(F *ECMAScriptFunction, homeObject ObjectType) {
	F.HomeObject = homeObject
}

type (
	PropertyKeyOrPrivateName interface {
		_propertyKeyOrPrivateName()
	}
	PropertyKeyOrPrivateNameName struct {
		PropertyKeyOrPrivateName
		PrivateName PrivateName
	}
)

// DefineMethodProperty
// spec: 10.2.8
// return PrivateElement or nil(UNUSED)
func DefineMethodProperty(homeObject ObjectType, key PropertyKeyOrPrivateName, closure ObjectType, enumerable bool) (c Completion[*PrivateElement]) {
	Assert(homeObject.IsExtensible())
	switch k := key.(type) {
	case PropertyKeyOrPrivateNameName:
		c.value = &PrivateElement{
			Key:   k.PrivateName,
			Kind:  PrivateElementKindMethod,
			Value: closure.ToValue(),
		}
		return
	case PropertyKey:
		desc := &PropertyDescriptor{
			Value:        closure.ToValue(),
			Writable:     true,
			Enumerable:   enumerable,
			Configurable: true,
		}
		homeObject.DefinePropertyOrThrow(k, desc)
		// unused
		return
	}
	panic("unreachable")
	return
}

// 10.2.9
func SetFunctionName(function ObjectType, key PropertyKeyOrPrivateName, prefix string) {
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
	if prefix != "" {
		name = prefix + " " + name
	}
	if isBuiltinFunction {
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
func SetFunctionLength(function ObjectType, length JSInt) {
	Assert(function.IsExtensible())
	Assert(!function.PropertyStorage().Has(NewStringPropertyKey("length")))

	function.DefinePropertyOrThrow(NewStringPropertyKey("length"), &PropertyDescriptor{
		Value:        NewNumberValue(length.ToNumber()),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
}

// MARK: - Internal

func (e *ECMAScriptFunction) String() string {
	return fmt.Sprintf("ECMAScriptFunction: %s", e.SourceText)
}
