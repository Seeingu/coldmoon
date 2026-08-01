package coldmoon

import "testing"

func TestCreatePerIterationEnvironmentCopiesBindingCells(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	outer := realm.GlobalEnv
	previous := NewDeclarativeEnvironment(outer)
	previous.CreateMutableBinding("index", false)
	previous.InitializeBinding("index", NewNumberValue(1))

	context := &ExecutionContext{
		Realm: realm,
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  previous,
			VariableEnvironment: previous,
		},
	}
	scope := agent.enterExecutionContext(context)
	defer scope.Leave()

	vm := context.VM
	result := vm.CreatePerIterationEnvironment([]string{"index"})
	if result.IsAbrupt() {
		t.Fatalf("CreatePerIterationEnvironment returned %v", result.Error())
	}

	current, ok := vm.RunningLexicalEnvironment().(*DeclarativeEnvironment)
	if !ok || current == previous {
		t.Fatalf("current environment = %#v, want a fresh declarative environment", current)
	}
	if current.OuterEnv() != outer {
		t.Fatal("fresh iteration environment did not retain the loop environment's outer environment")
	}
	current.SetMutableBinding("index", NewNumberValue(2), true)
	if got := ReturnAssertNormal(previous.GetBindingValue(agent, "index", true)).(*NumberValue).Data; got != 1 {
		t.Fatalf("previous iteration binding = %v, want 1", got)
	}
	if got := ReturnAssertNormal(current.GetBindingValue(agent, "index", true)).(*NumberValue).Data; got != 2 {
		t.Fatalf("current iteration binding = %v, want 2", got)
	}
}

func TestCreatePerIterationEnvironmentNoBindingsIsNoOp(t *testing.T) {
	agent, realm := newSourceTextModuleTestRealm(t)
	env := NewDeclarativeEnvironment(realm.GlobalEnv)
	context := &ExecutionContext{
		Realm: realm,
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  env,
			VariableEnvironment: env,
		},
	}
	scope := agent.enterExecutionContext(context)
	defer scope.Leave()

	result := context.VM.CreatePerIterationEnvironment(nil)
	if result.IsAbrupt() {
		t.Fatalf("CreatePerIterationEnvironment returned %v", result.Error())
	}
	if context.VM.RunningLexicalEnvironment() != env {
		t.Fatal("empty binding list replaced the running lexical environment")
	}
}
