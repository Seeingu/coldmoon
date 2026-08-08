package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/runtime"
)

const replHelp = `.help  Show REPL commands
.exit  Exit the REPL
`

// runREPL evaluates complete script submissions in one Realm. The parser's
// typed incomplete diagnostic, rather than delimiter counting, decides when a
// continuation line is required.
func runREPL(invocation Invocation, realm *coldmoon.Realm, tracker *rejectionTracker) int {
	reader := bufio.NewReader(invocation.Stdin)
	var sourceText strings.Builder
	submission := 1

	for {
		if invocation.StdinIsTTY {
			if sourceText.Len() == 0 {
				fmt.Fprint(invocation.Stdout, "> ")
			} else {
				fmt.Fprint(invocation.Stdout, "... ")
			}
		}

		line, readErr := reader.ReadString('\n')
		if sourceText.Len() == 0 {
			switch strings.TrimSpace(line) {
			case ".exit":
				return 0
			case ".help":
				fmt.Fprint(invocation.Stdout, replHelp)
				if readErr == io.EOF {
					return 0
				}
				continue
			}
		}

		if line != "" {
			sourceText.WriteString(line)
		}
		if sourceText.Len() == 0 {
			if readErr == io.EOF {
				return 0
			}
			if readErr != nil {
				fmt.Fprintf(invocation.Stderr, "<repl>: IOError: %v\n", readErr)
				return 1
			}
			continue
		}

		source := coldmoon.Source{
			Text:    sourceText.String(),
			Name:    fmt.Sprintf("<repl:%d>", submission),
			BaseDir: invocation.Cwd,
			Kind:    coldmoon.SourceScript,
		}
		parseErr := coldmoon.CheckSource(source, realm)
		if diagnosticIsIncomplete(parseErr) {
			if readErr == io.EOF {
				renderError(invocation.Stderr, parseErr)
				return 1
			}
			if readErr != nil {
				fmt.Fprintf(invocation.Stderr, "<repl>: IOError: %v\n", readErr)
				return 1
			}
			continue
		}

		if parseErr != nil {
			renderError(invocation.Stderr, parseErr)
		} else if strings.TrimSpace(source.Text) != "" {
			value, evaluateErr := coldmoon.EvaluateSource(source, realm)
			if evaluateErr != nil {
				renderError(invocation.Stderr, evaluateErr)
			} else {
				fmt.Fprintln(invocation.Stdout, runtime.FormatValue(value))
			}
			// Unhandled rejections are reported per submission; the REPL stays
			// alive to accept the next input.
			tracker.report(invocation.Stderr, source.Name)
		}

		sourceText.Reset()
		submission++
		if readErr == io.EOF {
			return 0
		}
		if readErr != nil {
			fmt.Fprintf(invocation.Stderr, "<repl>: IOError: %v\n", readErr)
			return 1
		}
	}
}

func diagnosticIsIncomplete(err error) bool {
	var diagnostic *coldmoon.Diagnostic
	return errors.As(err, &diagnostic) && diagnostic.Incomplete
}
