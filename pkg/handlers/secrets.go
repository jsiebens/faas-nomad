package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-hclog"
	"github.com/jsiebens/faas-nomad/pkg/services"
	ftypes "github.com/openfaas/faas-provider/types"
)

type SecretsResponse struct {
	StatusCode int
	Body       []byte
}

func MakeSecretHandler(secrets services.Secrets, logger hclog.Logger) http.HandlerFunc {
	log := logger.Named("secrets")

	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)

		switch r.Method {
		case http.MethodGet:
			getSecrets(secrets, w, log)
			return
		case http.MethodPost:
			setSecret(true, secrets, body, w, log)
			return
		case http.MethodPut:
			setSecret(false, secrets, body, w, log)
			return
		case http.MethodDelete:
			deleteSecret(secrets, body, w, log)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func getSecrets(vc services.Secrets, w http.ResponseWriter, log hclog.Logger) {
	secrets, err := vc.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		log.Error("Error listing secrets", "error", err.Error())
	}

	resultsJson, err := json.Marshal(secrets)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		log.Error("Error listing secrets", "error", err.Error())
	}

	writeJsonResponse(w, http.StatusOK, resultsJson)

	log.Trace("Secrets listed successfully")
}

func setSecret(create bool, vc services.Secrets, body []byte, w http.ResponseWriter, log hclog.Logger) {
	var secret ftypes.Secret

	if err := json.Unmarshal(body, &secret); err != nil {
		writeError(w, http.StatusBadRequest, err)
		log.Error("Error creating/updating secret", "error", err.Error())
	}

	if err := vc.Set(secret.Name, secret.Value); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		log.Error("Error creating/updating secret", "secret", secret.Name, "error", err.Error())
	}

	if create {
		writeJsonResponse(w, http.StatusCreated, nil)
		log.Debug("Secret created successfully", "secret", secret.Name)
	} else {
		writeJsonResponse(w, http.StatusOK, nil)
		log.Debug("Secret updated successfully", "secret", secret.Name)
	}
}

func deleteSecret(vc services.Secrets, body []byte, w http.ResponseWriter, log hclog.Logger) {
	var secret ftypes.Secret

	if err := json.Unmarshal(body, &secret); err != nil {
		writeError(w, http.StatusBadRequest, err)
		log.Error("Error deleting secret", "error", err.Error())
	}

	if err := vc.Delete(secret.Name); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		log.Error("Error deleting secret", "error", err.Error())
	}

	writeJsonResponse(w, http.StatusOK, nil)
	log.Debug("Secret deleted successfully", "secret", secret.Name)
}
