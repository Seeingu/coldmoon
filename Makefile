ICU4XGO_DIR := $(shell go list -f "{{.Dir}}" github.com/Seeingu/icu4xgo)

static:
	go mod download
	cd ${ICU4XGO_DIR} && sudo make rustlib

build: main.go coldmoon/
	go build -o main.exe .

build-all: static build

test: coldmoon tests
	go test ./...

fmt: coldmoon tests
	gofumpt -w -l .