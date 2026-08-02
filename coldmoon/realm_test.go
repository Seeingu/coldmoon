package coldmoon

import (
	"math/rand"
	"strings"
	"testing"
)

func TestCreateRealmPublishesCompleteIntrinsics(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	realm := CreateRealm(agent)

	if !realm.IsReady() {
		t.Fatal("CreateRealm returned a Realm that is not ready")
	}
	if missing := missingIntrinsicFields(realm.Intrinsics); len(missing) != 0 {
		t.Fatalf("published Realm has missing intrinsics: %v", missing)
	}
	for _, binding := range globalIntrinsicBindings {
		if realm.Intrinsics.Get(binding.intrinsic) == nil {
			t.Fatalf("global binding %q resolves to a nil intrinsic", binding.name)
		}
	}
}

// TestCreateRealmInitializesIndependentReplaceableRandomSources protects both
// production diversity and the deterministic RNG seam exposed to embedders.
func TestCreateRealmInitializesIndependentReplaceableRandomSources(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	first := CreateRealm(agent)
	second := CreateRealm(agent)

	allEqual := true
	for range 8 {
		firstValue := first.Rng.Float64()
		secondValue := second.Rng.Float64()
		if firstValue < 0 || firstValue >= 1 || secondValue < 0 || secondValue >= 1 {
			t.Fatalf("Realm random values out of range: %v, %v", firstValue, secondValue)
		}
		allEqual = allEqual && firstValue == secondValue
	}
	if allEqual {
		t.Fatal("independently created Realms shared the same random sequence")
	}

	first.Rng = *rand.New(rand.NewSource(42))
	want := first.Rng.Float64()
	first.Rng = *rand.New(rand.NewSource(42))
	if got := first.Rng.Float64(); got != want {
		t.Fatalf("replacement random source produced %v, want deterministic %v", got, want)
	}
}

func TestBuildingRealmCannotPublishGlobalState(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	draft := &Realm{
		Agent:      agent,
		Intrinsics: &Intrinsics{},
		state:      realmStateBuilding,
	}

	defer func() {
		if recover() == nil {
			t.Fatal("building Realm accepted a global object")
		}
	}()
	draft.SetRealmGlobalObject(nil, nil)
}

func TestIntrinsicBootstrapPhaseRejectsMissingDependencies(t *testing.T) {
	draft := &Realm{Intrinsics: &Intrinsics{}, state: realmStateBuilding}

	deferredPanic := ""
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				deferredPanic = recovered.(string)
			}
		}()
		draft.createIterationAndCallableIntrinsics()
	}()
	if !strings.Contains(deferredPanic, `phase "iteration and callable" requires ObjectPrototype`) {
		t.Fatalf("out-of-order phase panic = %q", deferredPanic)
	}
}
