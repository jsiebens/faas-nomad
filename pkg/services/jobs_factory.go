package services

import (
	"fmt"
	"github.com/hashicorp/nomad/api"
	"github.com/jsiebens/faas-nomad/pkg/types"
	ftypes "github.com/openfaas/faas-provider/types"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	EnvProcessName = "fprocess"
)

var (
	logFiles = 5
	logSize  = 2
)

type JobFactory interface {
	CreateJob(namespace string, fd ftypes.FunctionDeployment) *api.Job
}

func NewJobFactory(config *types.ProviderConfig) JobFactory {
	return &jobFactory{config: config}
}

type jobFactory struct {
	config *types.ProviderConfig
}

func (f *jobFactory) CreateJob(namespace string, fd ftypes.FunctionDeployment) *api.Job {

	region := f.config.Scheduling.Region
	constraints, datacenters := f.createConstraints(f.config, fd)
	name := fmt.Sprintf("%s%s", f.config.Scheduling.JobPrefix, fd.Service)
	priority := 50

	job := api.NewServiceJob(name, name, region, priority)
	job.Namespace = &namespace
	job.Meta = f.createAnnotations(fd)
	job.Update = f.createUpdateStrategy(fd)
	job.Datacenters = datacenters
	job.Constraints = constraints
	job.TaskGroups = f.createTaskGroups(fd)

	return job
}

func (f *jobFactory) createConstraints(config *types.ProviderConfig, r ftypes.FunctionDeployment) ([]*api.Constraint, []string) {
	var constraints []*api.Constraint
	var datacenters []string

	if r.Constraints == nil || len(r.Constraints) == 0 {
		return constraints, config.Scheduling.Datacenters
	}

	for _, requestConstraint := range r.Constraints {
		fields := strings.Fields(requestConstraint)

		if len(fields) < 3 {
			continue
		}

		if strings.Contains(fields[0], "datacenter") && (fields[1] == "==" || fields[1] == "=") {
			datacenters = append(datacenters, fields[2])
			continue
		}

		attribute := fields[0]
		operator := fields[1]
		value := strings.Join(fields[2:], " ")

		match, _ := regexp.MatchString("^\\${.*}$", attribute)
		if !match {
			attribute = fmt.Sprintf("${%v}", attribute)
		}

		if operator == "==" {
			operator = "="
		}

		constraints = append(constraints, &api.Constraint{
			LTarget: attribute,
			Operand: operator,
			RTarget: value,
		})
	}

	if len(datacenters) != 0 {
		return constraints, datacenters
	} else {
		return constraints, config.Scheduling.Datacenters
	}
}

func (f *jobFactory) createAnnotations(r ftypes.FunctionDeployment) map[string]string {
	annotations := map[string]string{}
	if r.Annotations != nil {
		for k, v := range *r.Annotations {
			annotations[k] = v
		}
	}
	return annotations
}

func (f *jobFactory) createUpdateStrategy(fd ftypes.FunctionDeployment) *api.UpdateStrategy {
	// Update Strategy
	stagger := types.ParseIntOrDurationValueFromMap(fd.Labels, "com.openfaas.nomad.update.stagger", 5*time.Second)
	maxParallel := types.ParseIntValueFromMap(fd.Labels, "com.openfaas.nomad.update.max_parallel", 3)
	healthCheck := types.ParseStringValueFromMap(fd.Labels, "com.openfaas.nomad.update.health_check", "checks")
	minHealthyTime := types.ParseIntOrDurationValueFromMap(fd.Labels, "com.openfaas.nomad.update.min_healthy_time", 5*time.Second)
	healthyDeadline := types.ParseIntOrDurationValueFromMap(fd.Labels, "com.openfaas.nomad.update.healthy_deadline", 2*time.Minute)
	progressDeadline := types.ParseIntOrDurationValueFromMap(fd.Labels, "com.openfaas.nomad.update.progress_deadline", 5*time.Minute)
	canary := types.ParseIntValueFromMap(fd.Labels, "com.openfaas.nomad.update.canary", 0)
	autoRevert := types.ParseBoolValueFromMap(fd.Labels, "com.openfaas.nomad.update.auto_revert", true)
	autoPromote := types.ParseBoolValueFromMap(fd.Labels, "com.openfaas.nomad.update.auto_promote", false)

	return &api.UpdateStrategy{
		Stagger:          &stagger,
		MaxParallel:      &maxParallel,
		HealthCheck:      &healthCheck,
		MinHealthyTime:   &minHealthyTime,
		HealthyDeadline:  &healthyDeadline,
		ProgressDeadline: &progressDeadline,
		Canary:           &canary,
		AutoRevert:       &autoRevert,
		AutoPromote:      &autoPromote,
	}
}

func (f *jobFactory) createTaskGroups(fd ftypes.FunctionDeployment) []*api.TaskGroup {
	count := f.getInitialCount(fd)

	network := &api.NetworkResource{
		Mode: f.config.Scheduling.NetworkingMode,
	}

	if f.config.Consul.ConnectAware {
		network.Mode = "bridge"
	} else {
		network.DynamicPorts = []api.Port{{Label: "http", To: 8080}}
	}

	gracePeriod := 5 * time.Second

	check := api.ServiceCheck{
		Name:                   fd.Service,
		Type:                   "http",
		Path:                   "/_/health",
		InitialStatus:          "critical",
		SuccessBeforePassing:   1,
		FailuresBeforeCritical: 3,
		Interval:               5 * time.Second,
		Timeout:                2 * time.Second,
		CheckRestart: &api.CheckRestart{
			Limit:          3,
			Grace:          &gracePeriod,
			IgnoreWarnings: false,
		},
	}

	if f.config.Consul.ConnectAware {
		check.Expose = true
	} else {
		check.PortLabel = "http"
	}

	service := &api.Service{
		Name:      fmt.Sprintf("%s%s", f.config.Scheduling.JobPrefix, fd.Service),
		PortLabel: "http",
		Tags:      []string{"http", "faas"},
		Checks:    []api.ServiceCheck{check},
	}

	if f.config.Consul.ConnectAware {
		service.PortLabel = "8080"
		service.Connect = &api.ConsulConnect{
			SidecarService: &api.ConsulSidecarService{},
		}
	}

	group := api.TaskGroup{
		Name:     &fd.Service,
		Count:    &count,
		Networks: []*api.NetworkResource{network},
		Services: []*api.Service{service},
		Tasks:    []*api.Task{f.createTask(fd)},
	}

	return []*api.TaskGroup{&group}
}

func (f *jobFactory) getInitialCount(fd ftypes.FunctionDeployment) int {
	return types.ParseIntValueFromMap(fd.Labels, "com.openfaas.scale.min", 1)
}

func (f *jobFactory) createTask(fd ftypes.FunctionDeployment) *api.Task {
	var task api.Task

	configMap := map[string]interface{}{
		"image":  fd.Image,
		"labels": createLabels(fd),
	}

	if !f.config.Consul.ConnectAware {
		configMap["ports"] = []string{"http"}
	}

	task = api.Task{
		Name:   fd.Service,
		Driver: "docker",
		Config: configMap,
		LogConfig: &api.LogConfig{
			MaxFiles:      &logFiles,
			MaxFileSizeMB: &logSize,
		},
		Env:       createEnvVars(fd),
		Resources: createTaskResources(fd),
	}

	if len(fd.Secrets) > 0 {
		task.Config["volumes"] = createSecretVolumes(fd.Secrets)
		task.Templates = createSecrets(f.config.Vault.SecretPathPrefix, fd.Secrets)
		task.Vault = &api.Vault{
			Policies: []string{f.config.Vault.Policy},
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

		embeddedTemplate := fmt.Sprintf(`{{with secret "%s"}}{{base64Decode .Data.value}}{{end}}`, path)
		template := &api.Template{
			DestPath:     &destPath,
			EmbeddedTmpl: &embeddedTemplate,
		}

		templates = append(templates, template)
	}

	return templates
}
