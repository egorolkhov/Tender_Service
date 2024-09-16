package routes

import (
	"avito.go/internal/app"
	"avito.go/internal/middleware"
	"github.com/gorilla/mux"
)

func NewRouter(App app.App) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/ping", middleware.Middleware(App.CheckerController.CheckServer)).Methods("GET")

	router.HandleFunc("/api/tenders", middleware.Middleware(App.TenderController.TendersInfo)).Methods("GET")
	router.HandleFunc("/api/tenders/my", middleware.Middleware(App.TenderController.TendersMy)).Methods("GET")
	router.HandleFunc("/api/tenders/new", middleware.Middleware(App.TenderController.CreateTender)).Methods("POST")

	router.HandleFunc("/api/tenders/{tenderId}/status", middleware.Middleware(App.TenderController.TenderStatus)).Methods("GET")
	router.HandleFunc("/api/tenders/{tenderId}/status", middleware.Middleware(App.TenderController.TenderUpdateStatus)).Methods("PUT")
	router.HandleFunc("/api/tenders/{tenderId}/edit", middleware.Middleware(App.TenderController.TenderEdit)).Methods("PATCH")
	router.HandleFunc("/api/tenders/{tenderId}/rollback/{version}", middleware.Middleware(App.TenderController.RollbackTender)).Methods("PUT")

	router.HandleFunc("/api/bids/new", middleware.Middleware(App.BidController.CreateBid)).Methods("POST")
	router.HandleFunc("/api/bids/my", middleware.Middleware(App.BidController.BidsMy)).Methods("GET")
	router.HandleFunc("/api/bids/{tenderId}/list", middleware.Middleware(App.BidController.BidsTenderList)).Methods("GET")
	router.HandleFunc("/api/bids/{bidId}/status", middleware.Middleware(App.BidController.BidStatus)).Methods("GET")
	router.HandleFunc("/api/bids/{bidId}/status", middleware.Middleware(App.BidController.BidUpdateStatus)).Methods("PUT")
	router.HandleFunc("/api/bids/{bidId}/edit", middleware.Middleware(App.BidController.BidEdit)).Methods("PATCH")
	router.HandleFunc("/api/bids/{bidId}/rollback/{version}", middleware.Middleware(App.BidController.RollbackTender)).Methods("PUT")
	router.HandleFunc("/api/bids/{bidId}/submit_decision", middleware.Middleware(App.BidController.BidSubmitDecision)).Methods("PUT")

	router.HandleFunc("/api/bids/{bidId}/feedback", middleware.Middleware(App.BidController.BidFeedback)).Methods("PUT")
	router.HandleFunc("/api/bids/{tenderId}/reviews", middleware.Middleware(App.BidController.BidsReviews)).Methods("GET")

	return router
}
