// Copyright © 2017 Tino Rusch
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	grpcServer "github.com/trusch/btrfaas/fgateway/grpc"
	httpServer "github.com/trusch/btrfaas/fgateway/http"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "fgateway",
	Short: "a gateway to your functions",
	Long:  `a gateway to your functions`,
	Run: func(cmd *cobra.Command, args []string) {
		go runHTTPServer(cmd)
		go runGRPCServer(cmd)
		select {}
	},
}

func runHTTPServer(cmd *cobra.Command) {
	httpAddr, _ := cmd.Flags().GetString("http-address")
	grpcPort, _ := cmd.Flags().GetUint16("grpc-default-port")
	handler := httpServer.NewFunctionDispatcher(grpcPort)
	log.Infof("start listening for HTTP requests on %v", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, handler))
}

func runGRPCServer(cmd *cobra.Command) {
	grpcAddr, _ := cmd.Flags().GetString("grpc-address")
	grpcPort, _ := cmd.Flags().GetUint16("grpc-default-port")
	grpcServer := grpcServer.NewServer(grpcAddr, grpcPort)
	log.Fatal(grpcServer.ListenAndServe())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.Flags().String("http-address", ":8080", "http listen address")
	RootCmd.Flags().String("grpc-address", ":2424", "grpc listen address")
	RootCmd.Flags().Uint16("grpc-default-port", 2424, "grpc default port")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".fgateway" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".fgateway")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
