package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
)

func MakeFunctionReader(config *types.ProviderConfig, jobs services.Jobs) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		options := &api.QueryOptions{
			Namespace: config.Scheduling.Namespace,
			Prefix:    config.Scheduling.JobPrefix,
		}

		list, _, err := jobs.List(options)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		functions, err := getFunctions(config, jobs, list, options)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set(HeaderContentType, TypeApplicationJson)
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)
	}
}

func getFunctions(config *types.ProviderConfig, client services.Jobs, jobs []*api.JobListStub, options *api.QueryOptions) ([]ftypes.FunctionStatus, error) {
	functions := make([]ftypes.FunctionStatus, 0)
	for _, j := range jobs {
		if j.Status == "running" || j.Status == "pending" {
			job, _, err := client.Info(j.ID, options)
			if err != nil {
				return functions, err
			}

			functions = append(functions, createFunctionStatus(job, config.Scheduling.JobPrefix))
		}
	}
	return functions, nil
}
