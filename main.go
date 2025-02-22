package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Seeingu/coldmoon/runtime"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

func main() {
	Debug.Enable()
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	runtime.RegisterTerminalRuntime(realm)
	// runtime.RegisterTest262Runtime(realm)

	// get args, if is file, read file and evaluate
	args := os.Args[1:]
	if len(args) > 0 {
		EvaluateModule(args[0], realm)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		Evaluate(input, realm)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading input:", err)
	}
}
