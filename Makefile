.PHONY: build build-all test test262 fmt cli-smoke

BINARY := bin/coldmoon$(shell go env GOEXE)

build:
	mkdir -p bin
	go build -o $(BINARY) .

build-all: build

test: coldmoon tests
	go test ./...

test262:
	COLDMOON_RUN_TEST262=1 go test ./tests -run '^Test262WithCoverage$$' -count=1 -v

fmt: coldmoon tests
	gofumpt -w -l .

cli-smoke: build
	$(BINARY) --version
	$(BINARY) --help
	$(BINARY) -e 'console.log("coldmoon cli smoke")'
