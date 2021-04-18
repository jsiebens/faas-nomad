package handlers

import (
	"encoding/json"
	"net/http"
)

func MakeListNamespaceHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonOut, marshalErr := json.Marshal([]string{})
		if marshalErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set(HeaderContentType, TypeApplicationJson)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonOut)
	}
}
