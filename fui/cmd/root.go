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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/grpc/credentials"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trusch/btrfaas/fgateway/grpc"
	g "google.golang.org/grpc"
)

var cli *grpc.Client

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "fui",
	Short: "BtrFaaS HTTP UI",
	Long:  `BtrFaaS HTTP UI`,
	Run: func(cmd *cobra.Command, args []string) {
		gateway, _ := cmd.Flags().GetString("gateway")
		// Create a certificate pool from the certificate authority
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile("/run/secrets/btrfaas-ca-cert.pem")
		if err != nil {
			log.Fatal(err)
		}

		// Append the certificates from the CA
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			log.Fatal(err)
		}

		creds := credentials.NewTLS(&tls.Config{
			ServerName: "fgateway",
			RootCAs:    certPool,
		})

		cli, err := grpc.NewClient(gateway, g.WithTransportCredentials(creds))
		if err != nil {
			log.Fatal(err)
		}

		http.HandleFunc("/api/invoke", func(w http.ResponseWriter, r *http.Request) {
			expr := r.Header.Get("X-Btrfaas-Chain")
			optionsStr := r.Header.Get("X-Btrfaas-Options")
			chain := strings.Split(expr, "|")
			options := [][]string{}
			if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			buf := &bytes.Buffer{}
			if err := cli.Run(context.Background(), chain, options, r.Body, buf); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			io.Copy(w, buf)
		})
		assetDir, _ := cmd.Flags().GetString("assets")
		http.Handle("/", http.FileServer(http.Dir(assetDir)))
		listen, _ := cmd.Flags().GetString("listen")
		if err := http.ListenAndServe(listen, nil); err != nil {
			log.Fatal(err)
		}
	},
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
	RootCmd.Flags().StringP("listen", "l", ":80", "http listen address")
	RootCmd.Flags().StringP("gateway", "g", "fgateway:2424", "gateway address")
	RootCmd.Flags().StringP("assets", "a", "assets", "asset directory")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
}
