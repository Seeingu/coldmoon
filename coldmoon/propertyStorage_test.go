package coldmoon

import (
	"reflect"
	"testing"
)

func TestPropertyStorageOwnKeyOrder(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	firstSymbol := NewSymbolPropertyKey(agent.CreateSymbol("first"))
	secondSymbol := NewSymbolPropertyKey(agent.CreateSymbol("second"))
	storage := NewPropertyStorage()
	descriptor := func() *PropertyDescriptor {
		return NewFrozenPropertyDescriptor(UndefinedValue)
	}

	storage.Set(NewStringPropertyKey("beta"), descriptor())
	storage.Set(NewStringPropertyKey("2"), descriptor())
	storage.Set(NewStringPropertyKey("1"), descriptor())
	storage.Set(NewStringPropertyKey("alpha"), descriptor())
	storage.Set(firstSymbol, descriptor())
	storage.Set(NewStringPropertyKey("01"), descriptor())
	storage.Set(secondSymbol, descriptor())

	want := []PropertyKey{
		NewStringPropertyKey("1"),
		NewStringPropertyKey("2"),
		NewStringPropertyKey("beta"),
		NewStringPropertyKey("alpha"),
		NewStringPropertyKey("01"),
		firstSymbol,
		secondSymbol,
	}
	if got := storage.OrderedKeys(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ordered keys = %#v, want %#v", got, want)
	}

	storage.Delete(NewStringPropertyKey("beta"))
	storage.Set(NewStringPropertyKey("beta"), descriptor())
	want = []PropertyKey{
		NewStringPropertyKey("1"),
		NewStringPropertyKey("2"),
		NewStringPropertyKey("alpha"),
		NewStringPropertyKey("01"),
		NewStringPropertyKey("beta"),
		firstSymbol,
		secondSymbol,
	}
	if got := storage.OrderedKeys(); !reflect.DeepEqual(got, want) {
		t.Fatalf("reinserted key order = %#v, want %#v", got, want)
	}
}
