static:
	cd ./thirdparty/icu4xgo && make rustlib
	cp ./thirdparty/icu4xgo/lib/*.a ./lib/

build: main.go coldmoon/
	go build -o main.exe .

build-all: static build

test: coldmoon tests
	go test ./...

fmt: coldmoon tests
	gofumpt -w -l .