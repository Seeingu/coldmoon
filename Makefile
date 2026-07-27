build: main.go coldmoon/
	go build -o main.exe .

build-all: build

test: coldmoon tests
	go test ./...

test262:
	COLDMOON_RUN_TEST262=1 go test ./tests -run '^Test262WithCoverage$$' -count=1 -v

fmt: coldmoon tests
	gofumpt -w -l .
