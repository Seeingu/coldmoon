package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPrintsEvalResult(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", "-p", "1 + 2"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader(""),
		Stdout:     &stdout,
		Stderr:     &stderr,
		Version:    "test",
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "3\n" {
		t.Fatalf("stdout = %q, want %q", got, "3\n")
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunHelpDoesNotStartTheRuntime(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "--help"},
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}
	if got := stdout.String(); !strings.Contains(got, "coldmoon [options] [file|-] [arguments...]") {
		t.Fatalf("help output %q does not contain usage", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunVersionUsesInjectedBuildVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv:    []string{"/tmp/coldmoon", "--version"},
		Stdout:  &stdout,
		Stderr:  &stderr,
		Version: "v1.2.3",
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}
	if got := stdout.String(); got != "coldmoon v1.2.3\n" {
		t.Fatalf("stdout = %q, want version line", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunRejectsUnknownOption(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "--bad"},
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 2 {
		t.Fatalf("Run() exit code = %d, want 2", exitCode)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	want := "coldmoon: error: unknown option \"--bad\"\nTry 'coldmoon --help' for more information.\n"
	if got := stderr.String(); got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestRunRejectsEvalAndPrintTogether(t *testing.T) {
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "-e", "1", "-p", "2"},
		Stdout: io.Discard,
		Stderr: &stderr,
	})

	if exitCode != 2 {
		t.Fatalf("Run() exit code = %d, want 2", exitCode)
	}
	if got := stderr.String(); !strings.Contains(got, "-e/--eval and -p/--print cannot be combined") {
		t.Fatalf("stderr = %q, want conflict diagnostic", got)
	}
}

func TestRunEvalUsesInjectedTerminalOutput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "-e", `console.log("hello")`},
		Cwd:    t.TempDir(),
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "hello\n" {
		t.Fatalf("stdout = %q, want %q", got, "hello\n")
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunEvalPassesNodeShapedProcessArgv(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv: []string{
			"/tmp/coldmoon",
			"-p",
			`process.argv.length + ":" + process.argv[0] + ":" + process.argv[1] + ":" + process.argv[2]`,
			"alpha",
			"beta",
		},
		Cwd:    t.TempDir(),
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	want := "3:/tmp/coldmoon:alpha:beta\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunExecutesFileAsModuleWithScriptArguments(t *testing.T) {
	directory := t.TempDir()
	dependency := filepath.Join(directory, "dep.js")
	entry := filepath.Join(directory, "app.js")
	if err := os.WriteFile(dependency, []byte("export const answer = 42;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := fmt.Sprintf(`
import { answer } from "./dep.js";
assert(answer === 42);
assert(process.argv[0] === %q);
assert(process.argv[1] === %q);
assert(process.argv[2] === "alpha");
console.log("ok");
`, "/tmp/coldmoon", entry)
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", entry, "alpha"},
		Cwd:    directory,
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "ok\n" {
		t.Fatalf("stdout = %q, want %q", got, "ok\n")
	}
}

func TestRunExecutesPipedStdinOnceWithoutPrompt(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader(`console.log(process.argv.length + ":stdin")`),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "1:stdin\n" {
		t.Fatalf("stdout = %q, want %q", got, "1:stdin\n")
	}
	if strings.Contains(stdout.String(), "> ") {
		t.Fatalf("piped stdin unexpectedly printed a prompt: %q", stdout.String())
	}
}

func TestRunExecutesExplicitStdinWithDashInProcessArgv(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv: []string{"/tmp/coldmoon", "-", "alpha"},
		Cwd:  t.TempDir(),
		Stdin: strings.NewReader(`
assert(process.argv.length === 3);
assert(process.argv[1] === "-");
assert(process.argv[2] === "alpha");
console.log("explicit-stdin");
`),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "explicit-stdin\n" {
		t.Fatalf("stdout = %q, want explicit stdin output", got)
	}
}

func TestRunRejectsInvalidInputType(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "--input-type", "commonjs", "-e", "1"},
		Stdout: io.Discard,
		Stderr: &stderr,
	})

	if exitCode != 2 {
		t.Fatalf("Run() exit code = %d, want 2", exitCode)
	}
	if got := stderr.String(); !strings.Contains(got, `--input-type must be "script" or "module"`) {
		t.Fatalf("stderr = %q, want input-type diagnostic", got)
	}
}

func TestRunRejectsCheckWithInteractive(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "--check", "-i", "-e", "1"},
		Stdout: io.Discard,
		Stderr: &stderr,
	})

	if exitCode != 2 {
		t.Fatalf("Run() exit code = %d, want 2", exitCode)
	}
	if got := stderr.String(); !strings.Contains(got, "--check and -i/--interactive cannot be combined") {
		t.Fatalf("stderr = %q, want check/interactive diagnostic", got)
	}
}

func TestRunRendersSyntaxDiagnosticWithoutGoStack(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "-e", "const value = };"},
		Cwd:    t.TempDir(),
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1", exitCode)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	got := stderr.String()
	for _, fragment := range []string{
		"<eval>:1:",
		`SyntaxError: unexpected token "}"`,
		"1 | const value = };",
		"|               ^",
	} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("stderr = %q, want fragment %q", got, fragment)
		}
	}
	if strings.Contains(got, "goroutine") || strings.Contains(got, ".go:") {
		t.Fatalf("stderr leaked a Go stack: %q", got)
	}
}

func TestRunCheckOnlyParsesEntryWithoutExecutingOrLoadingImports(t *testing.T) {
	directory := t.TempDir()
	entry := filepath.Join(directory, "check.js")
	if err := os.WriteFile(entry, []byte(`
import {} from "./missing.js";
console.log("must-not-run");
`), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "--check", entry},
		Cwd:    directory,
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, check unexpectedly executed source", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, check unexpectedly loaded imports", got)
	}
}

func TestRunValidatesConflictsBeforeFileArgumentBoundary(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "app.js")
	if err := os.WriteFile(entry, []byte("1;"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "--check", "-i", entry},
		Cwd:    filepath.Dir(entry),
		Stdout: io.Discard,
		Stderr: &stderr,
	})

	if exitCode != 2 {
		t.Fatalf("Run() exit code = %d, want 2", exitCode)
	}
	if got := stderr.String(); !strings.Contains(got, "--check and -i/--interactive cannot be combined") {
		t.Fatalf("stderr = %q, want conflict diagnostic", got)
	}
}

func TestRunREPLSupportsMultilineInputAndSharedRealm(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv: []string{"/tmp/coldmoon"},
		Cwd:  t.TempDir(),
		Stdin: strings.NewReader("function add(a, b) {\n" +
			"return a + b;\n" +
			"}\n" +
			"add(1, 2)\n" +
			".exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: true,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	want := "> ... ... undefined\n> 3\n> "
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunREPLContinuesAfterSyntaxError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("const value = };\n1 + 1\n.exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: true,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}
	if got := stdout.String(); got != "> > 2\n> " {
		t.Fatalf("stdout = %q, want recovered REPL output", got)
	}
	if got := stderr.String(); !strings.Contains(got, "<repl:1>:1:") || !strings.Contains(got, "SyntaxError:") {
		t.Fatalf("stderr = %q, want first-submission syntax diagnostic", got)
	}
}

func TestRunREPLContinuesAfterThrownPrimitive(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("throw \"boom\";\n2 + 2\n.exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: true,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}
	if got := stdout.String(); got != "> > 4\n> " {
		t.Fatalf("stdout = %q, want recovered REPL output", got)
	}
	if got := stderr.String(); !strings.Contains(got, "<repl:1>: Uncaught boom") {
		t.Fatalf("stderr = %q, want primitive throw diagnostic", got)
	}
}

func TestRunREPLReportsIncompleteInputAtEOF(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("function value() {\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: true,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1", exitCode)
	}
	if got := stdout.String(); got != "> ... " {
		t.Fatalf("stdout = %q, want primary and continuation prompts", got)
	}
	if got := stderr.String(); !strings.Contains(got, "<repl:1>:") || !strings.Contains(got, "SyntaxError:") {
		t.Fatalf("stderr = %q, want incomplete diagnostic", got)
	}
}

func TestRunREPLAcceptsSubmissionLargerThanScannerLimit(t *testing.T) {
	const size = 70 * 1024
	input := `"` + strings.Repeat("a", size) + `".length` + "\n.exit\n"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader(input),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: true,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	want := fmt.Sprintf("> %d\n> ", size)
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunForcedREPLOnPipeDoesNotPrintPrompts(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", "-i"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("1 + 1\n.exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "2\n" {
		t.Fatalf("stdout = %q, want result without prompts", got)
	}
}

func TestRunInteractiveAfterEvalReusesRealm(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", "-e", "let answer = 40;", "-i"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("answer + 2\n.exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "42\n" {
		t.Fatalf("stdout = %q, want REPL result from prelude state", got)
	}
}

func TestRunDoubleDashAllowsFileNameBeginningWithDash(t *testing.T) {
	directory := t.TempDir()
	entry := filepath.Join(directory, "-app.js")
	if err := os.WriteFile(entry, []byte(`console.log(process.argv[2])`), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "--", "-app.js", "alpha"},
		Cwd:    directory,
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "alpha\n" {
		t.Fatalf("stdout = %q, want script argument", got)
	}
}

func TestRunInlineModuleResolvesImportsFromInvocationDirectory(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "dep.js"), []byte(`export const answer = 42;`), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv: []string{
			"/tmp/coldmoon",
			"--input-type", "module",
			"-e", `import { answer } from "./dep.js"; console.log(answer);`,
		},
		Cwd:    directory,
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "42\n" {
		t.Fatalf("stdout = %q, want inline module output", got)
	}
}

// TestRunFileModulePublishesCanonicalEntryIdentity verifies that a dependency
// importing the file entry converges on the already parsed module record.
func TestRunFileModulePublishesCanonicalEntryIdentity(t *testing.T) {
	directory := t.TempDir()
	entry := filepath.Join(directory, "main.js")
	dependency := filepath.Join(directory, "dep.js")
	if err := os.WriteFile(entry, []byte(`
import { dep } from "./dep.js";
export const root = 1;
console.log(dep);
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dependency, []byte(`
import { root } from "./main.js";
export const dep = 41;
`), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", entry},
		Cwd:    directory,
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "41\n" {
		t.Fatalf("stdout = %q, want one entry-module evaluation", got)
	}
}

func TestRunInputTypeScriptOverridesFileModuleDefault(t *testing.T) {
	directory := t.TempDir()
	entry := filepath.Join(directory, "entry.js")
	if err := os.WriteFile(entry, []byte(`export const value = 1;`), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "--input-type=script", entry},
		Cwd:    directory,
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1", exitCode)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, script input unexpectedly executed module", got)
	}
	if got := stderr.String(); !strings.Contains(got, "SyntaxError:") {
		t.Fatalf("stderr = %q, want script grammar diagnostic", got)
	}
}

func TestRunReturnsOneForUncaughtJavaScriptError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "-e", `throw new TypeError("boom");`},
		Cwd:    t.TempDir(),
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1", exitCode)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "<eval>: TypeError: boom\n" {
		t.Fatalf("stderr = %q, want uncaught error diagnostic", got)
	}
}

// TestRunRendersClassHeritageTypeErrorWithoutStack verifies that an expected
// class-definition failure stays inside the CLI diagnostic boundary.
func TestRunRendersClassHeritageTypeErrorWithoutStack(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "-e", `class Invalid extends 1 {}`},
		Cwd:    t.TempDir(),
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1", exitCode)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	got := stderr.String()
	if !strings.Contains(got, "<eval>: TypeError: superclass is not a constructor") {
		t.Fatalf("stderr = %q, want class heritage diagnostic", got)
	}
	for _, leaked := range []string{"panic:", "goroutine", ".go:"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("stderr leaked Go implementation detail %q: %q", leaked, got)
		}
	}
}

func TestRunReturnsOneForMissingFileWithoutUsageOrStack(t *testing.T) {
	directory := t.TempDir()
	missing := filepath.Join(directory, "missing.js")
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", missing},
		Cwd:    directory,
		Stdin:  strings.NewReader(""),
		Stdout: io.Discard,
		Stderr: &stderr,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1", exitCode)
	}
	got := stderr.String()
	if !strings.Contains(got, missing+": IOError:") {
		t.Fatalf("stderr = %q, want missing-file diagnostic", got)
	}
	if strings.Contains(got, "Try 'coldmoon --help'") || strings.Contains(got, "goroutine") {
		t.Fatalf("stderr included usage or Go stack: %q", got)
	}
}

func TestRunRejectsInvalidContracts(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "check with print", args: []string{"--check", "-p", "1"}, want: "--check and -p/--print cannot be combined"},
		{name: "print module", args: []string{"--input-type", "module", "-p", "1"}, want: "-p/--print does not support module input"},
		{name: "interactive explicit stdin", args: []string{"-i", "-"}, want: "explicit stdin '-' and -i/--interactive cannot be combined"},
		{name: "removed test262 mode", args: []string{"-test262"}, want: `unknown option "-test262"`},
		{name: "removed test262 root", args: []string{"-test262-root", "test262"}, want: `unknown option "-test262-root"`},
		{name: "check without source", args: []string{"--check"}, want: "--check requires a file, eval source, or non-TTY stdin"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stderr bytes.Buffer
			exitCode := Run(Invocation{
				Argv:       append([]string{"/tmp/coldmoon"}, test.args...),
				Cwd:        t.TempDir(),
				Stdin:      strings.NewReader(""),
				Stdout:     io.Discard,
				Stderr:     &stderr,
				StdinIsTTY: true,
			})

			if exitCode != 2 {
				t.Fatalf("Run() exit code = %d, want 2; stderr = %q", exitCode, stderr.String())
			}
			if got := stderr.String(); !strings.Contains(got, test.want) {
				t.Fatalf("stderr = %q, want fragment %q", got, test.want)
			}
		})
	}
}

func TestRunRejectsTrailingTopLevelToken(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "--check", "-e", "const value = 1; }"},
		Cwd:    t.TempDir(),
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); !strings.Contains(got, "SyntaxError:") {
		t.Fatalf("stderr = %q, want trailing-token syntax diagnostic", got)
	}
}

func TestRunStopsParsingOptionsAtFileOperand(t *testing.T) {
	directory := t.TempDir()
	entry := filepath.Join(directory, "arguments.js")
	if err := os.WriteFile(entry, []byte(`console.log(process.argv[2] + ":" + process.argv[3]);`), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", entry, "--check", "-i"},
		Cwd:        directory,
		Stdin:      strings.NewReader(""),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: true,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "--check:-i\n" {
		t.Fatalf("stdout = %q, want file arguments unchanged", got)
	}
}

func TestRunDoubleDashPassesLeadingDashArgumentToEval(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "-p", "process.argv[1]", "--", "--flag"},
		Cwd:    t.TempDir(),
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "--flag\n" {
		t.Fatalf("stdout = %q, want leading-dash argument", got)
	}
}

func TestRunREPLHelpAndProcessArgv(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", "-i"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader(".help\nprocess.argv.length + ':' + process.argv[0]\n.exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	want := replHelp + "1:/tmp/coldmoon\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunREPLSupportsMultilineTemplateAndBlockComment(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv: []string{"/tmp/coldmoon", "-i"},
		Cwd:  t.TempDir(),
		Stdin: strings.NewReader("`hello\n" +
			"world`.length\n" +
			"/* open\n" +
			"closed */ 6 * 7\n" +
			".exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "11\n42\n" {
		t.Fatalf("stdout = %q, want multiline results", got)
	}
}

func TestRunDrainsSchedulerBeforeReturning(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "-e", `setTimeout(() => console.log("later"), 0); console.log("now");`},
		Cwd:    t.TempDir(),
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "now\nlater\n" {
		t.Fatalf("stdout = %q, want scheduler output before return", got)
	}
}

func TestRunReturnsOneForAsynchronousCallbackError(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "-e", `setTimeout(() => { throw new TypeError("later"); }, 0);`},
		Cwd:    t.TempDir(),
		Stdin:  strings.NewReader(""),
		Stdout: io.Discard,
		Stderr: &stderr,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1; stderr = %q", exitCode, stderr.String())
	}
	if got := stderr.String(); got != "<eval>: TypeError: later\n" {
		t.Fatalf("stderr = %q, want async callback diagnostic", got)
	}
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestRunReturnsOneForStdinReadError(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon"},
		Cwd:        t.TempDir(),
		Stdin:      errorReader{err: errors.New("read failed")},
		Stdout:     io.Discard,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1; stderr = %q", exitCode, stderr.String())
	}
	if got := stderr.String(); got != "<stdin>: IOError: read failed\n" {
		t.Fatalf("stderr = %q, want stdin I/O diagnostic", got)
	}
}

func TestRunHelpAndVersionStopOptionParsing(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		version    string
		wantStdout string
	}{
		{
			name:       "help",
			args:       []string{"--check", "--help", "/does/not/exist.js"},
			wantStdout: helpText,
		},
		{
			name:       "version",
			args:       []string{"--version", "ignored.js"},
			version:    "v9.8.7",
			wantStdout: "coldmoon v9.8.7\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := Run(Invocation{
				Argv:       append([]string{"/tmp/coldmoon"}, test.args...),
				Cwd:        t.TempDir(),
				Stdin:      errorReader{err: errors.New("must not read stdin")},
				Stdout:     &stdout,
				Stderr:     &stderr,
				Version:    test.version,
				StdinIsTTY: false,
			})

			if exitCode != 0 {
				t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
			}
			if got := stdout.String(); got != test.wantStdout {
				t.Fatalf("stdout = %q, want %q", got, test.wantStdout)
			}
			if got := stderr.String(); got != "" {
				t.Fatalf("stderr = %q, want empty", got)
			}
		})
	}
}

func TestRunMakesInjectedExecutableAbsolute(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"bin/coldmoon", "-p", "process.argv[0]"},
		Cwd:    ".",
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	want := filepath.Join(workingDirectory, "bin", "coldmoon") + "\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want absolute executable %q", got, want)
	}
}

func TestRunInputTypeDoesNotChangeREPLGrammar(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", "-i", "--input-type", "module"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("1 + 1\n.exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "2\n" {
		t.Fatalf("stdout = %q, want script REPL result", got)
	}
}

func TestRunCheckRejectsKnownInvalidSyntaxWithoutPanicking(t *testing.T) {
	for _, sourceText := range []string{
		"a++++",
		"new.foo",
		"1 = 2",
		"3!",
		"`head $",
		"const value = {",
		"class Value {",
		`const value = {"name"};`,
		"if (true)",
		"if (true) {} else",
		"while (true)",
		"for (value in source)",
		"label:",
	} {
		t.Run(sourceText, func(t *testing.T) {
			var stderr bytes.Buffer
			exitCode := Run(Invocation{
				Argv:   []string{"/tmp/coldmoon", "--check", "-e", sourceText},
				Cwd:    t.TempDir(),
				Stdin:  strings.NewReader(""),
				Stdout: io.Discard,
				Stderr: &stderr,
			})

			if exitCode != 1 {
				t.Fatalf("Run() exit code = %d, want 1; stderr = %q", exitCode, stderr.String())
			}
			got := stderr.String()
			if !strings.Contains(got, "SyntaxError:") || strings.Contains(got, "goroutine") || strings.Contains(got, "panic:") {
				t.Fatalf("stderr = %q, want typed syntax diagnostic without stack", got)
			}
		})
	}
}

func TestRunREPLReportsIncompleteArrowAtEOF(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("(value) =>\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: true,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "> ... " {
		t.Fatalf("stdout = %q, want continuation prompt", got)
	}
	got := stderr.String()
	if !strings.Contains(got, "SyntaxError:") || strings.Contains(got, "panic:") || strings.Contains(got, "goroutine") {
		t.Fatalf("stderr = %q, want incomplete arrow diagnostic without stack", got)
	}
}

func TestRunREPLReportsDollarTerminatedTemplateAtEOF(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("`head $"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: true,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "> " {
		t.Fatalf("stdout = %q, want primary prompt", got)
	}
	got := stderr.String()
	if !strings.Contains(got, "SyntaxError:") || strings.Contains(got, "panic:") || strings.Contains(got, "goroutine") {
		t.Fatalf("stderr = %q, want incomplete template diagnostic without stack", got)
	}
}

func TestRunREPLContinuesAfterLexicalRedeclaration(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", "-i"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("let value = 1\nlet value = 2\n1 + 1\n.exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "undefined\n2\n" {
		t.Fatalf("stdout = %q, want recovery result", got)
	}
	got := stderr.String()
	if !strings.Contains(got, "SyntaxError:") || strings.Contains(got, "panic:") || strings.Contains(got, "goroutine") {
		t.Fatalf("stderr = %q, want redeclaration diagnostic without stack", got)
	}
}

func TestRunREPLContinuesAfterVarLexicalCollision(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", "-i"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("let value = 1\nvar value\n3\n.exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "undefined\n3\n" {
		t.Fatalf("stdout = %q, want recovery result", got)
	}
	got := stderr.String()
	if !strings.Contains(got, "SyntaxError:") || strings.Contains(got, "panic:") || strings.Contains(got, "goroutine") {
		t.Fatalf("stderr = %q, want declaration-collision diagnostic without stack", got)
	}
}

func TestRunREPLContinuesAfterConstantAssignment(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", "-i"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader("const value = 1\nvalue = 2\n2 + 2\n.exit\n"),
		Stdout:     &stdout,
		Stderr:     &stderr,
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "undefined\n4\n" {
		t.Fatalf("stdout = %q, want recovery result", got)
	}
	got := stderr.String()
	if !strings.Contains(got, "TypeError:") || strings.Contains(got, "panic:") || strings.Contains(got, "goroutine") {
		t.Fatalf("stderr = %q, want constant-assignment diagnostic without stack", got)
	}
}

func TestRunReturnsOneForInvalidRegExpConstructor(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := Run(Invocation{
		Argv:   []string{"/tmp/coldmoon", "-e", `new RegExp("[")`},
		Cwd:    t.TempDir(),
		Stdin:  strings.NewReader(""),
		Stdout: io.Discard,
		Stderr: &stderr,
	})

	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1; stderr = %q", exitCode, stderr.String())
	}
	got := stderr.String()
	if !strings.Contains(got, "SyntaxError:") || strings.Contains(got, "panic:") || strings.Contains(got, "goroutine") {
		t.Fatalf("stderr = %q, want RegExp syntax diagnostic without stack", got)
	}
}

func TestRunCheckRejectsDeclarationEarlyErrors(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
	}{
		{
			name:      "duplicate lexical declaration",
			arguments: []string{"--check", "-e", "let value; let value;"},
		},
		{
			name:      "missing const initializer",
			arguments: []string{"--check", "-e", "const value;"},
		},
		{
			name:      "missing var declaration",
			arguments: []string{"--check", "-e", "var;"},
		},
		{
			name:      "lexical and var collision",
			arguments: []string{"--check", "-e", "let value; var value;"},
		},
		{
			name:      "duplicate module lexical declaration",
			arguments: []string{"--check", "--input-type", "module", "-e", "let value; let value;"},
		},
		{
			name:      "missing local module export",
			arguments: []string{"--check", "--input-type", "module", "-e", "export { missing };"},
		},
		{
			name:      "duplicate module default export",
			arguments: []string{"--check", "--input-type", "module", "-e", "export default 0; export default 0;"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stderr bytes.Buffer
			exitCode := Run(Invocation{
				Argv:   append([]string{"/tmp/coldmoon"}, test.arguments...),
				Cwd:    t.TempDir(),
				Stdin:  strings.NewReader(""),
				Stdout: io.Discard,
				Stderr: &stderr,
			})

			if exitCode != 1 {
				t.Fatalf("Run() exit code = %d, want 1; stderr = %q", exitCode, stderr.String())
			}
			got := stderr.String()
			if !strings.Contains(got, "SyntaxError:") || strings.Contains(got, "panic:") || strings.Contains(got, "goroutine") {
				t.Fatalf("stderr = %q, want early-error diagnostic without stack", got)
			}
		})
	}
}

// TestRunReportsUnhandledRejection covers the HostPromiseRejectionTracker
// hook: an async throw with no handler must print to stderr and exit non-zero
// instead of silently succeeding.
func TestRunReportsUnhandledRejection(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", "-e", "async function f() { throw new Error('async boom'); } f();"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader(""),
		Stdout:     &stdout,
		Stderr:     &stderr,
		Version:    "test",
		StdinIsTTY: false,
	})

	if exitCode == 0 {
		t.Fatalf("Run() exit code = 0, want non-zero for an unhandled rejection; stderr = %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Uncaught (in promise) Error: async boom") {
		t.Fatalf("stderr = %q, want an unhandled-rejection report", stderr.String())
	}
}

// TestRunHandledRejectionStaysSilent covers the Handle operation: a rejection
// that gains a handler must not be reported.
func TestRunHandledRejectionStaysSilent(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(Invocation{
		Argv:       []string{"/tmp/coldmoon", "-e", "async function f() { throw new Error('handled'); } f().catch(function () {});"},
		Cwd:        t.TempDir(),
		Stdin:      strings.NewReader(""),
		Stdout:     &stdout,
		Stderr:     &stderr,
		Version:    "test",
		StdinIsTTY: false,
	})

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0 for a handled rejection; stderr = %q", exitCode, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty for a handled rejection", stderr.String())
	}
}
