package coldmoon

import (
	"testing"
	"time"
)

func TestAtomicsWaitIsReleasedByNotify(t *testing.T) {
	agent := NewAgent()
	InitializeConstants()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	shared := AllocateSharedArrayBuffer(agent, realm.Intrinsics.SharedArrayBufferConstructor, 4, 0)
	typedArray := TypedArrayCreate(agent, TypedArrayNameInt32, realm.Intrinsics.Int32ArrayPrototype)
	initialized := InitializeTypedArrayFromArrayBuffer(
		agent,
		typedArray,
		NewArrayBufferLike(shared),
		UndefinedValue,
		UndefinedValue,
	)
	if initialized.IsAbrupt() {
		t.Fatal("initialize Int32Array over SharedArrayBuffer")
	}

	result := make(chan CompletionValue, 1)
	go func() {
		result <- atomicWaitOperation(
			agent,
			typedArray.ToValue(),
			NewNumberValue(0),
			NewNumberValue(0),
			NewNumberValue(5_000),
		)
	}()

	location := atomicWaitLocation{block: typedArray.ViewedArrayBuffer.Data(), byteIndex: 0}
	deadline := time.Now().Add(2 * time.Second)
	for {
		atomicsWaiters.Lock()
		waiting := len(atomicsWaiters.queues[location]) == 1
		atomicsWaiters.Unlock()
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Atomics.wait did not enter its waiter queue")
		}
		time.Sleep(time.Millisecond)
	}

	notified := atomicNotifyOperation(
		agent,
		typedArray.ToValue(),
		NewNumberValue(0),
		NewNumberValue(1),
	)
	if notified.IsAbrupt() || notified.Data().String() != "1" {
		t.Fatalf("Atomics.notify result = %v, want 1", notified.Data())
	}

	select {
	case completion := <-result:
		if completion.IsAbrupt() || completion.Data().String() != "ok" {
			t.Fatalf("Atomics.wait result = %v, want ok", completion.Data())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Atomics.wait was not released by Atomics.notify")
	}
}
