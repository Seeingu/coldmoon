package coldmoon

import (
	"reflect"
	"testing"
)

func TestTokenizerRecognizesPrivateIdentifier(t *testing.T) {
	tokenizer := NewTokenizer("#secret")
	if tokenizer.CurrentToken.Type != TPrivateIdentifier {
		t.Fatalf("token type = %v, want TPrivateIdentifier", tokenizer.CurrentToken.Type)
	}
	if tokenizer.CurrentToken.Value != "#secret" {
		t.Fatalf("token value = %q, want #secret", tokenizer.CurrentToken.Value)
	}
}

func TestParserBuildsPrivateClassElementsAndOptionalAccess(t *testing.T) {
	script := NewParser(`
class Secret {
    #value;
    #method() { return this.#value; }
    get #access() { return this.#value; }
    set #access(value) { this.#value = value; }
    inspect(other) { return other?.#value; }
}
`, ParserContext{}).Parse()
	declaration := script.StatementList[0].(*StatementListItemDeclaration).Declaration.(*ClassDeclaration)
	body := declaration.ClassTail.ClassBody
	wantNames := []PrivateIdentifierName{"#value", "#method", "#access", "#access"}
	if got := body.PrivateBoundIdentifiers(); !reflect.DeepEqual(got, wantNames) {
		t.Fatalf("PrivateBoundIdentifiers() = %v, want %v", got, wantNames)
	}

	method := body.ClassElementList.Items[1].(*ClassElementMethodDefinition).MethodDefinition
	if privateName, ok := method.PropertyName.(*PropertyNamePrivateIdentifier); !ok || privateName.Identifier != "#method" {
		t.Fatalf("private method property = %#v", method.PropertyName)
	}

	inspect := body.ClassElementList.Items[4].(*ClassElementMethodDefinition).MethodDefinition
	returnStatement := inspect.FunctionExpression.Body.StatementList[0].(*StatementListItemStatement).Statement.(*ReturnStatement)
	optional := returnStatement.Expression.(*OptionalExpression)
	if len(optional.Properties) != 1 || optional.Properties[0].PrivateIdentifier != "#value" {
		t.Fatalf("optional private chain = %#v", optional.Properties)
	}
	if got := optional.String(); got != "other?.#value" {
		t.Fatalf("optional private chain String() = %q", got)
	}
}

func TestParserRejectsPrivateNamesOutsideClassElements(t *testing.T) {
	for _, source := range []string{
		`const invalid = { #value() {} };`,
		`class Invalid { #constructor() {} }`,
	} {
		t.Run(source, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("parse unexpectedly succeeded")
				}
			}()
			NewParser(source, ParserContext{}).Parse()
		})
	}
}

func TestResolvePrivateIdentifierWalksOuterEnvironments(t *testing.T) {
	agent := NewAgent()
	symbol := agent.CreateSymbol("#outer")
	symbol.IsPrivate = true
	want := PrivateName{Symbol: symbol}
	outer := NewPrivateEnvironment(nil)
	outer.Names = append(outer.Names, want)
	inner := NewPrivateEnvironment(outer)

	if got := ResolvePrivateIdentifier(inner, "#outer"); !got.Equal(want) {
		t.Fatalf("ResolvePrivateIdentifier() = %#v, want %#v", got, want)
	}
}
