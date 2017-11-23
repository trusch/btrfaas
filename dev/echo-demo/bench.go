package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/trusch/btrfaas/fgateway/grpc"
	g "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type runSyncFunc func(context.Context, string, []byte, int) error

var certPool *x509.CertPool

func getTransportCredentials(target string) (g.DialOption, error) {
	if certPool == nil {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		ca, err := ioutil.ReadFile(filepath.Join(home, ".btrfaas", "btrfaas_default", "ca-cert.pem"))
		if err != nil {
			return nil, fmt.Errorf("could not read ca certificate: %s", err)
		}
		certPool = x509.NewCertPool()
		// Append the certificates from the CA
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, errors.New("failed to append ca certs")
		}
	}

	creds := credentials.NewTLS(&tls.Config{
		ServerName: target,
		RootCAs:    certPool,
	})

	return g.WithTransportCredentials(creds), nil
}

func runBtrfaasSync(ctx context.Context, fn string, data []byte, n int) error {
	creds, err := getTransportCredentials("fgateway")
	if err != nil {
		return err
	}
	cli, err := grpc.NewClient("127.0.0.1:2424", creds)
	if err != nil {
		log.Fatal("init:", err)
	}
	for i := 0; i < n; i++ {
		start := time.Now()
		if err := cli.Run(ctx, []string{fn}, [][]string{{}}, bytes.NewReader(data), ioutil.Discard); err != nil {
			fmt.Print(err)
		}
		end := time.Now()
		stats.WithLabelValues(fn).Observe(end.Sub(start).Seconds())
	}
	return nil
}

func runAsync(ctx context.Context, fn string, data []byte, sync runSyncFunc, p, n int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, p)
	for i := 0; i < p; i++ {
		go func() {
			done <- sync(ctx, fn, data, n)
		}()
	}
	for i := 0; i < p; i++ {
		err := <-done
		if err != nil {
			return err
		}
	}
	return nil
}

var c = flag.Int("c", 10, "concurrency level")
var n = flag.Int("n", 1000, "how many requests")
var size = flag.Int("size", 32, "payload size")
var fn = flag.String("function", "", "function to benchmark")

var stats = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name:       "rpc_duration_seconds",
		Help:       "The function call durations.",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001},
	},
	[]string{"function"},
)

func main() {
	flag.Parse()
	fns := []string{*fn}
	if fns[0] == "" {
		fns = []string{
			"echo-go",
			"echo-node",
			"echo-python",
			"echo-shell",
			"http://echo-openfaas",
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"function", "req/s", "q99", "q95", "q90", "q50"})
	for _, fn := range fns {
		stats.Reset()
		reg := prometheus.NewRegistry()
		reg.MustRegister(stats)
		data := make([]byte, *size)
		if _, err := rand.Read(data); err != nil {
			log.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		start := time.Now()
		if err := runAsync(ctx, fn, data, runBtrfaasSync, *c, *n); err != nil {
			log.Fatal(err)
		}
		end := time.Now()
		metricFamilies, _ := reg.Gather()
		summary := metricFamilies[0].GetMetric()[0].Summary
		q50 := summary.Quantile[0].GetValue() * 1000.
		q90 := summary.Quantile[1].GetValue() * 1000.
		q95 := summary.Quantile[2].GetValue() * 1000.
		q99 := summary.Quantile[3].GetValue() * 1000.
		reqPerSecond := float64(*c**n) / end.Sub(start).Seconds()
		fmt.Printf("finished with %v with %.2f req/s\n", fn, reqPerSecond)
		table.Append([]string{
			fn,
			fmt.Sprintf("%.2f", reqPerSecond),
			fmt.Sprintf("%.2f", q99),
			fmt.Sprintf("%.2f", q95),
			fmt.Sprintf("%.2f", q90),
			fmt.Sprintf("%.2f", q50),
		})
	}
	table.Render()
}
