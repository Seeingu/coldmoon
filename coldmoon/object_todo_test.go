package coldmoon

import "testing"

func TestObjectIsOrdinaryUsesInternalMethods(t *testing.T) {
	agent := NewAgent()
	prototype := NewObject(agent, nil, "Prototype")

	nullPrototypeObject := NewObject(agent, nil, "Object")
	if !nullPrototypeObject.IsOrdinary() {
		t.Fatal("an object with the ordinary internal methods and a null prototype must be ordinary")
	}

	exoticObject := NewObject(agent, prototype, "Exotic")
	exoticObject.internalMethods().Get = func(ObjectType, PropertyKey, Value) CompletionValue {
		return UndefinedValue.ToCompletion()
	}
	if exoticObject.IsOrdinary() {
		t.Fatal("an object that overrides an essential internal method must be exotic")
	}
}

func TestPrivateFieldAddCallsHostHookAndStoresElement(t *testing.T) {
	agent := NewAgent()
	hookCalls := 0
	agent.HostHooks.HostEnsureCanAddPrivateElement = func() {
		hookCalls++
	}

	object := NewObject(agent, nil, "Object")
	privateName := PrivateName{Symbol: agent.CreateSymbol("field")}
	want := NewStringValue("value")

	object.PrivateFieldAdd(privateName, want)

	if hookCalls != 1 {
		t.Fatalf("HostEnsureCanAddPrivateElement call count = %d, want 1", hookCalls)
	}
	element := object.PrivateElementFind(privateName)
	if element == nil {
		t.Fatal("private field was not stored")
	}
	if element.Kind != PrivateElementKindField {
		t.Fatalf("private element kind = %v, want field", element.Kind)
	}
	if element.Value != want {
		t.Fatalf("private field value = %v, want %v", element.Value, want)
	}
}

func newThrowingBuiltin(agent *Agent, reason Value) ObjectType {
	var behavior BehaviorFn = func(Value, []Value, ObjectType) CompletionConvertable[Value] {
		return CompletionValue{t: CompletionTypeThrow, err: reason}
	}
	return CreateBuiltinFunction(
		agent,
		behavior,
		0,
		CMString("throwing"),
		builtinFunctionArgs{realm: agent.CurrentRealm()},
	)
}

func TestGetCompletionPropagatesAccessorFailure(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	reason := NewStringValue("accessor failure")
	object := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
	object.DefinePropertyOrThrow(NewStringPropertyKey("valueOf"), &PropertyDescriptor{
		Get:             newThrowingBuiltin(agent, reason),
		GetSet:          true,
		Configurable:    true,
		ConfigurableSet: true,
	})
	object.DefinePropertyOrThrow(NewStringPropertyKey("toString"), &PropertyDescriptor{
		Get:             newThrowingBuiltin(agent, reason),
		GetSet:          true,
		Configurable:    true,
		ConfigurableSet: true,
	})

	result := object.OrdinaryToPrimitive(PreferredTypeNumber)
	if !result.IsAbrupt() {
		t.Fatal("throwing valueOf getter produced a normal completion")
	}
	if result.Error() != reason {
		t.Fatalf("getter failure = %v, want original reason %v", result.Error(), reason)
	}

	stringResult := ToStringCompletion(agent, object.ToValue())
	if !stringResult.IsAbrupt() || stringResult.Error() != reason {
		t.Fatalf("ToStringCompletion getter failure = %#v, want %v", stringResult, reason)
	}
}

func TestGetCompletionPropagatesProxyTrapFailure(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	reason := NewStringValue("proxy get failure")
	target := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
	handler := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
	handler.CreateDataPropertyOrThrow(
		NewStringPropertyKey("get"),
		newThrowingBuiltin(agent, reason).ToValue(),
	)
	proxy := ProxyCreate(agent, target.ToValue(), handler.ToValue())

	result := proxy.GetCompletion(NewStringPropertyKey("property"))
	if !result.IsAbrupt() {
		t.Fatal("throwing Proxy get trap produced a normal completion")
	}
	if result.Error() != reason {
		t.Fatalf("Proxy trap failure = %v, want original reason %v", result.Error(), reason)
	}

	// ToString first performs @@toPrimitive lookup. That lookup must retain the
	// same Proxy trap throw rather than escaping through a panic-style helper.
	stringResult := ToStringCompletion(agent, proxy.ToValue())
	if !stringResult.IsAbrupt() || stringResult.Error() != reason {
		t.Fatalf("ToStringCompletion proxy failure = %#v, want %v", stringResult, reason)
	}
}
