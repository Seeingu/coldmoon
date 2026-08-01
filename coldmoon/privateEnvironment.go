package coldmoon

// 9.2
type PrivateEnvironment struct {
	OuterPrivateEnvironment *PrivateEnvironment
	Names                   []PrivateName
}

// 9.2.1.1
func NewPrivateEnvironment(outerPrivateEnv *PrivateEnvironment) *PrivateEnvironment {
	return &PrivateEnvironment{
		OuterPrivateEnvironment: outerPrivateEnv,
	}
}

// ResolvePrivateIdentifier follows the lexical chain of private environments.
// Static semantics guarantee that a valid private identifier is present in
// this chain before evaluation reaches this operation.
// spec: 9.2.1.2
func ResolvePrivateIdentifier(privateEnv *PrivateEnvironment, identifier PrivateIdentifierName) PrivateName {
	for env := privateEnv; env != nil; env = env.OuterPrivateEnvironment {
		for _, privateName := range env.Names {
			if privateName.Symbol.Description == string(identifier) {
				return privateName
			}
		}
	}
	panic("unresolvable private identifier: " + string(identifier))
}
