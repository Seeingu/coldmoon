
build: main.go coldmoon/
	go build -o main.exe main.go

test: coldmoon tests
	go test ./...
