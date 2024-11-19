package tests

import (
	"fmt"
	. "github.com/Seeingu/coldmoon/coldmoon"
	"testing"
)

func testSource(t *testing.T, s string) {
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	//sourceText := "\t{true; false\u2028;;;}\r\nnull;debugger\uFEFF"
	sourceText := s
	script := ParseScript(sourceText, realm, nil)
	_ = script.Evaluate()

	{
		o := NewObject(agent, nil)
		key := StringPropertyKey{Value: "a"}
		o.InternalMethods().DefineOwnProperty(
			o,
			key,
			&PropertyDescriptor{
				Value: &NumberValue{Data: 12}},
		)
		keys := o.InternalMethods().OwnPropertyKeys(o)
		fmt.Println("Keys: ", keys)
		o2 := NewObject(agent, o)
		value := o2.InternalMethods().Get(o2, key, nil)
		fmt.Println("Value: ", value.(*NumberValue).Data)
	}

	booleanConstructor := realm.GlobalObject.Get(NewStringPropertyKey("Boolean"))

	oo := booleanConstructor.(*ObjectValue).Object.(*BuiltinFunction)
	booleanObject := ObjectConstruct(
		oo,
		[]Value{&BooleanValue{Data: false}}, nil)

	valueOf := booleanObject.ToObject().Get(NewStringPropertyKey("valueOf"))
	value := CallAssumeCallableNoArgs(valueOf, NewValueFromObject(booleanObject))
	fmt.Println("new Boolean(true).valueOf() = ", value.String())
}

func TestBaseline(t *testing.T) {
	//sourceText := "\t{true; false\u2028;;;}\r\nnull;debugger\uFEFF"
	sourceText := `
Date.UTC(2012);
`
	testSource(t, sourceText)

	sourceText = `
2 == 1;
2 > 1;
false || 1;
true && 1;
true ? 2 : 1;
2 ** 3;
() => 123;
`
	testSource(t, sourceText)
}
