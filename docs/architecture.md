# Coldmoon architecture

Coldmoon separates ECMAScript semantics in `coldmoon` from host bindings in
`runtime`. The core package owns language state such as Agents, Realms,
execution contexts, objects, promises, and module records. Host packages install
terminal and Test262 globals without owning language execution.

## CLI orchestration

`internal/cli` is the command-line boundary. Its sole public operation,
`Run(Invocation) int`, receives the complete argument vector, working
directory, stdin, stdout, stderr, TTY state, and version as injected values. It
parses the command contract, owns one Agent and Realm for the invocation, and
returns a stable process status. It never calls `os.Exit`.

The root `main` package is deliberately thin: it resolves operating-system
state, constructs the `Invocation`, and exits with the returned status. This
keeps command parsing and execution testable without mutating global process
state. A file, inline source, stdin batch, and REPL submission all pass through
the same source boundary. `--interactive` reuses the Realm created for the
successful preceding batch operation.

## Safe source execution

The `coldmoon` package describes an entry point as `Source`: text, display name,
base directory, and script-or-module kind. `CheckSource` parses only that entry
point and does not resolve or load its imports. `EvaluateSource` parses, links
when needed, executes, and drains the Agent scheduler before returning. The
older `ParseScript`, `ParseModule`, `Evaluate`, and `EvaluateModule` entry
points remain compatibility wrappers over the same implementation.

For module evaluation, a file source uses its containing directory as
`Source.BaseDir`; inline and stdin sources use the invocation working directory.
Relative import resolution starts from that base.

Expected failures cross this boundary as a typed `Diagnostic`. A diagnostic
preserves its category, JavaScript error name and message, source name,
optional span and source line, incomplete-input classification, original thrown
value, and cause. Syntax failures and known JavaScript abrupt completions are
converted; internal invariant failures remain Go panics. The CLI can therefore
render stable user diagnostics without hiding engine defects, and the REPL can
request continuation based on `Incomplete` rather than error-message matching.

## Terminal host ownership

The terminal host owns host-visible terminal bindings, not program execution or
process lifetime. `TerminalOptions` supplies stdout, stderr, and argv to
`RegisterTerminalRuntimeWithOptions`; `RegisterTerminalRuntime` remains the
compatibility adapter that uses OS defaults. Console methods, the global
`print`, CLI print mode, and REPL result display share the same JavaScript value
formatting contract.

The only Node-style process surface is `process.argv`, installed as an ordinary
mutable JavaScript Array. The CLI determines its entries from the input form,
while the terminal runtime only publishes the injected values. In the
CLI-facing path, console and global `print` output flows through the writers
owned by the host boundary; expected source failures return diagnostics rather
than writing directly to process output.

## Runtime scheduling

Every `Agent` owns one `Scheduler`. The scheduler is responsible for:

- draining promise jobs in FIFO order and exactly once;
- running timers by deadline and insertion order;
- tracking finite asynchronous tasks before their goroutines start;
- continuing until jobs, timers, and tracked tasks are all idle.

`Evaluate` and `EvaluateModule` run the scheduler before returning. Callers do
not separately poll an event loop or wait on a `sync.WaitGroup`.

Time is the scheduler's only replaceable seam. Production Agents use
`RealClock`; tests may construct an Agent with `NewAgentWithClock` to run timers
without wall-clock sleeps.

Host timer bindings validate and convert JavaScript arguments, then submit the
callback to the Agent scheduler. Promise implementations enqueue jobs through
the same scheduler, so callbacks created by timers and callbacks created by
promises share one turn model.

## Scheduler invariants

1. A queued promise job is removed before it runs.
2. Jobs queued by a running job execute later in the same drain.
3. Timers with equal deadlines retain insertion order.
4. A tracked asynchronous task increments the active count before launch.
5. `RunUntilIdle` returns only when no queued, timed, or tracked work remains.

These invariants are covered by scheduler contract tests in
`coldmoon/scheduler_test.go` and host-level regression tests in
`tests/simple_test.go`.

## Execution contexts and VMs

An execution context is entered through one Agent-owned transition. Ordinary
script, function, builtin, eval, and module execution receives an
`ExecutionContextScope` and defers `Leave`; early returns and panics therefore
cannot silently leak a frame. Generator and async lifecycles use the matching
resume and suspend transitions because their frame lifetime crosses a lexical
call boundary.

The transition creates the context's VM on first entry. `RunNode` reuses that
VM and only applies node-local strictness for the duration of an evaluation.
This gives mutable interpreter state one owner and preserves it across nested
evaluation without allowing it to leak into the next context.

Direct eval selects its lexical, variable, and private environments before
entering the eval context, then evaluates the parsed script in that context.
It does not call the top-level script entry point, which would replace those
environments with the Realm global environment.

## Builtin invocation and completions

Every builtin call crosses one boundary that owns its execution context,
normalizes an empty Go result to JavaScript `undefined`, and converts
language-level panics into throw completions. Unexpected internal panics remain
panics so implementation defects are not disguised as JavaScript exceptions.

`BuiltinInvocation` describes the receiver, arguments, and optional new target.
Its argument accessor implements the ECMAScript rule that an omitted argument
is `undefined`. Builtin implementations use the same `argumentAt` primitive,
eliminating direct constant-index reads that could leak a Go bounds panic.

Completion conversion preserves the completion type, target, and error while
changing only its generic payload type. This keeps `return`, `break`,
`continue`, and `throw` intact across helper boundaries.

## Iterator consumption boundaries

Iterator acquisition preserves language completions from the iterator method
and `next` lookup. Once a record exists, failures from `next`, `done`, and
`value` are returned directly and mark that record done; consumers do not call
`return` again for them.

Once a consumer has obtained an item, failures introduced while consuming that
item have a different boundary. `Map`, `Set`, and `Object.fromEntries` close the
iterator when entry validation, entry property access, or the collection adder
completes abruptly. The close is observable exactly once. An original throw
takes precedence over a failure from the iterator's `return` getter or method,
but the getter and method still run for their side effects.

Completion-aware callers use `GetMethodCompletion`,
`GetIteratorFromMethodCompletion`, `IteratorStepValue`, and `IteratorClose`.
The panic-style wrappers remain only for older algorithms that have not yet
migrated to explicit completion propagation.

## Module graph

Each Agent owns one `ModuleGraph`. The host first resolves a referrer and
specifier to a canonical identity; the graph checks its Realm-local cache; only
a cache miss asks the host to load source and invokes the parser. Records enter
the cache before dependency loading, so every edge in a cyclic graph converges
on the same record.

The language core knows neither filesystem paths nor file-reading APIs.
`runtime.FilesystemModuleLoader` implements the terminal/Test262 host seam and
uses absolute cleaned paths as identities. Other hosts can supply embedded or
network-backed loaders without changing parsing, linking, or evaluation.

Per-module `LoadedModules` maps retain specifier edges required by ECMAScript
algorithms, while canonical ownership and deduplication live only in
`ModuleGraph`.

## Realm bootstrap

A Realm has two internal states: building and ready. `CreateRealm` constructs
all intrinsics on a private draft, reflects over the completed intrinsic record
to reject any missing field, installs restricted function properties, and only
then marks the Realm ready. Global object/environment publication and default
global bindings reject a building Realm.

Intrinsic construction is split into dependency-checked phases: foundations,
iteration/callables, the standard library, typed arrays, utility namespaces,
and errors. Each phase validates the key products of earlier phases before it
runs, so an unsafe reorder fails at the phase boundary with the missing
intrinsic named explicitly.

Function calls currently use the VM's eager execution path. Proper tail calls
require a continuation trampoline and are intentionally not exposed through a
dormant call-site flag; adding them is an execution-engine change rather than
an expression-evaluator toggle.

Global constructor exposure is a declarative mapping from JavaScript names to
`IntrinsicName`. This keeps global publication on the same lookup path used by
the rest of the runtime instead of maintaining a second set of direct field
references.

## Objects and properties

Object property storage and internal-method tables are package-private
implementation details. Host packages interact through `ObjectType` semantic
operations and cannot mutate the backing map or replace dispatch functions.
Exotic object constructors inside the core may still override the specific
internal methods their ECMAScript semantics require.

`PropertyStorage` records insertion order alongside descriptor lookup.
`OrderedKeys` is the single ordinary-key ordering implementation: array-index
strings sort numerically, other strings retain insertion order, and Symbols
retain insertion order after all strings. Deleting and redefining a string or
Symbol gives it a new insertion position.

## Syntax semantics

`StaticSemantics` owns parse-time whole-program queries. Scripts cache
declaration snapshots, functions analyze parameters and declarations through
the same owner, and modules compute requests/imports/exports once when parsed.
Runtime instantiation consumes those records instead of rediscovering facts
through scattered AST assertions.

`RuntimeSemantics` owns Evaluation for the current execution context and its VM.
`RunNode` applies temporary strictness and delegates to that owner. Concrete AST
structs explicitly implement their runtime methods; they no longer anonymously
embed `ASTNode` merely to satisfy the interface through nil method promotion.

## Test262 lifecycle

`Test262Suite` requires an explicit, validated checkout root. A
`Test262Runtime` owns the harness set for one Realm, so its lifetime is ordinary
Go ownership rather than a global Realm-keyed map with manual release.

`Test262Runner` creates an isolated Agent/Realm per case and returns through a
buffered result channel with a per-file context deadline. Its cache keys include
the runner version, portable suite-relative path, and test contents. The full
coverage scan is opt-in through `make test262`; ordinary `go test ./...` remains
bounded.

## Verification policy

Each deep module has focused contract tests and participates in the ordinary
package regression and race checks. The Test262 coverage suite is a separate,
explicit compatibility run because its thousands of cases and persistent cache
have a different lifecycle from unit and integration tests.

Each module should expose a small interface, keep implementation details local,
and add an adapter seam only when at least two real implementations exist.
