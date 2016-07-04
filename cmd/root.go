// mqttbridge
// https://github.com/topfreegames/mqttbridge
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ConfigFile is the configuration file used for running a command
var ConfigFile string

// Verbose determines how verbose mqttbridge will run under
var Verbose int

// RootCmd is the root command for mqttbridge CLI application
var RootCmd = &cobra.Command{
	Use:   "mqttbridge",
	Short: "mqttbridge bridges http to mqtt",
	Long:  `mqttbridge bridges http to mqtt.`,
}

// Execute runs RootCmd to initialize mqttbridge CLI application
func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().IntVarP(
		&Verbose, "verbose", "v", 0,
		"Verbosity level => v0: Error, v1=Warning, v2=Info, v3=Debug",
	)

	RootCmd.PersistentFlags().StringVarP(
		&ConfigFile, "config", "c", "./config/local.yaml",
		"config file (default is ./config/local.yaml",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if ConfigFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(ConfigFile)
	}
	viper.SetEnvPrefix("mqttbridge")
	viper.SetConfigName(".mqttbridge") // name of config file (without extension)
	viper.AddConfigPath("$HOME")       // adding home directory as first search path
	viper.AutomaticEnv()               // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}
