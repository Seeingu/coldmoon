package coldmoon

import (
	"path"

	"github.com/Seeingu/coldmoon/pkg"
)

func fatalOnError(result Value) {
	if o, ok := result.GetObject(); ok {
		if e, ok := o.(*ErrorObject); ok {
			println("Return Error: ", e.Message)
			panic(e)
		}
	}
}

// TODO: resolve path
func EvaluateModule(filePath string, realm *Realm) {
	agent := realm.Agent
	hostDefined := HostDefined{
		FileName: path.Base(filePath),
		BaseDir:  path.Dir(filePath),
	}
	module := ParseModule(pkg.MustReadFile(filePath), realm, hostDefined)
	var result Value
	p := module.LoadRequestedModules(hostDefined)
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
	for _, job := range agent.QueuedPromiseJobs.Data() {
		previousRealm := agent.RunningExecutionContext().Realm
		agent.RunningExecutionContext().Realm = job.realm
		job.job.Fun(job.job.Captures)
		agent.RunningExecutionContext().Realm = previousRealm
	}
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
	for _, job := range agent.QueuedPromiseJobs.Data() {
		previousRealm := agent.RunningExecutionContext().Realm
		agent.RunningExecutionContext().Realm = job.realm
		job.job.Fun(job.job.Captures)
		agent.RunningExecutionContext().Realm = previousRealm
	}
}
