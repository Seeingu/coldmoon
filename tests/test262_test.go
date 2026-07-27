package tests

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"
	"testing"

	. "github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/runtime"
)

const expectedCoverage = 0.95

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
}

func Test262WithCoverage(t *testing.T) {
	if os.Getenv("COLDMOON_RUN_TEST262") != "1" {
		t.Skip("set COLDMOON_RUN_TEST262=1 to run the bounded Test262 coverage suite")
	}
	InitializeConstants()
	suite := repositoryTest262Suite(t)
	runner := runtime.NewTest262Runner(suite)
	cachePath := filepath.Join(currentTestDirectory(t), ".passed.v2.txt")
	passedCache := loadTest262Cache(t, cachePath)
	skipped := map[string]struct{}{
		"test/built-ins/Array/from/elements-deleted-after.js": {},
	}

	var passed []string
	var failedLog strings.Builder
	var failedCount int
	visitor := func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(filePath) != ".js" {
			return nil
		}
		relative, err := filepath.Rel(suite.Root, filePath)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if _, ok := skipped[relative]; ok {
			return nil
		}
		cacheKey, err := runner.CacheKey(filePath)
		if err != nil {
			return err
		}
		if passedCache[relative] == cacheKey {
			passed = append(passed, relative)
			return nil
		}

		t.Logf("Testing file: %s", relative)
		if err := runner.Run(context.Background(), filePath); err != nil {
			failedCount++
			fmt.Fprintf(&failedLog, "%s, Failed: %v\n", relative, err)
			return nil
		}
		passedCache[relative] = cacheKey
		passed = append(passed, relative)
		return nil
	}

	for _, feature := range supportFeatures {
		if err := filepath.WalkDir(suite.Path("test/"+feature), visitor); err != nil {
			t.Error(err)
		}
	}
	writeTest262Cache(t, cachePath, passedCache)
	if err := os.WriteFile(filepath.Join(currentTestDirectory(t), ".failed.v2.txt"), []byte(failedLog.String()), 0o644); err != nil {
		t.Error(err)
	}

	total := len(passed) + failedCount
	if total == 0 {
		t.Fatal("Test262 runner found no cases")
	}
	coverage := float64(len(passed)) / float64(total)
	t.Logf("Coverage: %.2f%%", coverage*100)
	if coverage < expectedCoverage {
		t.Errorf("coverage is less than expected: %.2f%% < %.2f%%", coverage*100, expectedCoverage*100)
	}
}

func repositoryTest262Suite(t *testing.T) *runtime.Test262Suite {
	t.Helper()
	suite, err := runtime.NewTest262Suite(filepath.Join(currentTestDirectory(t), "..", "test262"))
	if err != nil {
		t.Fatal(err)
	}
	return suite
}

func currentTestDirectory(t *testing.T) string {
	t.Helper()
	_, sourceFile, _, ok := goruntime.Caller(0)
	if !ok {
		t.Fatal("locate Test262 test source")
	}
	return filepath.Dir(sourceFile)
}

func loadTest262Cache(t *testing.T, filePath string) map[string]string {
	t.Helper()
	cache := make(map[string]string)
	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return cache
	}
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, relative, ok := strings.Cut(scanner.Text(), "\t")
		if ok && key != "" && relative != "" {
			cache[relative] = key
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return cache
}

func writeTest262Cache(t *testing.T, filePath string, cache map[string]string) {
	t.Helper()
	paths := make([]string, 0, len(cache))
	for relative := range cache {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	var content strings.Builder
	for _, relative := range paths {
		fmt.Fprintf(&content, "%s\t%s\n", cache[relative], relative)
	}
	if err := os.WriteFile(filePath, []byte(content.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}
