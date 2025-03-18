# arkadiko
# https://github.com/topfreegames/arkadiko
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

FROM golang:1.23-alpine AS build

LABEL app=arkadiko
LABEL builder=true
LABEL maintainer='TFG CO <backend@tfgco.com>'

WORKDIR /src

COPY . .

# Build a static binary.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o arkadiko .

# Verify if the binary is truly static.
RUN ldd /src/arkadiko 2>&1 | grep -q 'Not a valid dynamic program'

FROM alpine

COPY --from=build /src/arkadiko ./arkadiko
COPY --from=build /src/config ./config

ENV ARKADIKO_MQTTSERVER_HOST localhost
ENV ARKADIKO_MQTTSERVER_PORT 1883
ENV ARKADIKO_MQTTSERVER_USER admin
ENV ARKADIKO_MQTTSERVER_PASS admin
ENV USE_BASICAUTH false
ENV ARKADIKO_BASICAUTH_USERNAME ""
ENV ARKADIKO_BASICAUTH_PASSWORD ""

ENTRYPOINT ["./arkadiko", "start", "--bind", "0.0.0.0", "--port", "8890", "--config", "/config/local.yml"]
EXPOSE 8890

LABEL app=arkadiko
LABEL maintainer='TFG CO <backend@tfgco.com>'
