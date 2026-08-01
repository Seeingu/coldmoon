package coldmoon

import "testing"

func TestEncodeURIPreservesReservedCharactersAndRejectsLoneSurrogate(t *testing.T) {
	realm := newURITestRealm(t)
	encodeURI := NewEncodeURI(realm)

	assertURINormalResult(
		t,
		encodeURI.Call(UndefinedValue, []Value{NewStringValue("https://user@example.com/a b;v?x=✓&y=1+$,#片")}),
		"https://user@example.com/a%20b;v?x=%E2%9C%93&y=1+$,#%E7%89%87",
	)
	assertURIErrorCompletion(
		t,
		realm,
		encodeURI.Call(UndefinedValue, []Value{NewStringValue(string([]byte{0xed, 0xa0, 0x80}))}),
	)
}

func TestEncodeURIComponentEscapesReservedCharactersAndRejectsLoneSurrogate(t *testing.T) {
	realm := newURITestRealm(t)
	encodeURIComponent := NewEncodeURIComponent(realm)

	assertURINormalResult(
		t,
		encodeURIComponent.Call(UndefinedValue, []Value{NewStringValue(";/?:@&=+$,# %✓")}),
		"%3B%2F%3F%3A%40%26%3D%2B%24%2C%23%20%25%E2%9C%93",
	)
	assertURINormalResult(
		t,
		encodeURIComponent.Call(UndefinedValue, []Value{NewStringValue("AZaz09-_.!~*'()")}),
		"AZaz09-_.!~*'()",
	)
	assertURINormalResult(
		t,
		encodeURIComponent.Call(UndefinedValue, []Value{NewStringValue("�")}),
		"%EF%BF%BD",
	)
	assertURIErrorCompletion(
		t,
		realm,
		encodeURIComponent.Call(UndefinedValue, []Value{NewStringValue(string([]byte{0xed, 0xb0, 0x80}))}),
	)
}

func TestDecodeURIPreservesReservedEscapesAndRejectsMalformedUTF8(t *testing.T) {
	realm := newURITestRealm(t)
	decodeURI := NewDecodeURI(realm)

	assertURINormalResult(
		t,
		decodeURI.Call(UndefinedValue, []Value{NewStringValue("https://example.com/a%20b%3Fx=%E2%9C%93%23frag")}),
		"https://example.com/a b%3Fx=✓%23frag",
	)
	assertURINormalResult(
		t,
		decodeURI.Call(UndefinedValue, []Value{NewStringValue("%3B%2f%3F%3a%40%26%3D%2b%24%2C%23")}),
		"%3B%2f%3F%3a%40%26%3D%2b%24%2C%23",
	)
	for _, malformed := range []string{"%", "%GG", "%C0%AF", "%ED%A0%80"} {
		t.Run(malformed, func(t *testing.T) {
			assertURIErrorCompletion(
				t,
				realm,
				decodeURI.Call(UndefinedValue, []Value{NewStringValue(malformed)}),
			)
		})
	}
}

func TestDecodeURIComponentDecodesReservedEscapesAndRejectsMalformedUTF8(t *testing.T) {
	realm := newURITestRealm(t)
	decodeURIComponent := NewDecodeURIComponent(realm)

	assertURINormalResult(
		t,
		decodeURIComponent.Call(UndefinedValue, []Value{NewStringValue("a%20b%3Fx%3D%E2%9C%93%23")}),
		"a b?x=✓#",
	)
	for _, malformed := range []string{"%E2%28%A1", "%F4%90%80%80", "%80"} {
		t.Run(malformed, func(t *testing.T) {
			assertURIErrorCompletion(
				t,
				realm,
				decodeURIComponent.Call(UndefinedValue, []Value{NewStringValue(malformed)}),
			)
		})
	}
}

func newURITestRealm(t *testing.T) *Realm {
	t.Helper()
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	return agent.CurrentRealm()
}

func assertURINormalResult(t *testing.T, result CompletionValue, want string) {
	t.Helper()
	if result.IsAbrupt() {
		t.Fatalf("URI function threw %v", result.Error())
	}
	if got := result.Data().String(); got != want {
		t.Fatalf("URI function result = %q, want %q", got, want)
	}
}

func assertURIErrorCompletion(t *testing.T, realm *Realm, result CompletionValue) {
	t.Helper()
	if !result.IsAbrupt() || result.Error() == nil {
		t.Fatalf("URI function result = %v, want URIError completion", result.Data())
	}
	errorObject, ok := result.Error().GetObject()
	if !ok {
		t.Fatalf("URI function error = %T, want object", result.Error())
	}
	if errorObject.Prototype() != realm.Intrinsics.URIErrorPrototype {
		t.Fatalf("URI function error prototype = %v, want URIError.prototype", errorObject.Prototype())
	}
}
