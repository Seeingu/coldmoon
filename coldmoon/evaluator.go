package coldmoon

func Evaluate(source string, realm *Realm) {
	agent := realm.Agent
	result := ParseScript(source, realm, nil).Evaluate()
	if o, ok := ValueGetObject(result); ok {
		if e, ok := o.(*ErrorObject); ok {
			println("Return Error: ", e.Message)
			panic(e)
		}
	}
	for _, job := range agent.QueuedPromiseJobs.Data() {
		previousRealm := agent.runningExecutionContext().Realm
		agent.runningExecutionContext().Realm = job.realm
		job.job.Fun(job.job.Captures)
		agent.runningExecutionContext().Realm = previousRealm
	}
}
