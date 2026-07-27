package coldmoon

func fatalOnError(result Value) {
	if o, ok := result.GetObject(); ok {
		if e, ok := o.(*ErrorObject); ok {
			println("Return Error: ", e.Message)
			panic(e)
		}
	}
}

func EvaluateModule(filePath string, realm *Realm) {
	agent := realm.Agent
	loaded := agent.ModuleGraph.Load(realm.ToReferrer(), filePath, HostDefined{})
	if loaded.IsAbrupt() {
		fatalOnError(loaded.Error())
		return
	}
	module := loaded.Data().(*SourceTextModule)
	var result Value
	p := module.LoadRequestedModules()
	switch p.PromiseState {
	case PromiseStatePending:
		panic("unreachable")
	case PromiseStateFulfilled:
		module.Link()
		if agent.exception != nil {
			fatalOnError(agent.exception)
		}
		p = module.Evaluate()
		switch p.PromiseState {
		case PromiseStatePending:
			panic("unreachable")
		case PromiseStateFulfilled:
			result = UndefinedValue
		case PromiseStateRejected:
			result = p.PromiseResult
		}
	case PromiseStateRejected:
		result = p.PromiseResult
	}
	fatalOnError(result)
	agent.Scheduler.RunUntilIdle()
}

func Evaluate(source string, realm *Realm) {
	agent := realm.Agent
	result := ParseScript(source, realm, nil).Evaluate()
	if o, ok := result.GetObject(); ok {
		if e, ok := o.(*ErrorObject); ok {
			println("Return Error: ", e.Message)
			panic(e)
		}
	}
	agent.Scheduler.RunUntilIdle()
}
