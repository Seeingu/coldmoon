package coldmoon

// 9.2
type PrivateEnvironment struct {
	OuterPrivateEnvironment *PrivateEnvironment
	Names                   []string
}

// 9.2.1.1
func NewPrivateEnvironment(outerPrivateEnv *PrivateEnvironment) *PrivateEnvironment {
	return &PrivateEnvironment{
		OuterPrivateEnvironment: outerPrivateEnv,
		Names:                   []string{},
	}
}
