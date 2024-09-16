package tender

import (
	"avito.go/internal/models"
	"avito.go/internal/storage"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"net/http"
)

type ResponseDataUpdateStatus struct {
	Result models.Tender
}

type RequestDataStatus struct {
	TenderID string `schema:"tenderId" validate:"required,max=100"`
	Username string `schema:"username"`
}

type RequestDataUpdateStatus struct {
	TenderID string `schema:"tenderId" validate:"required,max=100"`
	Status   string `schema:"status" validate:"required,oneof=Created Published Closed"`
	Username string `schema:"username" validate:"required"`
}

func (tc *TenderController) TenderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID, ok := vars["tenderId"]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "The request parameters are incorrect."}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req RequestDataStatus
	decoder := schema.NewDecoder()
	validate := validator.New()

	err := decoder.Decode(&req, r.URL.Query())
	req.TenderID = tenderID
	errValidate := validate.Struct(req)
	if err != nil || errValidate != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "The request parameters are incorrect."}
		json.NewEncoder(w).Encode(response)
		return
	}

	// получаем актуальное значение статуса тендера
	// если пользователь не существует или некорректен - 401

	status, err := tc.Storage.GetStatus(r.Context(), req.TenderID, req.Username, serviceKey)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrRights):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)

			response := ErrorResponse{Reason: "Insufficient rights to perform the action."}
			json.NewEncoder(w).Encode(response)
			return
		case errors.Is(err, storage.ErrNoBid):
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

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(status))
	w.WriteHeader(http.StatusOK)
}

func (tc *TenderController) TenderUpdateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID, ok := vars["tenderId"]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "The request parameters are incorrect."}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req RequestDataUpdateStatus
	decoder := schema.NewDecoder()
	validate := validator.New()

	err := decoder.Decode(&req, r.URL.Query())
	req.TenderID = tenderID
	errValidate := validate.Struct(req)
	if err != nil || errValidate != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "The request parameters are incorrect."}
		json.NewEncoder(w).Encode(response)
		return
	}

	//Взаимодействие с бд, Обновляем статус тендера и возвращаем актуальное значение
	// если пользователь не существует или некорректен - 401
	// если недостаточно прав для выполнения действия - 403
	// если тендер не найден - 404

	tenderInterface, err := tc.Storage.UpdateStatus(r.Context(), req.TenderID, req.Status, req.Username, serviceKey)
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
	tender, ok := tenderInterface.(models.Tender)
	if !ok {
		fmt.Println("Failed to convert interface to []models.Tender")
		return
	}

	var resp ResponseDataUpdateStatus
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
