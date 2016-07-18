FROM golang:1.6.2-alpine

MAINTAINER TFG Co <backend@tfgco.com>

EXPOSE 8890

RUN apk update
RUN apk add git nginx supervisor apache2-utils make g++

RUN go get -u github.com/Masterminds/glide/...

ADD . /go/src/github.com/topfreegames/arkadiko

WORKDIR /go/src/github.com/topfreegames/arkadiko
RUN glide install
RUN go install github.com/topfreegames/arkadiko

ENV ARKADIKO_MQTTSERVER_HOST localhost
ENV ARKADIKO_MQTTSERVER_PORT 1883
ENV ARKADIKO_MQTTSERVER_USER admin
ENV ARKADIKO_MQTTSERVER_PASS admin
ENV ARKADIKO_REDIS_HOST localhost
ENV ARKADIKO_REDIS_PORT 6379
ENV ARKADIKO_REDIS_MAXPOLLSIZE 20
ENV ARKADIKO_REDIS_PASSWORD ""
ENV USE_BASICAUTH false
ENV BASICAUTH_USERNAME arkadiko
ENV BASICAUTH_PASSWORD arkadiko

RUN mkdir -p /etc/nginx/sites-enabled
ADD ./docker/supervisord-arkadiko.conf /etc/supervisord-arkadiko.conf
ADD ./docker/nginx_conf /etc/nginx/nginx.conf

CMD /bin/sh -l -c docker/start.sh
