package runtime

import (
	"fmt"
	"io"
	"math"
	"os"
	runtimedebug "runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Seeingu/coldmoon/coldmoon"
)

const consoleDefaultLabel = "default"

type consoleState struct {
	mu     sync.Mutex
	stdout io.Writer
	stderr io.Writer
	now    func() time.Time
	counts map[string]uint64
	groups []string
	timers map[string]time.Time
}

type consoleTableRow struct {
	index  string
	values map[string]string
}

func newConsoleState(stdout, stderr io.Writer, now func() time.Time) *consoleState {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if now == nil {
		now = time.Now
	}
	return &consoleState{
		stdout: stdout,
		stderr: stderr,
		now:    now,
		counts: make(map[string]uint64),
		timers: make(map[string]time.Time),
	}
}

func jsPrint(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
	parts := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		parts = append(parts, formatConsoleValue(argument))
	}
	fmt.Fprintln(os.Stdout, strings.Join(parts, " "))
	return coldmoon.UndefinedValue
}

// CreateConsole creates a WHATWG-style terminal console. Informational output
// is written to stdout, while warnings, errors, and failed assertions use
// stderr. Each returned object owns independent count, group, and timer state.
func CreateConsole(realm *coldmoon.Realm) coldmoon.ObjectType {
	return createConsole(realm, os.Stdout, os.Stderr, time.Now)
}

// CreateConsoleWithWriter creates a console whose complete output is sent to
// writer. Hosts that need separate severity streams should use CreateConsole;
// this variant is useful for embedding and deterministic output capture.
func CreateConsoleWithWriter(realm *coldmoon.Realm, writer io.Writer) coldmoon.ObjectType {
	return createConsole(realm, writer, writer, time.Now)
}

func createConsole(
	realm *coldmoon.Realm,
	stdout io.Writer,
	stderr io.Writer,
	now func() time.Time,
) coldmoon.ObjectType {
	agent := realm.Agent
	prototype := coldmoon.OrdinaryObjectCreate(agent, realm.Intrinsics.ObjectPrototype, nil)
	console := coldmoon.NewObject(agent, prototype, "console")
	state := newConsoleState(stdout, stderr, now)

	define := func(name string, behavior coldmoon.BehaviorFn) {
		coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString(name), console, behavior, 0)
	}
	defineLogger := func(name string) {
		define(name, func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
			state.print(name, formatConsoleArguments(agent, arguments))
			return coldmoon.UndefinedValue
		})
	}

	for _, name := range []string{"log", "info", "debug", "warn", "error", "dirxml"} {
		defineLogger(name)
	}

	define("assert", func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		if len(arguments) > 0 && arguments[0].ToBoolean() {
			return coldmoon.UndefinedValue
		}
		data := append([]coldmoon.Value(nil), arguments[min(1, len(arguments)):]...)
		if len(data) > 0 {
			if first, ok := data[0].(*coldmoon.StringValue); ok {
				data[0] = coldmoon.NewStringValue("Assertion failed: " + first.Data)
			} else {
				data = append([]coldmoon.Value{coldmoon.NewStringValue("Assertion failed")}, data...)
			}
		} else {
			data = []coldmoon.Value{coldmoon.NewStringValue("Assertion failed")}
		}
		state.print("assert", formatConsoleArguments(agent, data))
		return coldmoon.UndefinedValue
	})

	define("count", func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		label := consoleLabel(arguments)
		state.mu.Lock()
		state.counts[label]++
		count := state.counts[label]
		state.mu.Unlock()
		state.print("count", []string{fmt.Sprintf("%s: %d", label, count)})
		return coldmoon.UndefinedValue
	})

	define("countReset", func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		label := consoleLabel(arguments)
		state.mu.Lock()
		_, exists := state.counts[label]
		if exists {
			state.counts[label] = 0
		}
		state.mu.Unlock()
		if !exists {
			state.print("warn", []string{fmt.Sprintf("Count for %q does not exist", label)})
		}
		return coldmoon.UndefinedValue
	})

	defineGroup := func(name string) {
		define(name, func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
			parts := formatConsoleArguments(agent, arguments)
			label := "group"
			if len(parts) > 0 {
				label = strings.Join(parts, " ")
			}
			state.print(name, []string{label})
			state.mu.Lock()
			state.groups = append(state.groups, label)
			state.mu.Unlock()
			return coldmoon.UndefinedValue
		})
	}
	defineGroup("group")
	defineGroup("groupCollapsed")
	define("groupEnd", func(_ coldmoon.Value, _ []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		state.mu.Lock()
		if len(state.groups) > 0 {
			state.groups = state.groups[:len(state.groups)-1]
		}
		state.mu.Unlock()
		return coldmoon.UndefinedValue
	})

	define("time", func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		label := consoleLabel(arguments)
		state.mu.Lock()
		_, exists := state.timers[label]
		if !exists {
			state.timers[label] = state.now()
		}
		state.mu.Unlock()
		if exists {
			state.print("warn", []string{fmt.Sprintf("Timer %q already exists", label)})
		}
		return coldmoon.UndefinedValue
	})

	define("timeLog", func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		label := consoleLabel(arguments)
		state.mu.Lock()
		start, exists := state.timers[label]
		var elapsed time.Duration
		if exists {
			elapsed = state.now().Sub(start)
		}
		state.mu.Unlock()
		if !exists {
			state.print("warn", []string{fmt.Sprintf("Timer %q does not exist", label)})
			return coldmoon.UndefinedValue
		}
		parts := []string{fmt.Sprintf("%s: %s", label, formatConsoleDuration(elapsed))}
		if len(arguments) > 1 {
			parts = append(parts, formatConsoleArguments(agent, arguments[1:])...)
		}
		state.print("timeLog", parts)
		return coldmoon.UndefinedValue
	})

	define("timeEnd", func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		label := consoleLabel(arguments)
		state.mu.Lock()
		start, exists := state.timers[label]
		var elapsed time.Duration
		if exists {
			elapsed = state.now().Sub(start)
			delete(state.timers, label)
		}
		state.mu.Unlock()
		if !exists {
			state.print("warn", []string{fmt.Sprintf("Timer %q does not exist", label)})
			return coldmoon.UndefinedValue
		}
		state.print("timeEnd", []string{fmt.Sprintf("%s: %s", label, formatConsoleDuration(elapsed))})
		return coldmoon.UndefinedValue
	})

	define("clear", func(_ coldmoon.Value, _ []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		state.clear()
		return coldmoon.UndefinedValue
	})

	define("trace", func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		label := strings.Join(formatConsoleArguments(agent, arguments), " ")
		stack := strings.TrimSuffix(string(runtimedebug.Stack()), "\n")
		if label != "" {
			stack = label + "\n" + stack
		}
		state.print("trace", []string{stack})
		return coldmoon.UndefinedValue
	})

	define("dir", func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		var item coldmoon.Value = coldmoon.UndefinedValue
		if len(arguments) > 0 {
			item = arguments[0]
		}
		state.print("dir", []string{formatConsoleDirectory(item)})
		return coldmoon.UndefinedValue
	})

	define("table", func(_ coldmoon.Value, arguments []coldmoon.Value, _ coldmoon.ObjectType) coldmoon.CompletionConvertable[coldmoon.Value] {
		var item coldmoon.Value = coldmoon.UndefinedValue
		if len(arguments) > 0 {
			item = arguments[0]
		}
		var properties coldmoon.Value
		if len(arguments) > 1 {
			properties = arguments[1]
		}
		if table, ok := formatConsoleTable(item, properties); ok {
			state.print("table", []string{table})
		} else {
			state.print("log", formatConsoleArguments(agent, []coldmoon.Value{item}))
		}
		return coldmoon.UndefinedValue
	})

	return console
}

func (state *consoleState) print(level string, parts []string) {
	if len(parts) == 0 {
		return
	}
	message := strings.Join(parts, " ")
	prefix := consoleLevelPrefix(level)

	state.mu.Lock()
	defer state.mu.Unlock()
	writer := state.stdout
	if level == "warn" || level == "error" || level == "assert" {
		writer = state.stderr
	}
	indent := strings.Repeat("  ", len(state.groups))
	continuationPrefix := strings.Repeat(" ", utf8.RuneCountInString(prefix))
	for index, line := range strings.Split(message, "\n") {
		linePrefix := prefix
		if index > 0 {
			linePrefix = continuationPrefix
		}
		fmt.Fprintln(writer, indent+linePrefix+line)
	}
}

func (state *consoleState) clear() {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.groups = nil
	// ANSI erase-display and cursor-home is the portable terminal equivalent
	// of the Console Standard's implementation-defined clear operation.
	fmt.Fprint(state.stdout, "\x1b[2J\x1b[H")
}

func consoleLevelPrefix(level string) string {
	switch level {
	case "debug":
		return "[debug] "
	case "info":
		return "[info] "
	case "warn":
		return "[warn] "
	case "error", "assert":
		return "[error] "
	case "trace":
		return "[trace] "
	default:
		return ""
	}
}

func consoleLabel(arguments []coldmoon.Value) string {
	if len(arguments) == 0 {
		return consoleDefaultLabel
	}
	return string(arguments[0].ToString())
}

func formatConsoleDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	milliseconds := float64(duration) / float64(time.Millisecond)
	return strconv.FormatFloat(milliseconds, 'f', 3, 64) + "ms"
}

func formatConsoleArguments(agent *coldmoon.Agent, arguments []coldmoon.Value) []string {
	if len(arguments) == 0 {
		return nil
	}
	first, isString := arguments[0].(*coldmoon.StringValue)
	if !isString || len(arguments) == 1 {
		parts := make([]string, 0, len(arguments))
		for _, argument := range arguments {
			parts = append(parts, formatConsoleValue(argument))
		}
		return parts
	}

	var formatted strings.Builder
	nextArgument := 1
	for index := 0; index < len(first.Data); index++ {
		if first.Data[index] != '%' || index+1 >= len(first.Data) {
			formatted.WriteByte(first.Data[index])
			continue
		}
		specifier := first.Data[index+1]
		if specifier == '%' {
			formatted.WriteByte('%')
			index++
			continue
		}
		if nextArgument >= len(arguments) || !strings.ContainsRune("sdifoOc", rune(specifier)) {
			formatted.WriteByte(first.Data[index])
			continue
		}
		argument := arguments[nextArgument]
		nextArgument++
		index++
		switch specifier {
		case 's':
			formatted.WriteString(string(argument.ToString()))
		case 'd', 'i':
			formatted.WriteString(formatConsoleNumber(agent, argument, true))
		case 'f':
			formatted.WriteString(formatConsoleNumber(agent, argument, false))
		case 'o', 'O':
			formatted.WriteString(formatConsoleDirectory(argument))
		case 'c':
			// CSS styling has no terminal analogue; consume the style argument.
		}
	}

	parts := []string{formatted.String()}
	for _, argument := range arguments[nextArgument:] {
		parts = append(parts, formatConsoleValue(argument))
	}
	return parts
}

func formatConsoleNumber(agent *coldmoon.Agent, value coldmoon.Value, integer bool) string {
	if value.TypeString() == "symbol" {
		return "NaN"
	}
	converted := value.ToNumber(agent)
	if converted.IsAbrupt() || converted.Data() == nil {
		return "NaN"
	}
	number := converted.Data()
	if !integer || number.Data.IsNaN() || number.Data.IsInf() {
		return string(number.ToString())
	}
	return strconv.FormatFloat(math.Trunc(number.Data.ToFloat()), 'f', -1, 64)
}

func formatConsoleValue(value coldmoon.Value) string {
	if value == nil {
		return "undefined"
	}
	if object, ok := value.GetObject(); ok {
		return object.String()
	}
	return string(value.ToString())
}

func formatConsoleDirectory(value coldmoon.Value) string {
	if value == nil {
		return "undefined"
	}
	object, ok := value.GetObject()
	if !ok {
		return formatConsoleValue(value)
	}
	keys := consoleEnumerableStringKeys(object)
	if len(keys) == 0 {
		return object.String()
	}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", key.Value, formatConsoleValue(object.Get(key))))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func consoleEnumerableStringKeys(object coldmoon.ObjectType) []coldmoon.StringPropertyKey {
	var result []coldmoon.StringPropertyKey
	for _, key := range coldmoon.InternalOwnPropertyKeys(object) {
		stringKey, ok := key.(coldmoon.StringPropertyKey)
		if !ok {
			continue
		}
		descriptor := coldmoon.OrdinaryGetOwnProperty(object, key)
		if descriptor != nil && descriptor.Enumerable {
			result = append(result, stringKey)
		}
	}
	return result
}

func formatConsoleTable(value coldmoon.Value, propertiesValue coldmoon.Value) (string, bool) {
	object, ok := value.GetObject()
	if !ok {
		return "", false
	}

	propertyFilter := consoleTablePropertyFilter(propertiesValue)
	rows := make([]consoleTableRow, 0)
	columns := append([]string(nil), propertyFilter...)
	seenColumns := make(map[string]struct{}, len(columns))
	for _, column := range columns {
		seenColumns[column] = struct{}{}
	}

	for _, key := range consoleEnumerableStringKeys(object) {
		row := consoleTableRow{index: key.Value, values: make(map[string]string)}
		entry := object.Get(key)
		if entryObject, isObject := entry.GetObject(); isObject {
			for _, entryKey := range consoleEnumerableStringKeys(entryObject) {
				if len(propertyFilter) > 0 && !containsConsoleColumn(propertyFilter, entryKey.Value) {
					continue
				}
				row.values[entryKey.Value] = formatConsoleValue(entryObject.Get(entryKey))
				if _, seen := seenColumns[entryKey.Value]; !seen {
					columns = append(columns, entryKey.Value)
					seenColumns[entryKey.Value] = struct{}{}
				}
			}
		} else {
			const valueColumn = "Values"
			if len(propertyFilter) == 0 || containsConsoleColumn(propertyFilter, valueColumn) {
				row.values[valueColumn] = formatConsoleValue(entry)
				if _, seen := seenColumns[valueColumn]; !seen {
					columns = append(columns, valueColumn)
					seenColumns[valueColumn] = struct{}{}
				}
			}
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return "(empty table)", true
	}
	return renderConsoleTable(rows, columns), true
}

func consoleTablePropertyFilter(value coldmoon.Value) []string {
	if value == nil || value == coldmoon.UndefinedValue {
		return nil
	}
	object, ok := value.GetObject()
	if !ok || !coldmoon.IsArray(value) {
		return nil
	}
	length := object.LengthOfArrayLike()
	if length.IsAbrupt() {
		return nil
	}
	properties := make([]string, 0, int(length.Data()))
	for index := coldmoon.JSInt(0); index < length.Data(); index++ {
		property := object.Get(coldmoon.NewIntegerIndexPropertyKey(index))
		properties = append(properties, string(property.ToString()))
	}
	return properties
}

func containsConsoleColumn(columns []string, target string) bool {
	for _, column := range columns {
		if column == target {
			return true
		}
	}
	return false
}

func renderConsoleTable(rows []consoleTableRow, columns []string) string {
	headings := append([]string{"(index)"}, columns...)
	widths := make([]int, len(headings))
	for index, heading := range headings {
		widths[index] = utf8.RuneCountInString(heading)
	}
	for _, row := range rows {
		widths[0] = max(widths[0], utf8.RuneCountInString(row.index))
		for index, column := range columns {
			widths[index+1] = max(widths[index+1], utf8.RuneCountInString(row.values[column]))
		}
	}

	renderCells := func(cells []string) string {
		padded := make([]string, len(cells))
		for index, cell := range cells {
			cell = strings.ReplaceAll(cell, "\n", "\\n")
			padded[index] = cell + strings.Repeat(" ", widths[index]-utf8.RuneCountInString(cell))
		}
		return strings.Join(padded, " | ")
	}

	lines := []string{renderCells(headings)}
	separators := make([]string, len(widths))
	for index, width := range widths {
		separators[index] = strings.Repeat("-", width)
	}
	lines = append(lines, strings.Join(separators, "-+-"))
	for _, row := range rows {
		cells := []string{row.index}
		for _, column := range columns {
			cells = append(cells, row.values[column])
		}
		lines = append(lines, renderCells(cells))
	}
	return strings.Join(lines, "\n")
}

func definePrint(realm *coldmoon.Realm) {
	coldmoon.DefineBuiltinFunction(realm, coldmoon.CMString("print"), realm.GlobalObject, jsPrint, 1)
}
