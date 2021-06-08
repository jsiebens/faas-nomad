package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupDeployHandler(body []byte) (*services.MockJobs, http.HandlerFunc, *http.Request, *httptest.ResponseRecorder) {
	jobs := &services.MockJobs{}

	response := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/system/functions", bytes.NewReader(body))

	config, _ := types.DefaultConfig()
	handler := MakeDeployHandler(config, jobs, hclog.Default())

	return jobs, handler, request, response
}

func TestDeployHandlerReportsErrorWhenRequestIsInvalid(t *testing.T) {
	_, deployHandler, request, recorder := setupDeployHandler([]byte(""))

	deployHandler(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDeployHandlerReportsErrorWhenJobRegistrationFails(t *testing.T) {
	req := ftypes.FunctionDeployment{}
	req.Service = "Func123"
	body, _ := json.Marshal(req)

	jobs, deployHandler, request, recorder := setupDeployHandler(body)

	jobs.On("Register", mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("failure"))

	deployHandler(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestDeployHandlerReportsOKWhenJobIsRegistered(t *testing.T) {
	req := ftypes.FunctionDeployment{}
	req.Service = "Func123"
	body, _ := json.Marshal(req)

	jobs, deployHandler, request, recorder := setupDeployHandler(body)

	jobs.On("Register", mock.Anything, mock.Anything).Return(nil, nil, nil)

	deployHandler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	args := jobs.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	count := job.TaskGroups[0].Count

	assert.Equal(t, 1, *count)
}

func TestDeployHandlerWithInitialScaleCount(t *testing.T) {
	labels := map[string]string{
		"com.openfaas.scale.min": "3",
	}

	req := ftypes.FunctionDeployment{}
	req.Service = "Func123"
	req.Labels = &labels
	body, _ := json.Marshal(req)

	jobs, deployHandler, request, recorder := setupDeployHandler(body)

	jobs.On("Register", mock.Anything, mock.Anything).Return(nil, nil, nil)

	deployHandler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	args := jobs.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	count := job.TaskGroups[0].Count

	assert.Equal(t, 3, *count)
}
