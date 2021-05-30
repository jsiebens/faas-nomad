package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createMockJob(id string, status string) *api.Job {
	now := int64(time.Now().Nanosecond())
	name := "faas-fn-JOB123"
	namespace := "default"
	count := 1
	labels := []interface{}{}
	labels = append(labels, map[string]interface{}{"label": "test"})
	annotations := map[string]string{"topic": "test"}
	return &api.Job{
		ID:        &name,
		Name:      &name,
		Namespace: &namespace,
		Status:    &status,
		Meta:      annotations,
		SubmitTime: &now,
		TaskGroups: []*api.TaskGroup{{
			Count: &count,
			Tasks: []*api.Task{{
				Name: "Task" + id,
				Config: map[string]interface{}{"image": "docker",
					"labels": labels},
			}},
		},
		}}
}

func setupFunctionReader() (*services.MockJobs, http.HandlerFunc, *http.Request, *httptest.ResponseRecorder) {
	jobs := &services.MockJobs{}

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/system/functions", bytes.NewReader([]byte("")))

	config := &types.ProviderConfig{Scheduling: types.SchedulingConfig{
		JobPrefix: "faas-fn-",
	}}

	handler := MakeFunctionReader(config, jobs)

	return jobs, handler, request, response
}

func TestFunctionReaderReportsErrorWhenListingJobsFails(t *testing.T) {
	jobs, functionReader, request, recorder := setupFunctionReader()

	jobs.On("List", mock.Anything).Return([]*api.JobListStub{}, nil, fmt.Errorf("failure"))

	functionReader(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestFunctionReaderReportsErrorWhenGettingJobInfoFails(t *testing.T) {
	jobs, functionReader, request, recorder := setupFunctionReader()

	jobList := []*api.JobListStub{{ID: "123", Status: "running"}}

	jobs.On("List", mock.Anything).Return(jobList, nil, nil)
	jobs.On("Info", "123", mock.Anything).Return(nil, nil, fmt.Errorf("failure"))

	functionReader(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	jobs.AssertCalled(t, "Info", "123", mock.Anything)
}

func TestFunctionReaderReportsRunningFunctions(t *testing.T) {
	jobs, functionReader, request, recorder := setupFunctionReader()

	job1 := createMockJob("1234", "running")
	job2 := createMockJob("4567", "stopped")
	job3 := createMockJob("8929", "pending")

	d := make([]*api.JobListStub, 0)
	d = append(d, &api.JobListStub{ID: *job1.ID, Status: *job1.Status})
	d = append(d, &api.JobListStub{ID: *job2.ID, Status: *job2.Status})
	d = append(d, &api.JobListStub{ID: *job3.ID, Status: *job3.Status})

	jobs.On("List", mock.Anything).Return(d, nil, nil)
	jobs.On("Info", *job1.ID, mock.Anything).Return(job1, nil, nil)
	jobs.On("Info", *job2.ID, mock.Anything).Return(job2, nil, nil)

	functionReader(recorder, request)

	body, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		t.Fatal(err)
	}

	funcs := make([]ftypes.FunctionStatus, 0)
	json.Unmarshal(body, &funcs)

	assert.Equal(t, 2, len(funcs))
}
