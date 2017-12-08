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
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/trusch/btrfaas/template"
)

// functionInitCmd represents the functionInit command
var functionInitCmd = &cobra.Command{
	Use:   "init <target directory>",
	Short: "init a new function",
	Long:  `init a new function`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			os.Exit(1)
		}
		tmplID, _ := cmd.Flags().GetString("template")
		ref, _ := cmd.Flags().GetString("ref")
		if err := template.Load(tmplID, ref, args[0]); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	functionCmd.AddCommand(functionInitCmd)
	functionInitCmd.Flags().String("template", "go", "function template to use")
	functionInitCmd.Flags().String("ref", "refs/tags/v0.3.3", "repo reference to fetch")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// functionInitCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// functionInitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
