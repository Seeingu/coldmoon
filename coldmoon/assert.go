package coldmoon

func Assert(t bool) {
	if !t {
		panic("throw error from Assert")
	}
}
