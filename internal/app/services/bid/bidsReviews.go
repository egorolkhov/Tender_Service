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

type ResponseDataReviews struct {
	Result []models.FeedBack
}

type RequestDataReviews struct {
	TenderID          string `schema:"tenderId" validate:"required,max=100"`
	AuthorUsername    string `schema:"authorUsername" validate:"required,max=100"`
	RequesterUsername string `schema:"requesterUsername" validate:"required,max=100"`
	Limit             int    `schema:"limit" validate:"gte=1,lte=100"`
	Offset            int    `schema:"offset" validate:"gte=0"`
}

func (bc *BidController) BidsReviews(w http.ResponseWriter, r *http.Request) {
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

	req := RequestDataReviews{
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

	feedbacks, err := bc.Storage.GetFeedback(r.Context(), req.TenderID, req.AuthorUsername, req.RequesterUsername, req.Limit, req.Offset)
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

			response := ErrorResponse{Reason: "The bid does not exist."}
			json.NewEncoder(w).Encode(response)
			return
		case errors.Is(err, storage.ErrNoUser):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)

			response := ErrorResponse{Reason: "The user does not exist or is invalid."}
			json.NewEncoder(w).Encode(response)
			return
		case errors.Is(err, storage.ErrNoTender):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)

			response := ErrorResponse{Reason: "The tender does not exist."}
			json.NewEncoder(w).Encode(response)
			return
		case errors.Is(err, storage.ErrNoReviews):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)

			response := ErrorResponse{Reason: "Reviews does not exist."}
			json.NewEncoder(w).Encode(response)
			return
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	var resp ResponseDataReviews
	resp.Result = feedbacks
	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
	w.WriteHeader(http.StatusOK)
}
