package coldmoon

// 19.2.1.1
func PerformEval(agent *Agent, x Value, strictCaller bool, direct bool) Value {
	if !direct {
		Assert(!strictCaller)
	}

	stringValue, ok := x.(*StringValue)
	if !ok {
		return x
	}

	evalRealm := agent.CurrentRealm()

	agent.HostHooks.HostEnsureCanCompileStrings(evalRealm)

	script := ParseScript(stringValue.Data, evalRealm, nil)
	if len(script.ECMAScriptCode.StatementList) == 0 {
		return nil
	}

	strictEval := strictCaller || script.ECMAScriptCode.IsStrict()

	runningContext := agent.runningExecutionContext()
	var lexEnv EnvironmentRecord
	var varEnv EnvironmentRecord
	var privateEnv *PrivateEnvironment

	if direct {
		lexEnv = NewDeclarativeEnvironment(runningContext.ECMAScriptCode.LexicalEnvironment)
		varEnv = runningContext.ECMAScriptCode.VariableEnvironment
		privateEnv = runningContext.ECMAScriptCode.PrivateEnvironment
	} else {
		lexEnv = NewDeclarativeEnvironment(evalRealm.GlobalEnv)
		varEnv = evalRealm.GlobalEnv
		privateEnv = nil
	}
	if strictEval {
		varEnv = lexEnv
	}

	evalContext := &ExecutionContext{
		Realm:          evalRealm,
		ScriptOrModule: runningContext.ScriptOrModule,
		Function:       nil,
		ECMAScriptCode: &ExecutionContextAdditionalState{
			LexicalEnvironment:  lexEnv,
			VariableEnvironment: varEnv,
			PrivateEnvironment:  privateEnv,
		},
	}

	agent.ExecutionContextStack.Push(evalContext)

	result := script.Evaluate()
	agent.ExecutionContextStack.Pop()

	return result
}
