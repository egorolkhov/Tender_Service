package bid

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
	Result models.Bid
}

type RequestDataCreate struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"required,max=500"`
	TenderID    string `json:"tenderId" validate:"required,max=100"`
	AuthorType  string `json:"authorType" validate:"required,oneof=Organization User"`
	AuthorId    string `json:"authorId" validate:"required"`
}

func (bc *BidController) CreateBid(w http.ResponseWriter, r *http.Request) {
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

	bid := models.Bid{
		ID:          ID.GenerateCorrelationID(),
		Name:        req.Name,
		Description: req.Description,
		Status:      "Created",
		TenderID:    req.TenderID,
		AuthorType:  req.AuthorType,
		AuthorID:    req.AuthorId,
		Version:     1,
		CreatedAt:   time.Now(),
	}

	// проверка на то, валиден ли юзер
	// взоимодействие с бд. Создаем новое предложение
	err = bc.Storage.Add(r.Context(), bid, req.AuthorId, serviceKey)
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

	result, err := json.Marshal(bid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
	w.WriteHeader(http.StatusOK)
}
