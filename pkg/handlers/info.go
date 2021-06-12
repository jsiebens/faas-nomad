package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/openfaas/faas-provider/types"
)

const (
	OrchestrationIdentifier = "nomad"
	ProviderName            = "faas-nomad"
)

func MakeInfoHandler(version, sha string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providerInfo := types.ProviderInfo{
			Orchestration: OrchestrationIdentifier,
			Name:          ProviderName,
			Version: &types.VersionInfo{
				Release: version,
				SHA:     sha,
			},
		}

		jsonOut, marshalErr := json.Marshal(providerInfo)
		if marshalErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set(HeaderContentType, TypeApplicationJson)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonOut)
	}
}
