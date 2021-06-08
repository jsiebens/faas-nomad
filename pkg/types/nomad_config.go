package types

type NomadConfig struct {
	Addr          string
	ACLToken      string
	CACert        string
	ClientCert    string
	ClientKey     string
	TLSSkipVerify bool
}
