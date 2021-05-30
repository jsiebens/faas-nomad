package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
)

func MakeReplicaReader(config *types.ProviderConfig, client services.Jobs) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		functionName := vars["name"]

		options := &api.QueryOptions{
			Namespace: config.Scheduling.Namespace,
		}

		job, _, err := client.Info(fmt.Sprintf("%s%s", config.Scheduling.JobPrefix, functionName), options)

		if job == nil || err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// get the number of available allocations from the job
		readyCount, err := getAllocationReadyCount(client, job, options)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		status := createFunctionStatus(job, config.Scheduling.JobPrefix)
		status.AvailableReplicas = readyCount

		statusBytes, _ := json.Marshal(status)
		w.Header().Set(HeaderContentType, TypeApplicationJson)
		w.WriteHeader(http.StatusOK)
		w.Write(statusBytes)
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
