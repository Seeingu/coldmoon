package coldmoon

type devFeatures struct {
	enableNewVM bool
}

var DevFeatures = devFeatures{
	enableNewVM: false,
}

func (d *devFeatures) ToggleNewVM(v bool) {
	d.enableNewVM = v
}

func (d *devFeatures) IsNewVMEnabled() bool {
	return d.enableNewVM
}
