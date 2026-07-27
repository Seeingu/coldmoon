# coldmoon

[In Development], A JavaScript runtime

<picture>
    <img alt="Coldmoon"
         src="https://raw.githubusercontent.com/Seeingu/coldmoon/spec/_assets/icon.webp"
         width="50%">
</picture>

## Build

### 1. Clone the Repository

To clone the repository along with the test262 test suite, use the following command:

```bash
git clone --recurse-submodules https://github.com/Seeingu/coldmoon.git
```

### 2. Build and Test

To build the project and run the tests, use the following commands:

```bash
make build-all

make test
```

The regular test command is bounded and skips the full Test262 coverage scan.
Run that suite explicitly with:

```bash
make test262
```

The CLI Test262 host also requires an explicit checkout root through
`-test262-root` or `TEST262_ROOT`.

## Architecture

See [docs/architecture.md](docs/architecture.md) for runtime ownership,
scheduler invariants, and the deep-module roadmap.

## Inspiration

This project is highly inspired by [kiesel](https://codeberg.org/kiesel-js/kiesel).

## License

MIT
