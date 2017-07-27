# arkadiko
# https://github.com/topfreegames/arkadiko
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

FROM golang:1.8-alpine3.6

MAINTAINER TFG Co <backend@tfgco.com>

EXPOSE 8890

RUN apk update
RUN apk add bash git make g++ apache2-utils

# http://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ADD bin/arkadiko-linux-x86_64 /go/bin/arkadiko
RUN chmod +x /go/bin/arkadiko

RUN mkdir -p /home/arkadiko/
ADD ./config/local.yml /home/arkadiko/local.yml

ENV ARKADIKO_MQTTSERVER_HOST localhost
ENV ARKADIKO_MQTTSERVER_PORT 1883
ENV ARKADIKO_MQTTSERVER_USER admin
ENV ARKADIKO_MQTTSERVER_PASS admin
ENV ARKADIKO_REDIS_HOST localhost
ENV ARKADIKO_REDIS_PORT 6379
ENV ARKADIKO_REDIS_MAXPOLLSIZE 20
ENV ARKADIKO_REDIS_PASSWORD ""
ENV USE_BASICAUTH false
ENV ARKADIKO_BASICAUTH_USERNAME ""
ENV ARKADIKO_BASICAUTH_PASSWORD ""

CMD /go/bin/arkadiko start --bind 0.0.0.0 --port 8890 --config /home/arkadiko/local.yml
