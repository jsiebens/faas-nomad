package types

import (
	ftypes "github.com/openfaas/faas-provider/types"
)

func LoadConfig() (*ftypes.FaaSConfig, error) {
	return ftypes.ReadConfig{}.Read(ftypes.OsEnv{})
}
