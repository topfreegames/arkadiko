# Arkadiko

[![Build Status](https://travis-ci.org/topfreegames/arkadiko.svg?branch=master)](https://travis-ci.org/topfreegames/arkadiko)

A bridge from HTTP to MQTT developed in go

### Installing

`make setup`

### Usage

First you need to build:

`make build`

Then you can run arkadiko by running.

`./arkadiko start`

You can specify the configuration file and other parameters by setting the flags.

To specify host, port and configuration file you may call arkadiko as follows:

`./arkadiko start --bind 0.0.0.0 --port 8890 --config ./config/local.yml`


### Features

Arkadiko receives MQTT messages on the route `/sendmqtt` and expects the body to be a JSON payload with two parameters, `topic`, a string, and `payload`, which should be a JSON parameter.

#### Example

`echo '{"topic": "test2", "payload": {"message": "hello", "number": 1} }' | curl -d @- localhost:8890/sendmqtt`

Sends the MQTT message `{"message":"hello","number":1}` to the topic `test2`

### Testing

Run `make run-tests`

### The name

https://en.wikipedia.org/wiki/Arkadiko_Bridge
