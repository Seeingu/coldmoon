package coldmoon

import (
	"bytes"
	"testing"
)

func newStructuredDataTestRealm() (*Agent, *Realm) {
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	return agent, agent.CurrentRealm()
}

func constructArrayBufferForTest(
	t *testing.T,
	agent *Agent,
	realm *Realm,
	byteLength JSInt,
	maxByteLength *JSInt,
) *ArrayBufferLike {
	t.Helper()
	arguments := []Value{byteLength.ToValue()}
	if maxByteLength != nil {
		options := OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
		options.CreateDataPropertyOrThrow(NewStringPropertyKey("maxByteLength"), maxByteLength.ToValue())
		arguments = append(arguments, options.ToValue())
	}
	result := realm.Intrinsics.ArrayBufferConstructor.Construct(arguments, nil)
	if result.IsAbrupt() {
		t.Fatalf("construct ArrayBuffer(%d) completed abruptly: %v", byteLength, result.Error())
	}
	buffer, ok := result.Data().(*ArrayBufferLike)
	if !ok {
		t.Fatalf("ArrayBuffer constructor returned %T, want *ArrayBufferLike", result.Data())
	}
	return buffer
}

func callArrayBufferResize(t *testing.T, realm *Realm, buffer *ArrayBufferLike, byteLength JSInt) CompletionValue {
	t.Helper()
	resize := realm.Intrinsics.ArrayBufferPrototype.Get(NewStringPropertyKey("resize"))
	return MustGetObject(resize).Call(buffer.ToValue(), []Value{byteLength.ToValue()})
}

// TestArrayBufferResizePreservesTheCommonPrefix verifies both growth and
// shrinkage use a fresh zero-filled block while retaining the shared prefix.
func TestArrayBufferResizePreservesTheCommonPrefix(t *testing.T) {
	agent, realm := newStructuredDataTestRealm()
	maxByteLength := JSInt(8)
	buffer := constructArrayBufferForTest(t, agent, realm, 4, &maxByteLength)
	buffer.Data().Set(0, []byte{1, 2, 3, 4})

	if !buffer.Get(NewStringPropertyKey("resizable")).ToBoolean() {
		t.Fatal("resizable ArrayBuffer reported resizable=false")
	}
	if result := callArrayBufferResize(t, realm, buffer, 6); result.IsAbrupt() {
		t.Fatalf("grow to 6 completed abruptly: %v", result.Error())
	}
	if got := buffer.ByteLength(); got != 6 {
		t.Fatalf("byte length after growth = %d, want 6", got)
	}
	if got, want := buffer.Data().Slice(0, 6), []byte{1, 2, 3, 4, 0, 0}; !bytes.Equal(got, want) {
		t.Fatalf("bytes after growth = %v, want %v", got, want)
	}

	if result := callArrayBufferResize(t, realm, buffer, 2); result.IsAbrupt() {
		t.Fatalf("shrink to 2 completed abruptly: %v", result.Error())
	}
	if got, want := buffer.Data().Slice(0, 2), []byte{1, 2}; !bytes.Equal(got, want) {
		t.Fatalf("bytes after shrink = %v, want %v", got, want)
	}
	if result := callArrayBufferResize(t, realm, buffer, 9); !result.IsAbrupt() {
		t.Fatal("resize beyond maxByteLength succeeded")
	}
}

// TestFixedArrayBufferRejectsResize verifies absence of the resizable internal
// slot is represented by the fixed-length sentinel used by this runtime.
func TestFixedArrayBufferRejectsResize(t *testing.T) {
	agent, realm := newStructuredDataTestRealm()
	buffer := constructArrayBufferForTest(t, agent, realm, 4, nil)
	if buffer.Get(NewStringPropertyKey("resizable")).ToBoolean() {
		t.Fatal("fixed ArrayBuffer reported resizable=true")
	}
	if result := callArrayBufferResize(t, realm, buffer, 2); !result.IsAbrupt() {
		t.Fatal("fixed ArrayBuffer accepted resize")
	}
}

// TestSharedArrayBufferGrowUpdatesLogicalLength verifies a growable shared
// buffer exposes only its logical prefix and uses idempotent CAS growth.
func TestSharedArrayBufferGrowUpdatesLogicalLength(t *testing.T) {
	agent, realm := newStructuredDataTestRealm()
	object := AllocateSharedArrayBuffer(agent, realm.Intrinsics.SharedArrayBufferConstructor, 2, 6)
	buffer := object.(*SharedArrayBufferObject)
	buffer.ArrayBufferData.Set(0, []byte{3, 4})

	if got := ArrayBufferByteLength(NewArrayBufferLike(buffer), SeqCst); got != 2 {
		t.Fatalf("initial shared byte length = %d, want 2", got)
	}
	if result := sharedArrayBufferGrow(agent, buffer.ToValue(), JSInt(4).ToValue()); result.IsAbrupt() {
		t.Fatalf("grow to 4 completed abruptly: %v", result.Error())
	}
	if got := ArrayBufferByteLength(NewArrayBufferLike(buffer), SeqCst); got != 4 {
		t.Fatalf("shared byte length after growth = %d, want 4", got)
	}
	if got, want := buffer.ArrayBufferData.Slice(0, 4), []byte{3, 4, 0, 0}; !bytes.Equal(got, want) {
		t.Fatalf("shared bytes after growth = %v, want %v", got, want)
	}
	if result := sharedArrayBufferGrow(agent, buffer.ToValue(), JSInt(4).ToValue()); result.IsAbrupt() {
		t.Fatalf("idempotent grow completed abruptly: %v", result.Error())
	}
	if result := sharedArrayBufferGrow(agent, buffer.ToValue(), JSInt(3).ToValue()); !result.IsAbrupt() {
		t.Fatal("SharedArrayBuffer.grow accepted a smaller length")
	}
	if result := sharedArrayBufferGrow(agent, buffer.ToValue(), JSInt(7).ToValue()); !result.IsAbrupt() {
		t.Fatal("SharedArrayBuffer.grow accepted a length beyond maxByteLength")
	}
}

// TestCreateListFromArrayLikePropertyKeysRejectsOtherValues verifies the
// property-key mode accepts only String and Symbol elements.
func TestCreateListFromArrayLikePropertyKeysRejectsOtherValues(t *testing.T) {
	agent, _ := newStructuredDataTestRealm()
	arrayLike := CreateArrayFromList(agent, []Value{
		NewStringValue("name"),
		agent.CreateSymbol("key"),
	})

	result := CreateListFromArrayLike(agent, arrayLike.ToValue(), ArrayLikeElementTypesPropertyKey)
	if result.IsAbrupt() || len(result.Data()) != 2 {
		t.Fatalf("property-key list = %#v, abrupt=%v", result.Data(), result.IsAbrupt())
	}

	arrayLike.Set(NewIntegerIndexPropertyKey(1), NewNumberValue(1), setThrowTypeThrow)
	result = CreateListFromArrayLike(agent, arrayLike.ToValue(), ArrayLikeElementTypesPropertyKey)
	if !result.IsAbrupt() {
		t.Fatal("property-key list accepted a Number element")
	}

	allTypes := CreateListFromArrayLike(agent, arrayLike.ToValue())
	if allTypes.IsAbrupt() || len(allTypes.Data()) != 2 {
		t.Fatalf("all-types list = %#v, abrupt=%v", allTypes.Data(), allTypes.IsAbrupt())
	}
}

// TestTypedArrayDefineOwnPropertyChecksDescriptorPresence verifies omitted
// Boolean fields differ from explicitly supplied false attributes.
func TestTypedArrayDefineOwnPropertyChecksDescriptorPresence(t *testing.T) {
	agent, realm := newStructuredDataTestRealm()
	typedArray := TypedArrayCreate(agent, TypedArrayNameUint8, realm.Intrinsics.Uint8ArrayPrototype)
	AllocateTypedArrayBuffer(agent, typedArray, 1)
	key := NewIntegerIndexPropertyKey(0)
	define := func(desc *PropertyDescriptor) bool {
		t.Helper()
		result := typedArray.internalMethods().DefineOwnProperty(typedArray, key, desc)
		if result.IsAbrupt() {
			t.Fatalf("DefineOwnProperty completed abruptly: %v", result.Error())
		}
		return result.Data()
	}

	if !define(&PropertyDescriptor{Value: NewNumberValue(7)}) {
		t.Fatal("descriptor with omitted Boolean fields was rejected")
	}
	if got := TypedArrayGetElement(agent, typedArray, 0).(*NumberValue).Data; got != 7 {
		t.Fatalf("typed array element = %v, want 7", got)
	}

	rejected := []*PropertyDescriptor{
		{ConfigurableSet: true, Configurable: false},
		{EnumerableSet: true, Enumerable: false},
		{WritableSet: true, Writable: false},
		{GetSet: true},
	}
	for _, desc := range rejected {
		if define(desc) {
			t.Fatalf("explicitly incompatible descriptor was accepted: %#v", desc)
		}
	}

	if !define(&PropertyDescriptor{
		Value:           NewNumberValue(9),
		Writable:        true,
		WritableSet:     true,
		Enumerable:      true,
		EnumerableSet:   true,
		Configurable:    true,
		ConfigurableSet: true,
	}) {
		t.Fatal("compatible fully populated descriptor was rejected")
	}
}
