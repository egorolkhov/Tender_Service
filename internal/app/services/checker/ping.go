package checker

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Reason string `json:"reason"`
}

type CheckerController struct {
}

func (cc *CheckerController) CheckServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := Response{Reason: "Only GET requests are supported"}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
