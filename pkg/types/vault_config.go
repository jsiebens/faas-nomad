package types

type VaultConfig struct {
	Addr             string
	TLSSkipVerify    bool
	SecretPathPrefix string
}
