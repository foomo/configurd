install:
	go build -o /usr/local/bin/squadron cmd/main.go

build:
	mkdir -p bin
	go build -o bin/squadron cmd/main.go

test:
	go test ./...

test.update:
	go test -update ./...
