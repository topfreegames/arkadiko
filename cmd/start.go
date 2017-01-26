// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/topfreegames/arkadiko/api"
	"github.com/uber-go/zap"
)

var host string
var port int
var debug bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the arkadiko API server",
	Long: `Starts arkadiko server with the specified arguments. You can use
	environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		ll := zap.InfoLevel
		if debug {
			ll = zap.DebugLevel
		}
		logger := zap.New(
			zap.NewJSONEncoder(),
			ll,
		).With(
			zap.String("source", "app"),
		)
		app, err := api.GetApp(
			host,
			port,
			ConfigFile,
			debug,
			logger,
		)

		if err != nil {
			logger.Fatal("Could not get arkadiko application.", zap.Error(err))
		}

		err = app.Start()
		if err != nil {
			logger.Fatal("Could not start arkadiko application.", zap.Error(err))
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind arkadiko to")
	startCmd.Flags().IntVarP(&port, "port", "p", 8890, "Port to bind arkadiko to")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
}
