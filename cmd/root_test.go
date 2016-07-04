// mqttbridge
// https://github.com/topfreegames/mqttbridge
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"io"
	"os"
	"testing"

	. "github.com/franela/goblin"
	"github.com/spf13/cobra"
)

var out io.Writer = os.Stdout

func Test(t *testing.T) {
	g := Goblin(t)

	g.Describe("Root Cmd", func() {
		g.It("Should run command", func() {
			var rootCmd = &cobra.Command{
				Use:   "mqttbridge",
				Short: "mqttbridge bridges http to mqtt",
				Long:  `mqttbridge bridges http to mqtt.`,
			}
			Execute(rootCmd)
		})
	})
}
