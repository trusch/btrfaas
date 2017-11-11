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
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/trusch/btrfaas/fgateway/grpc"
	g "google.golang.org/grpc"
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
		expr := strings.Join(args, " ")
		gw, _ := cmd.Flags().GetString("gateway")
		cli, err := grpc.NewClient(gw, g.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		chain, opts, err := createCallRequest(expr)
		if err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()
		if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout != 0 {
			c, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			ctx = c
		}

		if err = cli.Run(ctx, chain, opts, os.Stdin, os.Stdout); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	functionCmd.AddCommand(invokeCmd)
	invokeCmd.Flags().String("gateway", "127.0.0.1:2424", "fgatway address")
	invokeCmd.Flags().Duration("timeout", 0*time.Second, "specify a timeout for the call")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// invokeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// invokeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func createCallRequest(expr string) (chain []string, opts []map[string]string, err error) {
	fnExpressions := strings.Split(expr, "|")
	chain = make([]string, len(fnExpressions))
	opts = make([]map[string]string, len(fnExpressions))
	for idx, fnExpression := range fnExpressions {
		parts := strings.Split(strings.Trim(fnExpression, " "), " ")
		if len(parts) < 1 {
			return nil, nil, errors.New("malformed expression")
		}
		fn := strings.Trim(parts[0], " ")
		chain[idx] = fn
		fnOpts := make(map[string]string)
		for i := 1; i < len(parts); i++ {
			pairSlice := strings.Split(parts[i], "=")
			if len(pairSlice) < 2 {
				return nil, nil, errors.New("malformed expression")
			}
			fnOpts[pairSlice[0]] = pairSlice[1]
		}
		opts[idx] = fnOpts
	}
	return chain, opts, nil
}
