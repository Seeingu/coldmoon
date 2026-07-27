package coldmoon

import "testing"

type vmCaptureNode struct {
	seen **VM
}

func (n *vmCaptureNode) String() string {
	return "vmCaptureNode"
}

func (n *vmCaptureNode) Evaluation(vm *VM) CompletionValue {
	*n.seen = vm
	return UndefinedValue.ToCompletion()
}

func TestExecutionContextOwnsVMForItsEntireLifetime(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	context := &ExecutionContext{
		Realm: realm,
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  realm.GlobalEnv,
			VariableEnvironment: realm.GlobalEnv,
		},
	}
	scope := agent.enterExecutionContext(context)
	defer scope.Leave()

	var first, second *VM
	RunNode(agent, &vmCaptureNode{seen: &first})
	RunNode(agent, &vmCaptureNode{seen: &second})

	if first == nil || first != second || first != context.VM {
		t.Fatalf("RunNode VMs = (%p, %p), context VM = %p", first, second, context.VM)
	}
}

func TestExecutionContextScopeRestoresStackAfterEarlyReturn(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	base := agent.RunningExecutionContext()

	func() {
		scope := agent.enterExecutionContext(&ExecutionContext{Realm: realm})
		defer scope.Leave()
	}()

	if agent.RunningExecutionContext() != base {
		t.Fatal("execution context scope did not restore the caller")
	}
}
