package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
	"io/ioutil"
	"net/http"
)

func MakeDeployHandler(config *types.ProviderConfig, jobFactory services.JobFactory, jobs services.Jobs, secrets services.Secrets, logger hclog.Logger) func(w http.ResponseWriter, r *http.Request) {
	log := logger.Named("deploy_handler")

	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)

		req := ftypes.FunctionDeployment{}
		err := json.Unmarshal(body, &req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		namespace := config.Scheduling.Namespace

		// validate secrets
		for _, s := range req.Secrets {
			if !secrets.Exists(s) {
				writeError(w, http.StatusBadRequest, fmt.Errorf("secret with key '%s' is not available", s))
				return
			}
		}

		job := jobFactory.CreateJob(namespace, req)

		// Use the Nomad API client to register the job
		writeOptions := &api.WriteOptions{Namespace: namespace}
		registerOptions := &api.RegisterOptions{
			PreserveCounts: true,
		}
		_, _, err = jobs.RegisterOpts(job, registerOptions, writeOptions)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			log.Error("Error registering function", "function", *job.Name, "namespace", *job.Namespace, "error", err.Error())
			return
		}

		log.Debug("Function registered successfully", "function", *job.Name, "namespace", *job.Namespace)
		w.WriteHeader(http.StatusOK)
	}
}
