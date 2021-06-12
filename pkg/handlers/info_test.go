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

	handler := MakeInfoHandler("1.2.3", "fa097935ca9d551d91fa78ed81ec05c6a1df249f")
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
	assert.Equal(t, OrchestrationIdentifier, info.Orchestration)
	assert.Equal(t, ProviderName, info.Name)
	assert.Equal(t, "1.2.3", info.Version.Release)
	assert.Equal(t, "fa097935ca9d551d91fa78ed81ec05c6a1df249f", info.Version.SHA)
	assert.Equal(t, "", info.Version.CommitMessage)
}
