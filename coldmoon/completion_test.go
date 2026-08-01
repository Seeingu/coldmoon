package coldmoon

import "testing"

func TestCompletionFromPreservesAbruptControlFlow(t *testing.T) {
	source := Completion[int]{
		t:      CompletionTypeContinue,
		target: "loop",
	}

	converted := CompletionFrom(Completion[string]{}, source)

	if converted.t != CompletionTypeContinue {
		t.Fatalf("completion type = %v, want continue", converted.t)
	}
	if converted.target != "loop" {
		t.Fatalf("completion target = %q, want loop", converted.target)
	}
}

func TestReturnIfAbruptPreservesSameTypedReturnValue(t *testing.T) {
	want := NewStringValue("returned")
	completion := CompletionValue{
		t:     CompletionTypeReturn,
		value: want,
	}

	_, abrupt, propagated := ReturnIfAbrupt(completion, CompletionValue{})
	if !abrupt {
		t.Fatal("return completion was not abrupt")
	}
	if propagated.t != CompletionTypeReturn {
		t.Fatalf("completion type = %v, want return", propagated.t)
	}
	if propagated.value != want {
		t.Fatalf("return value = %v, want %v", propagated.value, want)
	}
}

func TestUpdateEmptyPreservesAbruptCompletion(t *testing.T) {
	result := CompletionValue{
		t:      CompletionTypeContinue,
		target: "outer",
	}
	updated := UpdateEmpty(result, UndefinedValue)
	if updated.t != CompletionTypeContinue {
		t.Fatalf("completion type = %v, want continue", updated.t)
	}
	if updated.target != "outer" {
		t.Fatalf("target = %q, want outer", updated.target)
	}
	if updated.value != UndefinedValue {
		t.Fatalf("value = %#v, want undefined", updated.value)
	}
}

func TestBuiltinBoundaryConvertsThrownValuesAndNilResults(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)

	throws := CreateBuiltinFunction(agent, func(Value, []Value, ObjectType) CompletionConvertable[Value] {
		panic(agent.ThrowException(SyntaxError, "builtin failure"))
	}, 0, CMString("throws"), builtinFunctionArgs{})
	thrown := throws.Call(UndefinedValue, nil)
	if thrown.t != CompletionTypeThrow || thrown.Error() == nil {
		t.Fatalf("thrown completion = %#v, want a throw with an error value", thrown)
	}

	returnsEmpty := CreateBuiltinFunction(agent, func(Value, []Value, ObjectType) CompletionConvertable[Value] {
		return nil
	}, 0, CMString("returnsEmpty"), builtinFunctionArgs{})
	empty := returnsEmpty.Call(UndefinedValue, nil)
	if empty.IsAbrupt() || empty.Data() != UndefinedValue {
		t.Fatalf("nil builtin result = %#v, want normal undefined", empty)
	}
}
