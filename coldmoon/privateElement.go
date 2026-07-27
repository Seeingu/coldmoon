package coldmoon

type PrivateElementKind int

const (
	PrivateElementKindField PrivateElementKind = iota
	PrivateElementKindMethod
	PrivateElementKindAccessor
)

type PrivateElement struct {
	Key   PrivateName
	Kind  PrivateElementKind
	Value Value
	Get   ObjectType
	Set   ObjectType
}
