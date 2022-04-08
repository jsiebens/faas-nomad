package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/jsiebens/faas-nomad/pkg/resolver"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
)

func MakeReplicaReader(config *types.ProviderConfig, client services.Jobs, resolver resolver.ServiceResolver, logger hclog.Logger) http.HandlerFunc {
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

		if job.Status != nil && *job.Status == "dead" {
			status.Replicas = 0
			status.AvailableReplicas = 0
		} else {
			availableReplicas, _, err := resolver.ResolveAll(functionName)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				log.Error("Error reading function status", "function", functionName, "namespace", namespace, "error", err.Error())
				return
			}
			status.AvailableReplicas = uint64(len(availableReplicas))
		}

		statusBytes, _ := json.Marshal(status)
		w.Header().Set(HeaderContentType, TypeApplicationJson)
		w.WriteHeader(http.StatusOK)
		w.Write(statusBytes)

		log.Trace("Function status read successfully", "function", functionName, "namespace", namespace)
	}

}
