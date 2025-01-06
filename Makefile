ICU4XGO_DIR := $(shell go list -f "{{.Dir}}" github.com/Seeingu/icu4xgo)

static:
	go mod download
	cp ${ICU4XGO_DIR}/Cargo.toml ./
	cargo rustc -p icu_capi --crate-type staticlib --release
	sudo cp ./target/release/libicu_capi.a ${ICU4XGO_DIR}/lib

build: main.go coldmoon/
	go build -o main.exe .

build-all: static build

test: coldmoon tests
	go test ./...

fmt: coldmoon tests
	gofumpt -w -l .