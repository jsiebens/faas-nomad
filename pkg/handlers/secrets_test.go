package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/jsiebens/faas-nomad/pkg/services"
	ftypes "github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/assert"
)

func TestSecretsHandlerReportsAvailableSecrets(t *testing.T) {
	actualValues := []ftypes.Secret{
		{Name: "secret-a"},
		{Name: "secret-b"},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/system/secrets", bytes.NewReader([]byte("")))

	secrets := &services.MockSecrets{}
	secrets.On("List").Return(actualValues, nil)

	handler := MakeSecretHandler(secrets, hclog.Default())
	handler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, TypeApplicationJson, recorder.Header().Get(HeaderContentType))

	var arr []ftypes.Secret

	body, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		t.Fatal(err)
	}

	unmarshalErr := json.Unmarshal(body, &arr)

	assert.Nil(t, unmarshalErr, "Expected no error")
	assert.Equal(t, actualValues, arr)
}

func TestSecretsHandlerReportsErrorWhenListingSecrets(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/system/secrets", bytes.NewReader([]byte("")))

	secrets := &services.MockSecrets{}
	secrets.On("List").Return(nil, fmt.Errorf("error reading secrets"))

	handler := MakeSecretHandler(secrets, hclog.Default())
	handler(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestSecretsHandlerReportsCreatedWhenCreatingSecretSucceeds(t *testing.T) {
	value := "value-a"
	encoded := base64.StdEncoding.EncodeToString([]byte("value-a"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/system/secrets", bytes.NewReader(secretRequest("secret-a", value)))

	secrets := &services.MockSecrets{}
	secrets.On("Set", "secret-a", encoded).Return(nil)

	handler := MakeSecretHandler(secrets, hclog.Default())
	handler(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)
}

func TestSecretsHandlerReportsErrorWhenCreatingSecretFails(t *testing.T) {
	value := "value-a"
	encoded := base64.StdEncoding.EncodeToString([]byte("value-a"))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/system/secrets", bytes.NewReader(secretRequest("secret-a", value)))

	secrets := &services.MockSecrets{}
	secrets.On("Set", "secret-a", encoded).Return(fmt.Errorf("error reading secrets"))

	handler := MakeSecretHandler(secrets, hclog.Default())
	handler(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestSecretsHandlerReportsOKWhenUpdatingSecretSucceeds(t *testing.T) {
	value := "value-a"
	encoded := base64.StdEncoding.EncodeToString([]byte("value-a"))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("PUT", "/system/secrets", bytes.NewReader(secretRequest("secret-a", value)))

	secrets := &services.MockSecrets{}
	secrets.On("Set", "secret-a", encoded).Return(nil)

	handler := MakeSecretHandler(secrets, hclog.Default())
	handler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestSecretsHandlerReportsOKWhenCreatingBindarySecret(t *testing.T) {
	value := "value-a"
	encoded := base64.StdEncoding.EncodeToString([]byte("value-a"))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/system/secrets", bytes.NewReader(secretRequestWithRawValue("secret-a", value)))

	secrets := &services.MockSecrets{}
	secrets.On("Set", "secret-a", encoded).Return(nil)

	handler := MakeSecretHandler(secrets, hclog.Default())
	handler(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)
}

func TestSecretsHandlerReportsErrorWhenUpdatingSecretFails(t *testing.T) {
	value := "value-a"
	encoded := base64.StdEncoding.EncodeToString([]byte("value-a"))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("PUT", "/system/secrets", bytes.NewReader(secretRequest("secret-a", value)))

	secrets := &services.MockSecrets{}
	secrets.On("Set", "secret-a", encoded).Return(fmt.Errorf("error reading secrets"))

	handler := MakeSecretHandler(secrets, hclog.Default())
	handler(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestSecretsHandlerReportsOKWhenDeletingSecretSucceeds(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("DELETE", "/system/secrets", bytes.NewReader(deleteRequest("secret-a")))

	secrets := &services.MockSecrets{}
	secrets.On("Delete", "secret-a").Return(nil)

	handler := MakeSecretHandler(secrets, hclog.Default())
	handler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestSecretsHandlerReportsErrorWhenDeletingSecretFails(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("DELETE", "/system/secrets", bytes.NewReader(deleteRequest("secret-a")))

	secrets := &services.MockSecrets{}
	secrets.On("Delete", "secret-a").Return(fmt.Errorf("error reading secrets"))

	handler := MakeSecretHandler(secrets, hclog.Default())
	handler(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func deleteRequest(name string) []byte {
	req := ftypes.Secret{Name: name}
	data, _ := json.Marshal(req)
	return data
}

func secretRequest(name, value string) []byte {
	req := ftypes.Secret{Name: name, Value: value}
	data, _ := json.Marshal(req)
	return data
}

func secretRequestWithRawValue(name, value string) []byte {
	req := ftypes.Secret{Name: name, RawValue: []byte(value)}
	data, _ := json.Marshal(req)
	return data
}
