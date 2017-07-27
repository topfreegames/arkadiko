// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/topfreegames/arkadiko/api"
	"github.com/topfreegames/arkadiko/remote"
)

var host string
var port int
var debug bool
var rpc bool
var rpcHost string
var rpcPort int

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the arkadiko API server",
	Long: `Starts arkadiko server with the specified arguments. You can use
	environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		ll := log.InfoLevel
		switch {
		case Verbose == 0:
			ll = log.ErrorLevel
		case Verbose == 1:
			ll = log.WarnLevel
		case Verbose == 2:
			ll = log.InfoLevel
		case Verbose == 3:
			ll = log.DebugLevel
		}
		if debug {
			ll = log.DebugLevel
		}
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(ll)
		log.SetOutput(os.Stdout)

		logger := log.WithField("source", "app")
		logger.Debug("starting")
		app, err := api.GetApp(
			host,
			port,
			ConfigFile,
			debug,
			logger,
		)
		if err != nil {
			logger.WithError(err).Fatal("Could not get arkadiko application.")
		}
		logger = log.WithField("source", "rpc")
		if rpc {
			rpcServer, err := remote.NewServer(
				rpcHost,
				rpcPort,
				ConfigFile,
				debug,
				logger,
			)
			if err != nil {
				logger.WithError(err).Fatal("Could not get arkadiko RPC server.")
			}

			go func(rpcs *remote.Server) {
				err := rpcs.Start()
				if err != nil {
					logger.WithError(err).Fatal("Could not start arkadiko RPC server.")
				}
			}(rpcServer)
		}

		err = app.Start()
		if err != nil {
			logger.WithError(err).Fatal("Could not start arkadiko application.")
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind arkadiko to")
	startCmd.Flags().IntVarP(&port, "port", "p", 8890, "Port to bind arkadiko to")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
	startCmd.Flags().BoolVarP(&rpc, "rpc", "r", false, "Enable GRPC remote procedure call")
	startCmd.Flags().StringVarP(&rpcHost, "rpc-bind", "i", "0.0.0.0", "Host to bind arkadiko RPC to")
	startCmd.Flags().IntVarP(&rpcPort, "rpc-port", "t", 8891, "Port to bind arkadiko RPC to")
}
