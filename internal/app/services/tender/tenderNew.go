package tender

import (
	"avito.go/internal/models"
	"avito.go/internal/storage"
	ID "avito.go/pkg/uuid"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"net/http"
	"time"
)

type ResponseDataCreate struct {
	Result models.Tender
}

type RequestDataCreate struct {
	Name        string `json:"name" validate:"required,max=100"`                                        // Полное название тендера, обязательное поле, максимум 100 символов
	Description string `json:"description" validate:"required,max=500"`                                 // Описание тендера, обязательное поле, максимум 500 символов
	ServiceType string `json:"serviceType" validate:"required,oneof=Construction Delivery Manufacture"` // Вид услуги, одно из: Construction, Delivery, Manufacture
	//Status          string `json:"status" validate:"required,oneof=Created Published Closed"`               // Статус тендера, одно из: Created, Published, Closed
	OrganizationID  string `json:"organizationId" validate:"required,max=100"` // Уникальный идентификатор организации, максимум 100 символов
	CreatorUsername string `json:"creatorUsername" validate:"required"`        // Уникальный slug пользователя
}

func (tc *TenderController) CreateTender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "Only POST requests are supported"}
		json.NewEncoder(w).Encode(response)
		return
	}

	var req RequestDataCreate
	validate := validator.New()

	err := json.NewDecoder(r.Body).Decode(&req)
	errValidate := validate.Struct(req)
	if err != nil || errValidate != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ErrorResponse{Reason: "The request parameters are incorrect."}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer r.Body.Close()

	var resp ResponseDataCreate
	tender := models.Tender{
		ID:             ID.GenerateCorrelationID(),
		Name:           req.Name,
		Description:    req.Description,
		ServiceType:    req.ServiceType,
		Status:         "Created",
		OrganizationID: req.OrganizationID,
		Version:        1,
		CreatedAt:      time.Now(),
	}

	// Создаем новый тендер
	err = tc.Storage.Add(r.Context(), tender, req.CreatorUsername, serviceKey)
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
