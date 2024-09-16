package bid_test

import (
	"avito.go/internal/models"
	"context"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetMy(ctx context.Context, limit, offset int, username string, key int) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockStorage) GetStatus(ctx context.Context, Id, username string, key int) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockStorage) UpdateStatus(ctx context.Context, Id, status, username string, key int) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockStorage) RollbackVersion(ctx context.Context, Id, version, username string, key int) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockStorage) GetTenderBids(ctx context.Context, tenderId, username string, limit, offset int) ([]models.Bid, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockStorage) SubmitDecisionBid(ctx context.Context, bidId string, decision string, username string) (models.Bid, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockStorage) EditBid(ctx context.Context, bidId, username, bidName, description, status string) (models.Bid, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockStorage) AddFeedbackBid(ctx context.Context, bidId string, bidFeedback string, username string) (models.Bid, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockStorage) GetFeedback(ctx context.Context, tenderId, authorUsername, requesterUsername string, limit, offset int) ([]models.FeedBack, error) {
	//TODO implement me
	panic("implement me")
}
