package coldmoon

import "errors"

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
	_, err := evaluateModuleRecord(Source{Name: filePath, Kind: SourceModule}, realm, module)
	legacyFatal(err)
}

func Evaluate(source string, realm *Realm) {
	_, err := EvaluateSource(Source{
		Text: source,
		Name: "file.js",
		Kind: SourceScript,
	}, realm)
	legacyFatal(err)
}

func legacyFatal(err error) {
	if err == nil {
		return
	}
	var diagnostic *Diagnostic
	if errors.As(err, &diagnostic) && diagnostic.Thrown != nil {
		// Preserve the embedding entry points' historical behavior: Error
		// objects panic, while primitive throws are returned only by the safe
		// interface.
		fatalOnError(diagnostic.Thrown)
		return
	}
	panic(err)
}
