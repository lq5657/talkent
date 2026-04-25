.PHONY: build run clean

build:
	go build ./...

run:
	go run ./cmd/server/

clean:
	rm -f talkent talkent.db
