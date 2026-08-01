package coldmoon

import "testing"

func recordingCallable(agent *Agent, call func([]Value) CompletionValue) ObjectType {
	object := NewObject(agent, nil, "RecordingFunction")
	object.internalMethods().Call = func(_ ObjectType, _ Value, arguments []Value) CompletionValue {
		return call(arguments)
	}
	return object
}

func TestPromiseReactionJobPreservesRejectionCompletion(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	realm := &Realm{Agent: agent, Intrinsics: &Intrinsics{}}
	context := &ExecutionContext{Realm: realm}
	agent.resumeExecutionContext(context)
	defer agent.suspendExecutionContext(context)

	reason := NewStringValue("reason")
	tests := []struct {
		name    string
		handler *JobCallback
	}{
		{name: "missing rejection handler"},
		{
			name: "throwing handler",
			handler: &JobCallback{Callback: recordingCallable(agent, func([]Value) CompletionValue {
				return CompletionValue{t: CompletionTypeThrow, err: reason}
			})},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var resolved Value
			var rejected Value
			capability := &PromiseCapability{
				Resolve: recordingCallable(agent, func(arguments []Value) CompletionValue {
					resolved = argumentAt(arguments, 0)
					return UndefinedValue.ToCompletion()
				}),
				Reject: recordingCallable(agent, func(arguments []Value) CompletionValue {
					rejected = argumentAt(arguments, 0)
					return UndefinedValue.ToCompletion()
				}),
			}
			reaction := &PromiseReaction{
				Capability: capability,
				Type:       PromiseReactionTypeReject,
				Handler:    test.handler,
			}

			job := NewPromiseReactionJob(agent, reaction, reason)
			job.Job.Fun(job.Job.Captures)

			if resolved != nil {
				t.Fatalf("rejection reaction resolved with %v", resolved)
			}
			if rejected != reason {
				t.Fatalf("rejection reason = %v, want %v", rejected, reason)
			}
		})
	}
}

func TestAsyncGeneratorCompleteStepUsesProvidedRealmAndRestoresContext(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	currentPrototype := NewObject(agent, nil, "CurrentObjectPrototype")
	resultPrototype := NewObject(agent, nil, "ResultObjectPrototype")
	currentRealm := &Realm{
		Agent:      agent,
		Intrinsics: &Intrinsics{ObjectPrototype: currentPrototype},
	}
	resultRealm := &Realm{
		Agent:      agent,
		Intrinsics: &Intrinsics{ObjectPrototype: resultPrototype},
	}
	context := &ExecutionContext{Realm: currentRealm}
	agent.resumeExecutionContext(context)
	defer agent.suspendExecutionContext(context)

	var resolved ObjectType
	var realmDuringResolve *Realm
	capability := &PromiseCapability{
		Resolve: recordingCallable(agent, func(arguments []Value) CompletionValue {
			resolved = MustGetObject(argumentAt(arguments, 0))
			realmDuringResolve = agent.CurrentRealm()
			return UndefinedValue.ToCompletion()
		}),
		Reject: recordingCallable(agent, func([]Value) CompletionValue {
			t.Fatal("normal completion unexpectedly rejected")
			return UndefinedValue.ToCompletion()
		}),
	}
	generator := &AsyncGeneratorObject{Object: NewObject(agent, nil, "AsyncGenerator")}
	generator.ref = generator
	generator.AsyncGeneratorQueue.Enqueue(AsyncGeneratorRequest{Capability: capability})

	completion := NewStringValue("value").ToCompletion()
	AsyncGeneratorCompleteStep(agent, generator, completion, false, resultRealm)

	if context.Realm != currentRealm {
		t.Fatal("AsyncGeneratorCompleteStep did not restore the running context realm")
	}
	if realmDuringResolve != currentRealm {
		t.Fatal("promise resolve ran before the original realm was restored")
	}
	if resolved == nil {
		t.Fatal("promise was not resolved with an iterator result")
	}
	if resolved.Prototype() != resultPrototype {
		t.Fatal("iterator result was not created in the provided realm")
	}
}

func TestAsyncGeneratorCompleteStepRejectsWithThrowReason(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	reason := NewStringValue("reason")
	var rejected Value
	capability := &PromiseCapability{
		Resolve: recordingCallable(agent, func([]Value) CompletionValue {
			t.Fatal("throw completion unexpectedly resolved")
			return UndefinedValue.ToCompletion()
		}),
		Reject: recordingCallable(agent, func(arguments []Value) CompletionValue {
			rejected = argumentAt(arguments, 0)
			return UndefinedValue.ToCompletion()
		}),
	}
	generator := &AsyncGeneratorObject{Object: NewObject(agent, nil, "AsyncGenerator")}
	generator.ref = generator
	generator.AsyncGeneratorQueue.Enqueue(AsyncGeneratorRequest{Capability: capability})

	AsyncGeneratorCompleteStep(
		agent,
		generator,
		CompletionValue{t: CompletionTypeThrow, err: reason},
		true,
		nil,
	)

	if rejected != reason {
		t.Fatalf("rejection reason = %v, want %v", rejected, reason)
	}
}

func TestAsyncGeneratorDrainQueueRejectsThrowRequests(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	reason := NewStringValue("reason")
	var rejected Value
	capability := &PromiseCapability{
		Resolve: recordingCallable(agent, func([]Value) CompletionValue {
			t.Fatal("throw request unexpectedly resolved")
			return UndefinedValue.ToCompletion()
		}),
		Reject: recordingCallable(agent, func(arguments []Value) CompletionValue {
			rejected = argumentAt(arguments, 0)
			return UndefinedValue.ToCompletion()
		}),
	}
	generator := &AsyncGeneratorObject{
		Object:              NewObject(agent, nil, "AsyncGenerator"),
		AsyncGeneratorState: AsyncGeneratorStateCompleted,
	}
	generator.ref = generator
	generator.AsyncGeneratorQueue.Enqueue(AsyncGeneratorRequest{
		Completion: CompletionValue{t: CompletionTypeThrow, err: reason},
		Capability: capability,
	})

	AsyncGeneratorDrainQueue(agent, generator)

	if rejected != reason {
		t.Fatalf("rejection reason = %v, want %v", rejected, reason)
	}
	if !generator.AsyncGeneratorQueue.IsEmpty() {
		t.Fatal("drained async-generator queue is not empty")
	}
}
