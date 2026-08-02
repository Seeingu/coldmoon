# coldmoon

[In Development], A JavaScript runtime

<picture>
    <img alt="Coldmoon"
         src="https://raw.githubusercontent.com/Seeingu/coldmoon/spec/_assets/icon.webp"
         width="50%">
</picture>

## Build

Install the current module directly with Go:

```bash
go install github.com/Seeingu/coldmoon@latest
```

To build from a checkout, clone the repository with its Test262 submodule:

```bash
git clone --recurse-submodules https://github.com/Seeingu/coldmoon.git
cd coldmoon
make build
```

The executable is written to `bin/coldmoon` (`bin/coldmoon.exe` when Go's
`GOEXE` is `.exe`). `make build-all` remains an alias for `make build`.

Run the bounded test suite, race checks, and the CLI smoke test with:

```bash
make test
make race
make cli-smoke
```

The regular and race test commands are bounded and skip the full Test262
coverage scan. CI runs both before the CLI smoke test. Run the compatibility
suite explicitly with:

```bash
make test262
```

## Embedding

For Go embedders, `Source.Name` is now diagnostic-only. `EvaluateSource` treats
module input as temporary; callers that previously relied on `Source.Name` as a
module cache key should migrate to `EvaluateSourceWithModuleIdentity` and pass
the same canonical identity returned by their module loader.

## CLI

```text
coldmoon [options] [file|-] [arguments...]

Options:
  -e, --eval <source>          Evaluate source text
  -p, --print <source>         Evaluate source text and print its result
  -i, --interactive            Enter the REPL after successful evaluation
      --check                  Check syntax without executing source
      --input-type <type>      Treat batch input as script or module
  -h, --help                   Print usage
  -v, --version                Print the Coldmoon version
      --                       End option parsing
```

The first file operand ends option parsing; every remaining token is passed to
the program unchanged. With `--eval` or `--print`, the first non-option after
the source starts the program argument list. Use `--` before a program argument
that begins with `-`.

Version output comes from Go build metadata. Local builds are identified as
`devel` and include the short revision and dirty state when that information is
available.

Batch input uses these defaults:

| Input source | Default type | `process.argv` |
| --- | --- | --- |
| `file` | module | `[executableAbsolutePath, fileAbsolutePath, ...arguments]` |
| `--eval` / `--print` | script | `[executableAbsolutePath, ...arguments]` |
| explicit `-` stdin | script | `[executableAbsolutePath, "-", ...arguments]` |
| implicit non-TTY stdin | script | `[executableAbsolutePath]` |
| TTY REPL | script | `[executableAbsolutePath]` |

`--input-type script|module` overrides the type for batch input. The REPL is
always a script. With no file or inline source, Coldmoon reads all of stdin once
when stdin is not a TTY, and opens the REPL when it is a TTY. Piped input never
prints prompts. Relative imports in a file resolve from that file's directory;
relative imports in inline or stdin modules resolve from the invocation working
directory.

`--eval` and `--print` are mutually exclusive. `--check` cannot be combined
with `--print` or `--interactive`; `--print` cannot evaluate modules; and
explicit stdin (`-`) cannot be combined with `--interactive`. When
`--interactive` follows a file, eval, or print operation, the REPL opens in the
same Realm only if that operation succeeds. Syntax checking parses only the
entry source: it neither executes it nor loads imported modules.

### REPL

The REPL uses `> ` for a new submission and `... ` while a block, template, or
block comment is incomplete. Submissions share one Realm, successful values are
printed, and syntax or runtime errors do not end the session. `.help` lists the
REPL commands and `.exit` exits. Input is not limited to 64 KiB; history,
completion, raw-terminal line editing, and colors are intentionally outside the
current MVP.

Only a typed incomplete-syntax diagnostic requests another line. Reaching EOF
with no buffered source exits successfully; EOF in the middle of an incomplete
submission reports the diagnostic and exits with status 1.

### Diagnostics and exit status

Diagnostics are written to stderr without ANSI escape sequences. Source errors
include a source name, 1-based Unicode code-point line and column, the source
line, and a caret when that location is available. JavaScript `Error` values
are reported by name and message; throwing another value is reported as
`Uncaught <value>`. User source failures do not expose Go stacks.

| Status | Meaning |
| --- | --- |
| `0` | success, help, or version |
| `1` | source, I/O, linking, or runtime failure |
| `2` | invalid arguments or an invalid option combination |

### Host globals

The terminal host exposes a mutable JavaScript Array at `process.argv` using
the shapes above. No other Node.js `process` API is provided.

## Terminal console

`runtime.RegisterTerminalRuntime` installs a stateful WHATWG-style `console`
with logging, assertions, counters, groups, timers, traces, directory output,
and tables. `runtime.RegisterTerminalRuntimeWithOptions` lets embedders inject
stdout, stderr, and `process.argv`; the original registration function keeps
using the OS defaults. Embedders can call `runtime.CreateConsoleWithWriter`
when console output needs to be captured or redirected; each created console
keeps its own counter, group, and timer state.

## Architecture

See [docs/architecture.md](docs/architecture.md) for runtime ownership,
scheduler invariants, and the deep-module roadmap.

## Inspiration

This project is highly inspired by [kiesel](https://codeberg.org/kiesel-js/kiesel).

## License

MIT
