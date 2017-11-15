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

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/trusch/btrfaas/faas"
	yaml "gopkg.in/yaml.v2"
)

// functionDeployCmd represents the functionDeploy command
var functionDeployCmd = &cobra.Command{
	Use:   "deploy [<function spec>]",
	Short: "deploy a function",
	Long:  `deploy a function`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			os.Exit(1)
		}
		env, _ := cmd.Flags().GetString("env")
		cli := getFaaS(cmd)
		for _, arg := range args {
			bs, err := ioutil.ReadFile(arg)
			if err != nil {
				log.Fatal(err)
			}
			opts := faas.DeployFunctionOptions{}
			if err = yaml.Unmarshal(bs, &opts); err != nil {
				log.Fatal(err)
			}
			opts.EnvironmentID = env
			ctx := context.Background()
			err = cli.DeployFunction(ctx, &opts)
			if err != nil {
				log.Fatal(err)
			}
			log.Info("successfully deployed function ", opts.ID)
		}
	},
}

func init() {
	functionCmd.AddCommand(functionDeployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// functionDeployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// functionDeployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}