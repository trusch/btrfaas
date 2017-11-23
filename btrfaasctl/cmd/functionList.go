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
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/trusch/btrfaas/deployment"
	"github.com/trusch/btrfaas/faas"
)

// functionListCmd represents the functionList command
var functionListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list functions",
	Long:    `list functions`,
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Flags().GetString("env")
		cli := getFaaS(cmd)
		ctx := context.Background()
		functions, err := cli.ListFunctions(ctx, &faas.ListFunctionsOptions{
			ListServicesOptions: deployment.ListServicesOptions{
				EnvironmentID: env,
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		printFunctionTable(functions)
	},
}

func init() {
	functionCmd.AddCommand(functionListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// functionListCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// functionListCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func printFunctionTable(functions []*faas.FunctionInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"id", "image", "created", "scale"})
	for _, function := range functions {
		table.Append([]string{function.ID, function.Image, fmt.Sprint(function.CreatedAt), fmt.Sprint(function.Scale)})
	}
	table.Render()
}
