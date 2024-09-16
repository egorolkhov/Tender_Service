package app

import (
	"avito.go/internal/app/services/bid"
	"avito.go/internal/app/services/checker"
	"avito.go/internal/app/services/tender"
	"avito.go/internal/models"
	"context"
)

type Storage interface {
	Add(ctx context.Context, entity interface{}, username string, key int) error
	GetMy(ctx context.Context, limit, offset int, username string, key int) (interface{}, error)
	GetStatus(ctx context.Context, Id, username string, key int) (string, error)
	UpdateStatus(ctx context.Context, Id, status, username string, key int) (interface{}, error)
	RollbackVersion(ctx context.Context, Id, version, username string, key int) (interface{}, error)

	GetTenders(ctx context.Context, limit, offset int, serviceType []string) ([]models.Tender, error)
	EditTender(ctx context.Context, tenderId, username, tenderName, description, serviceType, status string) (models.Tender, error)

	GetTenderBids(ctx context.Context, tenderId, username string, limit, offset int) ([]models.Bid, error)
	SubmitDecisionBid(ctx context.Context, bidId string, decision string, username string) (models.Bid, error) // Отправить решение по биду
	EditBid(ctx context.Context, bidId, username, bidName, description, status string) (models.Bid, error)

	AddFeedbackBid(ctx context.Context, bidId string, bidFeedback string, username string) (models.Bid, error) // отправить отзыв по предложению.
	GetFeedback(ctx context.Context, tenderId, authorUsername, requesterUsername string, limit, offset int) ([]models.FeedBack, error)
}

type App struct {
	bid.BidController
	tender.TenderController
	checker.CheckerController
}

func NewApp(storage Storage) *App {
	bid := bid.BidController{Storage: storage}
	tender := tender.TenderController{Storage: storage}
	checker := checker.CheckerController{}

	return &App{BidController: bid, TenderController: tender, CheckerController: checker}
}
