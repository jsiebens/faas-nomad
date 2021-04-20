package services

import (
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/types"
)

type Jobs interface {
	List(q *api.QueryOptions) ([]*api.JobListStub, *api.QueryMeta, error)
	Info(jobID string, q *api.QueryOptions) (*api.Job, *api.QueryMeta, error)
	Register(*api.Job, *api.WriteOptions) (*api.JobRegisterResponse, *api.WriteMeta, error)
	Deregister(jobID string, purge bool, q *api.WriteOptions) (string, *api.WriteMeta, error)
	Allocations(jobID string, allAllocs bool, q *api.QueryOptions) ([]*api.AllocationListStub, *api.QueryMeta, error)
}

func NewNomadJobs(config types.NomadConfig) (Jobs, error) {
	nomadClient, err := api.NewClient(&api.Config{
		Address: config.Addr,
	})

	if err != nil {
		return nil, err
	}

	return nomadClient.Jobs(), nil
}
