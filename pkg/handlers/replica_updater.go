package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
)

func MakeReplicaUpdater(config *types.ProviderConfig, client services.Jobs) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		req := ftypes.ScaleServiceRequest{}
		err := json.Unmarshal(body, &req)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		job, _, err := client.Info(fmt.Sprintf("%s%s", config.Scheduling.JobPrefix, req.ServiceName), nil)

		if job == nil || err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		replicas := int(req.Replicas)
		job.TaskGroups[0].Count = &replicas

		_, _, err = client.Register(job, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
