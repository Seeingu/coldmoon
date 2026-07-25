package coldmoon

import "math"

func NewFunctionPrototypeWithIntrinsicsBinding(realm *Realm) ObjectType {
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		return UndefinedValue
	}
	f := CreateBuiltinFunction(realm.Agent, behavior, 0, CMString(""), builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.ObjectPrototype,
		prefix:        "",
		isConstructor: false,
	})

	realm.Intrinsics.FunctionPrototype = f
	initFunctionMethods(f, realm)

	return f
}

// initFunctionMethods depends on %Function.prototype% being defined
func initFunctionMethods(f ObjectType, realm *Realm) {
	agent := realm.Agent
	// 20.2.3.5 toString
	toString := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		o, ok := this.(*ObjectValue)
		if ok {
			ecmascriptFunction, ok := o.Object.(*ECMAScriptFunction)
			if ok {
				return NewStringValue(ecmascriptFunction.SourceText)
			}

			builtinFunction, ok := o.Object.(*BuiltinFunction)
			if ok {
				name := builtinFunction.InitialName
				sourceText := "function " + name + "() { [native code] }"
				return NewStringValue(sourceText)
			}
		}
		if IsCallable(this) {
			return NewStringValue("function () { [native code] }")
		}

		panic("TypeError")
	}
	f.defineBuiltinFunction(realm, CMString("toString"), toString, 0)

	// 20.2.3.3
	call := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		thisArg := Value(UndefinedValue)
		if len(argumentsList) > 0 {
			thisArg = argumentsList[0]
		}
		var args []Value
		if len(argumentsList) > 1 {
			args = argumentsList[1:]
		}
		fun := this
		if !IsCallable(fun) {
			var co CompletionValue
			return co.ThrowTypeError(agent, "Function.prototype.call receiver is not callable")
		}
		return fun.Call(agent, thisArg, args)
	}
	bind := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		var co CompletionValue
		thisArg := Value(UndefinedValue)
		if len(argumentsList) > 0 {
			thisArg = argumentsList[0]
		}
		var args []Value
		if len(argumentsList) > 1 {
			args = argumentsList[1:]
		}
		target := this
		if !IsCallable(target) {
			return co.ThrowTypeError(agent, "Function.prototype.bind receiver is not callable")
		}
		targetObject := MustGetObject(target)
		F := BoundFunctionCreate(realm.Agent, targetObject, thisArg, args)
		var L JSInt = 0
		targetHasLength := targetObject.HasProperty(NewStringPropertyKey("length"))
		if targetHasLength {
			targetLen := targetObject.Get(NewStringPropertyKey("length"))
			if n, ok := ValueGet[*NumberValue](targetLen); ok {
				if n.IsPositiveInf() {
					L = JSInt(math.Inf(1))
				} else if n.IsNegativeInf() {
					L = 0
				} else {
					targetLenAsInt, isAbrupt, rt := ReturnIfAbrupt(ToIntegerOrInfinity(agent, targetLen), co)
					if isAbrupt {
						return rt
					}
					Assert(!targetLenAsInt.IsInf())
					argCount := JSInt(len(args))
					L = (targetLenAsInt - argCount).Max(0)
				}
			}
		}

		SetFunctionLength(F, L)
		targetName := targetObject.Get(NewStringPropertyKey("name"))
		targetNameString := ""
		if s, ok := ValueGet[*StringValue](targetName); ok {
			targetNameString = s.Data
		}
		SetFunctionName(F, NewStringPropertyKey(targetNameString), "bound")
		return (F).ToValue()
	}
	f.defineBuiltinFunction(realm, CMString("call"), call, 1)
	f.defineBuiltinFunction(realm, CMString("bind"), bind, 1)

	apply := func(this Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		thisArg := Value(UndefinedValue)
		if len(argumentsList) > 0 {
			thisArg = argumentsList[0]
		}
		argArray := Value(UndefinedValue)
		if len(argumentsList) > 1 {
			argArray = argumentsList[1]
		}
		fun := this
		if !IsCallable(fun) {
			var co CompletionValue
			return co.ThrowTypeError(agent, "Function.prototype.apply receiver is not callable")
		}

		if argArray == UndefinedValue || argArray == NullValue {
			return fun.Call(agent, thisArg, []Value{})
		}

		var co CompletionValue
		argList, isAbrupt, rt := ReturnIfAbrupt(CreateListFromArrayLike(realm.Agent, argArray), co)
		if isAbrupt {
			return rt
		}
		return fun.Call(agent, thisArg, argList)
	}
	f.defineBuiltinFunction(realm, CMString("apply"), apply, 2)
}

type dynamicFunctionKind int

const (
	dynamicFunctionKindNormal dynamicFunctionKind = iota
	dynamicFunctionKindGenerator
	dynamicFunctionKindAsync
	dynamicFunctionKindAsyncGenerator
)

type GrammarSymbol[T any] struct {
	acceptFn func(*Parser) T
}

// 20.2.1.1.1
func CreateDynamicFunction(
	agent *Agent,
	constructor ObjectType,
	_newTarget ObjectType,
	kind dynamicFunctionKind,
	parameterList ArgumentsList,
	bodyArg Value,
) ObjectType {
	realm := agent.CurrentRealm()
	currentRealm := realm
	agent.HostHooks.HostEnsureCanCompileStrings(currentRealm)

	newTarget := _newTarget
	if _newTarget == nil {
		newTarget = constructor
	}

	var prefix string
	var fallbackPrototype IntrinsicName
	exprSym := &GrammarSymbol[Expression]{}
	bodySym := &GrammarSymbol[*FunctionBody]{}
	parameterSym := &GrammarSymbol[*FormalParameters]{}
	switch kind {
	case dynamicFunctionKindNormal:
		prefix = "function"
		exprSym.acceptFn = func(p *Parser) Expression {
			return p.functionExpression()
		}
		bodySym.acceptFn = func(p *Parser) *FunctionBody {
			return p.functionBody(FunctionTypeNormal)
		}
		parameterSym.acceptFn = func(p *Parser) *FormalParameters {
			return p.formalParameters()
		}
		fallbackPrototype = IntrinsicNameFunctionPrototype
	case dynamicFunctionKindGenerator:
		prefix = "function*"

		exprSym.acceptFn = func(p *Parser) Expression {
			return p.generatorExpression()
		}
		bodySym.acceptFn = func(p *Parser) *FunctionBody {
			return p.functionBody(FunctionTypeGenerator)
		}
		parameterSym.acceptFn = func(p *Parser) *FormalParameters {
			return p.formalParameters()
		}
		fallbackPrototype = IntrinsicNameGeneratorFunctionPrototype
	case dynamicFunctionKindAsyncGenerator:
		prefix = "async function*"

		exprSym.acceptFn = func(p *Parser) Expression {
			return p.asyncGeneratorExpression()
		}
		bodySym.acceptFn = func(p *Parser) *FunctionBody {
			return p.functionBody(FunctionTypeGenerator)
		}
		parameterSym.acceptFn = func(p *Parser) *FormalParameters {
			return p.formalParameters()
		}
		fallbackPrototype = IntrinsicNameAsyncGeneratorFunctionPrototype
	case dynamicFunctionKindAsync:
		prefix = "async function"

		exprSym.acceptFn = func(p *Parser) Expression {
			return p.asyncFunctionExpression()
		}
		bodySym.acceptFn = func(p *Parser) *FunctionBody {
			return p.functionBody(FunctionTypeAsync)
		}
		parameterSym.acceptFn = func(p *Parser) *FormalParameters {
			return p.formalParameters()
		}
		fallbackPrototype = "%AsyncFunction.prototype%"
	}

	argCount := len(parameterList)
	P := ""
	if argCount > 0 {
		P = parameterList[0].String()
		for i := 1; i < argCount; i++ {
			P += ", " + parameterList[i].String()
		}
	}

	bodyString := bodyArg.String()
	sourceString := prefix + " anonymous(" + P + "\n) {\n" + bodyString + "\n}"
	sourceText := sourceString

	parser := NewParser(P, ParserContext{
		FileName: "Function",
	})
	parameters := parameterSym.acceptFn(parser)

	parser = NewParser(bodyString, ParserContext{
		FileName: "Function",
	})
	body := bodySym.acceptFn(parser)
	body.Strict = body.FunctionBodyContainsUseStrict()

	parser = NewParser(sourceText, ParserContext{
		FileName: "Function",
	})
	_ = exprSym.acceptFn(parser)

	proto := GetPrototypeFromConstructor(newTarget, fallbackPrototype)

	env := realm.GlobalEnv

	var privateEnv *PrivateEnvironment

	function := OrdinaryFunctionCreate(
		agent,
		proto,
		sourceText,
		parameters,
		body,
		functionCreateThisModeNonLexical,
		env,
		privateEnv,
	)

	SetFunctionName(function, NewStringPropertyKey("anonymous"), "")
	switch kind {
	case dynamicFunctionKindNormal:
		MakeConstructor(function, true, nil)
	case dynamicFunctionKindGenerator:
		prototype := OrdinaryObjectCreate(agent, realm.Intrinsics.GeneratorFunctionPrototypePrototype, nil)
		function.defineBuiltinProperty(CMString("prototype"), &PropertyDescriptor{
			Value:        prototype.ToValue(),
			Writable:     true,
			Enumerable:   false,
			Configurable: false,
		})
	case dynamicFunctionKindAsyncGenerator:
		prototype := OrdinaryObjectCreate(agent, realm.Intrinsics.AsyncGeneratorFunctionPrototypePrototype, nil)
		function.defineBuiltinProperty(CMString("prototype"), &PropertyDescriptor{
			Value:        prototype.ToValue(),
			Writable:     true,
			Enumerable:   false,
			Configurable: false,
		})
	case dynamicFunctionKindAsync:
		panic("unimplemented")
	}
	return function
}

func NewFunctionConstructor(realm *Realm) ObjectType {
	// 20.2.1.1
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value] {
		agent := realm.Agent

		constructor := agent.ActiveFunctionObject()

		var parameters []Value
		var bodyArg Value = NewStringValue("")
		if len(argumentsList) > 0 {
			parameters = argumentsList[:len(argumentsList)-1]
			bodyArg = argumentsList[len(argumentsList)-1]
		}

		return CreateDynamicFunction(
			agent,
			constructor,
			newTarget,
			dynamicFunctionKindNormal,
			parameters,
			bodyArg,
		).ToValue()
	}
	f := CreateBuiltinFunction(realm.Agent, behavior, 1, CMString("Function"), builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		prefix:        "",
		isConstructor: true,
	})
	BindPrototypeAndConstructor(realm.Intrinsics.FunctionPrototype, f)
	return f
}
