package coldmoon

import "testing"

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
