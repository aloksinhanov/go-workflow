build:
	echo "Building"
	go build

test:
	echo "Executing unit tests"
	go test ./...

cover:
	echo "Running test with coverage"
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

compile:
	echo "Compiling for every OS and Platform"
	GOOS=freebsd GOARCH=386 go build
	GOOS=linux GOARCH=386 go build
	GOOS=windows GOARCH=386 go build

all: build cover
