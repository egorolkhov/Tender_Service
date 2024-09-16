package tender

import (
	"avito.go/internal/models"
	"avito.go/internal/storage"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"net/http"
)

type ResponseDataEdit struct {
	Result models.Tender
}

type RequestParamsEdit struct {
	TenderID string `schema:"tenderId" validate:"required,max=100"`
	Username string `schema:"username" validate:"required"`
}

type RequestBodyEdit struct {
	TenderName        string `json:"name" validate:"max=100"`
	TenderDescription string `json:"description" validate:"max=500"`
	TenderServiceType string `json:"serviceType" validate:"oneof=Construction Delivery Manufacture"`
}

func (tc *TenderController) TenderEdit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderId, ok := vars["tenderId"]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "The request parameters are incorrect."}
		json.NewEncoder(w).Encode(response)
		return
	}

	if r.Method != http.MethodPatch {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "Only PATCH requests are supported"}
		json.NewEncoder(w).Encode(response)
		return
	}
	decoder := schema.NewDecoder()
	validate := validator.New()

	var params RequestParamsEdit
	err := decoder.Decode(&params, r.URL.Query())
	params.TenderID = tenderId
	errValidate := validate.Struct(params)

	if err != nil || errValidate != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "The request parameters are incorrect."}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req RequestBodyEdit
	err = json.NewDecoder(r.Body).Decode(&req)
	defer r.Body.Close()
	errValidate = validate.Struct(req)

	if err != nil || errValidate != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "The request body are incorrect."}
		json.NewEncoder(w).Encode(response)
		return
	}

	// взаимодействие с бд
	// изменение параметров существующего тендера.
	// пользователь не существует или некорректен. - 401
	tender, err := tc.Storage.EditTender(r.Context(), params.TenderID, params.Username, req.TenderName, req.TenderDescription, req.TenderServiceType, "")
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrRights):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)

			response := ErrorResponse{Reason: "Insufficient rights to perform the action."}
			json.NewEncoder(w).Encode(response)
			return
		case errors.Is(err, storage.ErrNoTender):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)

			response := ErrorResponse{Reason: "The tender does not exist."}
			json.NewEncoder(w).Encode(response)
			return
		case errors.Is(err, storage.ErrNoUser):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)

			response := ErrorResponse{Reason: "The user does not exist or is invalid."}
			json.NewEncoder(w).Encode(response)
			return
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	var resp ResponseDataEdit
	resp.Result = tender
	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
	w.WriteHeader(http.StatusOK)
}
