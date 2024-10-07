package _parser

type Environment struct {
	locals    map[string]Object
	enclosing *Environment
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		enclosing: enclosing,
		locals:    make(map[string]Object),
	}
}

func (e *Environment) Exists(key string) bool {
	_, ok := e.Get(key)
	return ok
}

func (e *Environment) Get(key string) (expr Object, ok bool) {
	v, ok := e.locals[key]
	if !ok && e.enclosing != nil {
		return e.enclosing.Get(key)
	}
	return v, ok
}

func (e *Environment) Set(key string, value Object) {
	e.locals[key] = value
}
