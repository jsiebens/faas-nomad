package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
)

func MakeFunctionReader(config *types.ProviderConfig, jobs services.Jobs, logger hclog.Logger) func(w http.ResponseWriter, r *http.Request) {
	log := logger.Named("function_reader")

	return func(w http.ResponseWriter, r *http.Request) {
		namespace := config.Scheduling.Namespace

		options := &api.QueryOptions{
			Namespace: namespace,
			Prefix:    config.Scheduling.JobPrefix,
		}

		list, _, err := jobs.List(options)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			log.Error("Error listing functions", "namespace", namespace, "error", err.Error())
			return
		}

		functions, err := getFunctions(config, jobs, list, options)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			log.Error("Error listing functions", "namespace", namespace, "error", err.Error())
			return
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set(HeaderContentType, TypeApplicationJson)
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)

		log.Trace("Functions listed successfully", "namespace", namespace)
	}
}

func getFunctions(config *types.ProviderConfig, client services.Jobs, jobs []*api.JobListStub, options *api.QueryOptions) ([]ftypes.FunctionStatus, error) {
	functions := make([]ftypes.FunctionStatus, 0)
	for _, j := range jobs {
		job, _, err := client.Info(j.ID, options)
		if err != nil {
			return functions, err
		}

		functions = append(functions, createFunctionStatus(job, config.Scheduling.JobPrefix))
	}
	return functions, nil
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
	return
}

func writeJsonResponse(w http.ResponseWriter, code int, response []byte) {
	w.WriteHeader(code)
	w.Header().Set(HeaderContentType, TypeApplicationJson)
	w.Write(response)
}
