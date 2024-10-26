package coldmoon

type Symbol struct {
	Value
	Id          uint64
	Description string
}

func (s Symbol) String() string {
	panic("implement me")
}
