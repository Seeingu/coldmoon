package coldmoon

import "github.com/samber/lo"

type FunctionPrototype struct {
	*Object
}

func (f *FunctionPrototype) ToObject() *Object {
	return f.Object
}

func NewFunctionPrototype(realm *Realm) ObjectType {
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		return UndefinedValue
	}
	f := CreateBuiltinFunction(realm.Agent, behavior, 0, "", builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.ObjectPrototype,
		prefix:        "",
		isConstructor: false,
	})

	return f
}

func InitFunctionMethods(f ObjectType, realm *Realm) {
	// 20.2.3.5 toString
	var toString = func(this Value, argumentsList []Value, newTarget ObjectType) Value {
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
	DefineBuiltinFunction(f, "toString", toString, 0, realm)

	var call = func(this Value, argumentsList []Value, newTarget ObjectType) Value {
		thisArg := argumentsList[0]
		args := argumentsList[1:]
		fun := this
		if !IsCallable(fun) {
			panic("TypeError")
		}
		return fun.CallAssumeCallable(thisArg, args)
	}
	DefineBuiltinFunction(f, "call", call, 1, realm)
}

type FunctionConstructor struct {
	*Object
}

type dynamicFunctionKind int

const (
	dynamicFunctionKindNormal dynamicFunctionKind = iota
	dynamicFunctionKindGenerator
	dynamicFunctionKindAsync
	dynamicFunctionKindAsyncGenerator
)

// 20.2.1.1.1
func CreateDynamicFunction(
	agent *Agent,
	constructor ObjectType,
	newTarget ObjectType,
	kind dynamicFunctionKind,
	parameterList ArgumentsList,
	bodyArg Value,
) ObjectType {
	realm := agent.CurrentRealm()

	currentRealm := realm

	agent.HostHooks.HostEnsureCanCompileStrings(currentRealm)

	if newTarget == nil {
		newTarget = constructor
	}

	var prefix string
	var fallbackPrototype string
	switch kind {
	case dynamicFunctionKindNormal:
		prefix = "function"
		fallbackPrototype = "%Function.prototype%"

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

	sourceString := prefix + " " + P + " " + bodyString

	sourceText := sourceString

	parameters := NewParser(P, ParserContext{
		FileName: "Function",
	}).formalParameters()

	body := NewParser(bodyString, ParserContext{
		FileName: "Function",
	}).functionBody()

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
		MakeConstructor(function, false, nil)
	default:
		panic("unimplemented")
	}
	return function
}

func NewFunctionConstructor(realm *Realm) ObjectType {
	// 20.2.1.1
	var behavior BehaviorFn = func(thisArgument Value, argumentsList []Value, newTarget ObjectType) Value {
		parameters := argumentsList[0 : len(argumentsList)-1]
		agent := realm.Agent

		constructor := agent.ActiveFunctionObject()

		bodyArg, ok := lo.Last(argumentsList)
		if !ok {
			bodyArg = NewStringValue("")
		}

		return NewValueFromObject(CreateDynamicFunction(
			agent,
			constructor,
			newTarget,
			dynamicFunctionKindNormal,
			parameters,
			bodyArg,
		))
	}
	f := CreateBuiltinFunction(realm.Agent, behavior, 1, "Function", builtinFunctionArgs{
		realm:         realm,
		prototype:     realm.Intrinsics.FunctionPrototype,
		prefix:        "",
		isConstructor: true,
	})
	return f
}
