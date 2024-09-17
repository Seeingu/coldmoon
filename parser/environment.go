package parser

type Environment struct {
	locals    map[string]Expression
	enclosing *Environment
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		enclosing: enclosing,
		locals:    make(map[string]Expression),
	}
}

func (e *Environment) Exists(key string) bool {
	_, ok := e.Get(key)
	return ok
}

func (e *Environment) Get(key string) (expr Expression, ok bool) {
	v, ok := e.locals[key]
	if !ok && e.enclosing != nil {
		return e.enclosing.Get(key)
	}
	return v, ok
}

func (e *Environment) Set(key string, value Expression) {
	e.locals[key] = value
}
