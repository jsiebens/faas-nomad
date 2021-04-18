package main

import (
	"log"
	"net/http"

	"github.com/jsiebens/faas-nomad/pkg/handlers"
	"github.com/jsiebens/faas-nomad/pkg/types"
	fbootstrap "github.com/openfaas/faas-provider"
	ftypes "github.com/openfaas/faas-provider/types"
)

func main() {

	config, err := types.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	bootstrapHandlers := ftypes.FaaSHandlers{
		FunctionProxy:        unimplemented,
		FunctionReader:       unimplemented,
		DeployHandler:        unimplemented,
		DeleteHandler:        unimplemented,
		ReplicaReader:        unimplemented,
		ReplicaUpdater:       unimplemented,
		SecretHandler:        unimplemented,
		LogHandler:           unimplemented,
		UpdateHandler:        unimplemented,
		HealthHandler:        handlers.MakeHealthHandler(),
		InfoHandler:          handlers.MakeInfoHandler(),
		ListNamespaceHandler: handlers.MakeListNamespaceHandler(),
	}

	log.Printf("Listening on TCP port: %d\n", *config.TCPPort)
	fbootstrap.Serve(&bootstrapHandlers, config)
}

func unimplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
