static:
	cd ./thirdparty/icu4xgo && make rustlib
	cd ./thirdparty/icu4xgo && make install

build: main.go coldmoon/
	go build -o main.exe .

build-all: static build

test: coldmoon tests
	go test ./...

fmt: coldmoon tests
	gofumpt -w -l .