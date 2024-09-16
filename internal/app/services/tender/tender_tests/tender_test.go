package tender_test

import (
	"avito.go/internal/app/services/tender"
	"avito.go/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock для интерфейса Storage
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) RollbackVersion(ctx context.Context, Id, version, username string, key int) (interface{}, error) {
	tenderD := models.Tender{
		ID:   "1",
		Name: "Tender 1 version 2",
	}
	return tenderD, nil
}

func (m *MockStorage) EditTender(ctx context.Context, tenderId, username, tenderName, description, serviceType, status string) (models.Tender, error) {
	editedTender := models.Tender{
		ID:          "1",
		Name:        "Updated Tender",
		Description: "Updated Description",
		ServiceType: "Delivery",
	}
	return editedTender, nil
}

func (m *MockStorage) Add(ctx context.Context, entity interface{}, username string, key int) error {
	args := m.Called(ctx, entity, username, key)
	return args.Error(0)
}

func (m *MockStorage) GetMy(ctx context.Context, limit int, offset int, username string, key int) (interface{}, error) {
	args := m.Called(ctx, limit, offset, username, key)
	return args.Get(0), args.Error(1)
}

func (m *MockStorage) GetStatus(ctx context.Context, Id string, username string, key int) (string, error) {
	args := m.Called(ctx, Id, username, key)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) UpdateStatus(ctx context.Context, Id string, status string, username string, key int) (interface{}, error) {
	args := m.Called(ctx, Id, status, username, key)
	return args.Get(0), args.Error(1)
}

func (m *MockStorage) GetTenderBids(ctx context.Context, tenderId string, limit int, offset int, serviceType []string) ([]models.Bid, error) {
	args := m.Called(ctx, tenderId, limit, offset, serviceType)
	return args.Get(0).([]models.Bid), args.Error(1)
}

func (m *MockStorage) GetMyBids(ctx context.Context, limit int, offset int, username string) ([]models.Bid, error) {
	args := m.Called(ctx, limit, offset, username)
	return args.Get(0).([]models.Bid), args.Error(1)
}

func (m *MockStorage) SubmitDecisionBid(ctx context.Context, bidId string, decision string, username string) (models.Bid, error) {
	args := m.Called(ctx, bidId, decision, username)
	return args.Get(0).(models.Bid), args.Error(1)
}

func (m *MockStorage) GetBidStatus(ctx context.Context, bidId string, username string) (string, error) {
	args := m.Called(ctx, bidId, username)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) AddFeedbackBid(ctx context.Context, bidId string, bidFeedback string, username string) (models.Bid, error) {
	args := m.Called(ctx, bidId, bidFeedback, username)
	return args.Get(0).(models.Bid), args.Error(1)
}

func (m *MockStorage) GetFeedback(ctx context.Context, tenderId string, authorUsername string, requesterUsername string, limit string, offset string) ([]models.FeedBack, error) {
	args := m.Called(ctx, tenderId, authorUsername, requesterUsername, limit, offset)
	return args.Get(0).([]models.FeedBack), args.Error(1)
}

func (m *MockStorage) GetTenders(ctx context.Context, limit int, offset int, serviceType []string) ([]models.Tender, error) {
	args := m.Called(ctx, limit, offset, serviceType)
	return args.Get(0).([]models.Tender), args.Error(1)
}

func TestTendersInfo_Success(t *testing.T) {
	mockStorage := new(MockStorage)
	tc := tender.TenderController{Storage: mockStorage}

	req := httptest.NewRequest(http.MethodGet, "/api/tenders?limit=5&offset=0&service_type=Construction", nil)
	rr := httptest.NewRecorder()

	expectedTenders := []models.Tender{
		{ID: "1", Name: "Tender A", ServiceType: "Construction"},
		{ID: "2", Name: "Tender B", ServiceType: "Delivery"},
	}

	mockStorage.On("GetTenders", mock.Anything, 5, 0, []string{"Construction"}).Return(expectedTenders, nil)

	tc.TendersInfo(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response tender.ResponseDataInfo
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTenders, response.Result)
}

func TestCreateTender_Success(t *testing.T) {
	mockStorage := new(MockStorage)
	tc := tender.TenderController{Storage: mockStorage}

	reqBody := `{
		"name": "tender1",
		"description": "description",
		"serviceType": "Construction",
		"organizationId": "1",
		"creatorUsername": "user1"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/tenders/new", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	mockStorage.On("Add", mock.Anything, mock.AnythingOfType("models.Tender"), "user1", mock.Anything).Return(nil)

	tc.CreateTender(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response tender.ResponseDataCreate
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "tender1", response.Result.Name)
	assert.Equal(t, "description", response.Result.Description)
	assert.Equal(t, "Construction", response.Result.ServiceType)
	assert.Equal(t, "Created", response.Result.Status)
}

func TestTendersMy_Success(t *testing.T) {
	mockStorage := new(MockStorage)
	tc := tender.TenderController{Storage: mockStorage}

	tenders := []models.Tender{
		{ID: "1", Name: "Tender1", Description: "Description1", ServiceType: "Construction", Status: "Created"},
	}

	mockStorage.On("GetMy", mock.Anything, 5, 0, "user1", mock.Anything).Return(tenders, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/tenders/my?username=user1", nil)
	rr := httptest.NewRecorder()

	tc.TendersMy(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response tender.ResponseDataMy
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response.Result, 1)
	assert.Equal(t, "Tender1", response.Result[0].Name)
}

func TestTenderEdit_Success(t *testing.T) {
	mockStorage := &MockStorage{}
	tc := &tender.TenderController{
		Storage: mockStorage,
	}

	body := tender.RequestBodyEdit{
		TenderName:        "Updated Tender",
		TenderDescription: "Updated Description",
		TenderServiceType: "Delivery",
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	tenderD := models.Tender{
		ID: "1", Name: "Tender1", Description: "Description1", ServiceType: "Construction", Status: "Created",
	}

	editedTender, _ := mockStorage.EditTender(context.Background(), tenderD.ID, "1", tenderD.Name, tenderD.Description, tenderD.ServiceType, "status")

	req, err := http.NewRequest("PATCH", "/api/tenders/123", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatal(err)
	}

	vars := map[string]string{
		"tenderId": "123",
	}
	req = mux.SetURLVars(req, vars)

	q := req.URL.Query()
	q.Add("username", "user1")
	req.URL.RawQuery = q.Encode()

	rr := httptest.NewRecorder()

	tc.TenderEdit(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	var response tender.ResponseDataEdit
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Unable to parse response body: %v", err)
	}

	if response.Result.ID != editedTender.ID ||
		response.Result.Name != editedTender.Name ||
		response.Result.Description != editedTender.Description ||
		response.Result.ServiceType != editedTender.ServiceType {
		t.Errorf("Response does not match expected tender: got %+v, want %+v", response.Result, editedTender)
	}
}

func TestRollbackTender_Success(t *testing.T) {
	mockStorage := &MockStorage{}
	tc := &tender.TenderController{
		Storage: mockStorage,
	}

	RollbackTenderInterface, _ := mockStorage.RollbackVersion(context.Background(), "1", "creatred", "username", 2)

	RollbackTender, ok := RollbackTenderInterface.(models.Tender)
	if !ok {
		fmt.Println("Failed to convert interface to models.Tender")
	}

	req, err := http.NewRequest("PUT", "/api/tenders/1/rollback/2?username=user1", nil)
	if err != nil {
		t.Fatal(err)
	}

	vars := map[string]string{
		"tenderId": "1",
		"version":  "2",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()

	tc.RollbackTender(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	var response tender.ResponseDataRollback
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Unable to parse response body: %v", err)
	}

	if response.Result.ID != RollbackTender.ID || response.Result.Name != RollbackTender.Name {
		t.Errorf("Expected tender %+v, got %+v", RollbackTenderInterface, response.Result)
	}
}
