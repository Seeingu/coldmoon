package main

import (
	"fmt"
	"github.com/Seeingu/coldmoon/coldmoon"
)

func main() {
	agent := coldmoon.NewAgent()
	coldmoon.InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	_ = coldmoon.ParseScript("", realm, nil)

	o := coldmoon.NewObject(agent, nil)
	key := coldmoon.StringPropertyKey{Value: "a"}
	o.InternalMethods().DefineOwnProperty(
		o,
		key,
		&coldmoon.PropertyDescriptor{Value: &coldmoon.NumberValue{Data: 12}},
	)
	o2 := coldmoon.NewObject(agent, o)
	value := o2.InternalMethods().Get(o2, key, nil)
	fmt.Println("Value: ", value.(*coldmoon.NumberValue).Data)
}
