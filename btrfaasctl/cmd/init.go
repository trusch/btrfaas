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

	log "github.com/Sirupsen/logrus"
	"github.com/trusch/btrfaas/deployment"
	"github.com/trusch/btrfaas/faas"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init your faas",
	Long:  `init your faas`,
	Run: func(cmd *cobra.Command, args []string) {
		cli := getFaaS(cmd)
		env := viper.GetString("env")
		ctx := context.Background()
		gatewayImage, _ := cmd.Flags().GetString("gateway-image")
		err := cli.Init(ctx, &faas.InitOptions{
			PrepareEnvironmentOptions: deployment.PrepareEnvironmentOptions{
				ID: env,
			},
			GatewayImage: gatewayImage,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Info("successfully prepared faas ", env)
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().String("gateway-image", "btrfaas/fgateway:v0.3.3", "gateway image to use")
}
