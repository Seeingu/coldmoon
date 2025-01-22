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

type IReturnV2 struct{}

func (i *IReturnV2) String() string {
	return "Return "
}

type IGetValueV2 struct {
	name string
}

func (i *IGetValueV2) String() string {
	return "GetValue " + i.name
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

func (i *IR) Return() {
	i.instructions = append(i.instructions, &IReturnV2{})
}

func (i *IR) GetValue(name string) {
	i.instructions = append(i.instructions, &IGetValueV2{name})
	i.Return()
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

// List initialize a list value
func (i *IR) List(init string) {
	i.instructions = append(i.instructions, &IList{init: init})
	i.Return()
}

func (i *IR) BlockEvaluation(expr RuntimeSemanticsEvaluation, b *BytecodeContext) {
	i.AddInstruction(&IEnterBlock{})
	defer func() {
		i.AddInstruction(&ILeaveBlock{})
	}()
	expr.Evaluation(i, b)
}

type valueMap = map[string]Value

type VM2 struct {
	agent     *Agent
	blocks    []valueMap
	value     Value
	lastValue Value
	ip        int
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

func (v *VM2) Run(ir *IR) CompletionValue {
	for v.ip < len(ir.instructions) {
		i := ir.instructions[v.ip]
		switch ins := i.(type) {
		case *ILet:
			Assert(v.value != nil)
			v.valueMap()[ins.name] = v.value
			v.value = nil
		case *IReturnV2:
			Assert(v.lastValue != nil)
			v.value = v.lastValue
			v.lastValue = nil
		case *IGetValueV2:
			ref := v.valueMap()[ins.name]
			v.lastValue = ref.GetValue(v.agent)
		case *IThis:
			v.lastValue = ins.value
		case *IValue:
			v.lastValue = ins.value
		case *IList:
			var list []Value
			if ins.init != "" {
				list = []Value{v.valueMap()[ins.init]}
			}
			v.lastValue = NewListValue(list)
		case *IListConcatenation:
			list1 := v.valueMap()[ins.list1].(*ListValue).Values
			list2 := v.valueMap()[ins.list2].(*ListValue).Values
			v.lastValue = NewListValue(append(list1, list2...))
		case *IResolveBinding:
			v.lastValue = NewReferenceRecordValue(
				v.agent.ResolveBinding(string(ins.Name), nil, ins.Strict),
			)
		case *IEvaluateCall:
			v.lastValue = v.EvaluateCall(ins)
		case *IEnterBlock:
			v.blocks = append(v.blocks, make(valueMap))
		case *ILeaveBlock:
			v.blocks = v.blocks[:len(v.blocks)-1]
		case *IEvaluatePropertyAccessWithIdentifierKeyV2:
			baseValue := v.valueMap()[ins.baseValue]
			// TODO: use 13.1.2 Static Semantics: StringValue
			propertyNameString := ins.identifierName
			v.lastValue = NewReferenceRecordValue(
				NewReferenceRecord(
					&ReferenceRecordBase{value: baseValue},
					&ReferencedName{String: string(propertyNameString)},
					ins.strict,
					nil,
				))
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
	arguments := v.valueMap()[ins.arguments].(*ListValue).Values
	fun := v.valueMap()[ins.fun]
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
