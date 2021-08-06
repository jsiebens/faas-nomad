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

		status := createFunctionStatus(job, config.Scheduling.JobPrefix)

		deployment, _, err := client.LatestDeployment(fmt.Sprintf("%s%s", config.Scheduling.JobPrefix, functionName), options)
		if deployment != nil && err == nil {
			state := deployment.TaskGroups[functionName]
			status.Replicas = uint64(state.DesiredTotal)
			status.AvailableReplicas = uint64(state.HealthyAllocs)
		} else {
			writeError(w, http.StatusInternalServerError, err)
			log.Error("Error reading function status", "function", functionName, "namespace", namespace, "error", err.Error())
			return
		}

		statusBytes, _ := json.Marshal(status)
		w.Header().Set(HeaderContentType, TypeApplicationJson)
		w.WriteHeader(http.StatusOK)
		w.Write(statusBytes)

		log.Trace("Function status read successfully", "function", functionName, "namespace", namespace)
	}

}
