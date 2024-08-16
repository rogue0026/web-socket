.PHONY build:
build:
	go build -o ./cmd/bin/bot -race ./cmd/bot/main.go

.PHONY run:
run:
	./cmd/bin/bot

.PHONY all:
all: build run

.PHONY clean:
clean:
	rm ./cmd/bin/*