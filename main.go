package main

import (
	"fmt"
	. "github.com/Seeingu/coldmoon/coldmoon"
)

func main() {
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	//sourceText := "\t{true; false\u2028;;;}\r\nnull;debugger\uFEFF"
	sourceText := `
if (true) {
	  true;
} else {
	false;
}
`
	script := ParseScript(sourceText, realm, nil)
	fmt.Println("AST: ", script.ECMAScriptCode.String())
	result := script.Evaluate()
	fmt.Println("Result: ", result.String())

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
