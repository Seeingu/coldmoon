package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTest262CacheKeyIsPortableAndContentAddressed(t *testing.T) {
	createSuite := func(root string, source string) (*Test262Runner, string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(root, "harness"), 0o755); err != nil {
			t.Fatal(err)
		}
		testDirectory := filepath.Join(root, "test")
		if err := os.MkdirAll(testDirectory, 0o755); err != nil {
			t.Fatal(err)
		}
		filePath := filepath.Join(testDirectory, "case.js")
		if err := os.WriteFile(filePath, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		suite, err := NewTest262Suite(root)
		if err != nil {
			t.Fatal(err)
		}
		return NewTest262Runner(suite), filePath
	}

	firstRunner, firstPath := createSuite(t.TempDir(), "1 + 1;")
	secondRunner, secondPath := createSuite(t.TempDir(), "1 + 1;")
	firstKey, err := firstRunner.CacheKey(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	secondKey, err := secondRunner.CacheKey(secondPath)
	if err != nil {
		t.Fatal(err)
	}
	if firstKey != secondKey {
		t.Fatal("cache key depends on the machine-specific suite root")
	}

	if err := os.WriteFile(firstPath, []byte("1 + 2;"), 0o644); err != nil {
		t.Fatal(err)
	}
	changedKey, err := firstRunner.CacheKey(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	if changedKey == firstKey {
		t.Fatal("cache key did not change with test contents")
	}
}
