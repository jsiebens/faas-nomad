package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
)

func MakeReplicaUpdater(config *types.ProviderConfig, client services.Jobs, logger hclog.Logger) func(w http.ResponseWriter, r *http.Request) {
	log := logger.Named("replica_updater")

	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		req := ftypes.ScaleServiceRequest{}
		err := json.Unmarshal(body, &req)

		namespace := config.Scheduling.Namespace

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("Error updating function", "error", err.Error())
			return
		}

		options := &api.QueryOptions{
			Namespace: namespace,
		}

		job, _, err := client.Info(fmt.Sprintf("%s%s", config.Scheduling.JobPrefix, req.ServiceName), options)

		if job == nil || err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		replicas := int(req.Replicas)
		job.TaskGroups[0].Count = &replicas

		opts := &api.RegisterOptions{PreserveCounts: false}
		_, _, err = client.RegisterOpts(job, opts, nil)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			log.Error("Error updating function", "function", req.ServiceName, "namespace", namespace, "error", err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		log.Debug("Function updated successfully", "function", req.ServiceName, "namespace", namespace)
	}
}
