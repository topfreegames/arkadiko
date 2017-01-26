# arkadiko
# https://github.com/topfreegames/arkadiko
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

PACKAGES = $(shell glide novendor)
GODIRS = $(shell go list ./... | grep -v /vendor/ | sed s@github.com/topfreegames/arkadiko@.@g | egrep -v "^[.]$$")
OS = "$(shell uname | awk '{ print tolower($$0) }')"

setup:
	@go get -u github.com/Masterminds/glide/...
	@go get -v github.com/spf13/cobra/cobra
	@glide install

setup-ci:
	@sudo add-apt-repository -y ppa:masterminds/glide && sudo apt-get update
	@sudo apt-get install -y glide
	@go get github.com/topfreegames/goose/cmd/goose
	@go get github.com/mattn/goveralls
	@glide install

build:
	@go build $(PACKAGES)
	@go build

install:
	@go install

# get a redis instance up (localhost:4444)
redis:
	@redis-server ./redis.conf; sleep 1
	@redis-cli -p 4444 info > /dev/null

# kill this redis instance (localhost:4444)
kill-redis:
	@-redis-cli -p 4444 shutdown

run:
	@go run main.go start

run-containers:
	@cd test_containers && docker-compose up -d && cd ..

kill-containers:
	@cd test_containers && docker-compose stop && cd ..

test: run-tests

run-tests: kill-containers run-containers
	@make run-test coverage
	@make kill-containers

run-test unit:
	@ginkgo -r --cover .

test-coverage-run:
	@mkdir -p _build
	@-rm -rf _build/test-coverage-all.out
	@echo "mode: count" > _build/test-coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> _build/test-coverage-all.out; done'

test-coverage-func:
	@echo
	@echo "=-=-=-=-=-=-="
	@echo "Test Coverage"
	@echo "=-=-=-=-=-=-="
	@go tool cover -func=_build/test-coverage-all.out

test-coverage-html cover:
	@go tool cover -html=_build/test-coverage-all.out

test-coverage-write-html:
	@go tool cover -html=_build/test-coverage-all.out -o _build/test-coverage.html

coverage: test-coverage-run test-coverage-func

coverage-html:
	@go tool cover -html=coverage-all.out

cross: cross-linux cross-darwin

cross-linux:
	@mkdir -p ./bin
	@echo "Building for linux-i386..."
	@env GOOS=linux GOARCH=386 go build -o ./bin/arkadiko-linux-i386 ./main.go
	@echo "Building for linux-x86_64..."
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/arkadiko-linux-x86_64 ./main.go
	@$(MAKE) cross-exec

cross-darwin:
	@mkdir -p ./bin
	@echo "Building for darwin-i386..."
	@env GOOS=darwin GOARCH=386 go build -o ./bin/arkadiko-darwin-i386 ./main.go
	@echo "Building for darwin-x86_64..."
	@env GOOS=darwin GOARCH=amd64 go build -o ./bin/arkadiko-darwin-x86_64 ./main.go
	@$(MAKE) cross-exec

cross-exec:
	@chmod +x bin/*
