package coldmoon

import (
	"math"
	"math/big"
	"testing"
)

// TestValueHashImplementsSameValueZero verifies the key equivalence relation
// required by Map and Set without allowing unlike ECMAScript types to collide.
func TestValueHashImplementsSameValueZero(t *testing.T) {
	InitializeConstants()

	if NewNumberValue(JSNumber(math.NaN())).Hash() != NewNumberValue(JSNumber(math.NaN())).Hash() {
		t.Fatal("NaN values must share a SameValueZero hash")
	}
	if NewNumberValue(0).Hash() != NewNumberValue(JSNumber(math.Copysign(0, -1))).Hash() {
		t.Fatal("positive and negative zero must share a SameValueZero hash")
	}
	if NewNumberValue(1).Hash() == NewStringValue("1").Hash() {
		t.Fatal("number and string keys must not collide")
	}
	if NewNumberValue(1).Hash() == NewBigIntValue(big.NewInt(1)).Hash() {
		t.Fatal("number and bigint keys must not collide")
	}
}

// TestValueHashUsesObjectAndSymbolIdentity verifies that identity-bearing
// language values remain distinct while repeated wrappers stay stable.
func TestValueHashUsesObjectAndSymbolIdentity(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	first := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
	second := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)

	if first.ToValue().Hash() != first.ToValue().Hash() {
		t.Fatal("the same object must have a stable hash")
	}
	if first.ToValue().Hash() == second.ToValue().Hash() {
		t.Fatal("different objects must have different hashes")
	}
	if agent.CreateSymbol("same").Hash() == agent.CreateSymbol("same").Hash() {
		t.Fatal("different symbols with the same description must not collide")
	}
}

// TestWellKnownSymbolNamesCoversEveryKey ensures built-in function metadata can
// name every well-known symbol without reaching a fallback path.
func TestWellKnownSymbolNamesCoversEveryKey(t *testing.T) {
	want := map[WellKnownSymbolsKey]string{
		WellKnownSymbolsAsyncIterator:      "[Symbol.asyncIterator]",
		WellKnownSymbolsHasInstance:        "[Symbol.hasInstance]",
		WellKnownSymbolsIsConcatSpreadable: "[Symbol.isConcatSpreadable]",
		WellKnownSymbolsIterator:           "[Symbol.iterator]",
		WellKnownSymbolsMatch:              "[Symbol.match]",
		WellKnownSymbolsMatchAll:           "[Symbol.matchAll]",
		WellKnownSymbolsReplace:            "[Symbol.replace]",
		WellKnownSymbolsSearch:             "[Symbol.search]",
		WellKnownSymbolsSpecies:            "[Symbol.species]",
		WellKnownSymbolsSplit:              "[Symbol.split]",
		WellKnownSymbolsToPrimitive:        "[Symbol.toPrimitive]",
		WellKnownSymbolsToStringTag:        "[Symbol.toStringTag]",
		WellKnownSymbolsUnscopables:        "[Symbol.unscopables]",
	}

	for key, expected := range want {
		if got := key.ToName(); got != expected {
			t.Errorf("%q.ToName() = %q, want %q", key, got, expected)
		}
	}
}

// TestAnonymousFunctionDefinitionClassification covers the grammar forms that
// participate in named evaluation and their named counterparts.
func TestAnonymousFunctionDefinitionClassification(t *testing.T) {
	tests := []struct {
		name string
		expr Expression
		want bool
	}{
		{name: "arrow", expr: &ArrowFunction{}, want: true},
		{name: "async arrow", expr: &AsyncArrowFunction{}, want: true},
		{name: "anonymous function", expr: &FunctionExpression{}, want: true},
		{name: "named function", expr: &FunctionExpression{Identifier: "named"}, want: false},
		{name: "anonymous generator", expr: &GeneratorExpression{}, want: true},
		{name: "named generator", expr: &GeneratorExpression{IdentifierName: "named"}, want: false},
		{name: "anonymous async function", expr: &AsyncFunctionExpression{}, want: true},
		{name: "named async function", expr: &AsyncFunctionExpression{Identifier: "named"}, want: false},
		{name: "anonymous async generator", expr: &PrimaryExpressionAsyncGeneratorExpression{}, want: true},
		{name: "named async generator", expr: &PrimaryExpressionAsyncGeneratorExpression{IdentifierName: "named"}, want: false},
		{name: "anonymous class", expr: &ClassExpression{}, want: true},
		{name: "named class", expr: &ClassExpression{IdentifierName: "Named"}, want: false},
		{name: "ordinary expression", expr: &IdentifierReference{Identifier: "value"}, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsAnonymousFunctionDefinition(test.expr); got != test.want {
				t.Fatalf("IsAnonymousFunctionDefinition() = %t, want %t", got, test.want)
			}
		})
	}
}

// TestCreateDynamicAsyncFunction verifies that the AsyncFunction constructor
// creates a non-constructor function with the correct intrinsic prototype.
func TestCreateDynamicAsyncFunction(t *testing.T) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	function := CreateDynamicFunction(
		agent,
		realm.Intrinsics.AsyncFunctionConstructor,
		nil,
		dynamicFunctionKindAsync,
		nil,
		NewStringValue("return 1"),
	)

	if function.Prototype() != realm.Intrinsics.AsyncFunctionPrototype {
		t.Fatal("dynamic async function has the wrong prototype")
	}
	if function.propertyStorage().Has(NewStringPropertyKey("prototype")) {
		t.Fatal("async functions must not have a constructor prototype property")
	}
}
