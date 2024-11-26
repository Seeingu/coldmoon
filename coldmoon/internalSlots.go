package coldmoon

// [[PrivateMethods]]
type InternalSlotPrivateMethods interface {
	PrivateMethods() []*PrivateMethodDefinition
}

// [[Fields]]
type InternalSlotFields interface {
	Fields() []*ClassFieldDefinition
	SetFields(f []*ClassFieldDefinition)
}

type ClassFieldInitializerName PropertyKeyOrPrivateName

// [[ClassFieldInitializerName]]
type InternalSlotClassFieldInitializerName interface {
	ClassFieldInitializerName() ClassFieldInitializerName
	SetClassFieldInitializerName(n ClassFieldInitializerName)
}
