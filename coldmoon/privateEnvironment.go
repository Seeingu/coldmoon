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
