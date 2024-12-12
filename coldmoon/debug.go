package coldmoon

type debugConfig struct {
	PrintAST      bool
	PrintBytecode bool
}

func (d *debugConfig) Enable() {
	d.PrintBytecode = true
	d.PrintAST = true
}

func (d *debugConfig) Disable() {
	d.PrintBytecode = false
	d.PrintAST = false
}

var Debug = &debugConfig{
	PrintAST:      false,
	PrintBytecode: false,
}
