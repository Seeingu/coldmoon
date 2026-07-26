package coldmoon

import "strings"

type (
	BehaviorFn func(thisArgument Value, argumentsList []Value, newTarget ObjectType) CompletionConvertable[Value]
	// BuiltinInvocation is the normalized input to one builtin call or
	// construction. Argument returns undefined for an omitted argument.
	BuiltinInvocation struct {
		This      Value
		Arguments []Value
		NewTarget ObjectType
	}
	BuiltinFunction struct {
		*Object
		InternalSlotPrivateMethods
		InternalSlotFields
		Realm              *Realm
		InitialName        string
		Behavior           BehaviorFn
		RevocableProxy     ObjectType
		AdditionalFields   *AdditionalFields
		AdditionalFieldsV2 any
	}
)

// Argument returns the argument at index, or undefined when it was omitted.
func (i BuiltinInvocation) Argument(index int) Value {
	return argumentAt(i.Arguments, index)
}

// IsConstruct reports whether the builtin was invoked through [[Construct]].
func (i BuiltinInvocation) IsConstruct() bool {
	return i.NewTarget != nil
}

func argumentAt(arguments []Value, index int) Value {
	if index < 0 || index >= len(arguments) || arguments[index] == nil {
		return UndefinedValue
	}
	return arguments[index]
}

var (
	_ InternalSlotPrivateMethods = (*BuiltinFunction)(nil)
	_ InternalSlotFields         = (*BuiltinFunction)(nil)
)

func (b *BuiltinFunction) PrivateMethods() []*PrivateElement {
	return b.AdditionalFields.ClassConstructorFields.PrivateMethods
}

func (b *BuiltinFunction) Fields() []*ClassFieldDefinition {
	return b.AdditionalFields.ClassConstructorFields.Fields
}

func (b *BuiltinFunction) SetFields(f []*ClassFieldDefinition) {
	b.AdditionalFields.ClassConstructorFields.Fields = f
}

func (b *BuiltinFunction) Construct(
	argumentLists []Value,
	newTarget ObjectType,
) Completion[ObjectType] {
	if newTarget == nil {
		newTarget = b
	}
	return b.internalMethods().Construct(b, argumentLists, newTarget)
}

func (b *BuiltinFunction) ToObject() *Object {
	return b.Object
}

// 7.3.24
func (b *BuiltinFunction) GetFunctionRealm() *Realm {
	return b.Realm
}

// 10.3.1
func BuiltinCall(o ObjectType, thisArgument Value, argumentsList []Value) CompletionValue {
	return o.(*BuiltinFunction).BuiltinCallOrConstruct(thisArgument, argumentsList, nil)
}

// BuiltinConstruct
// [[Construct]]
// spec: 10.3.2
func BuiltinConstruct(
	b ObjectType,
	argumentsList []Value,
	newTarget ObjectType,
) (co Completion[ObjectType]) {
	r, isAbrupt, rt := ReturnIfAbrupt(b.(*BuiltinFunction).BuiltinCallOrConstruct(NullValue, argumentsList, newTarget), co)
	if isAbrupt {
		return rt
	}
	return r.ToObject(b.Agent())
}

// BuiltinCallOrConstruct
// spec: 10.3.3
func (b *BuiltinFunction) BuiltinCallOrConstruct(thisArgument Value, argumentsList []Value, newTarget ObjectType) (co CompletionValue) {
	a := b.Agent()
	calleeContext := &ExecutionContext{
		Function:       b.Object,
		Realm:          b.Realm,
		ch:             make(chan struct{}),
		ScriptOrModule: nil,
	}

	scope := a.enterExecutionContext(calleeContext)
	defer func() {
		scope.Leave()
		recovered := recover()
		if recovered != nil {
			switch value := recovered.(type) {
			case Value:
				co.t = CompletionTypeThrow
				co.err = value
			case CompletionValue:
				co = value
			case string:
				switch {
				case strings.HasPrefix(value, "TypeError"):
					co = co.ThrowTypeError(a, panicMessage(value, "TypeError"))
				case strings.HasPrefix(value, "RangeError"):
					co = co.ThrowRangeError(a, panicMessage(value, "RangeError"))
				default:
					panic(recovered)
				}
			default:
				panic(recovered)
			}
		}
	}()

	invocation := BuiltinInvocation{
		This:      thisArgument,
		Arguments: argumentsList,
		NewTarget: newTarget,
	}
	r := b.Behavior(invocation.This, invocation.Arguments, invocation.NewTarget)
	if r == nil {
		return UndefinedValue.ToCompletion()
	}

	return CompletionHandleV2(r)
}

func panicMessage(message string, prefix string) string {
	message = strings.TrimPrefix(message, prefix)
	message = strings.TrimPrefix(message, ":")
	return strings.TrimSpace(message)
}

type builtinFunctionArgs struct {
	realm              *Realm
	prototype          ObjectType
	prefix             string
	isConstructor      bool
	revocableProxy     ObjectType
	additionalFields   *AdditionalFields
	additionalFieldsV2 any
}

// CreateBuiltinFunction
// spec: 10.3.4
func CreateBuiltinFunction(
	agent *Agent,
	behavior BehaviorFn,
	length JSInt,
	name PropertyConvertable,
	args builtinFunctionArgs,
) *BuiltinFunction {
	realm := args.realm
	if realm == nil {
		realm = agent.CurrentRealm()
	}

	prototype := args.prototype
	if prototype == nil {
		prototype = realm.Intrinsics.FunctionPrototype
	}
	if prototype == nil {
		panic("FunctionPrototype is nil")
	}

	typeName := "BuiltinFunction"
	if nameString, ok := name.(CMString); ok {
		typeName += " " + string(nameString)
	}
	object := NewObject(agent, prototype, typeName)
	object.SetExtensible(true)
	function := &BuiltinFunction{
		Object:             object,
		Realm:              realm,
		Behavior:           behavior,
		InitialName:        "",
		RevocableProxy:     args.revocableProxy,
		AdditionalFields:   args.additionalFields,
		AdditionalFieldsV2: args.additionalFieldsV2,
	}
	function.ref = function
	function.internalMethods().Call = BuiltinCall
	if args.isConstructor {
		function.internalMethods().Construct = BuiltinConstruct
	}

	SetFunctionLength(function.Object, length)
	SetFunctionName(function, name.ToPropertyKey(), args.prefix)

	return function
}

// MARK: - Class

type ClassConstructorFields struct {
	ConstructorKind ConstructorKind
	SourceText      string
	PrivateMethods  []*PrivateElement
	Fields          []*ClassFieldDefinition
}
