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
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/trusch/btrfaas/faas"
)

// invokeCmd represents the invoke command
var invokeCmd = &cobra.Command{
	Use:   "invoke <function expression>",
	Short: "invoke a function",
	Long:  `invoke a function`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			os.Exit(1)
		}

		cli := getFaaS(cmd)

		ctx := context.Background()
		if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout != 0 {
			c, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			ctx = c
		}
		expr := strings.Join(args, " ")

		if err := cli.Invoke(ctx, &faas.InvokeOptions{
			GatewayAddress:     getGateway(cmd),
			FunctionExpression: expr,
			Input:              os.Stdin,
			Output:             os.Stdout,
		}); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	functionCmd.AddCommand(invokeCmd)
	invokeCmd.Flags().Duration("timeout", 0*time.Second, "specify a timeout for the call")
	invokeCmd.Flags().String("gateway", "", "gateway address")
}

func getGateway(cmd *cobra.Command) string {
	flags := cmd.Flags()
	gw, _ := flags.GetString("gateway")
	if gw == "" {
		faasProvider, _ := flags.GetString("faas-provider")
		switch faasProvider {
		case "btrfaas":
			gw = "127.0.0.1:2424"
		case "openfaas":
			gw = "127.0.0.1:8080"
		default:
			log.Fatal("unknown faas provider")
		}
	}
	return gw
}
