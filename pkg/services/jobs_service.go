package services

import (
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/types"
)

type Jobs interface {
	List(q *api.QueryOptions) ([]*api.JobListStub, *api.QueryMeta, error)
	Info(jobID string, q *api.QueryOptions) (*api.Job, *api.QueryMeta, error)
	LatestDeployment(jobID string, q *api.QueryOptions) (*api.Deployment, *api.QueryMeta, error)
	Register(*api.Job, *api.WriteOptions) (*api.JobRegisterResponse, *api.WriteMeta, error)
	Deregister(jobID string, purge bool, q *api.WriteOptions) (string, *api.WriteMeta, error)
	Allocations(jobID string, allAllocs bool, q *api.QueryOptions) ([]*api.AllocationListStub, *api.QueryMeta, error)
}

func NewNomadJobs(config types.NomadConfig) (Jobs, error) {
	c := api.DefaultConfig()

	c.Address = config.Addr
	c.SecretID = config.ACLToken
	c.TLSConfig.CACert = config.CACert
	c.TLSConfig.ClientCert = config.ClientCert
	c.TLSConfig.ClientKey = config.ClientKey
	c.TLSConfig.Insecure = config.TLSSkipVerify

	nomadClient, err := api.NewClient(c)

	if err != nil {
		return nil, err
	}

	return nomadClient.Jobs(), nil
}
