package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jsiebens/faas-nomad/pkg/services"
	ftypes "github.com/openfaas/faas-provider/types"
)

type SecretsResponse struct {
	StatusCode int
	Body       []byte
}

func MakeSecretHandler(secrets services.Secrets) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)

		var code int
		var response []byte
		var err error

		switch r.Method {
		case http.MethodGet:
			code, response, err = getSecrets(secrets)
			break
		case http.MethodPost:
			code, response, err = setSecret(true, secrets, body)
			break
		case http.MethodPut:
			code, response, err = setSecret(false, secrets, body)
			break
		case http.MethodDelete:
			code, response, err = deleteSecret(secrets, body)
			break
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err != nil {
			w.WriteHeader(code)
		} else {
			w.WriteHeader(code)
			w.Header().Set(HeaderContentType, TypeApplicationJson)
			w.Write(response)
		}
	}
}

func getSecrets(vc services.Secrets) (statusCode int, response []byte, err error) {
	secrets, err := vc.List()
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	resultsJson, err := json.Marshal(secrets)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, resultsJson, nil
}

func setSecret(create bool, vc services.Secrets, body []byte) (statusCode int, response []byte, err error) {
	var secret ftypes.Secret

	if err := json.Unmarshal(body, &secret); err != nil {
		return http.StatusBadRequest, nil, err
	}

	if err := vc.Set(secret.Name, secret.Value); err != nil {
		return http.StatusInternalServerError, nil, err
	}

	if create {
		return http.StatusCreated, nil, nil
	} else {
		return http.StatusOK, nil, nil
	}
}

func deleteSecret(vc services.Secrets, body []byte) (statusCode int, response []byte, err error) {
	var secret ftypes.Secret

	if err := json.Unmarshal(body, &secret); err != nil {
		return http.StatusBadRequest, nil, err
	}

	if err := vc.Delete(secret.Name); err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return http.StatusOK, nil, err
}
