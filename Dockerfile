FROM golang:1.6.2-alpine

MAINTAINER TFG Co <backend@tfgco.com>

EXPOSE 8890

RUN apk update
RUN apk add git

RUN go get -u github.com/Masterminds/glide/...

ADD . /go/src/github.com/topfreegames/arkadiko

WORKDIR /go/src/github.com/topfreegames/arkadiko
RUN glide install
RUN go install github.com/topfreegames/arkadiko

ENV ARKADIKO_MQTTSERVER_HOST localhost
ENV ARKADIKO_MQTTSERVER_PORT 1883
ENV ARKADIKO_MQTTSERVER_USER admin
ENV ARKADIKO_MQTTSERVER_PASS admin

CMD /go/bin/arkadiko start --bind 0.0.0.0 --port 8890 --config /go/src/github.com/topfreegames/arkadiko/config/local.yml
