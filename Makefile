build: main.go coldmoon/
	go build -o main.exe .

build-all: build

test: coldmoon tests
	go test ./...

fmt: coldmoon tests
	gofumpt -w -l .