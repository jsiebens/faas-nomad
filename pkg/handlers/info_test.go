package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	ftypes "github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/assert"
)

func TestInfoHandlerReportsProviderInfo(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/system/info", bytes.NewReader([]byte("")))

	handler := MakeInfoHandler()
	handler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, TypeApplicationJson, recorder.Header().Get(HeaderContentType))

	info := &ftypes.ProviderInfo{}

	body, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		t.Fatal(err)
	}

	unmarshalErr := json.Unmarshal(body, &info)

	assert.Nil(t, unmarshalErr, "Expected no error")
	assert.Equal(t, info.Orchestration, OrchestrationIdentifier)
	assert.Equal(t, info.Name, ProviderName)
	assert.Equal(t, info.Version.Release, "0.0.0")
	assert.Equal(t, info.Version.SHA, "")
	assert.Equal(t, info.Version.CommitMessage, "")
}
