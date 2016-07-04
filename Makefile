# mqttbridge
# https://github.com/topfreegames/mqttbridge
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

PACKAGES = $(shell glide novendor)
GODIRS = $(shell go list ./... | grep -v /vendor/ | sed s@github.com/topfreegames/mqttbridge@.@g | egrep -v "^[.]$$")
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

run:
	@go run main.go start

test:
	@go test $(PACKAGES)

coverage:
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg) && \
		tail -n +2 coverage.out >> coverage-all.out;)

coverage-html:
	@go tool cover -html=coverage-all.out
