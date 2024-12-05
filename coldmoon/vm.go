package coldmoon

import "C"
import (
	"fmt"
	"github.com/Seeingu/coldmoon/pkg"
	"math/big"
	"reflect"
	"strconv"
)

type VM struct {
	agent                    *Agent
	stack                    pkg.Stack[Value]
	result                   Value
	ip                       int
	reference                *ReferenceRecord
	referenceStack           pkg.Stack[*ReferenceRecord]
	exceptionJumpTargetStack pkg.Stack[int]
	exception                Value
	iterator                 *IteratorRecord
}

func NewVM(agent *Agent) *VM {
	return &VM{
		agent: agent,
	}
}

func (vm *VM) stackPush(v Value, m string) {
	if v == nil {
		return
	}
	if o, ok := vm.result.(*ObjectValue); ok {
		if oo, ok := o.Object.(*Object); ok && oo == nil {
			panic("TypeError: Cannot push undefined")
		}
	}
	vm.stack.Push(v)
	vm.debugPrintStack(m)
}

// stackPop pops a value from the stack.
// if the stack is empty, it returns undefined.
func (vm *VM) stackPop() Value {
	if vm.stack.IsEmpty() {
		if Debug.PrintBytecode {
			fmt.Printf("VM stack is empty, ip: %d\n", vm.ip)
		}
		return UndefinedValue
	}
	defer vm.debugPrintStack("stackPop")
	return vm.stack.Pop()
}

func (vm *VM) execute(executable *Executable, i Instruction) {
	agent := vm.agent
	switch ins := i.(type) {
	case *ILoad:
		vm.stackPush(vm.result, "ILoad")
	case *ILoadConstant:
		vm.stackPush(ins.Value, "ILoadConstant")
	case *IStore:
		vm.result = vm.stackPop()
	case *IStoreConstant:
		vm.result = ins.Value
	case *IResolveBinding:
		vm.reference = vm.agent.ResolveBinding(string(ins.Name), nil, ins.Strict)
	case *ICall:
		argumentCount := ins.ArgumentCount
		arguments := make([]Value, argumentCount)
		strict := ins.Strict
		for i := argumentCount - 1; i >= 0; i-- {
			arguments[i] = vm.stackPop()
		}
		this := vm.stackPop()
		function := vm.stackPop()

		realm := vm.agent.CurrentRealm()
		eval := realm.Intrinsics.Eval

		if !vm.referenceStack.IsEmpty() {
			ref := vm.referenceStack.Peek()
			refName, ok :=
				ref.ReferencedName.(*ReferencedNameString)
			if ref.IsPropertyReference() &&
				ok &&
				refName.String == "eval" &&
				pkg.FuncEqual(function.(*ObjectValue).Object, eval) {
				vm.result = directEval(vm.agent, arguments, strict)
				return
			}
		}

		vm.result = evaluateCall(
			vm.agent,
			function,
			this,
			arguments,
		)
	case *ILoadThisValue:
		this := evaluateCallGetThisValue(vm.referenceStack.Peek())
		vm.stackPush(this, "ILoadThisValue")
	case *ILoadThisValueSuper:
		env := agent.GetThisEnvironment()
		actualThis := env.GetThisBinding()
		vm.stackPush(actualThis, "ILoadThisValueSuper")
	case *IMakeSuperPropertyReference:
		propertyNameValue := vm.stackPop()
		strict := ins.Strict
		actualThis := vm.stackPop()
		propertyKey := ToPropertyKey(vm.agent, propertyNameValue)
		env := agent.GetThisEnvironment()
		Assert(env.HasSuperBinding())
		baseValue := env.(*FunctionEnvironment).GetSuperBase()
		vm.reference = &ReferenceRecord{
			Base: &ReferenceRecordBaseValue{
				Value: baseValue,
			},
			ReferencedName: propertyKey.ToReference(),
			Strict:         strict,
			ThisValue:      actualThis,
		}

	case *IResolveThisBinding:
		vm.result = vm.agent.ResolveThisBinding()
	case *IReturn:
		return
	case *IJump:
		vm.ip = ins.Target
	case *IJumpIfTrue:
		value := vm.result
		if value != nil && value.ToBoolean() {
			vm.ip = ins.Target
		} else {
			vm.ip = ins.TargetElse
		}
	case *IThrow:
		value := vm.result
		vm.agent.exception = value
		panic("CompletionTypeThrow")
	case *IInstantiateOrdinaryFunctionExpression:
		functionExpression := ins.FunctionExpression
		closure := InstantiateOrdinaryFunctionExpression(
			vm.agent,
			functionExpression,
			"",
		)
		vm.result = NewValueFromObject(closure)
	case *IInstantiateArrowFunctionExpression:
		arrowFunction := ins.FunctionExpression
		closure := InstantiateArrowFunctionExpression(
			vm.agent,
			arrowFunction,
			"",
		)
		vm.result = NewValueFromObject(closure)
	case *ITypeof:
		if vm.reference != nil {
			if vm.reference.IsUnresolvableReference() {
				vm.result = NewStringValue("undefined")
				return
			}
		}
		var value = vm.result
		if vm.reference != nil {
			value = vm.reference.GetValue(agent)
		}

		switch v := value.(type) {
		case *undefinedValue:
			vm.result = NewStringValue("undefined")
		case *nullValue:
			vm.result = NewStringValue("object")
		case *BooleanValue:
			vm.result = NewStringValue("boolean")
		case *NumberValue:
			vm.result = NewStringValue("number")
		case *StringValue:
			vm.result = NewStringValue("string")
		case *SymbolValue:
			vm.result = NewStringValue("symbol")
		case *BigIntValue:
			vm.result = NewStringValue("bigint")
		case *ObjectValue:
			if v.Object.InternalMethods().Call != nil {
				vm.result = NewStringValue("function")
			} else {
				vm.result = NewStringValue("object")
			}
		default:
			panic("unreachable")
		}

	case *IToNumber:
		value := vm.result
		vm.result = ToNumber(vm.agent, value)
	case *IToNumeric:
		value := vm.result
		vm.result = ToNumeric(vm.agent, value)
	case *IUnaryMinus:
		value := vm.result
		switch v := value.(type) {
		case *BigIntValue:
			vm.result = v.UnaryMinus()
		case *NumberValue:
			vm.result = v.UnaryMinus()
		default:
			panic("unreachable")
		}
	case *ILogicalNot:
		value := vm.result
		vm.result = NewBooleanValue(!value.ToBoolean())
	case *IObjectCreate:
		object := OrdinaryObjectCreate(vm.agent, vm.agent.CurrentRealm().Intrinsics.ObjectPrototype, nil)
		vm.result = NewValueFromObject(object)
	case *IObjectSetProperty:
		value := vm.stackPop()
		propertyKey := ToPropertyKey(vm.agent, vm.stackPop())
		object := vm.stackPop().(*ObjectValue).Object
		object.CreateDataPropertyOrThrow(propertyKey, value)
		vm.result = NewValueFromObject(object)
	case *IBitwiseNot:
		value := vm.result
		switch v := value.(type) {
		case *BigIntValue:
			vm.result = v.BitwiseNot()
		case *NumberValue:
			vm.result = v.BitwiseNot()
		default:
			panic("unreachable")
		}
	case *IEvaluatePropertyAccessWithExpressionKey:
		// 13.3.3
		propertyNameValue := vm.stackPop()
		strict := ins.Strict
		baseValue := vm.stackPop()
		Assert(baseValue != nil)
		propertyKey := ToPropertyKey(vm.agent, propertyNameValue)

		var referencedName ReferencedName
		switch p := propertyKey.(type) {
		case StringPropertyKey:
			referencedName = &ReferencedNameString{
				String: p.Value,
			}
		case SymbolPropertyKey:
			referencedName = &ReferencedNameSymbol{
				Symbol: p.Value,
			}
		case IntegerIndexPropertyKey:
			referencedName = &ReferencedNameString{
				String: strconv.Itoa(int(p.Value)),
			}
		}
		vm.reference = &ReferenceRecord{
			Base: &ReferenceRecordBaseValue{
				Value: baseValue,
			},
			ReferencedName: referencedName,
			Strict:         strict,
			ThisValue:      nil,
		}
	case *IEvaluatePropertyAccessWithIdentifierKey:
		// 13.3.4
		propertyNameString := ins.Name
		strict := ins.Strict
		baseValue := vm.stackPop()
		referencedName := &ReferencedNameString{
			String: string(propertyNameString),
		}
		vm.reference = &ReferenceRecord{
			Base: &ReferenceRecordBaseValue{
				Value: baseValue,
			},
			ReferencedName: referencedName,
			Strict:         strict,
			ThisValue:      nil,
		}
	case *IGetValue:
		if vm.reference != nil {
			vm.result = vm.reference.GetValue(agent)
		}
		vm.reference = nil
	case *IArrayCreate:
		vm.result = NewValueFromObject(ArrayCreate(vm.agent, 0, nil))
	case *IArraySetLength:
		length := JSInt(ins.Length)
		array := vm.result.(*ObjectValue).Object
		array.Set(NewStringPropertyKey("length"), NewNumberValue(length.ToNumber()), setThrowTypeThrow)
	case *IArraySetValue:
		index := JSInt(ins.Index)
		initValue := vm.stackPop()
		array := vm.stackPop().(*ObjectValue).Object
		array.CreateDataPropertyOrThrow(
			NewIntegerIndexPropertyKey(index),
			initValue,
		)
		vm.result = NewValueFromObject(array)
	case *IArrayPushValue:
		initValue := vm.stackPop()
		arrayValue := vm.stackPop()
		array := arrayValue.(*ObjectValue).Object
		length, _ := ValueGetLength(arrayValue)
		array.CreateDataPropertyOrThrow(
			NewIntegerIndexPropertyKey(length),
			initValue,
		)
		vm.result = NewValueFromObject(array)
	case *IArraySpread:
		spread := vm.stackPop()
		arrayValue := vm.stackPop()
		array := MustGetObject(arrayValue)
		iteratorRecord := GetIterator(vm.agent, spread, IteratorKindSync).Data()
		nextIndex, _ := ValueGetLength(arrayValue)
		for {
			next := iteratorRecord.IteratorStep()
			if next == nil {
				break
			}
			nextValue := IteratorValue(next)
			array.CreateDataPropertyOrThrow(
				NewIntegerIndexPropertyKey(nextIndex),
				nextValue,
			)
			nextIndex++
		}
		vm.result = NewValueFromObject(array)
	case *IGreaterThan:
		right := vm.stackPop()
		left := vm.stackPop()
		r := IsLessThan(vm.agent, left, right, IsLessThanOrderRightFirst)
		vm.result = NewBooleanValue(r)
	case *IGreaterThanEquals:
		right := vm.stackPop()
		left := vm.stackPop()
		r := IsLessThan(vm.agent, left, right, IsLessThanOrderRightFirst)
		vm.result = NewBooleanValue(!r)
	case *ILessThan:
		right := vm.stackPop()
		left := vm.stackPop()
		r := IsLessThan(vm.agent, left, right, IsLessThanOrderLeftFirst)
		vm.result = NewBooleanValue(r)
	case *ILessThanEquals:
		right := vm.stackPop()
		left := vm.stackPop()
		r := IsLessThan(vm.agent, left, right, IsLessThanOrderLeftFirst)
		vm.result = NewBooleanValue(!r)
	case *IHasProperty:
		right := vm.stackPop()
		left := vm.stackPop()
		rightObject, ok := right.(*ObjectValue)
		if !ok {
			panic("TypeError: right is not an object")
		}
		vm.result = NewBooleanValue(
			rightObject.Object.HasProperty(ToPropertyKey(vm.agent, left)))
	case *IInstanceOf:
		right := vm.stackPop()
		left := vm.stackPop()
		vm.result = NewBooleanValue(
			InstanceOfOperator(vm.agent, left, right),
		)
	case *ILooselyEqual:
		right := vm.stackPop()
		left := vm.stackPop()
		vm.result = NewBooleanValue(IsLooselyEqual(vm.agent, right, left))
	case *IStrictlyEqual:
		right := vm.stackPop()
		left := vm.stackPop()
		vm.result = NewBooleanValue(IsStrictlyEqual(right, left))
	case *IApplyStringOrNumericBinaryOperator:
		right := vm.stackPop()
		left := vm.stackPop()
		operator := ins.Operator
		vm.result = applyStringOrNumericBinaryOperator(
			vm.agent, left, right, operator,
		)
	case *INew:
		argumentCount := ins.ArgumentCount
		arguments := make([]Value, argumentCount)
		for i := argumentCount - 1; i >= 0; i-- {
			arguments[i] = vm.stackPop()
		}
		constructor := vm.stackPop()
		vm.result = evaluateNew(vm.agent, constructor, arguments)
	case *IPushExceptionJumpTarget:
		jumpTarget := ins.Target
		vm.exceptionJumpTargetStack.Push(jumpTarget)
	case *IPopExceptionJumpTarget:
		vm.exceptionJumpTargetStack.Pop()
	case *IRethrowExceptionIfAny:
		if vm.exception != nil {
			vm.agent.exception = vm.exception
			panic("CompletionTypeThrow")
		}
	case *IPushReference:
		Assert(vm.reference != nil)
		vm.referenceStack.Push(vm.reference)
	case *IPopReference:
		vm.referenceStack.Pop()
	case *IPutValue:
		lref := vm.referenceStack.Peek()
		rval := vm.result
		lref.PutValue(vm.agent, rval)
		vm.reference = nil
	case *ICreateCatchBinding:
		name := ins.IdentifierName
		thrownValue := vm.exception
		vm.exception = nil
		catchEnv := vm.agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		catchEnv.CreateMutableBinding(string(name), false)
		catchEnv.InitializeBinding(string(name), thrownValue)
	case *IDelete:
		ref := vm.reference
		if ref.IsUnresolvableReference() {
			Assert(!ref.Strict)
			vm.result = NewBooleanValue(true)
			return
		}
		if ref.IsPropertyReference() {
			Assert(!ref.IsPrivateReference())

			if ref.IsSuperReference() {
				panic("ReferenceError: cannot delete super")
			}

			baseObj := ValueToObject(agent, ref.Base.(*ReferenceRecordBaseValue).Value)
			var referencedName PropertyKey
			switch refName := ref.ReferencedName.(type) {
			case *ReferencedNameString:
				referencedName = NewStringPropertyKey(refName.String)
			case *ReferencedNameSymbol:
				referencedName = NewSymbolPropertyKey(refName.Symbol)
			default:
				panic("unreachable")
			}
			deleteStatus := baseObj.InternalMethods().Delete(baseObj, referencedName)
			if !deleteStatus && ref.Strict {
				panic("TypeError: cannot delete property")
			}
			vm.result = NewBooleanValue(deleteStatus)
		} else {
			base := ref.Base.(*ReferenceRecordBaseEnvironment)
			referencedName := ref.ReferencedName.(*ReferencedNameString).String
			deleteStatus := base.Environment.DeleteBinding(referencedName)
			vm.result = NewBooleanValue(deleteStatus)
		}
	case *IIncrement:
		value := vm.result
		switch v := value.(type) {
		case *BigIntValue:
			vm.result = v.Add(NewBigIntValue(big.NewInt(1)))
		case *NumberValue:
			vm.result = v.Add(NewNumberValue(1))
		default:
			panic("unreachable")
		}
	case *IDecrement:
		value := vm.result
		switch v := value.(type) {
		case *BigIntValue:
			vm.result = v.Subtract(NewBigIntValue(big.NewInt(1)))
		case *NumberValue:
			vm.result = v.Subtract(NewNumberValue(1))
		default:
			panic("unreachable")
		}
	case *IObjectDefineMethod:
		propertyName := vm.stack.Pop()
		object := MustGetObject(vm.stack.Pop())
		vm.MethodDefinitionEvaluation(methodDefinitionArgs{
			PropertyName:             propertyName,
			MethodType:               MethodDefinitionTypeMethod,
			FunctionExpression:       ins.FunctionExpression,
			GeneratorExpression:      ins.GeneratorExpression,
			AsyncFunctionExpression:  ins.AsyncFunctionExpression,
			AsyncGeneratorExpression: ins.AsyncGeneratorExpression,
		}, object, true)
	case *IObjectSpreadValue:
		fromValue := vm.stack.Pop()
		toValue := vm.stack.Pop()
		object := MustGetObject(toValue)
		var excludedNames []PropertyKey
		object.CopyDataProperties(fromValue, excludedNames)
		vm.result = NewValueFromObject(object)
	case *IInstantiateGeneratorFunctionExpression:
		functionExpression := ins.FunctionExpression
		closure := vm.InstantiateGeneratorFunctionExpression(functionExpression)
		vm.result = NewValueFromObject(closure)
	case *IInstantiateAsyncGeneratorFunctionExpression:
		functionExpression := ins.FunctionExpression
		closure := vm.InstantiateAsyncGeneratorFunctionExpression(functionExpression)
		vm.result = NewValueFromObject(closure)
	case *IInstantiateAsyncFunctionExpression:
		functionExpression := ins.FunctionExpression
		closure := vm.InstantiateAsyncFunctionExpression(functionExpression)
		vm.result = NewValueFromObject(closure)
	case *IGetNewTarget:
		t := vm.agent.GetNewTarget()
		if t != nil {
			vm.result = NewValueFromObject(t)
		} else {
			vm.result = UndefinedValue
		}
	case *IInstantiateAsyncArrowFunctionExpression:
		functionExpression := ins.FunctionExpression
		closure := vm.InstantiateAsyncArrowFunctionExpression(functionExpression, "")
		vm.result = NewValueFromObject(closure)
	case *IRegExpCreate:
		flags := vm.stack.Pop()
		pattern := vm.stack.Pop()
		vm.result = RegExpCreate(vm.agent, pattern, flags).Data().ToValue()
	case *IBindingClassDeclarationEvaluation:
		classDeclaration := ins.ClassDeclaration
		vm.result = vm.BindingClassDeclarationEvaluation(classDeclaration).ToValue()
	case *IClassDefinitionEvaluation:
		classExpression := ins.ClassExpression
		if classExpression.IdentifierName != "" {
			className := string(classExpression.IdentifierName)
			value := vm.ClassDefinitionEvaluation(classExpression.ClassTail, className, className)
			if f, ok := value.(*ECMAScriptFunction); ok {
				f.SourceText = classExpression.SourceText
			} else if b, ok := value.(*BuiltinFunction); ok {
				b.AdditionalFields.ClassConstructorFields.SourceText = classExpression.SourceText
			} else {
				panic("unreachable")
			}
			vm.result = value.ToValue()
		} else {
			value := vm.ClassDefinitionEvaluation(classExpression.ClassTail, "", "")
			if f, ok := value.(*ECMAScriptFunction); ok {
				f.SourceText = classExpression.SourceText
			} else if b, ok := value.(*BuiltinFunction); ok {
				b.AdditionalFields.ClassConstructorFields.SourceText = classExpression.SourceText
			} else {
				panic("unreachable")
			}
			vm.result = value.ToValue()
		}
	case *IEvaluateSuperCall:
		argumentCount := ins.ArgumentCount
		newTarget := agent.GetNewTarget()
		fun := vm.getSuperConstructor()
		arguments := make([]Value, argumentCount)
		for i := argumentCount - 1; i >= 0; i-- {
			arguments[i] = vm.stack.Pop()
		}
		if !IsConstructor(fun) {
			panic("TypeError: fun is not a constructor")
		}
		result := MustGetObject(fun).Construct(arguments, newTarget)
		thisER := agent.GetThisEnvironment()
		thisER.(*FunctionEnvironment).BindThisValue(result.ToValue())
		F := thisER.(*FunctionEnvironment).FunctionObject
		result.InitializeInstanceElements(F)
		vm.result = result.ToValue()
	case *IGetOrCreateImportMeta:
		module := agent.GetActiveScriptOrModule().(*SourceTextModule)
		if module.ImportMeta == nil {
			importMeta := OrdinaryObjectCreate(agent, nil, nil)
			importMetaValues := agent.HostHooks.HostGetImportMetaProperties(module)
			for k, v := range importMetaValues {
				importMeta.CreateDataPropertyOrThrow(k, v)
			}

			agent.HostHooks.HostFinalizeImportMeta(importMeta, module)
			module.ImportMeta = importMeta
			vm.result = importMeta.ToValue()
		} else {
			vm.result = module.ImportMeta.ToValue()
		}
	case *IImportCall:
		realm := agent.CurrentRealm()
		var referrer ImportedModuleReferrer
		if a := agent.GetActiveScriptOrModule(); a != nil {
			referrer = a.toReferrer()
		} else {
			referrer = realm.ToReferrer()
		}

		specifier := vm.stack.Pop()
		promiseCapability := NewPromiseCapability(agent, realm.Intrinsics.Promise.ToValue())
		specifierString := ToString(agent, specifier)

		agent.HostHooks.HostLoadImportedModule(
			agent,
			referrer,
			specifierString.Data,
			nil,
			promiseCapability.ToImportedModulePayload(),
		)
		vm.result = promiseCapability.Promise.ToValue()
	case *IForInIterator:
		value := vm.result
		obj := MustGetObject(value)
		iterator := CreateForInIterator(agent, obj)
		nextMethod := iterator.Get(NewStringPropertyKey("next"))
		vm.iterator = &IteratorRecord{
			Iterator:   iterator,
			NextMethod: nextMethod,
		}
	case *IGetIterator:
		vm.iterator = GetIterator(agent, vm.result, ins.IteratorKind).Data()
	case *ILoadIterator:
		vm.stackPush(vm.iterator.NextMethod, "ILoadIterator")
		vm.stackPush(NewValueFromObject(vm.iterator.Iterator), "ILoadIterator")
	}
}

func (vm *VM) getSuperConstructor() Value {
	agent := vm.agent
	envRec := agent.GetThisEnvironment()
	funEnv := envRec.(*FunctionEnvironment)
	activeFunction := funEnv.FunctionObject
	superConstructor := activeFunction.InternalMethods().GetPrototypeOf(activeFunction)
	return superConstructor.ToValue()
}

// 15.7.3
func (vm *VM) ClassElementEvaluation(classElement ClassElement, object ObjectType) (classFieldDefinition *ClassFieldDefinition, staticBlockDefinition *ClassStaticBlockDefinition, err error) {
	switch ce := classElement.(type) {
	case *ClassElementStaticBlock:
		staticBlockDefinition = vm.ClassStaticBlockDefinitionEvaluation(ce, object)
		return
	case *ClassElementFieldDefinition, *ClassElementStaticFieldDefinition:
		var fieldDefinition *FieldDefinition
		if c, ok := ce.(*ClassElementFieldDefinition); ok {
			fieldDefinition = c.FieldDefinition
		} else {
			fieldDefinition = ce.(*ClassElementStaticFieldDefinition).FieldDefinition
		}
		classFieldDefinition = vm.ClassFieldDefinitionEvaluation(fieldDefinition, object)
		return
	case *ClassElementStaticMethodDefinition, *ClassElementMethodDefinition:
		var methodDefinition *MethodDefinition
		if c, ok := ce.(*ClassElementStaticMethodDefinition); ok {
			methodDefinition = c.MethodDefinition
		} else {
			methodDefinition = ce.(*ClassElementMethodDefinition).MethodDefinition
		}
		propertyName := GenerateAndRunBytecode(vm.agent, methodDefinition.PropertyName)
		vm.MethodDefinitionEvaluation(methodDefinitionArgs{
			PropertyName:       propertyName.Data(),
			FunctionExpression: methodDefinition.FunctionExpression,
			MethodType:         methodDefinition.Type,
		}, object, true)
		return
	case *ClassElementEmpty:
		return
	}
	panic("unreachable")
}

// 15.7.14
func (vm *VM) ClassDefinitionEvaluation(classTail *ClassTail, classBinding string, className string) ObjectType {
	agent := vm.agent
	realm := agent.CurrentRealm()
	env := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
	classEnv := NewDeclarativeEnvironment(env)
	if classBinding != "" {
		classEnv.CreateImmutableBinding(classBinding, true)
	}

	outerPrivateEnvironment := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
	classPrivateEnvironment := NewPrivateEnvironment(outerPrivateEnvironment)

	if len(classTail.ClassBody.ClassElementList.Items) > 0 {
		// TODO
	}

	var protoParent ObjectType
	var constructorParent ObjectType
	if classTail.ClassHeritage == nil {
		protoParent = realm.Intrinsics.ObjectPrototype
		constructorParent = realm.Intrinsics.FunctionPrototype
	} else {
		agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment = classEnv
		superclassRef := GenerateAndRunBytecode(agent, classTail.ClassHeritage)
		agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment = env
		superclass := superclassRef.Data()
		if superclass == nil {
			protoParent = nil
			constructorParent = realm.Intrinsics.FunctionPrototype
		} else if !IsConstructor(superclass) {
			panic("TypeError: superclass is not a constructor")
		} else {
			protoParentValue := MustGetObject(superclass).Get(NewStringPropertyKey("prototype"))
			if !ValueIsObject(protoParentValue) {
				panic("TypeError: prototype is not an object")
			}
			protoParent = MustGetObject(protoParentValue)
			constructorParent = MustGetObject(superclass)
		}
	}

	proto := OrdinaryObjectCreate(agent, protoParent, nil)
	constructor := classTail.ClassBody.ConstructorMethod()
	agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment = classEnv
	agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment = classPrivateEnvironment

	var function ObjectType
	if constructor == nil {
		var defaultConstructor BehaviorFn = func(this Value, arguments []Value, newTarget ObjectType) Value {
			args := arguments
			if newTarget == nil {
				panic("TypeError: newTarget is nil")
			}

			F := agent.ActiveFunctionObject()
			classConstructorFields := F.(*BuiltinFunction).AdditionalFields.ClassConstructorFields
			var result ObjectType
			if classConstructorFields.ConstructorKind == ConstructorKindDerived {
				fun := function.InternalMethods().GetPrototypeOf(function)
				if !IsConstructor(fun.ToValue()) {
					panic("TypeError: prototype is not a constructor")
				}
				result = fun.Construct(args, newTarget)
			} else {
				result = OrdinaryCreateFromConstructor(agent, newTarget, "%Object.prototype", nil)
			}
			return result.ToValue()
		}

		function = CreateBuiltinFunction(
			agent,
			defaultConstructor,
			0,
			"constructor",
			builtinFunctionArgs{
				prototype:     constructorParent,
				realm:         realm,
				isConstructor: true,
				additionalFields: &AdditionalFields{
					ClassConstructorFields: &ClassConstructorFields{},
				},
			})
	} else {
		constructorInfo := DefineMethod(agent, constructor.FunctionExpression, NewStringValue("constructor"), proto, constructorParent)
		F := constructorInfo.Closure
		MakeClassConstructor(F.(*ECMAScriptFunction))
		SetFunctionName(F, NewStringPropertyKey(className), "")
		function = F
	}

	MakeConstructor(function, false, proto)
	if classTail.ClassHeritage != nil {
		if f, ok := function.(*ECMAScriptFunction); ok {
			f.ConstructorKind = ConstructorKindDerived
		} else if b, ok := function.(*BuiltinFunction); ok {
			b.AdditionalFields.ClassConstructorFields.ConstructorKind = ConstructorKindDerived
		} else {
			panic("unreachable")
		}
	}

	DefineMethodProperty(proto, NewStringPropertyKey("constructor"), function, false)

	elements := classTail.ClassBody.NonConstructorElements()

	instanceFields := make([]*ClassFieldDefinition, 0)

	staticClassFields := make([]*ClassFieldDefinition, 0)
	staticStaticBlocks := make([]*ClassStaticBlockDefinition, 0)

	for _, classElement := range elements {
		var classFieldDefinition *ClassFieldDefinition
		var staticBlockDefinition *ClassStaticBlockDefinition
		var err error
		if ClassElementIsStatic(classElement) {
			classFieldDefinition, staticBlockDefinition, err = vm.ClassElementEvaluation(classElement, proto)
		} else {
			classFieldDefinition, staticBlockDefinition, err = vm.ClassElementEvaluation(classElement, function)
		}
		if err != nil {
			agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment = env
			agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment = outerPrivateEnvironment
		} else if classFieldDefinition != nil {
			if !ClassElementIsStatic(classElement) {
				instanceFields = append(instanceFields, classFieldDefinition)
			} else {
				staticClassFields = append(staticClassFields, classFieldDefinition)
			}
		} else if staticBlockDefinition != nil {
			staticStaticBlocks = append(staticStaticBlocks, staticBlockDefinition)
		} else {
			panic("unreachable")
		}
	}

	agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment = env
	if classBinding != "" {
		classEnv.InitializeBinding(classBinding, function.ToValue())
	}

	function.(InternalSlotFields).SetFields(instanceFields)

	for _, element := range staticClassFields {
		function.DefineField(element)
	}
	for _, block := range staticStaticBlocks {
		CallAssumeCallableNoArgs(block.BodyFunction.ToValue(), function.ToValue())
	}

	agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment = outerPrivateEnvironment

	return function
}

// 15.7.15
func (vm *VM) BindingClassDeclarationEvaluation(classDeclaration *DeclarationClass) ObjectType {
	agent := vm.agent
	if classDeclaration.IdentifierName != "" {
		className := string(classDeclaration.IdentifierName)
		value := vm.ClassDefinitionEvaluation(classDeclaration.ClassTail, className, className)
		if f, ok := value.(*ECMAScriptFunction); ok {
			f.SourceText = classDeclaration.SourceText
		} else if b, ok := value.(*BuiltinFunction); ok {
			b.AdditionalFields.ClassConstructorFields.SourceText = classDeclaration.SourceText
		} else {
			panic("unreachable")
		}

		env := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		vm.InitializeBoundName(className, value.ToValue(), env)
		return value
	} else {
		value := vm.ClassDefinitionEvaluation(classDeclaration.ClassTail, "", "default")
		if f, ok := value.(*ECMAScriptFunction); ok {
			f.SourceText = classDeclaration.SourceText
		} else if b, ok := value.(*BuiltinFunction); ok {
			b.AdditionalFields.ClassConstructorFields.SourceText = classDeclaration.SourceText
		} else {
			panic("unreachable")
		}
		return value
	}
}

func (vm *VM) InitializeBoundName(name string, value Value, env EnvironmentRecord) {
	if env == nil {
		lhs := vm.agent.ResolveBinding(name, nil, true)
		lhs.PutValue(vm.agent, value)
	} else {
		env.InitializeBinding(name, value)
	}
}

func (vm *VM) InstantiateAsyncArrowFunctionExpression(functionExpression *PrimaryExpressionAsyncArrowFunction, name string) ObjectType {
	realm := vm.agent.CurrentRealm()
	env := vm.agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
	privateEnv := vm.agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
	sourceText := functionExpression.SourceText
	closure := OrdinaryFunctionCreate(
		vm.agent,
		realm.Intrinsics.AsyncFunctionPrototype,
		sourceText,
		functionExpression.FormalParameters,
		functionExpression.Body,
		functionCreateThisModeLexical,
		env,
		privateEnv,
	)
	SetFunctionName(closure, NewStringPropertyKey(name), "")
	return closure
}

func (vm *VM) InstantiateAsyncFunctionExpression(functionExpression *PrimaryExpressionAsyncFunctionExpression) ObjectType {
	realm := vm.agent.CurrentRealm()
	if functionExpression.Identifier != "" {
		name := string(functionExpression.Identifier)
		outerEnv := vm.agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		funcEnv := NewDeclarativeEnvironment(outerEnv)
		funcEnv.CreateImmutableBinding(name, false)
		privateEnv := vm.agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := functionExpression.SourceText
		closure := OrdinaryFunctionCreate(
			vm.agent,
			realm.Intrinsics.AsyncFunctionPrototype,
			sourceText,
			functionExpression.FormalParameters,
			functionExpression.Body,
			functionCreateThisModeNonLexical,
			funcEnv,
			privateEnv,
		)
		SetFunctionName(closure, NewStringPropertyKey(name), "")
		funcEnv.InitializeBinding(name, NewValueFromObject(closure))
		return closure
	} else {
		env := vm.agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := vm.agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := functionExpression.SourceText
		closure := OrdinaryFunctionCreate(
			vm.agent,
			realm.Intrinsics.AsyncFunctionPrototype,
			sourceText,
			functionExpression.FormalParameters,
			functionExpression.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		SetFunctionName(closure, NewStringPropertyKey(""), "")
		return closure
	}
}

func (vm *VM) InstantiateAsyncGeneratorFunctionExpression(functionExpression *PrimaryExpressionAsyncGeneratorExpression) ObjectType {
	realm := vm.agent.CurrentRealm()
	if functionExpression.IdentifierName != "" {
		name := string(functionExpression.IdentifierName)
		outerEnv := vm.agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		funcEnv := NewDeclarativeEnvironment(outerEnv)
		funcEnv.CreateImmutableBinding(name, false)
		privateEnv := vm.agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := functionExpression.SourceText
		closure := OrdinaryFunctionCreate(
			vm.agent,
			realm.Intrinsics.AsyncGeneratorFunctionPrototype,
			sourceText,
			functionExpression.FormalParameters,
			functionExpression.Body,
			functionCreateThisModeNonLexical,
			funcEnv,
			privateEnv,
		)
		SetFunctionName(closure, NewStringPropertyKey(name), "")
		prototype := OrdinaryObjectCreate(vm.agent, realm.Intrinsics.AsyncGeneratorFunctionPrototypePrototype, nil)

		closure.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
			Value:        NewValueFromObject(prototype),
			Writable:     true,
			Enumerable:   false,
			Configurable: false,
		})

		funcEnv.InitializeBinding(name, NewValueFromObject(closure))
		return closure
	} else {
		env := vm.agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := vm.agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := functionExpression.SourceText
		closure := OrdinaryFunctionCreate(
			vm.agent,
			realm.Intrinsics.AsyncGeneratorFunctionPrototype,
			sourceText,
			functionExpression.FormalParameters,
			functionExpression.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		SetFunctionName(closure, NewStringPropertyKey(""), "")
		prototype := OrdinaryObjectCreate(vm.agent, realm.Intrinsics.AsyncGeneratorFunctionPrototypePrototype, nil)

		closure.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
			Value:        NewValueFromObject(prototype),
			Writable:     true,
			Enumerable:   false,
			Configurable: false,
		})
		return closure
	}
}

func (vm *VM) ClassStaticBlockDefinitionEvaluation(classStaticBlock *ClassElementStaticBlock, homeObject ObjectType) *ClassStaticBlockDefinition {
	agent := vm.agent
	realm := agent.CurrentRealm()
	lex := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
	privateEnv := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
	sourceText := ""
	formalParameters := &FormalParameters{}
	var bodyFunction *ECMAScriptFunction
	functionBody := &FunctionBody{
		StatementList: classStaticBlock.StatementList,
		Strict:        true,
		Type:          FunctionTypeNormal,
	}
	bodyFunction = OrdinaryFunctionCreate(
		agent,
		realm.Intrinsics.FunctionPrototype,
		sourceText,
		formalParameters,
		functionBody,
		functionCreateThisModeNonLexical,
		lex,
		privateEnv,
	)
	MakeMethod(bodyFunction, homeObject)
	return &ClassStaticBlockDefinition{
		BodyFunction: bodyFunction,
	}
}

func (vm *VM) ClassFieldDefinitionEvaluation(fieldDefinition *FieldDefinition, homeObject ObjectType) *ClassFieldDefinition {
	agent := vm.agent
	realm := agent.CurrentRealm()
	var name PropertyKeyOrPrivateName
	value := GenerateAndRunBytecode(agent, fieldDefinition.PropertyName)
	if value.Data() != nil {
		name = ToPropertyKey(agent, value.Data())
	}
	var initializer ObjectType
	if fieldDefinition.Initializer != nil {
		formalParameterList := &FormalParameters{}
		env := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		var sourceText = ""
		functionBody := &FunctionBody{
			StatementList: StatementList{
				&StatementListItemStatement{
					Statement: &StatementReturn{
						Expression: fieldDefinition.Initializer,
					},
				},
			},
			Strict: true,
			Type:   FunctionTypeNormal,
		}
		initializer = OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.FunctionPrototype,
			sourceText,
			formalParameterList,
			functionBody,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		MakeMethod(initializer.(*ECMAScriptFunction), homeObject)
		initializer.(InternalSlotClassFieldInitializerName).SetClassFieldInitializerName(name)
	} else {
		return &ClassFieldDefinition{
			Name: name,
		}
	}

	return &ClassFieldDefinition{
		Name:        name,
		Initializer: initializer.(*ECMAScriptFunction),
	}
}

func (vm *VM) InstantiateGeneratorFunctionExpression(functionExpression *PrimaryExpressionGeneratorExpression) ObjectType {
	realm := vm.agent.CurrentRealm()
	if functionExpression.IdentifierName != "" {
		name := string(functionExpression.IdentifierName)
		outerEnv := vm.agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		funcEnv := NewDeclarativeEnvironment(outerEnv)
		funcEnv.CreateImmutableBinding(name, false)
		privateEnv := vm.agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := functionExpression.SourceText
		closure := OrdinaryFunctionCreate(
			vm.agent,
			realm.Intrinsics.GeneratorFunctionPrototype,
			sourceText,
			functionExpression.FormalParameters,
			functionExpression.Body,
			functionCreateThisModeNonLexical,
			funcEnv,
			privateEnv,
		)
		SetFunctionName(closure, NewStringPropertyKey(name), "")
		prototype := OrdinaryObjectCreate(vm.agent, realm.Intrinsics.GeneratorFunctionPrototypePrototype, nil)

		closure.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
			Value:        NewValueFromObject(prototype),
			Writable:     true,
			Enumerable:   false,
			Configurable: false,
		})

		funcEnv.InitializeBinding(name, NewValueFromObject(closure))
		return closure
	} else {
		env := vm.agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := vm.agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := functionExpression.SourceText
		closure := OrdinaryFunctionCreate(
			vm.agent,
			realm.Intrinsics.GeneratorFunctionPrototype,
			sourceText,
			functionExpression.FormalParameters,
			functionExpression.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		SetFunctionName(closure, NewStringPropertyKey(""), "")
		prototype := OrdinaryObjectCreate(vm.agent, realm.Intrinsics.GeneratorFunctionPrototypePrototype, nil)

		closure.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
			Value:        NewValueFromObject(prototype),
			Writable:     true,
			Enumerable:   false,
			Configurable: false,
		})
		return closure
	}

}

type methodDefinitionArgs struct {
	PropertyName             Value
	MethodType               MethodDefinitionType
	FunctionExpression       *PrimaryExpressionFunctionExpression
	GeneratorExpression      *PrimaryExpressionGeneratorExpression
	AsyncFunctionExpression  *PrimaryExpressionAsyncFunctionExpression
	AsyncGeneratorExpression *PrimaryExpressionAsyncGeneratorExpression
}

// 15.4.5
func (vm *VM) MethodDefinitionEvaluation(methodDefinition methodDefinitionArgs, object ObjectType, enumerable bool) *PrivateElement {
	agent := vm.agent
	realm := agent.CurrentRealm()
	functionExpression := methodDefinition.FunctionExpression
	methodType := methodDefinition.MethodType
	propertyName := vm.stack.Pop()
	switch methodType {
	case MethodDefinitionTypeMethod:
		methodDef := DefineMethod(
			vm.agent,
			functionExpression,
			propertyName,
			object,
			nil,
		)
		SetFunctionName(methodDef.Closure, methodDef.Key, "")
		DefineMethodProperty(object, methodDef.Key, methodDef.Closure, enumerable)
	case MethodDefinitionTypeGet:
		propKeyOrPrivateName := propertyName
		env := vm.agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := vm.agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := functionExpression.SourceText
		formalParameterList := &FormalParameters{}
		closure := OrdinaryFunctionCreate(
			vm.agent,
			vm.agent.CurrentRealm().Intrinsics.FunctionPrototype,
			sourceText,
			formalParameterList,
			functionExpression.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		MakeMethod(closure, object)

		propKey := ToPropertyKey(vm.agent, propertyName)
		SetFunctionName(closure, propKey, "get")
		if _, ok := GetPrivateName(agent, propKeyOrPrivateName); ok {
			return &PrivateElement{
				Get:  closure,
				Kind: PrivateElementKindAccessor,
				Set:  nil,
			}
		} else {
			desc := &PropertyDescriptor{
				Get:          closure,
				Enumerable:   enumerable,
				Configurable: true,
			}
			object.DefinePropertyOrThrow(propKey, desc)

			return nil
		}
	case MethodDefinitionTypeSet:
		propKeyOrPrivateName := propertyName
		propKey := ToPropertyKey(vm.agent, propertyName)
		env := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := functionExpression.SourceText
		closure := OrdinaryFunctionCreate(
			agent,
			agent.CurrentRealm().Intrinsics.FunctionPrototype,
			sourceText,
			functionExpression.FormalParameters,
			functionExpression.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		MakeMethod(closure, object)
		SetFunctionName(closure, propKey, "set")
		if _, ok := GetPrivateName(agent, propKeyOrPrivateName); ok {
			return &PrivateElement{
				Get:  nil,
				Kind: PrivateElementKindAccessor,
				Set:  closure,
			}
		} else {
			desc := &PropertyDescriptor{
				Set:          closure,
				Enumerable:   enumerable,
				Configurable: true,
			}
			object.DefinePropertyOrThrow(propKey, desc)
			return nil
		}
	case MethodDefinitionTypeGenerator:
		propKey := ToPropertyKey(vm.agent, propertyName)
		env := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		generatorExpression := methodDefinition.GeneratorExpression
		sourceText := generatorExpression.SourceText
		closure := OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.GeneratorFunctionPrototype,
			sourceText,
			generatorExpression.FormalParameters,
			generatorExpression.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)

		MakeMethod(closure, object)
		SetFunctionName(closure, propKey, "")
		prototype := OrdinaryObjectCreate(agent, realm.Intrinsics.GeneratorFunctionPrototypePrototype, nil)

		closure.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
			Value:        NewValueFromObject(prototype),
			Writable:     true,
			Enumerable:   false,
			Configurable: false,
		})
		DefineMethodProperty(object, propKey, closure, enumerable)
	case MethodDefinitionTypeAsync:
		a := methodDefinition.AsyncFunctionExpression
		propKey := ToPropertyKey(vm.agent, propertyName)
		env := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := a.SourceText
		closure := OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.AsyncFunctionPrototype,
			sourceText,
			a.FormalParameters,
			a.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		MakeMethod(closure, object)
		SetFunctionName(closure, propKey, "")
		DefineMethodProperty(object, propKey, closure, enumerable)
	case MethodDefinitionTypeAsyncGenerator:
		a := methodDefinition.AsyncGeneratorExpression
		propKey := ToPropertyKey(vm.agent, propertyName)
		env := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := a.SourceText
		closure := OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.AsyncGeneratorFunctionPrototype,
			sourceText,
			a.FormalParameters,
			a.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)
		MakeMethod(closure, object)
		SetFunctionName(closure, propKey, "")
		prototype := OrdinaryObjectCreate(agent, realm.Intrinsics.AsyncGeneratorFunctionPrototypePrototype, nil)
		closure.DefinePropertyOrThrow(NewStringPropertyKey("prototype"), &PropertyDescriptor{
			Value:        NewValueFromObject(prototype),
			Writable:     true,
			Enumerable:   false,
			Configurable: false,
		})
		DefineMethodProperty(object, propKey, closure, enumerable)
	}
	panic("unreachable")
}

func (vm *VM) Run(executable *Executable) CompletionValue {
	for vm.ip < len(executable.Instructions) {
		i := executable.Instructions[vm.ip]
		vm.execute(executable, i)
		if _, ok := i.(*IReturn); ok {
			return NewCompletionReturnValue(vm.result)
		}
		vm.ip += 1
	}
	if vm.result == nil {
		return UndefinedValue.ToCompletion()
	}
	return vm.result.ToCompletion()
}

// 13.3.6.2
func evaluateCall(agent *Agent, function Value, this Value, arguments []Value) Value {
	if _, ok := function.(*ObjectValue); !ok {
		panic("TypeError: function is not an object")
	}
	if !IsCallable(function) {
		panic("TypeError: function is not callable")
	}
	return function.CallAssumeCallable(this, arguments)
}

func evaluateCallGetThisValue(reference *ReferenceRecord) Value {
	if reference == nil {
		return UndefinedValue
	}
	if reference.IsPropertyReference() {
		return reference.GetThisValue()
	}
	refEnv := reference.Base.(*ReferenceRecordBaseEnvironment).Environment
	if o := refEnv.WithBaseObject(); o != nil {
		return NewValueFromObject(o)
	}
	return UndefinedValue
}

func directEval(agent *Agent, arguments []Value, strict bool) Value {
	if len(arguments) == 0 {
		return nil
	}
	evalArg := arguments[0]
	strictCaller := strict
	return PerformEval(agent, evalArg, strictCaller, true)
}

// 13.3.5.1.1
func evaluateNew(agent *Agent, constructor Value, arguments []Value) Value {
	if !IsConstructor(constructor) {
		panic("TypeError: constructor is not a constructor")
	}
	o := MustGetObject(constructor)
	return NewValueFromObject(
		o.Construct(arguments, nil),
	)
}

// 13.10.2
func InstanceOfOperator(agent *Agent, value Value, target Value) bool {
	if _, ok := target.(*ObjectValue); !ok {
		panic("TypeError: target is not an object")
	}
	symbol := WellKnownSymbols[WellKnownSymbolsHasInstance]
	instOfHandler := GetMethod(
		agent,
		target,
		NewSymbolPropertyKey(symbol))
	if instOfHandler != nil {
		return instOfHandler.ToValue().CallAssumeCallable(target, []Value{value}).ToBoolean()
	}

	if !IsCallable(target) {
		panic("TypeError: target is not callable")
	}
	return OrdinaryHasInstance(agent, target, value).Data()
}

// 13.15.3
func applyStringOrNumericBinaryOperator(
	agent *Agent,
	lhs Value,
	rhs Value,
	op BinaryOperator,
) Value {
	var finalLval = lhs
	var finalRval = rhs
	if op == BinaryOperatorAddition {
		lprim := ToPrimitive(agent, lhs, PreferredTypeDefault)
		rprim := ToPrimitive(agent, rhs, PreferredTypeDefault)
		_, lprimIsString := lprim.(*StringValue)
		_, rprimIsString := rprim.(*StringValue)
		if lprimIsString || rprimIsString {
			lstr := lprim.String()
			rstr := rprim.String()
			return NewStringValue(lstr + rstr)
		}

		finalLval = lprim
		finalRval = rprim
	}

	lnum := ToNumeric(agent, finalLval)
	rnum := ToNumeric(agent, finalRval)
	if reflect.TypeOf(lnum) != reflect.TypeOf(rnum) {
		panic("TypeError: lnum and rnum are not the same type")
	}

	lNumber, isNumber := lnum.(*NumberValue)
	lBigInt, _ := lnum.(*BigIntValue)
	rNumber, _ := rnum.(*NumberValue)
	rBigInt, _ := rnum.(*BigIntValue)

	switch op {
	case BinaryOperatorExponentiation:
		if isNumber {
			return lNumber.Exponentiate(rNumber)
		} else {
			return lBigInt.Exponentiate(rBigInt)
		}
	case BinaryOperatorMultiplication:
		if isNumber {
			return lNumber.Multiply(rNumber)
		} else {
			return lBigInt.Multiply(rBigInt)
		}
	case BinaryOperatorAddition:
		if isNumber {
			return lNumber.Add(rNumber)
		} else {
			return lBigInt.Add(rBigInt)
		}
	case BinaryOperatorSubtraction:
		if isNumber {
			return lNumber.Subtract(rNumber)
		} else {
			return lBigInt.Subtract(rBigInt)
		}
	case BinaryOperatorDivision:
		if isNumber {
			return lNumber.Divide(rNumber)
		} else {
			return lBigInt.Divide(rBigInt)
		}
	case BinaryOperatorRemainder:
		if isNumber {
			return lNumber.Remainder(rNumber)
		} else {
			return lBigInt.Remainder(rBigInt)
		}
	case BinaryOperatorLeftShift:
		if isNumber {
			return lNumber.LeftShift(rNumber)
		} else {
			return lBigInt.LeftShift(rBigInt)
		}
	case BinaryOperatorRightShift:
		if isNumber {
			return lNumber.SignedRightShift(rNumber)
		} else {
			return lBigInt.SignedRightShift(rBigInt)
		}
	case BinaryOperatorUnsignedRightShift:
		if isNumber {
			return lNumber.UnsignedRightShift(rNumber)
		} else {
			return lBigInt.UnsignedRightShift(rBigInt)
		}
	case BinaryOperatorBitwiseAnd:
		if isNumber {
			return lNumber.BitwiseAnd(rNumber)
		} else {
			return lBigInt.BitwiseAnd(rBigInt)
		}
	case BinaryOperatorBitwiseOr:
		if isNumber {
			return lNumber.BitwiseOr(rNumber)
		} else {
			return lBigInt.BitwiseOr(rBigInt)
		}
	case BinaryOperatorBitwiseXor:
		if isNumber {
			return lNumber.BitwiseXor(rNumber)
		} else {
			return lBigInt.BitwiseXor(rBigInt)
		}
	}
	panic("unreachable")
}

// 15.2.5
func InstantiateOrdinaryFunctionExpression(
	agent *Agent,
	functionExpression *PrimaryExpressionFunctionExpression,
	name string,
) ObjectType {
	realm := agent.CurrentRealm()
	if functionExpression.Identifier != "" {
		Assert(name == "")
		name = string(functionExpression.Identifier)
		outerEnv := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		funcEnv := NewDeclarativeEnvironment(outerEnv)
		funcEnv.CreateImmutableBinding(name, false)
		privateEnv := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := functionExpression.SourceText
		closure := OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.FunctionPrototype,
			sourceText,
			functionExpression.FormalParameters,
			functionExpression.Body,
			functionCreateThisModeNonLexical,
			funcEnv,
			privateEnv,
		)
		SetFunctionName(closure, NewStringPropertyKey(name), "")
		MakeConstructor(closure, false, nil)

		funcEnv.InitializeBinding(name, NewValueFromObject(closure))
		return closure
	} else {
		env := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
		privateEnv := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
		sourceText := functionExpression.SourceText
		closure := OrdinaryFunctionCreate(
			agent,
			realm.Intrinsics.FunctionPrototype,
			sourceText,
			functionExpression.FormalParameters,
			functionExpression.Body,
			functionCreateThisModeNonLexical,
			env,
			privateEnv,
		)

		SetFunctionName(closure, NewStringPropertyKey(name), "")
		MakeConstructor(closure, false, nil)
		return closure
	}
}

// 15.3.4
func InstantiateArrowFunctionExpression(agent *Agent, arrowFunction *PrimaryExpressionArrowFunction, name string) ObjectType {
	realm := agent.CurrentRealm()

	env := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment

	privateEnv := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment

	sourceText := arrowFunction.SourceText

	closure := OrdinaryFunctionCreate(
		agent,
		realm.Intrinsics.FunctionPrototype,
		sourceText,
		arrowFunction.FormalParameters,
		arrowFunction.Body,
		functionCreateThisModeLexical,
		env,
		privateEnv,
	)

	SetFunctionName(closure, NewStringPropertyKey(name), "")
	return closure
}

// 15.4.4
type DefineMethodRecord struct {
	Key     PropertyKey
	Closure ObjectType
}

func DefineMethod(agent *Agent, functionExpression *PrimaryExpressionFunctionExpression, propertyName Value, object ObjectType, proto ObjectType) *DefineMethodRecord {
	realm := agent.CurrentRealm()
	propKey := ToPropertyKey(agent, propertyName)
	env := agent.runningExecutionContext().ECMAScriptCode.LexicalEnvironment
	privateEnv := agent.runningExecutionContext().ECMAScriptCode.PrivateEnvironment
	var prototype ObjectType
	if proto == nil {
		prototype = realm.Intrinsics.FunctionPrototype
	} else {
		prototype = proto
	}
	sourceText := functionExpression.SourceText
	closure := OrdinaryFunctionCreate(agent,
		prototype,
		sourceText,
		functionExpression.FormalParameters,
		functionExpression.Body,
		functionCreateThisModeNonLexical,
		env,
		privateEnv,
	)
	MakeMethod(closure, object)
	return &DefineMethodRecord{
		Key:     propKey,
		Closure: closure,
	}
}

// MARK: - Debug

func (vm *VM) debugPrintStack(msg string) {
	if !Debug.PrintBytecode || true {
		return
	}
	fmt.Printf("Stack(%s): size: %d\n", msg, vm.stack.Len())

	for i, v := range vm.stack.Data() {
		fmt.Printf("%d: %s\n", i, printValue(v))
	}
}
