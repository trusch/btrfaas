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
	"context"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trusch/btrfaas/deployment"
)

// scaleServiceCmd represents the scaleService command
var scaleServiceCmd = &cobra.Command{
	Use:   "scale <service id> <scale>",
	Short: "scale a service",
	Long:  `scale a service`,
	Run: func(cmd *cobra.Command, args []string) {
		cli := getDeploymentPlatform(cmd)
		if len(args) != 2 {
			cmd.Help()
			os.Exit(1)
		}
		env := viper.GetString("env")
		serviceID := args[0]
		scale, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()
		err = cli.ScaleService(ctx, &deployment.ScaleServiceOptions{
			EnvironmentID: env,
			ID:            serviceID,
			Scale:         scale,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("successfully scaled service %v to %v instances", serviceID, scale)
	},
}

func init() {
	serviceCmd.AddCommand(scaleServiceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scaleServiceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scaleServiceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
