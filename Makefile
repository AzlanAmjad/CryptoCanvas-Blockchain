APP_NAME=DreamscapeCanvas

build:
	go build -o bin/$(APP_NAME)

run: build
	./bin/$(APP_NAME)

test:
	go test -race ./...