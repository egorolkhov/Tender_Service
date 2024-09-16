package bid

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

type ResponseDataList struct {
	Result []models.Bid
}

type RequestDataList struct {
	TenderID string `schema:"tenderId" validate:"required,max=100"`
	Username string `schema:"username" validate:"required"`
	Limit    int    `schema:"limit" validate:"gte=1,lte=100"` // Параметр limit (min 1, max 100)
	Offset   int    `schema:"offset" validate:"gte=0"`        // Параметр offset (минимум 0)
}

func (bc *BidController) BidsTenderList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID, ok := vars["tenderId"]
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

	req := RequestDataList{
		Limit:  5,
		Offset: 0,
	}

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

	var bids []models.Bid

	// Работа с бд, Список тендеров с возможностью фильтрации по типу услуг.
	//
	//Если фильтры не заданы, возвращаются все тендеры.

	bids, err = bc.Storage.GetTenderBids(r.Context(), req.TenderID, req.Username, req.Limit, req.Offset)
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
		case errors.Is(err, storage.ErrNoBid):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)

			response := ErrorResponse{Reason: "The bid does not exist."}
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
	//TODO: отсортировать по алфавиту по названию
	//TODO: добавить логику обработки данных, учесть лимит и сдвиг

	var resp ResponseDataList
	resp.Result = bids

	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
	w.WriteHeader(http.StatusOK)
}
