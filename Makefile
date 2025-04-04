# arkadiko
# https://github.com/topfreegames/arkadiko
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

OS = "$(shell uname | awk '{ print tolower($$0) }')"

build:
	@go build -mod vendor -a -installsuffix cgo -o . .

vendor:
	@go mod vendor

tidy:
	@go mod tidy


run:
	@go run main.go start --rpc --rpc-port=52345 -d

run-containers:
	@cd test_containers && docker compose up -d && cd ..
	@/bin/bash -c "until curl -s localhost:8081 > /dev/null; do echo 'Waiting for EMQTT...' && sleep 1; done"

kill-containers:
	@cd test_containers && docker compose stop && cd ..

test: run-tests

run-tests: kill-containers run-containers
	@make run-test coverage
	@make kill-containers

run-test unit:
	@go run github.com/onsi/ginkgo/ginkgo -r --cover .

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

build_proto:
	@go get -v github.com/golang/protobuf/protoc-gen-go
	@protoc --go_out=plugins=grpc:. ./remote/mqtt.proto

kill-bg:
	@ps aux | egrep main.+start.+rpc | egrep -v egrep | awk ' { print $$2 } ' | xargs kill -9

run-bg: kill-bg
	@go run main.go start -v0 --port=52344 --rpc --rpc-port=52345 &

bench: kill-containers run-containers run-bg
	@sleep 3
	@${MAKE} bench-run

bench-run:
	@go test -benchmem -bench . -benchtime 5s ./bench/...

mock-lib:
	@mockgen github.com/topfreegames/arkadiko/lib ArkadikoInterface | sed 's/mock_lib/mocks/' > lib/mocks/arkadiko.go
