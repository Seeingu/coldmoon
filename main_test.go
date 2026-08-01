package main

import (
	"os"
	"runtime/debug"
	"testing"
)

func TestVersionFromBuildInfoIncludesDevelopmentRevision(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890abcdef"},
			{Key: "vcs.modified", Value: "true"},
		},
	}

	if got := versionFromBuildInfo(info); got != "devel+1234567890ab.dirty" {
		t.Fatalf("versionFromBuildInfo() = %q", got)
	}
}

func TestVersionFromBuildInfoPrefersModuleVersion(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.2.3"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "ffffffffffffffff"},
			{Key: "vcs.modified", Value: "true"},
		},
	}

	if got := versionFromBuildInfo(info); got != "v1.2.3" {
		t.Fatalf("versionFromBuildInfo() = %q, want module version", got)
	}
}

func TestVersionFromBuildInfoFallsBackToDevelopment(t *testing.T) {
	for _, info := range []*debug.BuildInfo{nil, {Main: debug.Module{Version: "(devel)"}}} {
		if got := versionFromBuildInfo(info); got != "devel" {
			t.Fatalf("versionFromBuildInfo(%#v) = %q, want devel", info, got)
		}
	}
}

func TestVersionFromBuildInfoTreatsLocalPseudoVersionAsDevelopment(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v0.0.0-20260801100051-81de91b9834f+dirty"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "81de91b9834f7e643d6b0857c16616a463d2b6cd"},
			{Key: "vcs.modified", Value: "true"},
		},
	}

	if got := versionFromBuildInfo(info); got != "devel+81de91b9834f.dirty" {
		t.Fatalf("versionFromBuildInfo() = %q, want local development version", got)
	}
}

func TestIsTerminalRejectsNonTTYCharacterDevice(t *testing.T) {
	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if isTerminal(file) {
		t.Fatalf("isTerminal(%s) = true, want false", os.DevNull)
	}
}
