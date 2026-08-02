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
	source := Source{Name: filePath, Kind: SourceModule}
	var evaluationErr error
	defer func() {
		if recovered := recover(); recovered != nil {
			switch recovered := recovered.(type) {
			case *parseFailure:
				evaluationErr = syntaxDiagnostic(source, recovered)
			case Value:
				evaluationErr = thrownDiagnostic(DiagnosticRuntime, source, recovered, nil)
			default:
				panic(recovered)
			}
		}
		// This outer boundary also covers root loading and parsing, which occur
		// before evaluateModuleRecord takes ownership of scheduler draining.
		if scheduledFailure := agent.Scheduler.drainUntilIdle(); evaluationErr == nil && scheduledFailure != nil {
			evaluationErr = thrownDiagnostic(DiagnosticRuntime, source, scheduledFailure, nil)
		}
		legacyFatal(evaluationErr)
	}()

	loaded := agent.ModuleGraph.Load(realm.ToReferrer(), filePath, HostDefined{})
	if loaded.IsAbrupt() {
		thrown := loaded.Error()
		if thrown == nil {
			panic("module graph returned an invalid abrupt completion")
		}
		evaluationErr = thrownDiagnostic(DiagnosticLink, source, thrown, nil)
		return
	}
	module := loaded.Data().(*SourceTextModule)
	_, evaluationErr = evaluateModuleRecord(source, realm, module)
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
