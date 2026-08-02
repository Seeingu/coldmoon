// Package cli implements Coldmoon's user-facing command contract.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/runtime"
)

const helpText = `Usage:
  coldmoon [options] [file|-] [arguments...]
  coldmoon -e <code> [arguments...]
  coldmoon -p <code> [arguments...]

Options:
  -e, --eval <code>             Evaluate code
  -p, --print <code>            Evaluate code and print its result
  -i, --interactive             Enter the REPL after initial input
      --check                   Parse input without executing it
      --input-type <type>       Treat input as script or module
  -h, --help                    Show this help
  -v, --version                 Show version information
      --                        End option parsing
`

// Invocation contains the process state needed to run one CLI invocation.
// Keeping these values explicit lets command tests use the same interface as
// main without mutating process-global arguments, streams, or working state.
type Invocation struct {
	Argv       []string
	Cwd        string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	Version    string
	StdinIsTTY bool
}

// Run executes a Coldmoon command and returns its process exit code without
// terminating the caller. main is responsible for passing the result to
// os.Exit.
func Run(invocation Invocation) int {
	invocation = normalizeInvocation(invocation)
	options, err := parseOptions(invocation.Argv[1:])
	if err != nil {
		fmt.Fprintf(invocation.Stderr, "coldmoon: error: %s\n", err)
		fmt.Fprintln(invocation.Stderr, "Try 'coldmoon --help' for more information.")
		return 2
	}
	if options.showHelp {
		fmt.Fprint(invocation.Stdout, helpText)
		return 0
	}
	if options.showVersion {
		fmt.Fprintf(invocation.Stdout, "coldmoon %s\n", invocation.Version)
		return 0
	}

	prepared, err := prepareInput(invocation, options)
	if err != nil {
		renderError(invocation.Stderr, err)
		return 1
	}
	if !prepared.hasSource {
		if options.check {
			return renderUsageError(invocation.Stderr, "--check requires a file, eval source, or non-TTY stdin")
		}
		realm := newTerminalRealm(invocation, prepared.processArgv)
		return runREPL(invocation, realm)
	}

	realm := newTerminalRealm(invocation, prepared.processArgv)
	if options.check {
		if err := coldmoon.CheckSource(prepared.source, realm); err != nil {
			renderError(invocation.Stderr, err)
			return 1
		}
		return 0
	}

	var value coldmoon.Value
	if prepared.moduleIdentity != "" {
		value, err = coldmoon.EvaluateSourceWithModuleIdentity(prepared.source, prepared.moduleIdentity, realm)
	} else {
		value, err = coldmoon.EvaluateSource(prepared.source, realm)
	}
	if err != nil {
		renderError(invocation.Stderr, err)
		return 1
	}
	if options.hasPrint {
		fmt.Fprintln(invocation.Stdout, runtime.FormatValue(value))
	}
	if options.interactive {
		return runREPL(invocation, realm)
	}
	return 0
}

type preparedInput struct {
	source         coldmoon.Source
	moduleIdentity string
	processArgv    []string
	hasSource      bool
}

func prepareInput(invocation Invocation, options commandOptions) (preparedInput, error) {
	executable := absolutePath(invocation.Cwd, invocation.Argv[0])
	if options.hasEval || options.hasPrint {
		text := options.evalCode
		if options.hasPrint {
			text = options.printCode
		}
		return preparedInput{
			source: coldmoon.Source{
				Text:    text,
				Name:    "<eval>",
				BaseDir: invocation.Cwd,
				Kind:    selectedSourceKind(options.inputType, coldmoon.SourceScript),
			},
			processArgv: append([]string{executable}, options.arguments...),
			hasSource:   true,
		}, nil
	}
	if options.hasFile {
		path := absolutePath(invocation.Cwd, options.file)
		text, err := os.ReadFile(path)
		if err != nil {
			return preparedInput{}, fmt.Errorf("%s: IOError: %w", path, err)
		}
		return preparedInput{
			source: coldmoon.Source{
				Text:    string(text),
				Name:    path,
				BaseDir: filepath.Dir(path),
				Kind:    selectedSourceKind(options.inputType, coldmoon.SourceModule),
			},
			moduleIdentity: path,
			processArgv:    append([]string{executable, path}, options.arguments...),
			hasSource:      true,
		}, nil
	}
	if options.stdin || (!invocation.StdinIsTTY && !options.interactive) {
		text, err := io.ReadAll(invocation.Stdin)
		if err != nil {
			return preparedInput{}, fmt.Errorf("<stdin>: IOError: %w", err)
		}
		argv := []string{executable}
		if options.stdin {
			argv = append(argv, "-")
			argv = append(argv, options.arguments...)
		}
		return preparedInput{
			source: coldmoon.Source{
				Text:    string(text),
				Name:    "<stdin>",
				BaseDir: invocation.Cwd,
				Kind:    selectedSourceKind(options.inputType, coldmoon.SourceScript),
			},
			processArgv: argv,
			hasSource:   true,
		}, nil
	}
	return preparedInput{processArgv: []string{executable}}, nil
}

func selectedSourceKind(inputType string, fallback coldmoon.SourceKind) coldmoon.SourceKind {
	if inputType == "script" {
		return coldmoon.SourceScript
	}
	if inputType == "module" {
		return coldmoon.SourceModule
	}
	return fallback
}

func normalizeInvocation(invocation Invocation) Invocation {
	if len(invocation.Argv) == 0 {
		invocation.Argv = []string{"coldmoon"}
	}
	if invocation.Cwd == "" {
		invocation.Cwd = "."
	}
	if cwd, err := filepath.Abs(invocation.Cwd); err == nil {
		invocation.Cwd = cwd
	}
	if invocation.Stdin == nil {
		invocation.Stdin = strings.NewReader("")
	}
	if invocation.Stdout == nil {
		invocation.Stdout = io.Discard
	}
	if invocation.Stderr == nil {
		invocation.Stderr = io.Discard
	}
	if invocation.Version == "" {
		invocation.Version = "devel"
	}
	return invocation
}

func renderUsageError(writer io.Writer, message string) int {
	fmt.Fprintf(writer, "coldmoon: error: %s\n", message)
	fmt.Fprintln(writer, "Try 'coldmoon --help' for more information.")
	return 2
}

func renderError(writer io.Writer, err error) {
	var diagnostic *coldmoon.Diagnostic
	if !errors.As(err, &diagnostic) {
		fmt.Fprintln(writer, err)
		return
	}
	if diagnostic.Span == nil {
		if diagnostic.ErrorName == "Uncaught" {
			fmt.Fprintf(writer, "%s: Uncaught %s\n", diagnostic.SourceName, diagnostic.Message)
			return
		}
		fmt.Fprintf(writer, "%s: %s\n", diagnostic.SourceName, diagnostic.Error())
		return
	}

	line := diagnostic.Span.Start.Line
	column := diagnostic.Span.Start.Column
	fmt.Fprintf(writer, "%s:%d:%d: %s\n", diagnostic.SourceName, line, column, diagnostic.Error())
	lineLabel := strconv.Itoa(line)
	fmt.Fprintf(writer, "  %s | %s\n", lineLabel, diagnostic.SourceLine)
	indent := strings.Repeat(" ", len(lineLabel)+3)
	caretPadding := strings.Repeat(" ", max(column-1, 0))
	fmt.Fprintf(writer, "%s| %s^\n", indent, caretPadding)
}

func absolutePath(cwd, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(cwd, path))
}

func newTerminalRealm(invocation Invocation, argv []string) *coldmoon.Realm {
	coldmoon.InitializeConstants()
	agent := coldmoon.NewAgent()
	coldmoon.InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	runtime.RegisterTerminalRuntimeWithOptions(realm, runtime.TerminalOptions{
		Stdout: invocation.Stdout,
		Stderr: invocation.Stderr,
		Argv:   argv,
	})
	return realm
}
