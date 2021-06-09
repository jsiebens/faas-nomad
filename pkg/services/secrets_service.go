package services

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/vault/api"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
)

type Secrets interface {
	List() ([]ftypes.Secret, error)
	Set(key, value string) error
	Delete(key string) error
}

func NewVaultSecrets(config types.VaultConfig) (Secrets, error) {

	clientConfig := api.DefaultConfig()
	clientConfig.Address = config.Addr

	tlsConfig := api.TLSConfig{
		CACert:     config.CACert,
		ClientCert: config.ClientCert,
		ClientKey:  config.ClientKey,
		Insecure:   config.TLSSkipVerify,
	}

	if err := clientConfig.ConfigureTLS(&tlsConfig); err != nil {
		return nil, err
	}

	vaultClient, err := api.NewClient(clientConfig)

	if err != nil {
		return nil, err
	}

	vs := &VaultSecrets{
		client: vaultClient,
		prefix: config.SecretPathPrefix,
	}

	if err := vs.login(); err != nil {
		return nil, err
	}

	return vs, nil
}

type VaultSecrets struct {
	client *api.Client
	prefix string
}

func (vs *VaultSecrets) List() ([]ftypes.Secret, error) {
	secretList, err := vs.client.Logical().List(fmt.Sprintf("%s", vs.prefix))

	if err != nil || secretList == nil {
		return []ftypes.Secret{}, err
	}

	var secrets []ftypes.Secret
	for _, k := range secretList.Data["keys"].([]interface{}) {
		secrets = append(secrets, ftypes.Secret{Name: k.(string)})
	}

	return secrets, nil
}

func (vs *VaultSecrets) Set(key, value string) error {
	_, err := vs.client.Logical().Write(fmt.Sprintf("%s/%s", vs.prefix, key), map[string]interface{}{"value": value})
	return err
}

func (vs *VaultSecrets) Delete(key string) error {
	_, err := vs.client.Logical().Delete(fmt.Sprintf("%s/%s", vs.prefix, key))
	return err
}

// Gets and sets the initial access token from Vault
func (vs *VaultSecrets) login() error {
	token := vs.readToken()
	vs.client.SetToken(token)
	go vs.renew()
	return nil
}

func (vs *VaultSecrets) renew() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGUSR1)

	for {
		<-sigs
		token := vs.readToken()
		vs.client.SetToken(token)
	}
}

func (vs *VaultSecrets) readToken() string {
	file, err := ioutil.ReadFile("secrets/vault_token")
	if err != nil {
		return os.Getenv("VAULT_TOKEN")
	}
	return string(file)
}
