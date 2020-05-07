pack-example:
	go-bindata -prefix=example -o exampledata/exampledata.go -pkg exampledata example/...

build: pack-example
	go build -o /usr/local/bin/configurd cmd/main.go

test:
	go test ./...