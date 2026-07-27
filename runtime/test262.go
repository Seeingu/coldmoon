package runtime

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Seeingu/coldmoon/coldmoon"
)

const test262RunnerVersion = "2"

var (
	test262IncludesPattern   = regexp.MustCompile(`(?m)^includes:\s*\[([^\]]*)\]`)
	test262OnlyStrictPattern = regexp.MustCompile(`(?m)^flags:\s*\[[^\]]*\bonlyStrict\b`)
)

var baseTest262Harnesses = []string{
	"sta.js",
	"assert.js",
	"isConstructor.js",
	"nans.js",
	"assertRelativeDateMs.js",
	"propertyHelper.js",
	"testTypedArray.js",
	"testAtomics.js",
	"compareArray.js",
}

// Test262Suite is an explicitly located Test262 checkout.
type Test262Suite struct {
	Root string
}

// NewTest262Suite validates and canonicalizes root. Callers choose the root;
// the runtime never guesses it from the process working directory.
func NewTest262Suite(root string) (*Test262Suite, error) {
	if root == "" {
		return nil, errors.New("test262 root is required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	for _, required := range []string{"harness", "test"} {
		info, statErr := os.Stat(filepath.Join(absolute, required))
		if statErr != nil {
			return nil, fmt.Errorf("invalid test262 root %q: %w", absolute, statErr)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("invalid test262 root %q: %s is not a directory", absolute, required)
		}
	}
	return &Test262Suite{Root: absolute}, nil
}

// Path returns an absolute path inside the suite.
func (s *Test262Suite) Path(relative string) string {
	return filepath.Join(s.Root, filepath.FromSlash(relative))
}

// Test262Runtime owns harness state for exactly one Realm. Dropping this value
// drops the state; no global Realm-keyed registry or manual release is needed.
type Test262Runtime struct {
	suite  *Test262Suite
	realm  *coldmoon.Realm
	loaded map[string]struct{}
}

// NewRuntime installs the Test262 host globals and base harnesses in realm.
func (s *Test262Suite) NewRuntime(realm *coldmoon.Realm) (*Test262Runtime, error) {
	runtime := &Test262Runtime{
		suite:  s,
		realm:  realm,
		loaded: make(map[string]struct{}),
	}
	RegisterFilesystemModuleLoader(realm)
	runtime.installGlobals()
	for _, file := range baseTest262Harnesses {
		if err := runtime.loadHarness(file); err != nil {
			return nil, err
		}
	}
	return runtime, nil
}

func (r *Test262Runtime) installGlobals() {
	realm := r.realm
	global := realm.GlobalObject
	console := CreateConsole(realm)
	global.CreateDataProperty(coldmoon.CMString("console").ToPropertyKey(), console.ToValue())

	test262 := coldmoon.OrdinaryObjectCreate(realm.Agent, realm.Intrinsics.ObjectPrototype, nil)
	var createRealm coldmoon.BehaviorFn = func(
		_ coldmoon.Value,
		_ []coldmoon.Value,
		_ coldmoon.ObjectType,
	) coldmoon.CompletionConvertable[coldmoon.Value] {
		otherRealm := coldmoon.CreateRealm(realm.Agent)
		otherGlobal := coldmoon.OrdinaryObjectCreate(
			realm.Agent,
			otherRealm.Intrinsics.ObjectPrototype,
			nil,
		)
		otherRealm.SetRealmGlobalObject(otherGlobal, nil)
		otherRealm.SetDefaultGlobalBindings()

		record := coldmoon.OrdinaryObjectCreate(
			realm.Agent,
			realm.Intrinsics.ObjectPrototype,
			nil,
		)
		record.CreateDataPropertyOrThrow(
			coldmoon.CMString("global").ToPropertyKey(),
			otherGlobal.ToValue(),
		)
		return record.ToValue()
	}
	coldmoon.DefineBuiltinFunction(
		realm,
		coldmoon.CMString("createRealm"),
		test262,
		createRealm,
		0,
	)
	global.CreateDataProperty(
		coldmoon.CMString("$262").ToPropertyKey(),
		test262.ToValue(),
	)
}

func (r *Test262Runtime) loadHarness(file string) error {
	if _, ok := r.loaded[file]; ok {
		return nil
	}
	content, err := os.ReadFile(r.suite.Path(filepath.Join("harness", file)))
	if err != nil {
		return fmt.Errorf("load test262 harness %q: %w", file, err)
	}
	coldmoon.ParseScript(string(content), r.realm, nil).Evaluate()
	r.loaded[file] = struct{}{}
	return nil
}

// RegisterIncludes evaluates optional harness files declared by frontmatter.
func (r *Test262Runtime) RegisterIncludes(source string) error {
	match := test262IncludesPattern.FindStringSubmatch(source)
	if len(match) != 2 {
		return nil
	}
	for _, entry := range strings.Split(match[1], ",") {
		file := strings.Trim(strings.TrimSpace(entry), `"'`)
		if file == "" {
			continue
		}
		if err := r.loadHarness(file); err != nil {
			return err
		}
	}
	return nil
}

// Test262Runner executes isolated test files with bounded waiting and
// content-addressed cache keys.
type Test262Runner struct {
	Suite   *Test262Suite
	Timeout time.Duration
}

// NewTest262Runner creates a runner with a five-second per-file timeout.
func NewTest262Runner(suite *Test262Suite) *Test262Runner {
	return &Test262Runner{Suite: suite, Timeout: 5 * time.Second}
}

// CacheKey hashes runner semantics, the portable suite-relative path, and file
// contents so stale or machine-specific path-only successes cannot be reused.
func (r *Test262Runner) CacheKey(filePath string) (string, error) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(r.Suite.Root, filePath)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(test262RunnerVersion + "\x00" + filepath.ToSlash(relative) + "\x00" + string(source)))
	return fmt.Sprintf("%x", sum[:]), nil
}

// Run executes one file in a new Agent and Realm. The worker owns all mutable
// state, so returning on context cancellation cannot race with another case.
func (r *Test262Runner) Run(ctx context.Context, filePath string) error {
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}
	done := make(chan error, 1)
	go func() {
		done <- r.run(filePath)
	}()
	select {
	case <-ctx.Done():
		return fmt.Errorf("test262 timeout: %s: %w", filePath, ctx.Err())
	case err := <-done:
		return err
	}
}

func (r *Test262Runner) run(filePath string) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panic: %v", recovered)
		}
	}()
	source, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	agent := coldmoon.NewAgent()
	coldmoon.InitializeHostDefinedRealm(agent, nil)
	realmRuntime, err := r.Suite.NewRuntime(agent.CurrentRealm())
	if err != nil {
		return err
	}
	if err := realmRuntime.RegisterIncludes(string(source)); err != nil {
		return err
	}
	coldmoon.Evaluate(PrepareTest262Source(string(source)), agent.CurrentRealm())
	return nil
}

// PrepareTest262Source applies execution-mode flags from test262 frontmatter.
func PrepareTest262Source(source string) string {
	if test262OnlyStrictPattern.MatchString(source) {
		return "\"use strict\";\n" + source
	}
	return source
}
