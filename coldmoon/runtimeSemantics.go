package coldmoon

// spec: 8.1
type RuntimeSemanticsEvaluation interface {
	Evaluation(vm *VM) (co CompletionValue)
}

// 8.4.5
type RuntimeSemanticsNamedEvaluation interface {
	NamedEvaluation(name string, expr Expression) ObjectType
}

// 13.2.5.5
type RuntimeSemanticsPropertyDefinitionEvaluation interface {
	// PropertyDefinitionEvaluation returns UNUSED
	PropertyDefinitionEvaluation(vm *VM, obj ObjectType)
}

// 13.3.8.1
type RuntimeSemanticsArgumentListEvaluation interface {
	ArgumentListEvaluation(vm *VM) []Value
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
	ClassFieldDefinitionEvaluation(vm *VM, homeObject ObjectType) (field *ClassFieldDefinition, err Value)
}

// MARK: - Array

// RuntimeSemanticsArrayAccumulation
// spec: 13.2.4.1
type RuntimeSemanticsArrayAccumulation interface {
	ArrayAccumulation(vm *VM, array *ArrayObject, nextIndex JSInt) (index JSInt, err Value)
}

// MARK: - Loop

// RuntimeSemanticsForInOfLoopEvaluation
// spec: 13.7.5.5
type RuntimeSemanticsForInOfLoopEvaluation interface {
	ForInOfLoopEvaluation(vm *VM, labelSet []string) (value Value, err Value)
}

// RuntimeSemanticsWhileLoopEvaluation
// spec: 14.7.3.2
type RuntimeSemanticsWhileLoopEvaluation interface {
	WhileLoopEvaluation(vm *VM, labelSet []string) (value Value, err Value)
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
	EvaluateBody(agent *Agent, functionObject *ECMAScriptFunction, argumentList []Value) (value Value, err Value)
}

// MARK: - Try Catch

// RuntimeSemanticsCatchClauseEvaluation
// spec: 14.15.2
type RuntimeSemanticsCatchClauseEvaluation interface {
	CatchClauseEvaluation(vm *VM, thrownValue Value) (c Completion[Value])
}
