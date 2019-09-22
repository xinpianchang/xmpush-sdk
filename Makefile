SRC_DIR := $(shell ls -d */|grep -vE 'vendor|script|tmp')

all: test

fmt:
	# gofmt code
	gofmt -s -l -w $(SRC_DIR) .

test:
	go test -coverprofile .cover.out -v ./...
	# cover
	go tool cover -func=.cover.out
	go tool cover -html=.cover.out -o .cover.html


.PHONY: all fmt test
