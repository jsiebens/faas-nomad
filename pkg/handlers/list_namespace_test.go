package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListNamespaceHandlerReportsAvailableNamespaces(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/system/namespaces", bytes.NewReader([]byte("")))

	handler := MakeListNamespaceHandler()
	handler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, TypeApplicationJson, recorder.Header().Get(HeaderContentType))

	var arr []string

	body, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		t.Fatal(err)
	}

	unmarshalErr := json.Unmarshal(body, &arr)

	assert.Nil(t, unmarshalErr, "Expected no error")
	assert.Empty(t, arr)
}
