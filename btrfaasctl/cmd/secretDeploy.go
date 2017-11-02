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
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/trusch/btrfaas/deployment"

	"github.com/spf13/cobra"
)

// secretDeployCmd represents the secretDeploy command
var secretDeployCmd = &cobra.Command{
	Use:     "deploy <secret id> <secret content>",
	Aliases: []string{"create", "add"},
	Short:   "deploy a secret",
	Long:    `deploy a secret`,
	Run: func(cmd *cobra.Command, args []string) {
		secretID, secretValue := getIDAndValue(cmd, args)
		env, _ := cmd.Flags().GetString("env")
		cli, err := deployment.NewSwarmPlatform()
		if err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()
		err = cli.DeploySecret(ctx, &deployment.DeploySecretOptions{
			EnvironmentID: env,
			ID:            secretID,
			Value:         secretValue,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Info("successfully deployed secret ", secretID)
	},
}

func getIDAndValue(cmd *cobra.Command, args []string) (id, val string) {
	if len(args) < 1 {
		cmd.Help()
		os.Exit(1)
	}
	secretID := args[0]
	secretValue := ""
	filePath, err := cmd.Flags().GetString("file")
	if err != nil || filePath == "" {
		if len(args) < 2 {
			cmd.Help()
			os.Exit(1)
		}
		secretValue = args[1]
	}
	if secretValue == "" {
		bs, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}
		secretValue = string(bs)
	}
	return secretID, secretValue
}

func init() {
	secretCmd.AddCommand(secretDeployCmd)
	secretDeployCmd.Flags().StringP("file", "f", "", "")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// secretDeployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// secretDeployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
