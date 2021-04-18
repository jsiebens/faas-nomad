package types

import (
	ftypes "github.com/openfaas/faas-provider/types"
)

type ProviderConfig struct {
	FaaS ftypes.FaaSConfig

	Vault VaultConfig
}

func LoadConfig() (*ProviderConfig, error) {
	env := ftypes.OsEnv{}

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
		},
	}

	return providerConfig, err
}
