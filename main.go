package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/Seeingu/coldmoon/runtime"

	. "github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/pkg"
)

func main() {
	isTest262 := flag.Bool("test262", false, "register test262 runtime")
	flag.Parse()
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	if *isTest262 {
		fmt.Println("register test262 runtime")
		runtime.RegisterTest262Runtime(realm)
	} else {
		runtime.RegisterTerminalRuntime(realm)
	}
	Debug.IsReady = true

	// get args, if is a file, read and evaluate
	// else run as repl
	args := flag.Args()
	if len(args) > 0 {
		if *isTest262 {
			source := pkg.MustReadFile(args[0])
			runtime.RegisterTest262Includes(realm, source)
			Evaluate(runtime.PrepareTest262Source(source), realm)
		} else {
			EvaluateModule(args[0], realm)
		}
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
