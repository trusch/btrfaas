package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/trusch/btrfaas/fgateway/grpc"
	g "google.golang.org/grpc"
)

type DevNull struct{}

func (d *DevNull) Write(data []byte) (bs int, err error) {
	return len(data), nil
}

func runBtrfaasSync(ctx context.Context, fn string, data []byte, n int) error {
	cli, err := grpc.NewClient("127.0.0.1:2424", g.WithInsecure())
	if err != nil {
		log.Fatal("init:", err)
	}
	for i := 0; i < n; i++ {
		if err := cli.Run(ctx, []string{fn}, []map[string]string{map[string]string{}}, bytes.NewReader(data), &DevNull{}); err != nil {
			log.Print(err)
		}
	}
	return nil
}

func runOpenfaasSync(ctx context.Context, fn string, data []byte, n int) error {
	url := fmt.Sprintf("http://127.0.0.1:8080/function/%v", fn)
	for i := 0; i < n; i++ {
		req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
		req = req.WithContext(ctx)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Print(err)
			continue
		}
		resp.Body.Close()
	}
	return nil
}

type RunSyncFunc func(context.Context, string, []byte, int) error

func runAsync(ctx context.Context, fn string, data []byte, sync RunSyncFunc, p, n int) error {
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

func race(ctx context.Context, fn1 string, runSync1 RunSyncFunc, fn2 string, runSync2 RunSyncFunc, data []byte, p, n int) {
	var (
		done = make(chan error, 2)
	)
	go func() {
		err := runAsync(ctx, fn1, data, runSync1, p, n)
		done <- err
	}()
	go func() {
		err := runAsync(ctx, fn2, data, runSync2, p, n)
		done <- err
	}()
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			log.Print(err)
		}
	}
}

func main() {
	var (
		data = []byte("hello world")
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 32)
	pVec := []int{1, 5, 10}
	nVec := []int{10, 100, 500}
	for _, fn := range []string{"echo-shell", "echo-go", "echo-node", "echo-python"} {
		btrfaasFn := fn
		go func() {
			for _, p := range pVec {
				for _, n := range nVec {
					done <- runAsync(ctx, btrfaasFn, data, runBtrfaasSync, p, n)
				}
			}
		}()
	}

	go func() {
		for _, p := range pVec {
			for _, n := range nVec {
				done <- runAsync(ctx, "echo", data, runOpenfaasSync, p, n)
			}
		}
	}()

	for i := 0; i < 5*len(pVec)*len(nVec); i++ {
		if err := <-done; err != nil {
			log.Print(err)
		}
	}
}
