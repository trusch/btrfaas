package swarmintegrationtests_test

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	retries    = 10
	retryDelay = 1 * time.Second
)

func runScript(script string) (res string) {
	for i := 0; i < retries; i++ {
		cmd := exec.Command("bash", "-c", script)
		buf := &bytes.Buffer{}
		cmd.Stdout = buf
		cmd.Stderr = buf
		if err := cmd.Run(); err != nil {
			if i == retries-1 {
				Fail(fmt.Sprintf("%v: %v", err.Error(), buf.String()))
			} else {
				time.Sleep(retryDelay)
				continue
			}
		}
		res = buf.String()
		break
	}
	return
}

var _ = Describe("Smoke Test", func() {
	It("should be possible to init btrfaas with --platform=swarm", func() {
		runScript("btrfaasctl --platform=swarm init")
	})
	It("should be possible to deploy and call the echo tests", func() {
		runScript(`btrfaasctl --platform swarm function deploy ../../examples/echo-shell.yaml`)
		runScript(`btrfaasctl --platform swarm function deploy ../../examples/echo-go/function.yaml`)
		runScript(`btrfaasctl --platform swarm function deploy ../../examples/echo-python/function.yaml`)
		runScript(`btrfaasctl --platform swarm function deploy ../../examples/echo-node/function.yaml`)
		res := runScript(`echo -n foobar | btrfaasctl function invoke "echo-go | echo-node | echo-python | echo-shell"`)
		Expect(res).To(Equal("foobar"))
	})
	It("should be possible to deploy a function with an environement variable", func() {
		runScript("btrfaasctl --platform swarm function deploy ../../examples/with-env.yaml")
		res := runScript(`btrfaasctl function invoke with-env`)
		Expect(res).To(Equal("VALUE"))
	})
	It("should be possible to deploy and access a secret", func() {
		runScript("btrfaasctl --platform swarm secret deploy example-secret secret-value")
		runScript("btrfaasctl --platform swarm function deploy ../../examples/with-secret.yaml")
		res := runScript(`btrfaasctl function invoke with-secret`)
		Expect(res).To(Equal("secret-value"))
	})
	It("should be possible to deploy and access a function with options", func() {
		runScript("btrfaasctl --platform swarm function deploy ../../examples/sed.yaml")
		res := runScript(`echo -n foo | btrfaasctl function invoke "sed s/foo/bar/g"`)
		Expect(res).To(Equal("bar"))
	})
	It("should be possible to teardown btrfaas", func() {
		runScript("btrfaasctl --platform=swarm teardown")
	})
})
