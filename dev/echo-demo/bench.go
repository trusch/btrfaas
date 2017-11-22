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
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/trusch/btrfaas/fgateway/grpc"
	g "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type devNull struct{}

func (d *devNull) Write(data []byte) (bs int, err error) {
	return len(data), nil
}

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
		if err := cli.Run(ctx, []string{fn}, [][]string{{}}, bytes.NewReader(data), &devNull{}); err != nil {
			log.Print(err)
		}
	}
	return nil
}

func runAsync(ctx context.Context, fn string, data []byte, sync runSyncFunc, p, n int) error {
	start := time.Now()
	defer func() {
		end := time.Now()
		log.Printf("%v(%v/%v):\t%v req/s", fn, p, n, 1./(end.Sub(start).Seconds()/float64(p*n)))
	}()
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
var fn = flag.String("function", "echo-go", "function to benchmark")

func main() {
	flag.Parse()
	data := make([]byte, *size)
	if _, err := rand.Read(data); err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runAsync(ctx, *fn, data, runBtrfaasSync, *c, *n); err != nil {
		log.Fatal(err)
	}
}
