package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/jsiebens/faas-nomad/pkg/handlers"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	"github.com/jsiebens/faas-nomad/version"
	fbootstrap "github.com/openfaas/faas-provider"
	ftypes "github.com/openfaas/faas-provider/types"
)

var (
	configFile = flag.String("config", "", "Path to the configuration file.")
)

func main() {
	flag.Parse()

	config, err := types.LoadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	logger := setupLogging(config.Log)

	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)

	jobFactory := services.NewJobFactory(config)

	secrets, err := services.NewVaultSecrets(config.Vault)
	if err != nil {
		log.Fatal(err)
	}

	jobs, err := services.NewNomadJobs(config.Nomad)
	if err != nil {
		log.Fatal(err)
	}

	proxy, err := services.NewProxyHandler(*config, logger)
	if err != nil {
		log.Fatal(err)
	}

	bootstrapHandlers := ftypes.FaaSHandlers{
		FunctionProxy:        proxy.Handler(),
		FunctionReader:       handlers.MakeFunctionReader(config, jobs, logger),
		DeployHandler:        handlers.MakeDeployHandler(config, jobFactory, jobs, secrets, logger),
		DeleteHandler:        handlers.MakeDeleteHandler(config, jobs, logger),
		ReplicaReader:        handlers.MakeReplicaReader(config, jobs, proxy.Resolver(), logger),
		ReplicaUpdater:       handlers.MakeReplicaUpdater(config, jobs, logger),
		SecretHandler:        handlers.MakeSecretHandler(secrets, logger),
		LogHandler:           unimplemented,
		UpdateHandler:        handlers.MakeDeployHandler(config, jobFactory, jobs, secrets, logger),
		HealthHandler:        handlers.MakeHealthHandler(),
		InfoHandler:          handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommit),
		ListNamespaceHandler: handlers.MakeListNamespaceHandler(config),
	}

	logger.Info(fmt.Sprintf("Listening on TCP port: %d", *config.FaaS.TCPPort))

	fbootstrap.Serve(&bootstrapHandlers, &config.FaaS)
}

func unimplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func setupLogging(config types.LogConfig) hclog.Logger {
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:       "faas-nomad",
		Level:      hclog.LevelFromString(config.Level),
		JSONFormat: strings.ToLower(config.Format) == "json",
		Output:     createLogFile(config),
	})
	return appLogger
}

func createLogFile(config types.LogConfig) *os.File {
	if config.File != "" {
		f, err := os.OpenFile(config.File, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			return f
		}
		log.Printf("Unable to open file for output, defaulting to std out: %s\n", err.Error())
	}
	return os.Stdout
}
