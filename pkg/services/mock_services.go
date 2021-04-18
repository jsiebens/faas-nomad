package services

import (
	ftypes "github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/mock"
)

type MockSecrets struct {
	mock.Mock
}

func (ms *MockSecrets) List() ([]ftypes.Secret, error) {
	args := ms.Called()

	var resp []ftypes.Secret
	if r := args.Get(0); r != nil {
		resp = r.([]ftypes.Secret)
	}

	return resp, args.Error(1)
}

func (ms *MockSecrets) Set(key, value string) error {
	args := ms.Called(key, value)
	return args.Error(0)
}

func (ms *MockSecrets) Delete(key string) error {
	args := ms.Called(key)
	return args.Error(0)
}
