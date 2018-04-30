SRC_DIR := $(shell ls -d */|grep -vE 'vendor|script|tmp')

all: test

deps:
	# install deps
	@hash dep > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u github.com/golang/dep/cmd/dep; \
	fi
	@dep ensure -v

fmt:
	# gofmt code
	gofmt -s -l -w $(SRC_DIR) .

test:
	go test -coverprofile .cover.out -v ./...
	# cover
	go tool cover -func=.cover.out
	go tool cover -html=.cover.out -o .cover.html


.PHONY: all deps fmt test