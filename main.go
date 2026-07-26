package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/Seeingu/coldmoon/runtime"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

func main() {
	isTest262 := flag.Bool("test262", false, "register test262 runtime")
	test262Root := flag.String("test262-root", os.Getenv("TEST262_ROOT"), "explicit Test262 checkout root")
	flag.Parse()
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	var test262Runtime *runtime.Test262Runtime
	if *isTest262 {
		fmt.Println("register test262 runtime")
		suite, err := runtime.NewTest262Suite(*test262Root)
		if err != nil {
			panic(err)
		}
		test262Runtime, err = suite.NewRuntime(realm)
		if err != nil {
			panic(err)
		}
	} else {
		runtime.RegisterTerminalRuntime(realm)
	}
	Debug.IsReady = true

	// get args, if is a file, read and evaluate
	// else run as repl
	args := flag.Args()
	if len(args) > 0 {
		if *isTest262 {
			source, err := os.ReadFile(args[0])
			if err != nil {
				panic(err)
			}
			if err := test262Runtime.RegisterIncludes(string(source)); err != nil {
				panic(err)
			}
			Evaluate(runtime.PrepareTest262Source(string(source)), realm)
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
