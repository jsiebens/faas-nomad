package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
)

func MakeReplicaReader(config *types.ProviderConfig, client services.Jobs, logger hclog.Logger) http.HandlerFunc {
	log := logger.Named("replica_reader")

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		functionName := vars["name"]
		namespace := config.Scheduling.Namespace

		options := &api.QueryOptions{
			Namespace: namespace,
		}

		job, _, err := client.Info(fmt.Sprintf("%s%s", config.Scheduling.JobPrefix, functionName), options)

		if job == nil || err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// get the number of available allocations from the job
		readyCount, err := getAllocationReadyCount(client, job, options)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			log.Error("Error reading function status", "function", functionName, "namespace", namespace, "error", err.Error())
			return
		}

		status := createFunctionStatus(job, config.Scheduling.JobPrefix)
		status.AvailableReplicas = readyCount

		statusBytes, _ := json.Marshal(status)
		w.Header().Set(HeaderContentType, TypeApplicationJson)
		w.WriteHeader(http.StatusOK)
		w.Write(statusBytes)

		log.Trace("Function status read successfully", "function", functionName, "namespace", namespace)
	}

}

func getAllocationReadyCount(client services.Jobs, job *api.Job, q *api.QueryOptions) (uint64, error) {
	allocations, _, err := client.Allocations(*job.ID, true, q)
	var readyCount uint64

	for _, a := range allocations {
		for _, ts := range a.TaskStates {
			if ts.State == "running" {
				readyCount += 1
			}
		}
	}

	return readyCount, err
}
