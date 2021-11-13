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

func MakeDeleteHandler(config *types.ProviderConfig, jobs services.Jobs, logger hclog.Logger) func(w http.ResponseWriter, r *http.Request) {
	log := logger.Named("delete_handler")

	return func(w http.ResponseWriter, r *http.Request) {

		body, _ := ioutil.ReadAll(r.Body)
		req := ftypes.DeleteFunctionRequest{}
		err := json.Unmarshal(body, &req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		namespace := config.Scheduling.Namespace
		jobName := fmt.Sprintf("%s%s", config.Scheduling.JobPrefix, req.FunctionName)

		_, _, err = jobs.Deregister(jobName, true, &api.WriteOptions{Namespace: namespace})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			log.Error("Error deregistering function", "function", jobName, "namespace", namespace, "error", err.Error())
			return
		}

		log.Debug("Function deregistered successfully", "function", jobName, "namespace", namespace)
		w.WriteHeader(http.StatusOK)
	}

}
