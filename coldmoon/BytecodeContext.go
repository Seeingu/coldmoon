package coldmoon

import "github.com/Seeingu/coldmoon/pkg"

type BytecodeContext struct {
	vm                        *VM
	exe                       *Executable
	agent                     *Agent
	containedInStrictCode     bool
	labelContinueJumpIndexMap map[string]pkg.Stack[*IJump]
	labelBreakJumpIndexMap    map[string]pkg.Stack[*IJump]
	continueJumpIndices       pkg.Stack[*IJump]
	breakJumpIndices          pkg.Stack[*IJump]
	Label                     string
}

func (b *BytecodeContext) Run() CompletionValue {
	return b.vm.Run(b.exe)
}
