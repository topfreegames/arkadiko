// mqttbridge
// https://github.com/topfreegames/mqttbridge
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package main

import (
	"github.com/topfreegames/mqttbridge/cmd"
)

func main() {
	cmd.Execute(cmd.RootCmd)
}
