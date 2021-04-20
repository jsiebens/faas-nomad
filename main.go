package main

import (
	"log"
	"net/http"

	"github.com/jsiebens/faas-nomad/pkg/handlers"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	fbootstrap "github.com/openfaas/faas-provider"
	ftypes "github.com/openfaas/faas-provider/types"
)

func main() {

	config, err := types.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	secrets, err := services.NewVaultSecrets(config.Vault)
	if err != nil {
		log.Fatal(err)
	}

	jobs, err := services.NewNomadJobs(config.Nomad)
	if err != nil {
		log.Fatal(err)
	}

	bootstrapHandlers := ftypes.FaaSHandlers{
		FunctionProxy:        unimplemented,
		FunctionReader:       handlers.MakeFunctionReader(config, jobs),
		DeployHandler:        handlers.MakeDeployHandler(config, jobs),
		DeleteHandler:        handlers.MakeDeleteHandler(config, jobs),
		ReplicaReader:        unimplemented,
		ReplicaUpdater:       unimplemented,
		SecretHandler:        handlers.MakeSecretHandler(secrets),
		LogHandler:           unimplemented,
		UpdateHandler:        handlers.MakeDeployHandler(config, jobs),
		HealthHandler:        handlers.MakeHealthHandler(),
		InfoHandler:          handlers.MakeInfoHandler(),
		ListNamespaceHandler: handlers.MakeListNamespaceHandler(),
	}

	log.Printf("Listening on TCP port: %d\n", *config.FaaS.TCPPort)
	fbootstrap.Serve(&bootstrapHandlers, &config.FaaS)
}

func unimplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
