package coldmoon

// RuntimeSemanticsEvaluation
// spec: 8.1
type RuntimeSemanticsEvaluation interface {
	Evaluation(vm *VM) (co CompletionValue)
}

// RuntimeSemanticsNamedEvaluation
// spec: 8.4.5
type RuntimeSemanticsNamedEvaluation interface {
	NamedEvaluation(name string, expr Expression) ObjectType
}

// RuntimeSemanticsInstantiateFunctionObject
// 8.6.1
type RuntimeSemanticsInstantiateFunctionObject interface {
	// InstantiateFunctionObject return a FunctionObject
	InstantiateFunctionObject(env EnvironmentRecord, privateEnv *PrivateEnvironment) ObjectType
}

// RuntimeSemanticsBindingInitialization
// 8.6.2
type RuntimeSemanticsBindingInitialization interface {
	// BindingInitialization
	// env is EnvironmentRecord or undefined
	// return UNUSED or abrupt
	BindingInitialization(vm *VM, value Value, env EnvironmentRecord) CompletionValue
}

// 13.2.5.5
type RuntimeSemanticsPropertyDefinitionEvaluation interface {
	// PropertyDefinitionEvaluation returns UNUSED or abrupt
	PropertyDefinitionEvaluation(vm *VM, obj ObjectType) (co CompletionValue)
}

// 13.3.8.1
type RuntimeSemanticsArgumentListEvaluation interface {
	ArgumentListEvaluation(vm *VM) []Value
}

// RuntimeSemanticsDestructuringAssignmentEvaluation
// spec: 13.15.5.2
type RuntimeSemanticsDestructuringAssignmentEvaluation interface {
	DestructuringAssignmentEvaluation(vm *VM, value Value) (co CompletionValue)
}

// RuntimeSemanticsPropertyDestructuringAssignmentEvaluation
// spec: 13.15.5.3
type RuntimeSemanticsPropertyDestructuringAssignmentEvaluation interface {
	PropertyDestructuringAssignmentEvaluation(vm *VM, value Value) Completion[[]PropertyKey]
}

// RuntimeSemanticsRestDestructuringAssignmentEvaluation
// spec: 13.15.5.4
type RuntimeSemanticsRestDestructuringAssignmentEvaluation interface {
	// RestDestructuringAssignmentEvaluation
	// returns UNUSED or abrupt
	RestDestructuringAssignmentEvaluation(vm *VM, value Value, excludedNames []PropertyKey) CompletionValue
}

// RuntimeSemanticsIteratorDestructuringAssignmentEvaluation
// spec: 13.15.5.5
type RuntimeSemanticsIteratorDestructuringAssignmentEvaluation interface {
	// IteratorDestructuringAssignmentEvaluation
	// returns UNUSED or abrupt
	IteratorDestructuringAssignmentEvaluation(vm *VM, iteratorRecord *IteratorRecord) CompletionValue
}

// 15.3.4
type RuntimeSemanticsInstantiateArrowFunctionExpression interface {
	InstantiateArrowFunctionExpression(vm *VM, name string) Value
}

// RuntimeSemanticsMethodDefinitionEvaluation
// spec: 15.4.5
type RuntimeSemanticsMethodDefinitionEvaluation interface {
	// MethodDefinitionEvaluation returns PrivateElement or an abrupt Completion
	MethodDefinitionEvaluation(vm *VM, obj ObjectType, enumerable bool) Completion[*PrivateElement]
}

// RuntimeSemanticsChainEvaluation
// 13.3.9.2
type RuntimeSemanticsChainEvaluation interface {
	ChainEvaluation(vm *VM, baseValue Value, baseReference Value) (co CompletionValue)
}

type DefineMethodRecord struct {
	Key     PropertyKey
	Closure ObjectType
}

// RuntimeSemanticsDefineMethod
// spec: 15.4.4
type RuntimeSemanticsDefineMethod interface {
	// DefineMethod
	// proto is optional
	DefineMethod(vm *VM, obj ObjectType, proto ObjectType) Completion[*DefineMethodRecord]
}

// MARK: - Class Related

// 15.7.15
type RuntimeSemanticsBindingClassDeclarationEvaluation interface {
	// BindingClassDeclarationEvaluation returns a function object or an abrupt Completion
	BindingClassDeclarationEvaluation(vm *VM) (co Completion[ObjectType])
}

// RuntimeSemanticsInstantiateOrdinaryFunctionExpression
// spec: 15.2.5
type RuntimeSemanticsInstantiateOrdinaryFunctionExpression interface {
	// InstantiateOrdinaryFunctionExpression
	// name is optional
	InstantiateOrdinaryFunctionExpression(vm *VM, name PropertyKeyOrPrivateName) (fun ObjectType)
}

// RuntimeSemanticsClassDefinitionEvaluation
// spec: 15.7.14
type RuntimeSemanticsClassDefinitionEvaluation interface {
	// ClassDefinitionEvaluation
	// classBinding is optional
	ClassDefinitionEvaluation(vm *VM, classBinding string, className PropertyKeyOrPrivateName) (c Completion[ObjectType])
}

// classEvaluationResult enum
type classEvaluationResult struct {
	classFieldDefinition  *ClassFieldDefinition
	staticBlockDefinition *ClassStaticBlockDefinition
	privateElement        *PrivateElement
}

// RuntimeSemanticsClassElementEvaluation
// spec: 15.7.13
type RuntimeSemanticsClassElementEvaluation interface {
	ClassElementEvaluation(vm *VM, obj ObjectType) (result Completion[*classEvaluationResult])
}

// 15.7.10
type RuntimeSemanticsClassFieldDefinitionEvaluation interface {
	ClassFieldDefinitionEvaluation(vm *VM, homeObject ObjectType) (co Completion[*ClassFieldDefinition])
}

// MARK: - Array

// RuntimeSemanticsArrayAccumulation
// spec: 13.2.4.1
type RuntimeSemanticsArrayAccumulation interface {
	ArrayAccumulation(vm *VM, array *ArrayObject, nextIndex JSInt) (co Completion[JSInt])
}

// MARK: - Loop

// RuntimeSemanticsForInOfLoopEvaluation
// spec: 13.7.5.5
type RuntimeSemanticsForInOfLoopEvaluation interface {
	ForInOfLoopEvaluation(vm *VM, labelSet []string) (co CompletionValue)
}

// RuntimeSemanticsWhileLoopEvaluation
// spec: 14.7.3.2
type RuntimeSemanticsWhileLoopEvaluation interface {
	WhileLoopEvaluation(vm *VM, labelSet []string) (co CompletionValue)
}

// RuntimeSemanticsForDeclarationBindingInstantiation
// spec: 14.7.5.4
type RuntimeSemanticsForDeclarationBindingInstantiation interface {
	ForDeclarationBindingInstantiation(vm *VM, env EnvironmentRecord)
}

// MARK: - Generator

// RuntimeSemanticsInstantiateGeneratorFunctionExpression
// spec: 15.5.4
// name is optional
type RuntimeSemanticsInstantiateGeneratorFunctionExpression interface {
	InstantiateGeneratorFunctionExpression(vm *VM, name PropertyKeyOrPrivateName) (fun ObjectType)
}

// MARK: - Async

type RuntimeSemanticsInstantiateAsyncArrowFunctionExpression interface {
	InstantiateAsyncArrowFunctionExpression(vm *VM, name PropertyKeyOrPrivateName) (fun ObjectType)
}

// MARK: - Function

// RuntimeSemanticsEvaluateBody
// spec: 10.2.1.3
type RuntimeSemanticsEvaluateBody interface {
	EvaluateBody(agent *Agent, functionObject *ECMAScriptFunction, argumentList []Value) (co CompletionValue)
}

// MARK: - Try Catch

// RuntimeSemanticsCatchClauseEvaluation
// spec: 14.15.2
type RuntimeSemanticsCatchClauseEvaluation interface {
	CatchClauseEvaluation(vm *VM, thrownValue Value) (c Completion[Value])
}

// MARK: - Switch

// RuntimeSemanticsCaseBlockEvaluation
// spec: 14.12.2
type RuntimeSemanticsCaseBlockEvaluation interface {
	CaseBlockEvaluation(vm *VM, input Value) (co CompletionValue)
}
