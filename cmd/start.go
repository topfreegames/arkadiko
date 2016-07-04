// mqttbridge
// https://github.com/topfreegames/mqttbridge
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/topfreegames/mqttbridge/api"
)

var host string
var port int
var debug bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the mqttbridge API server",
	Long: `Starts mqttbridge server with the specified arguments. You can use
	environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		app := api.GetApp(
			host,
			port,
			ConfigFile,
			debug,
		)

		app.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind mqttbridge to")
	startCmd.Flags().IntVarP(&port, "port", "p", 8890, "Port to bind mqttbridge to")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
}
