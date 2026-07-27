package coldmoon

//go:generate stringer -type=ExceptionType
type ExceptionType int

func (e ExceptionType) ToIntrinsicName() IntrinsicName {
	switch e {
	case EvalError:
		return IntrinsicNameEvalError
	case RangeError:
		return IntrinsicNameRangeError
	case ReferenceError:
		return IntrinsicNameReferenceError
	case SyntaxError:
		return IntrinsicNameSyntaxError
	case TypeError:
		return IntrinsicNameTypeError
	case URIError:
		return IntrinsicNameURIError
	case AggregateError:
		return IntrinsicNameAggregateError
	default:
		panic("unknown ExceptionType")
	}
}

const (
	EvalError ExceptionType = iota
	RangeError
	ReferenceError
	SyntaxError
	TypeError
	URIError
	AggregateError
)
