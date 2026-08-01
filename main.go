package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/Seeingu/coldmoon/internal/cli"
	"golang.org/x/term"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coldmoon: determine working directory: %v\n", err)
		os.Exit(1)
	}

	executable, err := os.Executable()
	if err != nil {
		executable = os.Args[0]
	}
	argv := append([]string{executable}, os.Args[1:]...)
	exitCode := cli.Run(cli.Invocation{
		Argv:       argv,
		Cwd:        cwd,
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Version:    currentVersion(),
		StdinIsTTY: isTerminal(os.Stdin),
	})
	os.Exit(exitCode)
}

func isTerminal(file *os.File) bool {
	return file != nil && term.IsTerminal(int(file.Fd()))
}

func currentVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}
	return versionFromBuildInfo(info)
}

func versionFromBuildInfo(info *debug.BuildInfo) string {
	var revision string
	modified := false
	if info != nil {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = setting.Value
			case "vcs.modified":
				modified = setting.Value == "true"
			}
		}
	}
	if info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" && !isLocalPseudoVersion(info.Main.Version, revision) {
		return info.Main.Version
	}
	if revision == "" {
		return "devel"
	}
	if len(revision) > 12 {
		revision = revision[:12]
	}
	version := "devel+" + strings.ToLower(revision)
	if modified {
		version += ".dirty"
	}
	return version
}

// Go may stamp a checkout build with a generated pseudo-version. Treat that
// as a development build while preserving real tagged or installed versions.
func isLocalPseudoVersion(version, revision string) bool {
	if revision == "" {
		return false
	}
	shortRevision := strings.ToLower(revision)
	if len(shortRevision) > 12 {
		shortRevision = shortRevision[:12]
	}
	version = strings.TrimSuffix(strings.ToLower(version), "+dirty")
	parts := strings.Split(version, "-")
	if len(parts) < 3 || parts[len(parts)-1] != shortRevision {
		return false
	}
	timestamp := parts[len(parts)-2]
	if len(timestamp) != 14 {
		return false
	}
	for _, character := range timestamp {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
