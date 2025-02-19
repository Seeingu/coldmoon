package coldmoon

type debugConfig struct {
	PrintAST bool
}

func (d *debugConfig) Enable() {
	d.PrintAST = true
}

func (d *debugConfig) Disable() {
	d.PrintAST = false
}

var Debug = &debugConfig{
	PrintAST: false,
}
