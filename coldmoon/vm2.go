package coldmoon

import (
	"fmt"
	"strings"
)

type ILet struct {
	name string
}

var _ Instruction = (*ILet)(nil)

func (i *ILet) String() string {
	return "Let " + i.name
}

type IInternalGetLastValue struct{}

func (i *IInternalGetLastValue) String() string {
	return "IInternalGetLastValue "
}

type IGetValueFromName struct {
	name string
}

func (i *IGetValueFromName) String() string {
	return "GetValueFromName " + i.name
}

type IReturnV2 struct {
	value string
}

func (i *IReturnV2) String() string {
	return "ReturnV2 " + i.value
}

type IOr struct {
	x string
	y string
}

func (i *IOr) String() string {
	return "Or " + i.x + " " + i.y
}

type IGetValueV2 struct {
	name string
}

func (i *IGetValueV2) String() string {
	return "GetValue " + i.name
}

type IStringValue struct {
	value string
}

func (i *IStringValue) String() string {
	return "StringValue " + i.value
}

type IIsAnonymousFunctionDefinition struct {
	// expr is an ExpressionValue
	expr string
}

func (i *IIsAnonymousFunctionDefinition) String() string {
	return "IsAnonymousFunctionDefinition " + i.expr
}

type INamedEvaluation struct {
	expr string
	// name is a PropertyKey or a PrivateName
	name string
}

func (i *INamedEvaluation) String() string {
	return "NamedEvaluation " + i.expr
}

type IThis struct {
	value Value
}

func (i *IThis) String() string {
	return "This " + i.value.String()
}

// IEvaluateCall
type IEvaluateCall struct {
	// fun is function name
	fun          string
	ref          string
	arguments    string
	tailPosition string
}

func (i *IEvaluateCall) String() string {
	return "EvaluateCall " + i.fun
}

type IIsLessThan struct {
	x     string
	y     string
	order isLessThanOrder
}

func (i *IIsLessThan) String() string {
	return "IsLessThan " + i.x + " " + i.y
}

type IIfTrue struct {
	condition string
}

func (i *IIfTrue) String() string {
	return "IfTrue " + i.condition
}

type IEnterBlock struct{}

func (i *IEnterBlock) String() string {
	return "Block"
}

type IList struct {
	// init is nullable
	init string
}

func (i *IList) String() string {
	return "List"
}

type IListConcatenation struct {
	list1 string
	list2 string
}

func (i *IListConcatenation) String() string {
	return "ListConcatenation"
}

type ILeaveBlock struct{}

func (i *ILeaveBlock) String() string {
	return "LeaveBlock"
}

type IValue struct {
	value Value
}

func (i *IValue) String() string {
	return "Value " + i.value.String()
}

type IEvaluatePropertyAccessWithExpressionKeyV2 struct {
	baseValue  string
	expression string
	strict     bool
}

func (i *IEvaluatePropertyAccessWithExpressionKeyV2) String() string {
	return "EvaluatePropertyAccessWithExpressionKey " + i.baseValue
}

// 13.3.4
type IEvaluatePropertyAccessWithIdentifierKeyV2 struct {
	baseValue      string
	identifierName IdentifierName
	strict         bool
}

func (i *IEvaluatePropertyAccessWithIdentifierKeyV2) String() string {
	return "EvaluatePropertyAccessWithIdentifierKey " + i.baseValue
}

type IIsLooselyEqual struct {
	x string
	y string
}

func (i *IIsLooselyEqual) String() string {
	return "IsLooselyEqual " + i.x + " " + i.y
}

type IInitializeReferencedBindingV2 struct {
	// referenceRecord is ReferenceRecordValue
	referenceRecord string
	value           string
}

func (i *IInitializeReferencedBindingV2) String() string {
	return "InitializeReferencedBinding " + i.referenceRecord
}

type IIsStrictlyEqual struct {
	x string
	y string
}

func (i *IIsStrictlyEqual) String() string {
	return "IsStrictlyEqual " + i.x + " " + i.y
}

type ILogicalNotV2 struct {
	value string
}

func (i *ILogicalNotV2) String() string {
	return "LogicalNot " + i.value
}

type IResolveBindingV2 struct {
	name   string
	strict bool
}

func (i *IResolveBindingV2) String() string {
	return "ResolveBinding " + i.name
}

type IEvaluateStringOrNumericBinaryExpression struct {
	left     string
	operator BinaryOperator
	right    string
}

func (i *IEvaluateStringOrNumericBinaryExpression) String() string {
	return "EvaluateStringOrNumericBinaryExpression " + i.left + " " + i.operator.String() + " " + i.right
}

// IMark is an ip position anchor
// mark shouldn't be executed in vm
type IMark struct {
	mark mark
}

func (i *IMark) String() string {
	return "Mark " + string(i.mark)
}

// MARK: - IR

type IR struct {
	instructions []Instruction
}

func (i *IR) String() string {
	var s strings.Builder
	for index, ins := range i.instructions {
		s.WriteString(
			fmt.Sprintf("\t%d: %s\n", index, ins.String()),
		)
	}
	return s.String()
}

func NewIR() *IR {
	return &IR{}
}

// TODO(SM): Let should able to receive assignment
func (i *IR) Let(name string) {
	i.instructions = append(i.instructions, &ILet{name: name})
}

func (i *IR) Or(x, y string) {
	i.instructions = append(i.instructions, &IOr{x, y})
	i.GetLastValue()
}

func (i *IR) ResolveBinding(name IdentifierName, strict bool) {
	i.instructions = append(i.instructions, &IResolveBindingV2{name, strict})
	i.GetLastValue()
}

func (i *IR) InitializeReferencedBinding(referenceRecord string, value string) {
	i.instructions = append(i.instructions, &IInitializeReferencedBindingV2{referenceRecord, value})
}

func (i *IR) StringValue(s string) {
	i.instructions = append(i.instructions, &IStringValue{s})
	i.GetLastValue()
}

func (i *IR) IsAnonymousFunctionDefinition(expr string) {
	i.instructions = append(i.instructions, &IIsAnonymousFunctionDefinition{expr})
	i.GetLastValue()
}

func (i *IR) GetLastValue() {
	i.instructions = append(i.instructions, &IInternalGetLastValue{})
}

func (i *IR) GetValueFromName(name string) {
	i.instructions = append(i.instructions, &IGetValueFromName{name})
}

func (i *IR) Return(value string) {
	i.instructions = append(i.instructions, &IReturnV2{value})
}

func (i *IR) mark(mark mark) {
	i.instructions = append(i.instructions, &IMark{mark})
}

type mark string

// TODO: make marks global block uniquely
var (
	markIfTrue    mark = "if-true"
	markIfTrueEnd mark = "if-true-end"
	markIfElse    mark = "if-else"
	markIfElseEnd mark = "if-else-end"
)

// IfTrue matches
// if condition is true, execute trueCondition
// else execute falseCondition
//   - notice that both trueCondition and falseCondition
//     are not running in block,
//     then it will have same behavior as well as in spec
//   - falseCondition is optional
func (i *IR) IfTrue(condition string, trueCondition, falseCondition func()) {
	i.instructions = append(i.instructions, &IIfTrue{condition})
	i.mark(markIfTrue)
	trueCondition()
	i.mark(markIfTrueEnd)
	if falseCondition != nil {
		i.mark(markIfElse)
		falseCondition()
		i.mark(markIfElseEnd)
	}
}

func (i *IR) GetValue(name string) {
	i.instructions = append(i.instructions, &IGetValueV2{name})
	i.GetLastValue()
}

func (i *IR) AddInstruction(ins Instruction) {
	i.instructions = append(i.instructions, ins)
}

func (i *IR) This(v Value) {
	i.instructions = append(i.instructions, &IThis{v})
}

func (i *IR) Value(v Value) {
	i.instructions = append(i.instructions, &IValue{value: v})
}

func (i *IR) IsLooselyEqual(x, y string) {
	i.instructions = append(i.instructions, &IIsLooselyEqual{x: x, y: y})
	i.GetLastValue()
}

func (i *IR) IsStrictlyEqual(x, y string) {
	i.instructions = append(i.instructions, &IIsStrictlyEqual{x: x, y: y})
	i.GetLastValue()
}

func (i *IR) LogicalNot(x string) {
	i.instructions = append(i.instructions, &ILogicalNotV2{value: x})
	i.GetLastValue()
}

func (i *IR) IsLessThan(x, y string, order isLessThanOrder) {
	i.instructions = append(i.instructions, &IIsLessThan{x: x, y: y, order: order})
	i.GetLastValue()
}

// List initialize a list value
func (i *IR) List(init string) {
	i.instructions = append(i.instructions, &IList{init: init})
	i.GetLastValue()
}

func (i *IR) BlockEvaluation(expr RuntimeSemanticsEvaluation, b *BytecodeContext) {
	i.AddInstruction(&IEnterBlock{})
	defer func() {
		i.AddInstruction(&ILeaveBlock{})
	}()
	expr.Evaluation(i, b)
}

func (i *IR) RunInstructionsInBlock(f func()) {
	i.AddInstruction(&IEnterBlock{})
	defer func() {
		i.AddInstruction(&ILeaveBlock{})
	}()
	f()
}

func (i *IR) NamedEvaluation(expr string, name string) {
	i.instructions = append(i.instructions, &INamedEvaluation{expr, name})
	i.GetLastValue()
}

type valueMap = map[string]Value

type VM2 struct {
	agent     *Agent
	blocks    []valueMap
	value     Value
	lastValue Value
	ip        int
	// condition is a flag for if branch
	// TODO: scope control
	condition bool
}

func NewVM2(agent *Agent) *VM2 {
	// should have at least one block
	blocks := []valueMap{make(valueMap)}
	return &VM2{
		agent:  agent,
		blocks: blocks,
	}
}

func (v *VM2) valueMap() valueMap {
	return v.blocks[len(v.blocks)-1]
}

// getValue get value from valueMap
func (v *VM2) getValue(name string) Value {
	switch name {
	case "undefined":
		return UndefinedValue
	case "null":
		return NullValue
	case "true":
		return TrueValue
	case "false":
		return FalseValue
	}
	return v.valueMap()[name]
}

func (v *VM2) Run(ir *IR) CompletionValue {
loop:
	for v.ip < len(ir.instructions) {
		i := ir.instructions[v.ip]
		switch ins := i.(type) {
		case *ILet:
			Assert(v.value != nil)
			v.valueMap()[ins.name] = v.value
			v.value = nil
		case *IInternalGetLastValue:
			Assert(v.lastValue != nil)
			v.value = v.lastValue
			v.lastValue = nil
		case *IGetValueV2:
			ref := v.getValue(ins.name)
			v.lastValue = ref.GetValue(v.agent)
		case *IThis:
			v.lastValue = ins.value
		case *IValue:
			v.lastValue = ins.value
		case *IList:
			var list []Value
			if ins.init != "" {
				list = []Value{v.getValue(ins.init)}
			}
			v.lastValue = NewListValue(list)
		case *IListConcatenation:
			list1 := v.getValue(ins.list1).(*ListValue).Values
			list2 := v.getValue(ins.list2).(*ListValue).Values
			v.lastValue = NewListValue(append(list1, list2...))
		case *IResolveBindingV2:
			v.lastValue = NewReferenceRecordValue(
				v.agent.ResolveBinding(
					v.getValue(ins.name).String(),
					nil, ins.strict),
			)
		case *IEvaluateCall:
			r := v.EvaluateCall(ins)
			v.lastValue = r
		case *IEnterBlock:
			v.blocks = append(v.blocks, make(valueMap))
		case *ILeaveBlock:
			v.blocks = v.blocks[:len(v.blocks)-1]
		case *IEvaluatePropertyAccessWithIdentifierKeyV2:
			baseValue := v.getValue(ins.baseValue)
			// TODO: use 13.1.2 Static Semantics: StringValue
			propertyNameString := ins.identifierName
			v.lastValue = NewReferenceRecordValue(
				NewReferenceRecord(
					&ReferenceRecordBase{value: baseValue},
					&ReferencedName{String: string(propertyNameString)},
					ins.strict,
					nil,
				))
		case *IIsLooselyEqual:
			x := v.getValue(ins.x)
			y := v.getValue(ins.y)
			v.lastValue = NewBooleanValue(IsLooselyEqual(v.agent, x, y))
		case *IIsStrictlyEqual:
			x := v.getValue(ins.x)
			y := v.getValue(ins.y)
			v.lastValue = NewBooleanValue(IsStrictlyEqual(x, y))
		case *ILogicalNotV2:
			value := v.getValue(ins.value)
			v.lastValue = NewBooleanValue(!value.ToBoolean())
		case *IIsLessThan:
			x := v.getValue(ins.x)
			y := v.getValue(ins.y)
			v.lastValue = IsLessThanV2(v.agent, x, y, ins.order)
		case *IIfTrue:
			condition := v.getValue(ins.condition)
			v.condition = condition.ToBoolean()
		case *IGetValueFromName:
			v.lastValue = v.getValue(ins.name)
		case *IStringValue:
			v.lastValue = NewStringValue(ins.value)
		case *IIsAnonymousFunctionDefinition:
			value := v.getValue(ins.expr)
			if exprValue, ok := value.(*ExpressionValue); ok {
				v.lastValue = NewBooleanValue(exprValue.IsAnonymousFunctionDefinition())
			} else {
				Assert(false)
			}
		case *IReturnV2:
			v.value = v.getValue(ins.value)
			return NewCompletionReturnValue(v.value)
		case *IMark:
			// TODO: refactor
			if v.condition {
				if ins.mark == markIfElse {
					for i := v.ip; i < len(ir.instructions); i++ {
						if m, ok := ir.instructions[i].(*IMark); ok {
							if m.mark == markIfElseEnd {
								v.ip = i + 1
								continue loop
							}
						}
					}
				}
			} else {
				if ins.mark == markIfTrue {
					for i := v.ip; i < len(ir.instructions); i++ {
						if m, ok := ir.instructions[i].(*IMark); ok {
							if m.mark == markIfTrueEnd {
								v.ip = i
								continue loop
							}
						}
					}
				}
			}
		case *INamedEvaluation:
			value := v.getValue(ins.expr)
			// FIXME: handle named evaluation
			panic(value)
		case *IInitializeReferencedBindingV2:
			refRecordValue, ok := v.getValue(ins.referenceRecord).(*ReferenceRecordValue)
			Assert(ok)
			refRecordValue.ReferenceRecord.InitializeReferencedBinding(v.getValue(ins.value))
		case *IEvaluateStringOrNumericBinaryExpression:
			left := v.getValue(ins.left)
			right := v.getValue(ins.right)
			v.lastValue = ApplyStringOrNumericBinaryOperator(
				v.agent, left, right, ins.operator,
			)
		case *IOr:
			x := v.getValue(ins.x)
			if x.ToBoolean() {
				v.lastValue = x
			} else {
				y := v.getValue(ins.y)
				v.lastValue = y
			}
		default:
			panic("unknown instruction")
		}
		v.ip++
	}

	if v.agent.exception != nil {
		return NewCompletionValueError(v.agent.exception)
	}

	return v.value.ToCompletion()
}

func (v *VM2) skipBlock(ir *IR) {
	for i := v.ip; i < len(ir.instructions); i++ {
		if _, ok := ir.instructions[i].(*ILeaveBlock); ok {
			v.ip = i - 1
			return
		}
	}
}

// EvaluateCall ( func, ref, arguments, tailPosition )
// 13.3.6.2
func (v *VM2) EvaluateCall(ins *IEvaluateCall) Value {
	agent := v.agent
	var thisValue Value
	ref := v.valueMap()[ins.ref]
	if r, ok := ref.(*ReferenceRecordValue); ok {
		rr := r.ReferenceRecord
		if rr.IsPropertyReference() {
			thisValue = rr.GetThisValue()
		} else {
			refEnv, ok := rr.Base.Env()
			Assert(ok)
			// TODO(SM): we should provide undefined object
			if o := refEnv.WithBaseObject(); o != nil {
				thisValue = o.ToValue()
			} else {
				thisValue = UndefinedValue
			}
		}
	} else {
		thisValue = UndefinedValue
	}
	arguments := v.getValue(ins.arguments).(*ListValue).Values
	fun := v.getValue(ins.fun)
	if !ValueIsObject(fun) {
		return agent.ThrowTypeError("function is not an object")
	}
	if !IsCallable(fun) {
		return agent.ThrowTypeError("function is not callable")
	}
	// TODO: WIP: tailPosition
	// TODO(SM): argumentsList
	return evaluateCall(agent, fun, thisValue, arguments)
}
