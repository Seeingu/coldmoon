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
const expectedCoverage = 0.2

// supportFeatures is a list of features that are tested
// and gather coverage information.
var supportFeatures = []string{
	//"built-ins/Array/isArray",
	"built-ins/DataView",
	"built-ins/TypedArray/Symbol.species",
	"built-ins/SharedArrayBuffer/prototype",
	"built-ins/TypedArrayConstructors/BigInt64Array",
	"built-ins/Boolean",
}

func Test262WithCoverage(t *testing.T) {
	passedFiles := loadPassedResultFiles()
	agent := NewAgent()
	InitializeConstants()
	InitializeHostDefinedRealm(agent, nil)
	realm := agent.CurrentRealm()
	runtime.RegisterTest262Runtime(realm)

	var passed []string
	var failed []string
	var visitor fs.WalkDirFunc = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := passedFiles[path]; ok {
			fmt.Println("Skip testing file: ", path)
			passed = append(passed, path)
			return nil
		}
		fmt.Println("Testing file: ", path)
		err = evaluate(path, realm)
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed to evaluate file: %s, error: %v", path, err))
			failed = append(failed, path)
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
	coverage := float64(len(passed)) / float64(len(passed)+len(failed))
	fmt.Printf("Coverage: %.2f%%\n", coverage*100)
	if coverage < expectedCoverage {
		t.Errorf("Coverage is less than expected: %.2f%% < %.2f%%", coverage*100, expectedCoverage*100)
	}
}

// MARK: - utils

const PassedResultsFile = ".passed.txt"

func loadPassedResultFiles() map[string]bool {
	m := make(map[string]bool)
	_, err := os.Stat(PassedResultsFile)
	if err != nil {
		return nil
	}
	content := pkg.MustReadFile(PassedResultsFile)
	for _, f := range strings.Split(content, "\n") {
		m[f] = true
	}
	return m
}

func writePassedResultFiles(files []string) {
	content := strings.Join(files, "\n")
	os.WriteFile(PassedResultsFile, []byte(content), 0o644)
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
