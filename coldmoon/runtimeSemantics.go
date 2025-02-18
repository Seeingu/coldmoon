package coldmoon

// 13.1.3
type RuntimeSemanticsEvaluation interface {
	Evaluation(vm *VM2) Value
}

// 8.4.5
type RuntimeSemanticsNamedEvaluation interface {
	NamedEvaluation(name string, expr Expression) ObjectType
}

// 13.2.5.5
type RuntimeSemanticsPropertyDefinitionEvaluation interface {
	// PropertyDefinitionEvaluation returns UNUSED
	PropertyDefinitionEvaluation(vm *VM2, obj ObjectType)
}

// 13.3.8.1
type RuntimeSemanticsArgumentListEvaluation interface {
	ArgumentListEvaluation(vm *VM2) []Value
}

// 15.3.4
type RuntimeSemanticsInstantiateArrowFunctionExpression interface {
	InstantiateArrowFunctionExpression(vm *VM2, name string) Value
}

// 15.4.5
type RuntimeSemanticsMethodDefinitionEvaluation interface {
	// MethodDefinitionEvaluation returns PrivateElement or an abrupt Completion
	MethodDefinitionEvaluation(vm *VM2, obj ObjectType, enumerable bool) (pe *PrivateElement, err Value)
}

// RuntimeSemanticsChainEvaluation
// 13.3.9.2
type RuntimeSemanticsChainEvaluation interface {
	ChainEvaluation(vm *VM2, baseValue Value, baseReference Value) (value Value, err Value)
}

// MARK: - DefineMethod 15.4.4

type DefineMethodRecord struct {
	Key     PropertyKey
	Closure ObjectType
}
type RuntimeSemanticsDefineMethod interface {
	// proto is optional
	DefineMethod(vm *VM2, obj ObjectType, proto ObjectType) (record DefineMethodRecord, err Value)
}

// MARK: - Class Related

// 15.7.15
type RuntimeSemanticsBindingClassDeclarationEvaluation interface {
	// BindingClassDeclarationEvaluation returns a function object or an abrupt Completion
	BindingClassDeclarationEvaluation(vm *VM2) (obj ObjectType, err Value)
}

// 15.7.14
type RuntimeSemanticsClassDefinitionEvaluation interface {
	// classBinding is optional
	ClassDefinitionEvaluation(vm *VM2, classBinding string, className PropertyKeyOrPrivateName) (obj ObjectType, err Value)
}

// classEvaluationResult enum
type classEvaluationResult struct {
	classFieldDefinition  *ClassFieldDefinition
	staticBlockDefinition *ClassStaticBlockDefinition
	privateElement        *PrivateElement
}

// 15.7.13
type RuntimeSemanticsClassElementEvaluation interface {
	ClassElementEvaluation(vm *VM2, obj ObjectType) (result classEvaluationResult, err Value)
}

// 15.7.10
type RuntimeSemanticsClassFieldDefinitionEvaluation interface {
	ClassFieldDefinitionEvaluation(vm *VM2, homeObject ObjectType) (field *ClassFieldDefinition, err Value)
}

// MARK: - Array

// RuntimeSemanticsArrayAccumulation
// spec: 13.2.4.1
type RuntimeSemanticsArrayAccumulation interface {
	ArrayAccumulation(vm *VM2, array *ArrayObject, nextIndex JSInt) (index JSInt, err Value)
}

// MARK: - Loop

// RuntimeSemanticsForInOfLoopEvaluation
// spec: 13.7.5.5
type RuntimeSemanticsForInOfLoopEvaluation interface {
	ForInOfLoopEvaluation(vm *VM2, labelSet []string) (value Value, err Value)
}

// RuntimeSemanticsForDeclarationBindingInstantiation
// spec: 14.7.5.4
type RuntimeSemanticsForDeclarationBindingInstantiation interface {
	ForDeclarationBindingInstantiation(vm *VM2, env EnvironmentRecord)
}

// MARK: - Generator

// RuntimeSemanticsInstantiateGeneratorFunctionExpression
// spec: 15.5.4
// name is optional
type RuntimeSemanticsInstantiateGeneratorFunctionExpression interface {
	InstantiateGeneratorFunctionExpression(vm *VM2, name PropertyKeyOrPrivateName) (fun ObjectType)
}
