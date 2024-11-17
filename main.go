package main

import (
	"bufio"
	"fmt"
	"os"

	. "github.com/Seeingu/coldmoon/coldmoon"
)

func main() {
	agent := NewAgent()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		ParseScript(input, realm, nil).Evaluate()
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading input:", err)
	}
}
