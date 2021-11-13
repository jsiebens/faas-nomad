package types

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"os"
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
	Token            string
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
	Proxy      ProxyConfig
	Log        LogConfig
}

type ProxyConfig struct {
	Strategy string
}

func DefaultConfig() (*ProviderConfig, error) {
	return doLoadConfig(emptyEnv{})
}

func LoadConfig(filename string) (*ProviderConfig, error) {
	properties, err := readPropertiesFile(filename)
	if err != nil {
		return nil, err
	} else {
		return doLoadConfig(properties)
	}
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
			Token:            ftypes.ParseString(env.Getenv("vault_token"), ""),
			SecretPathPrefix: ftypes.ParseString(env.Getenv("vault_secret_path_prefix"), "openfaas-fn"),
			CACert:           ftypes.ParseString(env.Getenv("vault_tls_ca"), ""),
			ClientCert:       ftypes.ParseString(env.Getenv("vault_tls_cert"), ""),
			ClientKey:        ftypes.ParseString(env.Getenv("vault_tls_key"), ""),
			TLSSkipVerify:    ftypes.ParseBoolValue(env.Getenv("vault_tls_skip_verify"), false),
			Policy:           ftypes.ParseString(env.Getenv("vault_policy"), "openfaas-fn"),
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
			HttpCheck:      ftypes.ParseBoolValue(env.Getenv("job_http_check"), true),
		},

		Proxy: ProxyConfig{
			Strategy: ftypes.ParseString(env.Getenv("proxy_strategy"), "roundrobin"),
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

type viperEnv struct {
}

func (p viperEnv) Getenv(key string) string {
	s := viper.GetString(key)
	if len(s) != 0 {
		return s
	}
	return os.Getenv(key)
}

func readPropertiesFile(filename string) (ftypes.HasEnv, error) {
	if len(filename) == 0 {
		return ftypes.OsEnv{}, nil
	}

	res, err := homedir.Expand(filename)
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(res)
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return viperEnv{}, nil
}

func expandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}
