package main

import (
	"context"
	"log"

	"github.com/trusch/btrfaas/deployment"
)

func main() {
	platform, err := deployment.NewDockerPlatform()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = platform.PrepareEnvironment(ctx, &deployment.PrepareEnvironmentOptions{
		EnvironmentID: "test-env",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Print("prepared environment")

	err = platform.DeploySecret(ctx, &deployment.DeploySecretOptions{
		EnvironmentID: "test-env",
		SecretID:      "my-secret",
		Value:         "this is secret",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = platform.DeployService(ctx, &deployment.DeployServiceOptions{
		EnvironmentID: "test-env",
		ServiceID:     "my-debian",
		Image:         "debian",
		Cmd:           []string{"tail", "-f", "/dev/null"},
		Env:           map[string]string{"FOO": "BAR"},
		Secrets:       []string{"my-secret"},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = platform.UndeployService(ctx, &deployment.UndeployServiceOptions{
		EnvironmentID: "test-env",
		ServiceID:     "my-debian",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Print("undeployed service")

	err = platform.UndeploySecret(ctx, &deployment.UndeploySecretOptions{
		EnvironmentID: "test-env",
		SecretID:      "my-secret",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Print("undeployed secret")

}
