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

## Planned deep modules

The remaining architecture work is sequenced by dependency:

1. centralize execution-context and VM lifetime;
2. deepen builtin invocation and completion propagation;
3. consolidate module graph loading and identity;
4. make Realm and intrinsic construction atomic;
5. internalize the property model and narrow object dispatch;
6. give static and runtime syntax semantics explicit owners;
7. collapse the Test262 lifecycle into one runner.

Each module should expose a small interface, keep implementation details local,
and add an adapter seam only when at least two real implementations exist.
