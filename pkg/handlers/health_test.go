package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthHandlerReportsOK(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/healthz", bytes.NewReader([]byte("")))

	handler := MakeHealthHandler()
	handler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}
