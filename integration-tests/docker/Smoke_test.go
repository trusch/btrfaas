package dockerintegrationtests_test

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
				continue
			}
		}
		res = buf.String()
		break
	}
	return
}

var _ = Describe("Smoke Test", func() {
	It("should be possible to init btrfaas with --platform=docker", func() {
		runScript("btrfaasctl --platform=docker init")
		runScript("btrfaasctl --platform=docker service deploy ../../core-services/fgateway.yaml")
	})
	It("should be possible to deploy and call the echo tests", func() {
		runScript("btrfaasctl --platform docker service deploy ../../examples/services/echo/*")
		res := runScript(`echo -n foobar | btrfaasctl function invoke "echo-go | echo-node | echo-python | echo-shell"`)
		Expect(res).To(Equal("foobar"))
	})
	It("should be possible to deploy a service with an environement variable", func() {
		runScript("btrfaasctl --platform docker service deploy ../../examples/services/with-env.yaml")
		res := runScript(`btrfaasctl function invoke with-env`)
		Expect(res).To(Equal("VALUE"))
	})
	It("should be possible to deploy and access a secret", func() {
		runScript("btrfaasctl --platform docker secret deploy example-secret secret-value")
		runScript("btrfaasctl --platform docker service deploy ../../examples/services/with-secret.yaml")
		res := runScript(`btrfaasctl function invoke with-secret`)
		Expect(res).To(Equal("secret-value"))
	})
	It("should be possible to deploy and access a function with options", func() {
		runScript("btrfaasctl --platform docker service deploy ../../examples/services/sed.yaml")
		res := runScript(`echo -n foo | btrfaasctl function invoke "sed e=s/foo/bar/g"`)
		Expect(res).To(Equal("bar"))
	})
	It("should be possible to teardown btrfaas", func() {
		runScript("btrfaasctl --platform=docker teardown")
	})
})
