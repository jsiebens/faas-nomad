package main

import (
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/jsiebens/faas-nomad/pkg/handlers"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	fbootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/proxy"
	ftypes "github.com/openfaas/faas-provider/types"
)

func main() {

	config, err := types.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := setupLogging()

	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)

	secrets, err := services.NewVaultSecrets(config.Vault)
	if err != nil {
		log.Fatal(err)
	}

	jobs, err := services.NewNomadJobs(config.Nomad)
	if err != nil {
		log.Fatal(err)
	}

	resolver, err := services.NewConsulResolver(config, logger)
	if err != nil {
		log.Fatal(err)
	}

	bootstrapHandlers := ftypes.FaaSHandlers{
		FunctionProxy:        proxy.NewHandlerFunc(config.FaaS, resolver),
		FunctionReader:       handlers.MakeFunctionReader(config, jobs, logger),
		DeployHandler:        handlers.MakeDeployHandler(config, jobs, logger),
		DeleteHandler:        handlers.MakeDeleteHandler(config, jobs, resolver, logger),
		ReplicaReader:        handlers.MakeReplicaReader(config, jobs, logger),
		ReplicaUpdater:       handlers.MakeReplicaUpdater(config, jobs, logger),
		SecretHandler:        handlers.MakeSecretHandler(secrets, logger),
		LogHandler:           unimplemented,
		UpdateHandler:        handlers.MakeDeployHandler(config, jobs, logger),
		HealthHandler:        handlers.MakeHealthHandler(),
		InfoHandler:          handlers.MakeInfoHandler(),
		ListNamespaceHandler: handlers.MakeListNamespaceHandler(config),
	}

	log.Printf("Listening on TCP port: %d\n", *config.FaaS.TCPPort)
	fbootstrap.Serve(&bootstrapHandlers, &config.FaaS)
}

func unimplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func setupLogging() hclog.Logger {
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:       "faas-nomad",
		Level:      hclog.LevelFromString("debug"),
		JSONFormat: false,
		Output:     createLogFile(),
	})
	return appLogger
}

func createLogFile() *os.File {
	return os.Stdout
}
