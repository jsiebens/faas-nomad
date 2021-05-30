package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jsiebens/faas-nomad/pkg/types"
)

func MakeListNamespaceHandler(config *types.ProviderConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonOut, marshalErr := json.Marshal([]string{config.Scheduling.Namespace})
		if marshalErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set(HeaderContentType, TypeApplicationJson)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonOut)
	}
}
