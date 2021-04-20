package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupDeleteHandler(body []byte) (*services.MockJobs, http.HandlerFunc, *http.Request, *httptest.ResponseRecorder) {
	jobs := &services.MockJobs{}

	response := httptest.NewRecorder()
	request := httptest.NewRequest("DELETE", "/system/functions", bytes.NewReader(body))

	config := &types.ProviderConfig{Scheduling: types.SchedulingConfig{
		JobPrefix: "faas-fn-",
	}}

	handler := MakeDeleteHandler(config, jobs)

	return jobs, handler, request, response
}

func TestDeleteHandlerReportsErrorWhenRequestIsInvalid(t *testing.T) {
	_, deleteHandler, request, recorder := setupDeleteHandler([]byte(""))

	deleteHandler(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDeleteHandlerReportsErrorWhenDeregisterFails(t *testing.T) {
	req := ftypes.DeleteFunctionRequest{}
	req.FunctionName = "func123"
	data, _ := json.Marshal(req)

	jobs, deleteHandler, request, recorder := setupDeleteHandler(data)
	jobs.On("Deregister", "faas-fn-func123", mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("failure"))

	deleteHandler(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	jobs.AssertCalled(t, "Deregister", "faas-fn-func123", mock.Anything, mock.Anything)
}

func TestDeleteHandlerDeregisterJob(t *testing.T) {
	req := ftypes.DeleteFunctionRequest{}
	req.FunctionName = "func123"
	data, _ := json.Marshal(req)

	jobs, deleteHandler, request, recorder := setupDeleteHandler(data)
	jobs.On("Deregister", "faas-fn-func123", mock.Anything, mock.Anything).Return(nil, nil, nil)

	deleteHandler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	jobs.AssertCalled(t, "Deregister", "faas-fn-func123", mock.Anything, mock.Anything)
}
