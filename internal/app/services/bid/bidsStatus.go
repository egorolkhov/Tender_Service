package bid

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
	Result models.Bid
}

type RequestDataStatus struct {
	BidID    string `schema:"bidId" validate:"required,max=100"`
	Username string `schema:"username" validate:"required,max=100"`
}

type RequestDataUpdateStatus struct {
	BidID    string `schema:"bidId" validate:"required,max=100"`
	Status   string `schema:"status" validate:"required,oneof=Created Published Closed"`
	Username string `schema:"username" validate:"required"`
}

func (bc *BidController) BidStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID, ok := vars["bidId"]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "The request parameters are incorrect."}
		json.NewEncoder(w).Encode(response)
		return
	}

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "Only GET requests are supported"}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req RequestDataStatus
	decoder := schema.NewDecoder()
	validate := validator.New()

	err := decoder.Decode(&req, r.URL.Query())
	req.BidID = bidID
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
	status, err := bc.Storage.GetStatus(r.Context(), req.BidID, req.Username, serviceKey)
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

func (bc *BidController) BidUpdateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID, ok := vars["bidId"]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "The request parameters are incorrect."}
		json.NewEncoder(w).Encode(response)
		return
	}

	if r.Method != http.MethodPut {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "Only PUT requests are supported"}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req RequestDataUpdateStatus
	decoder := schema.NewDecoder()
	validate := validator.New()

	err := decoder.Decode(&req, r.URL.Query())
	req.BidID = bidID
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

	bidInterface, err := bc.Storage.UpdateStatus(r.Context(), req.BidID, req.Status, req.Username, serviceKey)
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
	bid, ok := bidInterface.(models.Bid)
	if !ok {
		fmt.Println("Failed to convert interface to []models.Tender")
		return
	}

	var resp ResponseDataUpdateStatus
	resp.Result = bid
	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
	w.WriteHeader(http.StatusOK)
}
