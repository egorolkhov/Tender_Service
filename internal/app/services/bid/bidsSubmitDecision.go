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

type ResponseDataSubmitDecision struct {
	Result models.Bid
}

type RequestDataSubmitDecision struct {
	BidID    string `schema:"bidId" validate:"required,max=100"`
	Decision string `schema:"decision" validate:"required,oneof=Approved Rejected"`
	Username string `schema:"username" validate:"required"`
}

func (bc *BidController) BidSubmitDecision(w http.ResponseWriter, r *http.Request) {
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

		response := ErrorResponse{Reason: "Only Put requests are supported"}
		json.NewEncoder(w).Encode(response)
		return
	}
	var req RequestDataSubmitDecision
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

	//обращение к базе данных, обновляем значение decision
	bid, err := bc.Storage.SubmitDecisionBid(r.Context(), req.BidID, req.Decision, req.Username)
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

	var resp ResponseDataSubmitDecision
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
