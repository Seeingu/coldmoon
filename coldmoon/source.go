package coldmoon

import (
	"fmt"
	"strings"
)

type parseFailure struct {
	message    string
	start      int
	end        int
	incomplete bool
	committed  bool
	sourceName string
	sourceText string
}

func (f *parseFailure) Error() string { return f.message }

// SourceKind selects the ECMAScript grammar used for a Source.
type SourceKind uint8

const (
	SourceScript SourceKind = iota
	SourceModule
)

// Source is a complete ECMAScript input together with its host identity.
type Source struct {
	Text    string
	Name    string
	BaseDir string
	Kind    SourceKind
}

// DiagnosticCategory classifies failures at the safe execution seam.
type DiagnosticCategory string

const (
	DiagnosticSyntax  DiagnosticCategory = "syntax"
	DiagnosticLink    DiagnosticCategory = "link"
	DiagnosticRuntime DiagnosticCategory = "runtime"
)

// Position identifies a Unicode code-point offset and its 1-based line and
// column in a Source.
type Position struct {
	Offset int
	Line   int
	Column int
}

// SourceSpan identifies a half-open range in a Source.
type SourceSpan struct {
	Start Position
	End   Position
}

// Diagnostic is a stable, non-panicking representation of an expected source
// or JavaScript execution failure.
type Diagnostic struct {
	Category   DiagnosticCategory
	ErrorName  string
	Message    string
	SourceName string
	SourceLine string
	Span       *SourceSpan
	Incomplete bool
	Thrown     Value
	Cause      error
}

func (d *Diagnostic) Error() string {
	if d == nil {
		return ""
	}
	if d.ErrorName == "" {
		return d.Message
	}
	if d.Message == "" {
		return d.ErrorName
	}
	return d.ErrorName + ": " + d.Message
}

func (d *Diagnostic) Unwrap() error {
	if d == nil {
		return nil
	}
	return d.Cause
}

// CheckSource parses source without loading module dependencies or executing
// user code.
func CheckSource(source Source, _ *Realm) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			failure, ok := recovered.(*parseFailure)
			if !ok {
				panic(recovered)
			}
			err = syntaxDiagnostic(source, failure)
		}
	}()
	parseSource(source, nil)
	return nil
}

// EvaluateSource parses and executes source, translating expected language
// failures into Diagnostic while allowing implementation panics to surface.
func EvaluateSource(source Source, realm *Realm) (value Value, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			switch recovered := recovered.(type) {
			case *parseFailure:
				value = nil
				err = syntaxDiagnostic(source, recovered)
			case Value:
				value = nil
				err = thrownDiagnostic(DiagnosticRuntime, source, recovered, nil)
			default:
				panic(recovered)
			}
		}
	}()

	script, module := parseSource(source, realm)
	if script != nil {
		result := script.evaluateCompletion()
		if result.IsAbrupt() {
			thrown := result.Error()
			if thrown == nil {
				thrown = result.Data()
			}
			return nil, thrownDiagnostic(DiagnosticRuntime, source, thrown, nil)
		}
		realm.Agent.Scheduler.RunUntilIdle()
		return result.Data(), nil
	}

	module = realm.Agent.ModuleGraph.cacheSource(realm, source.Name, module)
	return evaluateModuleRecord(source, realm, module)
}

func evaluateModuleRecord(source Source, realm *Realm, module *SourceTextModule) (Value, error) {
	loaded := module.LoadRequestedModules()
	if loaded.PromiseState == PromiseStatePending {
		realm.Agent.Scheduler.RunUntilIdle()
	}
	if loaded.PromiseState == PromiseStateRejected {
		return nil, thrownDiagnostic(DiagnosticLink, source, loaded.PromiseResult, nil)
	}
	if loaded.PromiseState != PromiseStateFulfilled {
		panic("module loading did not settle")
	}

	if thrown := module.Link(); thrown != nil {
		return nil, thrownDiagnostic(DiagnosticLink, source, thrown, nil)
	}
	evaluated := module.Evaluate()
	realm.Agent.Scheduler.RunUntilIdle()
	switch evaluated.PromiseState {
	case PromiseStateFulfilled:
		return UndefinedValue, nil
	case PromiseStateRejected:
		return nil, thrownDiagnostic(DiagnosticRuntime, source, evaluated.PromiseResult, nil)
	default:
		panic("module evaluation did not settle")
	}
}

func parseSource(source Source, realm *Realm) (*ScriptRecord, *SourceTextModule) {
	hostDefined := HostDefined{FileName: source.Name, BaseDir: source.BaseDir}
	switch source.Kind {
	case SourceScript:
		return parseScript(source.Text, realm, &hostDefined, ParserContext{
			FileName: source.Name,
			BaseDir:  source.BaseDir,
		}), nil
	case SourceModule:
		return nil, ParseModule(source.Text, realm, hostDefined)
	default:
		panic(fmt.Sprintf("unknown source kind %d", source.Kind))
	}
}

func syntaxDiagnostic(source Source, failure *parseFailure) *Diagnostic {
	if failure.sourceName != "" {
		source.Name = failure.sourceName
		source.Text = failure.sourceText
	}
	start, end := failure.start, failure.end
	if end < start {
		end = start
	}
	startPosition, sourceLine := sourcePosition(source.Text, start)
	endPosition, _ := sourcePosition(source.Text, end)
	return &Diagnostic{
		Category:   DiagnosticSyntax,
		ErrorName:  "SyntaxError",
		Message:    failure.message,
		SourceName: source.Name,
		SourceLine: sourceLine,
		Span: &SourceSpan{
			Start: startPosition,
			End:   endPosition,
		},
		Incomplete: failure.incomplete,
	}
}

func thrownDiagnostic(category DiagnosticCategory, source Source, thrown Value, cause error) *Diagnostic {
	diagnostic := &Diagnostic{
		Category:   category,
		ErrorName:  "Uncaught",
		SourceName: source.Name,
		Thrown:     thrown,
		Cause:      cause,
	}
	if thrown == nil {
		diagnostic.Message = "unknown exception"
		return diagnostic
	}
	if object, ok := thrown.GetObject(); ok {
		if exception, ok := object.(*ErrorObject); ok {
			if diagnostic.Cause == nil {
				diagnostic.Cause = exception.hostCause
			}
			diagnostic.ErrorName = exception.Name
			if diagnostic.ErrorName == "" {
				diagnostic.ErrorName = "Error"
			}
			diagnostic.Message = exception.Message
			return diagnostic
		}
	}
	diagnostic.Message = thrown.String()
	return diagnostic
}

func sourcePosition(text string, offset int) (Position, string) {
	runes := []rune(text)
	if offset < 0 {
		offset = 0
	}
	if offset > len(runes) {
		offset = len(runes)
	}

	line, column, lineStart := 1, 1, 0
	for index := 0; index < offset; index++ {
		switch runes[index] {
		case '\r':
			line++
			column = 1
			lineStart = index + 1
			if index+1 < offset && runes[index+1] == '\n' {
				index++
				lineStart = index + 1
			}
		case '\n', '\u2028', '\u2029':
			line++
			column = 1
			lineStart = index + 1
		default:
			column++
		}
	}

	lineEnd := lineStart
	for lineEnd < len(runes) && !strings.ContainsRune("\r\n\u2028\u2029", runes[lineEnd]) {
		lineEnd++
	}
	return Position{Offset: offset, Line: line, Column: column}, string(runes[lineStart:lineEnd])
}
