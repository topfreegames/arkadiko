// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"io"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/spf13/cobra"
)

var out io.Writer = os.Stdout

func Test(t *testing.T) {
	Describe("Root Cmd", func() {
		It("Should run command", func() {
			var rootCmd = &cobra.Command{
				Use:   "arkadiko",
				Short: "arkadiko bridges http to mqtt",
				Long:  `arkadiko bridges http to mqtt.`,
			}
			Execute(rootCmd)
		})
	})
}
