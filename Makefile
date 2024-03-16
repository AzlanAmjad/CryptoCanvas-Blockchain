APP_NAME=DreamscapeCanvas

build:
	go build -o bin/$(APP_NAME)

run: build
	./bin/$(APP_NAME)

run-verbose: build
	./bin/$(APP_NAME) -v

test:
	go test -race ./...

test-verbose:
	go test -v -race ./...