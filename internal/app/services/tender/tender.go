package tender

import (
	"avito.go/internal/models"
	"context"
)

const serviceKey = 2

type Storage interface {
	Add(ctx context.Context, entity interface{}, username string, key int) error
	GetMy(ctx context.Context, limit, offset int, username string, key int) (interface{}, error)
	GetStatus(ctx context.Context, Id, username string, key int) (string, error)
	UpdateStatus(ctx context.Context, Id, status, username string, key int) (interface{}, error)
	RollbackVersion(ctx context.Context, Id, version, username string, key int) (interface{}, error)

	GetTenders(ctx context.Context, limit, offset int, serviceType []string) ([]models.Tender, error)
	EditTender(ctx context.Context, tenderId, username, tenderName, description, serviceType, status string) (models.Tender, error)
}

type TenderController struct {
	Storage Storage
}

type ErrorResponse struct {
	Reason string `json:"reason"`
}
