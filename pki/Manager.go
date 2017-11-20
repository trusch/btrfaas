package pki

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/trusch/btrfaas/deployment"
	"github.com/trusch/pki"
)

var (
	curve = "P521"
)

func init() {

}

// Manager manages the public key infrastructure of a deployment
type Manager struct {
	platform deployment.SecretPlatform
	env      string
	ca       *pki.CA
}

// NewManager returns a new manager instance and creates keys if necessary
func NewManager(ctx context.Context, platform deployment.SecretPlatform, env string) (*Manager, error) {
	var ca *pki.CA
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	caCrtFile := filepath.Join(home, fmt.Sprintf(".btrfaas/%v/ca-cert.pem", env))
	caKeyFile := filepath.Join(home, fmt.Sprintf(".btrfaas/%v/ca-key.pem", env))
	if _, err = os.Stat(caCrtFile); err != nil {
		ca, err = pki.NewSelfSignedCA("btrfaas", curve, 0)
		if err != nil {
			return nil, err
		}
		cert, getCertErr := ca.GetCertAsPEM()
		if getCertErr != nil {
			return nil, err
		}
		key, getKeyErr := ca.GetKeyAsPEM()
		if getKeyErr != nil {
			return nil, err
		}
		if err = os.MkdirAll(fmt.Sprintf("%v/.btrfaas/%v", home, env), 0755); err != nil {
			return nil, err
		}
		if err = ioutil.WriteFile(caCrtFile, cert, 0600); err != nil {
			return nil, err
		}
		if err = ioutil.WriteFile(caKeyFile, key, 0600); err != nil {
			return nil, err
		}
	} else {
		certBs, e := ioutil.ReadFile(caCrtFile)
		if e != nil {
			return nil, e
		}
		keyBs, e := ioutil.ReadFile(caKeyFile)
		if e != nil {
			return nil, e
		}
		ca, err = pki.NewCA(certBs, keyBs, nil)
		if err != nil {
			return nil, err
		}
	}
	if _, err = platform.GetSecret(ctx, &deployment.GetSecretOptions{
		EnvironmentID: env,
		ID:            "btrfaas-ca-cert",
	}); err != nil {
		cert, err := ca.GetCertAsPEM()
		if err != nil {
			return nil, err
		}
		if err = platform.DeploySecret(ctx, &deployment.DeploySecretOptions{
			EnvironmentID: env,
			ID:            "btrfaas-ca-cert",
			Value:         cert,
		}); err != nil {
			return nil, err
		}
	}
	return &Manager{platform, env, ca}, nil
}

// IssueClient issues a new client certificate and saves it as secret
func (manager *Manager) IssueClient(ctx context.Context, id string) error {
	cert, key, err := manager.ca.IssueClient(id, curve, 0)
	if err != nil {
		return err
	}
	return saveCertAndKeyAsSecret(ctx, manager.platform, manager.env, id, cert, key)
}

// IssueServer issues a new server certificate and saves it as secret
func (manager *Manager) IssueServer(ctx context.Context, id string) error {
	cert, key, err := manager.ca.IssueServer(id, curve, 0)
	if err != nil {
		return err
	}
	return saveCertAndKeyAsSecret(ctx, manager.platform, manager.env, id, cert, key)
}

func saveCertAndKeyAsSecret(ctx context.Context, platform deployment.SecretPlatform, env, id string, cert, key []byte) error {
	if err := platform.DeploySecret(ctx, &deployment.DeploySecretOptions{
		EnvironmentID: env,
		ID:            id + "-key",
		Value:         key,
	}); err != nil {
		return err
	}
	return platform.DeploySecret(ctx, &deployment.DeploySecretOptions{
		EnvironmentID: env,
		ID:            id + "-cert",
		Value:         cert,
	})
}
