package services

import (
	"net/url"

	"github.com/hashicorp/nomad/api"
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

func (ms *MockSecrets) Exists(key string) bool {
	args := ms.Called(key)
	return args.Bool(0)
}

func (ms *MockSecrets) Delete(key string) error {
	args := ms.Called(key)
	return args.Error(0)
}

type MockJobs struct {
	mock.Mock
}

func (m *MockJobs) List(q *api.QueryOptions) ([]*api.JobListStub, *api.QueryMeta, error) {
	args := m.Called(q)

	var jobs []*api.JobListStub
	if j := args.Get(0); j != nil {
		jobs = j.([]*api.JobListStub)
	}

	var meta *api.QueryMeta
	if r := args.Get(1); r != nil {
		meta = r.(*api.QueryMeta)
	}

	return jobs, meta, args.Error(2)
}

func (m *MockJobs) Info(jobID string, q *api.QueryOptions) (*api.Job, *api.QueryMeta, error) {
	args := m.Called(jobID, q)

	var job *api.Job
	if j := args.Get(0); j != nil {
		job = j.(*api.Job)
	}

	var meta *api.QueryMeta
	if r := args.Get(1); r != nil {
		meta = r.(*api.QueryMeta)
	}

	return job, meta, args.Error(2)
}

func (m *MockJobs) LatestDeployment(jobID string, q *api.QueryOptions) (*api.Deployment, *api.QueryMeta, error) {
	args := m.Called(jobID, q)

	var deployment *api.Deployment
	if j := args.Get(0); j != nil {
		deployment = j.(*api.Deployment)
	}

	var meta *api.QueryMeta
	if r := args.Get(1); r != nil {
		meta = r.(*api.QueryMeta)
	}

	return deployment, meta, args.Error(2)
}

func (m *MockJobs) RegisterOpts(job *api.Job, opts *api.RegisterOptions, w *api.WriteOptions) (*api.JobRegisterResponse, *api.WriteMeta, error) {

	args := m.Called(job, opts, w)

	var resp *api.JobRegisterResponse
	if r := args.Get(0); r != nil {
		resp = r.(*api.JobRegisterResponse)
	}

	var meta *api.WriteMeta
	if r := args.Get(1); r != nil {
		meta = r.(*api.WriteMeta)
	}

	return resp, meta, args.Error(2)

}

func (m *MockJobs) Deregister(jobID string, purge bool, q *api.WriteOptions) (string, *api.WriteMeta, error) {

	args := m.Called(jobID, purge, q)

	return "", nil, args.Error(2)
}

func (m *MockJobs) Allocations(jobID string, allAllocs bool, q *api.QueryOptions) ([]*api.AllocationListStub, *api.QueryMeta, error) {
	args := m.Called(jobID, allAllocs, q)

	var allocs []*api.AllocationListStub
	if a := args.Get(0); a != nil {
		allocs = a.([]*api.AllocationListStub)
	}

	var meta *api.QueryMeta
	if r := args.Get(1); r != nil {
		meta = r.(*api.QueryMeta)
	}

	return allocs, meta, args.Error(2)
}

type MockResolver struct {
	mock.Mock
}

func (mr *MockResolver) Resolve(functionName string) (url.URL, error) {
	args := mr.Called(functionName)

	var resp url.URL
	if r := args.Get(0); r != nil {
		resp = r.(url.URL)
	}

	return resp, args.Error(2)
}

func (mr *MockResolver) RemoveCacheItem(functionName string) {
	mr.Called(functionName)
}
