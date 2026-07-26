# Coldmoon architecture

Coldmoon separates ECMAScript semantics in `coldmoon` from host bindings in
`runtime`. The core package owns language state such as Agents, Realms,
execution contexts, objects, promises, and module records. Host packages install
terminal and Test262 globals without owning language execution.

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

## Planned deep modules

The remaining architecture work is sequenced by dependency:

1. consolidate module graph loading and identity;
2. make Realm and intrinsic construction atomic;
3. internalize the property model and narrow object dispatch;
4. give static and runtime syntax semantics explicit owners;
5. collapse the Test262 lifecycle into one runner.

Each module should expose a small interface, keep implementation details local,
and add an adapter seam only when at least two real implementations exist.
