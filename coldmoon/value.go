package coldmoon

type Value interface {
}

type undefinedValue struct {
	Value
}

var UndefinedValue = undefinedValue{}

type nullValue struct {
	Value
}

var NullValue = nullValue{}
