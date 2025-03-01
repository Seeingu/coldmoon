package tests

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/pkg"
	"github.com/Seeingu/coldmoon/runtime"
)

// MARK: - Coverage config

// expectedCoverage is the expected coverage rate.
const expectedCoverage = 0.3

// supportFeatures is a list of features that are tested
// and gather coverage information.
var supportFeatures = []string{
	"built-ins/Array",
	"built-ins/Boolean",
	"built-ins/String",
	"built-ins/Symbol",
	"built-ins/Number",
	"built-ins/BigInt",
	"built-ins/Date",
	"built-ins/Function",
	"built-ins/Object",
	// "built-ins/RegExp",
	// "built-ins/DataView",
	// "built-ins/TypedArray",
	// "built-ins/SharedArrayBuffer",
	// "built-ins/TypedArrayConstructors",
	// "built-ins",
	// "harness",
	// "language",
}

func Test262WithCoverage(t *testing.T) {
	passedFiles := loadPassedResultFiles()
	skipped := getSkippedFiles()
	agent := NewAgent()
	InitializeConstants()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	runtime.RegisterTest262Runtime(realm)

	var passed []string
	var failedLog strings.Builder
	var failedCount int
	var skippedCount int
	var visitor fs.WalkDirFunc = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := skipped[path]; ok {
			// fmt.Println("Skip testing file: ", path)
			skippedCount++
			return nil
		}
		if _, ok := passedFiles[path]; ok {
			// fmt.Println("Skip testing passed file: ", path)
			passed = append(passed, path)
			skippedCount++
			return nil
		}
		fmt.Println("Testing file: ", path)
		err = evaluate(path, realm)
		if err != nil {
			failedCount++
			msg := fmt.Sprintf("Failed : %v", err)
			fmt.Println(msg)
			failedLog.WriteString(
				fmt.Sprintf(`%s, %s`, path, msg) + "\n")
		} else {
			passed = append(passed, path)
		}
		return nil
	}
	for _, feature := range supportFeatures {
		dir := runtime.MakeTest262Path("test/" + feature)
		err := filepath.WalkDir(dir, visitor)
		if err != nil {
			t.Error(err)
		}
	}
	writePassedResultFiles(passed)
	writeFailedLog(failedLog)
	coverage := float64(len(passed)) / float64(len(passed)+failedCount)
	fmt.Printf("Coverage: %.2f%%\n", coverage*100)
	if coverage < expectedCoverage {
		t.Errorf("Coverage is less than expected: %.2f%% < %.2f%%", coverage*100, expectedCoverage*100)
	}
}

// MARK: - utils

const (
	PassedResultsFile = ".passed.txt"
	FailedResultsFile = ".failed.txt"
)

var skippedFiles = []string{
	// global this reference is not supported
	runtime.MakeTest262Path("test/built-ins/Array/from/elements-deleted-after.js"),
}

func loadResultFile(fileName string) map[string]bool {
	m := make(map[string]bool)
	_, err := os.Stat(fileName)
	if err != nil {
		return nil
	}
	content := pkg.MustReadFile(fileName)
	for _, f := range strings.Split(content, "\n") {
		m[f] = true
	}
	return m
}

func getSkippedFiles() map[string]bool {
	m := make(map[string]bool)
	for _, f := range skippedFiles {
		m[f] = true
	}
	return m
}

func loadPassedResultFiles() map[string]bool {
	return loadResultFile(PassedResultsFile)
}

func loadFailedResultFiles() map[string]bool {
	return loadResultFile(FailedResultsFile)
}

func writePassedResultFiles(files []string) {
	content := strings.Join(files, "\n")
	os.WriteFile(PassedResultsFile, []byte(content), 0o644)
}

func writeFailedLog(log strings.Builder) {
	os.WriteFile(FailedResultsFile, []byte(log.String()), 0o644)
}

func evaluate(fileName string, realm *Realm) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("panic" + fmt.Sprint(r))
		}
	}()
	Evaluate(pkg.MustReadFile(fileName), realm)
	return
}

// TODO: performance
func evaluateAsync(ctx context.Context, fileName string, realm *Realm) (err error) {
	evaluateChan := make(chan struct{}, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New("panic" + fmt.Sprint(r))
			} else {
				evaluateChan <- struct{}{}
			}
		}()
		Evaluate(pkg.MustReadFile(fileName), realm)
	}()
	select {
	case <-ctx.Done():
		err = errors.New("timeout")
		return
	case <-evaluateChan:
		return
	}
}
