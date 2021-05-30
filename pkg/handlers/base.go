package handlers

import (
	"strings"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/openfaas/faas-provider/types"
)

const (
	HeaderContentType   = "Content-Type"
	TypeApplicationJson = "application/json"

	EnvProcessName = "fprocess"
)

func createFunctionStatus(job *api.Job, jobPrefix string) types.FunctionStatus {
	var labels = map[string]string{}
	task := job.TaskGroups[0].Tasks[0]

	if task.Config["labels"] != nil {
		labels = parseLabels(task.Config["labels"].([]interface{}))
	}

	var annotations = map[string]string{}
	if job.Meta != nil {
		annotations = job.Meta
	}

	return types.FunctionStatus{
		Name:            sanitiseJobName(job, jobPrefix),
		Namespace:       *job.Namespace,
		Image:           task.Config["image"].(string),
		Replicas:        uint64(*job.TaskGroups[0].Count),
		InvocationCount: 0,
		Labels:          &labels,
		Annotations:     &annotations,
		EnvProcess:      getEnvProcess(task.Env),
		CreatedAt:       time.Unix(0, *job.SubmitTime),
	}
}

func getEnvProcess(m map[string]string) string {
	if m == nil {
		return ""
	}
	s := m[EnvProcessName]
	if len(s) != 0 {
		return s
	}
	return ""
}

func parseLabels(labels []interface{}) map[string]string {
	newLabels := map[string]string{}
	if len(labels) > 0 {
		for _, l := range labels {
			for k, v := range l.(map[string]interface{}) {
				newLabels[k] = v.(string)
			}
		}
	}
	return newLabels
}

func sanitiseJobName(job *api.Job, jobPrefix string) string {
	return strings.Replace(*job.Name, jobPrefix, "", -1)
}
