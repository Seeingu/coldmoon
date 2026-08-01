package coldmoon

import "testing"

func TestDefineBuiltinFunctionHostAdapter(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	hostObject := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)

	behavior := func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		if len(arguments) == 0 {
			return UndefinedValue
		}
		return arguments[0]
	}
	DefineBuiltinFunction(realm, CMString("hostCall"), hostObject, behavior, 1)

	key := NewStringPropertyKey("hostCall")
	descriptor := OrdinaryGetOwnProperty(hostObject, key)
	if descriptor == nil || !descriptor.Writable || descriptor.Enumerable || !descriptor.Configurable {
		t.Fatalf("host built-in descriptor = %+v", descriptor)
	}
	function := hostObject.Get(key)
	if !IsCallable(function) {
		t.Fatal("DefineBuiltinFunction did not install a callable")
	}
	argument := NewStringValue("result")
	result := function.Call(agent, hostObject.ToValue(), []Value{argument})
	if result.IsAbrupt() || result.Data() != argument {
		t.Fatalf("host built-in call = %#v", result)
	}
}
