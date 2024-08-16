.PHONY build:
build:
	go build -o ./cmd/bin/bot ./cmd/bot/main.go

.PHONY run:
run:
	./cmd/bot/bot

.PHONY all:
all: build run

.PHONY clear:
clear:
	rm ./cmd/bin/*