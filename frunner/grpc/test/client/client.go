package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/trusch/frunner/grpc"
	g "google.golang.org/grpc"
)

func main() {
	cli, err := grpc.NewClient("localhost:2424", g.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	input := bytes.NewBufferString("foobar")
	output := &bytes.Buffer{}
	err = cli.Run(context.Background(), input, output)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(output.String())
}
