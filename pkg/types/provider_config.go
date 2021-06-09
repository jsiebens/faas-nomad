package types

import (
	"strings"

	ftypes "github.com/openfaas/faas-provider/types"
)

type ConsulConfig struct {
	Addr          string
	ACLToken      string
	CACert        string
	ClientCert    string
	ClientKey     string
	TLSSkipVerify bool
}

type NomadConfig struct {
	Addr          string
	ACLToken      string
	CACert        string
	ClientCert    string
	ClientKey     string
	TLSSkipVerify bool
}

type VaultConfig struct {
	Addr             string
	CACert           string
	ClientCert       string
	ClientKey        string
	TLSSkipVerify    bool
	SecretPathPrefix string
	Policy           string
}

type SchedulingConfig struct {
	Region         string
	Datacenters    []string
	Namespace      string
	JobPrefix      string
	NetworkingMode string
	Purge          bool
	HttpCheck      bool
}

type LogConfig struct {
	Level  string
	Format string
	File   string
}

type ProviderConfig struct {
	FaaS ftypes.FaaSConfig

	Vault      VaultConfig
	Consul     ConsulConfig
	Nomad      NomadConfig
	Scheduling SchedulingConfig
	Log        LogConfig
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
			CACert:           ftypes.ParseString(env.Getenv("vault_tls_ca"), ""),
			ClientCert:       ftypes.ParseString(env.Getenv("vault_tls_cert"), ""),
			ClientKey:        ftypes.ParseString(env.Getenv("vault_tls_key"), ""),
			TLSSkipVerify:    ftypes.ParseBoolValue(env.Getenv("vault_tls_skip_verify"), false),
			Policy:           ftypes.ParseString(env.Getenv("vault_policy"), "openfaas"),
		},

		Consul: ConsulConfig{
			Addr:          ftypes.ParseString(env.Getenv("consul_addr"), "http://localhost:8500"),
			ACLToken:      ftypes.ParseString(env.Getenv("consul_token"), ""),
			CACert:        ftypes.ParseString(env.Getenv("consul_tls_ca"), ""),
			ClientCert:    ftypes.ParseString(env.Getenv("consul_tls_cert"), ""),
			ClientKey:     ftypes.ParseString(env.Getenv("consul_tls_key"), ""),
			TLSSkipVerify: ftypes.ParseBoolValue(env.Getenv("consul_tls_skip_verify"), false),
		},

		Nomad: NomadConfig{
			Addr:          ftypes.ParseString(env.Getenv("nomad_addr"), "http://localhost:4646"),
			ACLToken:      ftypes.ParseString(env.Getenv("nomad_token"), ""),
			CACert:        ftypes.ParseString(env.Getenv("nomad_tls_ca"), ""),
			ClientCert:    ftypes.ParseString(env.Getenv("nomad_tls_cert"), ""),
			ClientKey:     ftypes.ParseString(env.Getenv("nomad_tls_key"), ""),
			TLSSkipVerify: ftypes.ParseBoolValue(env.Getenv("nomad_tls_skip_verify"), false),
		},

		Scheduling: SchedulingConfig{
			Region:         ftypes.ParseString(env.Getenv("job_region"), "global"),
			Datacenters:    strings.Split(ftypes.ParseString(env.Getenv("job_datacenters"), "dc1"), ","),
			Namespace:      ftypes.ParseString(env.Getenv("job_namespace"), "default"),
			JobPrefix:      ftypes.ParseString(env.Getenv("job_name_prefix"), "faas-fn-"),
			NetworkingMode: ftypes.ParseString(env.Getenv("job_network_mode"), "host"),
			Purge:          ftypes.ParseBoolValue(env.Getenv("job_purge"), false),
			HttpCheck:      ftypes.ParseBoolValue(env.Getenv("job_http_check"), true),
		},

		Log: LogConfig{
			Level:  ftypes.ParseString(env.Getenv("log_level"), "info"),
			Format: ftypes.ParseString(env.Getenv("log_format"), "text"),
			File:   ftypes.ParseString(env.Getenv("log_file"), ""),
		},
	}

	return providerConfig, err
}

type emptyEnv struct {
}

func (emptyEnv) Getenv(key string) string {
	return ""
}
