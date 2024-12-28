package coldmoon

func Evaluate(source string, realm *Realm) {
	result := ParseScript(source, realm, nil).Evaluate()
	if o, ok := ValueGetObject(result); ok {
		if e, ok := o.(*ErrorObject); ok {
			println("Return Error: ", e.Message)
			panic(e)
		}
	}
}
