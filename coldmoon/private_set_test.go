package coldmoon

import "testing"

func TestPrivateReferencePutValueUpdatesField(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)

	object := NewObject(agent, nil, "Object")
	privateName := PrivateName{Symbol: agent.CreateSymbol("field")}
	object.PrivateFieldAdd(privateName, NewNumberValue(1))
	reference := NewReferenceRecord(
		NewReferenceRecordBaseValue(object.ToValue()),
		&ReferencedName{PrivateName: &privateName},
		true,
		nil,
	)

	want := NewNumberValue(2)
	if result := reference.PutValue(agent, want); result.IsAbrupt() {
		t.Fatalf("PutValue completed abruptly: %#v", result)
	}
	if got := object.PrivateElementFind(privateName).Value; got != want {
		t.Fatalf("private field value = %v, want %v", got, want)
	}
}

func TestPrivateSetAccessorAndNonWritableKinds(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	object := NewObject(agent, nil, "Object")

	var received Value
	setter := CreateBuiltinFunction(agent, func(_ Value, arguments []Value, _ ObjectType) CompletionConvertable[Value] {
		received = argumentAt(arguments, 0)
		return UndefinedValue
	}, 1, CMString("set value"), builtinFunctionArgs{})
	accessorName := PrivateName{Symbol: agent.CreateSymbol("accessor")}
	object.PrivateMethodOrAccessorAdd(&PrivateElement{
		Key:  accessorName,
		Kind: PrivateElementKindAccessor,
		Set:  setter,
	})
	want := NewStringValue("updated")
	if result := object.PrivateSet(accessorName, want); result.IsAbrupt() {
		t.Fatalf("accessor PrivateSet completed abruptly: %#v", result)
	}
	if received != want {
		t.Fatalf("setter received %v, want %v", received, want)
	}

	methodName := PrivateName{Symbol: agent.CreateSymbol("method")}
	object.PrivateMethodOrAccessorAdd(&PrivateElement{
		Key:   methodName,
		Kind:  PrivateElementKindMethod,
		Value: setter.ToValue(),
	})
	if result := object.PrivateSet(methodName, want); !result.IsAbrupt() {
		t.Fatal("writing a private method did not throw")
	}
}
