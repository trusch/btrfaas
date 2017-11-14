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
	"errors"
	"fmt"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trusch/btrfaas/deployment"
	"github.com/trusch/btrfaas/deployment/docker"
	"github.com/trusch/btrfaas/deployment/swarm"
	"github.com/trusch/btrfaas/faas"
	"github.com/trusch/btrfaas/faas/btrfaas"
	"github.com/trusch/btrfaas/faas/openfaas"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "btrfaasctl",
	Short: "control your btrfaas cluster",
	Long:  `control your btrfaas cluster`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringP("env", "e", "btrfaas_default", "environment to use")
	RootCmd.PersistentFlags().String("platform", "docker", "deployment platform (docker, swarm)")
	RootCmd.PersistentFlags().String("faas-provider", "btrfaas", "faas provider (btrfaas, openfaas)")
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

		// Search config in home directory with name ".btrfaasctl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".btrfaasctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func getDeploymentPlatform(cmd *cobra.Command) deployment.Platform {
	platformID, err := cmd.Flags().GetString("platform")
	if err != nil {
		log.Fatal(err)
	}
	var platform deployment.Platform
	switch platformID {
	case "docker":
		platform, err = docker.NewPlatform()
	case "swarm":
		platform, err = swarm.NewPlatform()
	default:
		err = errors.New("deployment platform unsupported")
	}
	if err != nil {
		log.Fatal(err)
	}
	return platform
}

func getFaaS(cmd *cobra.Command) faas.FaaS {
	faasID, err := cmd.Flags().GetString("faas-provider")
	if err != nil {
		log.Fatal(err)
	}
	var result faas.FaaS
	switch faasID {
	case "btrfaas":
		result = btrfaas.New(getDeploymentPlatform(cmd))
	case "openfaas":
		result = openfaas.New(getDeploymentPlatform(cmd))
	default:
		err = errors.New("faas provider unsupported")
	}
	if err != nil {
		log.Fatal(err)
	}
	return result
}
