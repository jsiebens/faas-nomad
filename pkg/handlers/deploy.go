package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/services"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
)

var (
	logFiles = 5
	logSize  = 2

	// Update Strategy
	updateAutoRevert      = true
	updateMinHealthyTime  = 5 * time.Second
	updateHealthyDeadline = 30 * time.Second
	updateStagger         = 5 * time.Second
)

func MakeDeployHandler(config *types.ProviderConfig, jobs services.Jobs, logger hclog.Logger) func(w http.ResponseWriter, r *http.Request) {
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

		job := createJob(config, namespace, req)

		// Use the Nomad API client to register the job
		_, _, err = jobs.Register(job, &api.WriteOptions{Namespace: namespace})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			log.Error("Error registering function", "function", *job.Name, "namespace", *job.Namespace, "error", err.Error())
			return
		}

		log.Debug("Function registered successfully", "function", *job.Name, "namespace", *job.Namespace)
		w.WriteHeader(http.StatusOK)
	}
}

func createJob(config *types.ProviderConfig, namespace string, fd ftypes.FunctionDeployment) *api.Job {

	region := config.Scheduling.Region
	datacenters := config.Scheduling.Datacenters
	name := fmt.Sprintf("%s%s", config.Scheduling.JobPrefix, fd.Service)
	priority := 50

	job := api.NewServiceJob(name, name, region, priority)
	job.Namespace = &namespace
	job.Meta = createAnnotations(fd)
	job.Update = createUpdateStrategy()
	job.Datacenters = datacenters
	job.TaskGroups = createTaskGroups(config, fd)

	return job
}

func createTaskGroups(config *types.ProviderConfig, fd ftypes.FunctionDeployment) []*api.TaskGroup {
	count := getInitialCount(fd)

	network := api.NetworkResource{
		Mode:         config.Scheduling.NetworkingMode,
		DynamicPorts: []api.Port{{Label: "http", To: 8080}},
	}

	group := api.TaskGroup{
		Name:     &fd.Service,
		Count:    &count,
		Networks: []*api.NetworkResource{&network},
		Tasks:    []*api.Task{createTask(config, fd)},
	}

	return []*api.TaskGroup{&group}
}

func getInitialCount(fd ftypes.FunctionDeployment) int {
	if fd.Labels != nil {
		m := *fd.Labels
		count, err := strconv.ParseInt(m["com.openfaas.scale.min"], 10, 32)
		if err == nil {
			return int(count)
		}
	}
	return 1
}

func createTask(config *types.ProviderConfig, fd ftypes.FunctionDeployment) *api.Task {
	var task api.Task
	task = api.Task{
		Name:   fd.Service,
		Driver: "docker",
		Config: map[string]interface{}{
			"image":  fd.Image,
			"ports":  []string{"http"},
			"labels": createLabels(fd),
		},
		Services: []*api.Service{
			{
				Name:      fmt.Sprintf("%s%s", config.Scheduling.JobPrefix, fd.Service),
				PortLabel: "http",
			},
		},
		LogConfig: &api.LogConfig{
			MaxFiles:      &logFiles,
			MaxFileSizeMB: &logSize,
		},
		Env:       createEnvVars(fd),
		Resources: createTaskResources(fd),
	}

	if len(fd.Secrets) > 0 {
		task.Config["volumes"] = createSecretVolumes(fd.Secrets)
		task.Templates = createSecrets(config.Vault.SecretPathPrefix, fd.Secrets)
		task.Vault = &api.Vault{
			Policies: []string{config.Vault.Policy},
		}
	}

	return &task
}

func createTaskResources(fd ftypes.FunctionDeployment) *api.Resources {
	taskMemory := 128
	taskCPU := 100

	if fd.Limits != nil {
		cpu, err := strconv.ParseInt(fd.Limits.CPU, 10, 32)
		if err == nil {
			taskCPU = int(cpu)
		}

		mem, err := strconv.ParseInt(fd.Limits.Memory, 10, 32)
		if err == nil {
			taskMemory = int(mem)
		}
	}

	return &api.Resources{
		MemoryMB: &taskMemory,
		CPU:      &taskCPU,
	}
}

func createAnnotations(r ftypes.FunctionDeployment) map[string]string {
	annotations := map[string]string{}
	if r.Annotations != nil {
		for k, v := range *r.Annotations {
			annotations[k] = v
		}
	}
	return annotations
}

func createLabels(r ftypes.FunctionDeployment) []map[string]interface{} {
	var labels = make(map[string]interface{})
	if r.Labels != nil {
		for k, v := range *r.Labels {
			labels[k] = v
		}
	}
	return []map[string]interface{}{labels}
}

func createEnvVars(r ftypes.FunctionDeployment) map[string]string {
	envVars := map[string]string{}

	if r.EnvVars != nil {
		envVars = r.EnvVars
	}

	if r.EnvProcess != "" {
		envVars[EnvProcessName] = r.EnvProcess
	}

	return envVars
}

func createUpdateStrategy() *api.UpdateStrategy {
	return &api.UpdateStrategy{
		MinHealthyTime:  &updateMinHealthyTime,
		AutoRevert:      &updateAutoRevert,
		Stagger:         &updateStagger,
		HealthyDeadline: &updateHealthyDeadline,
	}
}

func createSecretVolumes(secrets []string) []string {
	newVolumes := []string{}
	for _, s := range secrets {
		destPath := "secrets/" + s + ":/var/openfaas/secrets/" + s
		newVolumes = append(newVolumes, destPath)
	}
	return newVolumes
}

func createSecrets(vaultPrefix string, secrets []string) []*api.Template {
	var templates []*api.Template

	for _, s := range secrets {
		path := fmt.Sprintf("%s/%s", vaultPrefix, s)
		destPath := "secrets/" + s

		embeddedTemplate := fmt.Sprintf(`{{with secret "%s"}}{{.Data.value}}{{end}}`, path)
		template := &api.Template{
			DestPath:     &destPath,
			EmbeddedTmpl: &embeddedTemplate,
		}

		templates = append(templates, template)
	}

	return templates
}
