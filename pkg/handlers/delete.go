package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
)

func MakeDeleteHandler(config *types.ProviderConfig, jobs services.Jobs, resolver services.ServiceResolver) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		body, _ := ioutil.ReadAll(r.Body)
		req := ftypes.DeleteFunctionRequest{}
		err := json.Unmarshal(body, &req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		jobName := fmt.Sprintf("%s%s", config.Scheduling.JobPrefix, req.FunctionName)

		_, _, err = jobs.Deregister(jobName, false, &api.WriteOptions{Namespace: config.Scheduling.Namespace})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resolver.RemoveCacheItem(jobName)

		w.WriteHeader(http.StatusOK)
	}

}
