package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/Seeingu/coldmoon/runtime"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

func main() {
	Debug.Enable()
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	runtime.RegisterTerminalRuntime(realm)

	// get args, if is file, read file and evaluate
	args := os.Args[1:]
	if len(args) > 0 {
		file, err := os.ReadFile(args[0])
		if err != nil {
			log.Fatalf("Failed to read file, %v\n", err)
		}
		Evaluate(string(file), realm)
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
