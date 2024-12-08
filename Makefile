
build: main.go coldmoon/
	go build -o main.exe main.go

test: coldmoon tests
	go test ./...

fmt: coldmoon tests
	gofumpt -w -l .