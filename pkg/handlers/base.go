package handlers

import (
	"strings"

	"github.com/hashicorp/nomad/api"
	"github.com/openfaas/faas-provider/types"
)

const (
	HeaderContentType = "Content-Type"

	TypeApplicationJson = "application/json"
)

func createFunctionStatus(job *api.Job, jobPrefix string) types.FunctionStatus {
	var labels = map[string]string{}
	if job.TaskGroups[0].Tasks[0].Config["labels"] != nil {
		labels = parseLabels(job.TaskGroups[0].Tasks[0].Config["labels"].([]interface{}))
	}

	var annotations = map[string]string{}
	if job.Meta != nil {
		annotations = job.Meta
	}

	return types.FunctionStatus{
		Name:            sanitiseJobName(job, jobPrefix),
		Namespace:       *job.Namespace,
		Image:           job.TaskGroups[0].Tasks[0].Config["image"].(string),
		Replicas:        uint64(*job.TaskGroups[0].Count),
		InvocationCount: 0,
		Labels:          &labels,
		Annotations:     &annotations,
	}
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
