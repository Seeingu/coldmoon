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
	flag.Parse()
	InitializeConstants()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	eventLoop := runtime.NewEventLoop(agent)
	if *isTest262 {
		fmt.Println("register test262 runtime")
		runtime.RegisterTest262Runtime(realm)
	} else {
		runtime.RegisterTerminalRuntime(eventLoop, realm)
	}
	Debug.IsReady = true

	// get args, if is a file, read and evaluate
	// else run as repl
	args := flag.Args()
	if len(args) > 0 {
		EvaluateModule(args[0], realm)
		eventLoop.Poll()
		agent.WG.Wait()
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
		eventLoop.Poll()
		agent.WG.Wait()
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading input:", err)
	}
}
