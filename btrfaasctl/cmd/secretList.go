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
	"fmt"

	yaml "gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/trusch/btrfaas/deployment"
)

// secretListCmd represents the secretList command
var secretListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list deployed secrets",
	Long:    `list deployed secrets`,
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Flags().GetString("env")
		cli, err := deployment.NewSwarmPlatform()
		if err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()
		secrets, err := cli.ListSecrets(ctx, &deployment.ListSecretsOptions{
			EnvironmentID: env,
		})
		if err != nil {
			log.Fatal(err)
		}
		if len(secrets) == 0 {
			log.Info("no secrets are deployed")
		} else {
			bs, _ := yaml.Marshal(secrets)
			fmt.Print(string(bs))
		}
	},
}

func init() {
	secretCmd.AddCommand(secretListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// secretListCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// secretListCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
