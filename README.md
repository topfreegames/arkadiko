# Arkadiko

[![Build Status](https://travis-ci.org/topfreegames/arkadiko.svg?branch=master)](https://travis-ci.org/topfreegames/arkadiko)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/arkadiko/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/arkadiko?branch=master)

A bridge from HTTP to MQTT developed in go

### Installing

```
make setup
```

### Usage

First you need to build:

```
make build
```

Then you can run arkadiko by running.

```
./arkadiko start
```

You can specify the configuration file and other parameters by setting the flags.

To specify host, port and configuration file you may call arkadiko as follows:

`./arkadiko start --bind 0.0.0.0 --port 8890 --config ./config/local.yml`

You can also run Arkadiko with:

```
make run
```

It will start Arkadiko in port 8890 with config file `/config/local.yml`

### Features

Arkadiko receives MQTT messages on the route `/sendmqtt/:topic` and expects the body to be a JSON payload. This payload is the message that will be sent to the MQTT server.

There is a small gotcha to sending hierarchical topics, the slashes must be escaped in the URL, so the called URL should be something like `/sendmqtt/top%2Flevel`.

#### Example

`echo '{"message": "hello", "number": 1}' | curl -d @- localhost:8890/sendmqtt/topic`

Sends the MQTT message `{"message":"hello","number":1}` to the topic `topic`

### Retained Messages

Arkadiko supports sending retained messages. In order to specify that the message being published should be retained you just need to send a querystring parameter of `retained=true`, like:

`echo '{"message": "hello", "number": 1}' | curl -d @- localhost:8890/sendmqtt/topic?retained=true`

Any other value passed to retained (`retained=false, retained=else, retained=`) will be treated as false.

### Testing

Run `make test`

### Coverage

Run `make coverage coverage-html`

### The name

https://en.wikipedia.org/wiki/Arkadiko_Bridge
