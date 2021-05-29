package types

import (
	"strings"

	ftypes "github.com/openfaas/faas-provider/types"
)

type SchedulingConfig struct {
	Region         string
	Datacenters    []string
	Namespace      string
	JobPrefix      string
	NetworkingMode string
}

type ProviderConfig struct {
	FaaS ftypes.FaaSConfig

	Vault      VaultConfig
	Consul     ConsulConfig
	Nomad      NomadConfig
	Scheduling SchedulingConfig
}

func DefaultConfig() (*ProviderConfig, error) {
	return doLoadConfig(emptyEnv{})
}

func LoadConfig() (*ProviderConfig, error) {
	return doLoadConfig(ftypes.OsEnv{})
}

func doLoadConfig(env ftypes.HasEnv) (*ProviderConfig, error) {
	faasConfig, err := ftypes.ReadConfig{}.Read(env)

	if err != nil {
		return nil, err
	}

	providerConfig := &ProviderConfig{
		FaaS: *faasConfig,

		Vault: VaultConfig{
			Addr:             ftypes.ParseString(env.Getenv("vault_addr"), "http://localhost:8200"),
			SecretPathPrefix: ftypes.ParseString(env.Getenv("vault_secret_path_prefix"), "openfaas"),
			TLSSkipVerify:    ftypes.ParseBoolValue(env.Getenv("vault_tls_skip_verify"), false),
			Policy:           ftypes.ParseString(env.Getenv("vault_policy"), "openfaas"),
		},

		Consul: ConsulConfig{
			Addr: ftypes.ParseString(env.Getenv("consul_addr"), "http://localhost:8500"),
		},
		
		Nomad: NomadConfig{
			Addr: ftypes.ParseString(env.Getenv("nomad_addr"), "http://localhost:4646"),
		},

		Scheduling: SchedulingConfig{
			Region:         ftypes.ParseString(env.Getenv("job_region"), "global"),
			Datacenters:    strings.Split(ftypes.ParseString(env.Getenv("job_datacenters"), "dc1"), ","),
			Namespace:      ftypes.ParseString(env.Getenv("job_namespace"), "default"),
			JobPrefix:      ftypes.ParseString(env.Getenv("job_name_prefix"), "faas-fn-"),
			NetworkingMode: ftypes.ParseString(env.Getenv("job_network_mode"), "host"),
		},
	}

	return providerConfig, err
}

type emptyEnv struct {
}

func (emptyEnv) Getenv(key string) string {
	return ""
}
