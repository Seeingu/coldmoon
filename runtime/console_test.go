package runtime

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Seeingu/coldmoon/coldmoon"
)

func newConsoleTestRealm(t *testing.T) *coldmoon.Realm {
	t.Helper()
	coldmoon.InitializeConstants()
	agent := coldmoon.NewAgent()
	coldmoon.InitializeHostDefinedRealm(agent, nil)
	return agent.CurrentRealm()
}

func callConsoleMethod(
	t *testing.T,
	realm *coldmoon.Realm,
	console coldmoon.ObjectType,
	name string,
	arguments ...coldmoon.Value,
) {
	t.Helper()
	method := console.Get(coldmoon.NewStringPropertyKey(name))
	if !coldmoon.IsCallable(method) {
		t.Fatalf("console.%s is not callable", name)
	}
	result := method.Call(realm.Agent, console.ToValue(), arguments)
	if result.IsAbrupt() {
		t.Fatalf("console.%s returned abrupt completion: %v", name, result.Error())
	}
}

func TestConsoleLoggingCountingAndGrouping(t *testing.T) {
	realm := newConsoleTestRealm(t)
	var output bytes.Buffer
	console := createConsole(realm, &output, &output, time.Now)

	for _, method := range []string{
		"log", "info", "debug", "warn", "error", "assert", "count", "countReset",
		"group", "groupCollapsed", "groupEnd", "time", "timeLog", "timeEnd", "clear",
		"trace", "dir", "dirxml", "table",
	} {
		if !coldmoon.IsCallable(console.Get(coldmoon.NewStringPropertyKey(method))) {
			t.Fatalf("console.%s was not installed", method)
		}
	}

	callConsoleMethod(t, realm, console, "log", coldmoon.NewStringValue("value=%d"), coldmoon.NewNumberValue(7))
	callConsoleMethod(t, realm, console, "info", coldmoon.NewStringValue("ready"))
	callConsoleMethod(t, realm, console, "debug", coldmoon.NewStringValue("details"))
	callConsoleMethod(t, realm, console, "warn", coldmoon.NewStringValue("careful"))
	callConsoleMethod(t, realm, console, "error", coldmoon.NewStringValue("broken"))
	callConsoleMethod(t, realm, console, "assert", coldmoon.TrueValue, coldmoon.NewStringValue("not printed"))
	callConsoleMethod(t, realm, console, "assert", coldmoon.FalseValue, coldmoon.NewStringValue("expected %d"), coldmoon.NewNumberValue(7))

	callConsoleMethod(t, realm, console, "count", coldmoon.NewStringValue("job"))
	callConsoleMethod(t, realm, console, "count", coldmoon.NewStringValue("job"))
	callConsoleMethod(t, realm, console, "countReset", coldmoon.NewStringValue("job"))
	callConsoleMethod(t, realm, console, "count", coldmoon.NewStringValue("job"))
	callConsoleMethod(t, realm, console, "countReset", coldmoon.NewStringValue("missing"))

	callConsoleMethod(t, realm, console, "group", coldmoon.NewStringValue("outer"))
	callConsoleMethod(t, realm, console, "log", coldmoon.NewStringValue("inside"))
	callConsoleMethod(t, realm, console, "groupCollapsed", coldmoon.NewStringValue("nested"))
	callConsoleMethod(t, realm, console, "warn", coldmoon.NewStringValue("deep"))
	callConsoleMethod(t, realm, console, "groupEnd")
	callConsoleMethod(t, realm, console, "groupEnd")
	callConsoleMethod(t, realm, console, "groupEnd")

	want := "" +
		"value=7\n" +
		"[info] ready\n" +
		"[debug] details\n" +
		"[warn] careful\n" +
		"[error] broken\n" +
		"[error] Assertion failed: expected 7\n" +
		"job: 1\n" +
		"job: 2\n" +
		"job: 1\n" +
		"[warn] Count for \"missing\" does not exist\n" +
		"outer\n" +
		"  inside\n" +
		"  nested\n" +
		"    [warn] deep\n"
	if output.String() != want {
		t.Fatalf("console output mismatch\n--- got ---\n%s--- want ---\n%s", output.String(), want)
	}
}

func TestConsoleTimersUsePerConsoleState(t *testing.T) {
	realm := newConsoleTestRealm(t)
	var output bytes.Buffer
	start := time.Unix(100, 0)
	times := []time.Time{start, start.Add(1500 * time.Microsecond), start.Add(4 * time.Millisecond)}
	nextTime := 0
	console := createConsole(realm, &output, &output, func() time.Time {
		if nextTime >= len(times) {
			t.Fatal("console clock called more often than expected")
		}
		result := times[nextTime]
		nextTime++
		return result
	})

	callConsoleMethod(t, realm, console, "time", coldmoon.NewStringValue("load"))
	callConsoleMethod(t, realm, console, "time", coldmoon.NewStringValue("load"))
	callConsoleMethod(t, realm, console, "timeLog", coldmoon.NewStringValue("load"), coldmoon.NewStringValue("checkpoint"))
	callConsoleMethod(t, realm, console, "timeEnd", coldmoon.NewStringValue("load"))
	callConsoleMethod(t, realm, console, "timeLog", coldmoon.NewStringValue("load"))

	want := "" +
		"[warn] Timer \"load\" already exists\n" +
		"load: 1.500ms checkpoint\n" +
		"load: 4.000ms\n" +
		"[warn] Timer \"load\" does not exist\n"
	if output.String() != want {
		t.Fatalf("timer output mismatch\n--- got ---\n%s--- want ---\n%s", output.String(), want)
	}
	if nextTime != len(times) {
		t.Fatalf("clock calls = %d, want %d", nextTime, len(times))
	}

	var otherOutput bytes.Buffer
	other := createConsole(realm, &otherOutput, &otherOutput, time.Now)
	callConsoleMethod(t, realm, console, "count")
	callConsoleMethod(t, realm, other, "count")
	if !strings.HasSuffix(output.String(), "default: 1\n") || otherOutput.String() != "default: 1\n" {
		t.Fatal("console instances did not keep independent counter state")
	}
}

func TestConsoleClearTraceDirAndTable(t *testing.T) {
	realm := newConsoleTestRealm(t)
	var output bytes.Buffer
	console := createConsole(realm, &output, &output, time.Now)

	callConsoleMethod(t, realm, console, "group", coldmoon.NewStringValue("outer"))
	callConsoleMethod(t, realm, console, "clear")
	callConsoleMethod(t, realm, console, "log", coldmoon.NewStringValue("after clear"))
	if output.String() != "outer\n\x1b[2J\x1b[Hafter clear\n" {
		t.Fatalf("clear output/group reset = %q", output.String())
	}

	output.Reset()
	callConsoleMethod(t, realm, console, "trace", coldmoon.NewStringValue("failure path"))
	if !strings.Contains(output.String(), "[trace] failure path\n") || !strings.Contains(output.String(), "goroutine ") {
		t.Fatalf("trace output lacks label or stack: %q", output.String())
	}

	first := coldmoon.OrdinaryObjectCreate(realm.Agent, realm.Intrinsics.ObjectPrototype, nil)
	first.CreateDataProperty(coldmoon.NewStringPropertyKey("name"), coldmoon.NewStringValue("Ada"))
	first.CreateDataProperty(coldmoon.NewStringPropertyKey("score"), coldmoon.NewNumberValue(7))
	second := coldmoon.OrdinaryObjectCreate(realm.Agent, realm.Intrinsics.ObjectPrototype, nil)
	second.CreateDataProperty(coldmoon.NewStringPropertyKey("name"), coldmoon.NewStringValue("Lin"))
	second.CreateDataProperty(coldmoon.NewStringPropertyKey("score"), coldmoon.NewNumberValue(9))

	output.Reset()
	callConsoleMethod(t, realm, console, "dir", first.ToValue())
	if output.String() != "{name: Ada, score: 7}\n" {
		t.Fatalf("dir output = %q", output.String())
	}

	rows := coldmoon.CreateArrayFromList(realm.Agent, []coldmoon.Value{first.ToValue(), second.ToValue()})
	propertyFilter := coldmoon.CreateArrayFromList(realm.Agent, []coldmoon.Value{coldmoon.NewStringValue("name")})
	output.Reset()
	callConsoleMethod(t, realm, console, "table", rows.ToValue(), propertyFilter.ToValue())
	table := output.String()
	for _, fragment := range []string{"(index) | name", "0       | Ada", "1       | Lin"} {
		if !strings.Contains(table, fragment) {
			t.Fatalf("table output %q lacks %q", table, fragment)
		}
	}
	if strings.Contains(strings.SplitN(table, "\n", 2)[0], "score") {
		t.Fatalf("table property filter was ignored: %q", table)
	}
}
