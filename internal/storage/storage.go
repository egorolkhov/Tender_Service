package storage

import (
	"avito.go/internal/models"
	"avito.go/pkg/uuid"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"math"
	"sort"
	"time"
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
	GetFeedback(ctx context.Context, tenderId, authorUsername, requesterUsername, limit, offset int) ([]models.FeedBack, error)
}

type DB struct {
	DB *sql.DB
}

func NewStorage(DatabaseDSN string) *DB {
	fmt.Println("STORAGE", DatabaseDSN)
	db, err := sql.Open("pgx", DatabaseDSN)
	if err != nil {
		log.Println(err)
	}
	err = CreateTable(db)
	if err != nil {
		log.Println(err)
	}
	return &DB{db}
}

func (db *DB) Add(ctx context.Context, entity interface{}, username string, key int) error {
	switch key {
	case 1:
		bid, ok := entity.(models.Bid)
		if !ok {
			return errors.New("invalid entity type")
		}
		exist, _ := GetUserByID(ctx, db, username)
		if !exist {
			return ErrNoUser
		}
		check, _ := isUserResponsibleForTender(ctx, db, GetUsernameByID(ctx, db, username), bid.TenderID)
		if check {
			return ErrRights
		}
		TenderExist, _ := GetTender(ctx, db, bid.TenderID)
		if !TenderExist {
			return ErrNoTender
		}

		query := squirrel.Insert("bid").
			Columns("id", "name", "description", "status", "tender_id", "author_type", "author_id", "version", "created_at").
			Values(bid.ID, bid.Name, bid.Description, bid.Status, bid.TenderID, bid.AuthorType, bid.AuthorID, bid.Version, bid.CreatedAt).
			PlaceholderFormat(squirrel.Dollar)

		sql, args, err := query.ToSql()
		if err != nil {
			return err
		}
		_, err = db.DB.ExecContext(ctx, sql, args...)
		if err != nil {
			return err
		}
		return nil
	case 2:
		tender, ok := entity.(models.Tender)
		if !ok {
			return errors.New("invalid entity type")
		}
		exist, _ := GetUser(ctx, db, username)
		if !exist {
			return ErrNoUser
		}
		check, _ := IsUserResponsibleForOrganization(ctx, db, username, tender.OrganizationID)
		if !check {
			return ErrRights
		}

		query := squirrel.Insert("tender").
			Columns("id", "name", "description", "service_type", "status", "organization_id", "version", "created_at").
			Values(tender.ID, tender.Name, tender.Description, tender.ServiceType, tender.Status, tender.OrganizationID, tender.Version, tender.CreatedAt).
			PlaceholderFormat(squirrel.Dollar)

		sql, args, err := query.ToSql()
		if err != nil {
			return err
		}
		_, err = db.DB.ExecContext(ctx, sql, args...)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (db *DB) GetMy(ctx context.Context, limit, offset int, username string, key int) (interface{}, error) {
	switch key {
	case 1:
		exist, _ := GetUser(ctx, db, username)
		if !exist && username != "" {
			return nil, ErrNoUser
		}

		var bids []models.Bid

		query := squirrel.Select("bid.id", "bid.name", "bid.description", "bid.status", "bid.tender_id", "bid.author_type", "bid.author_id", "bid.version", "bid.created_at", "bid.updated_at").
			From("bid").
			Join("employee e ON bid.author_id = e.id").
			Where(squirrel.Eq{"e.username": username}).
			Limit(uint64(limit)).
			Offset(uint64(offset)).
			PlaceholderFormat(squirrel.Dollar)

		sql, args, err := query.ToSql()
		if err != nil {
			return nil, err
		}
		rows, err := db.DB.QueryContext(ctx, sql, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var bid models.Bid
			if err = rows.Scan(
				&bid.ID,
				&bid.Name,
				&bid.Description,
				&bid.Status,
				&bid.TenderID,
				&bid.AuthorType,
				&bid.AuthorID,
				&bid.Version,
				&bid.CreatedAt,
				&bid.UpdatedAt); err != nil {
				return nil, err
			}
			bids = append(bids, bid)
		}
		defer rows.Close()
		sort.Slice(bids, func(i, j int) bool {
			return bids[i].Name < bids[j].Name
		})
		return bids, nil
	case 2:
		exist, _ := GetUser(ctx, db, username)
		if !exist && username != "" {
			return nil, ErrNoUser
		}

		var tenders []models.Tender

		query := squirrel.Select("tender.id", "tender.name", "tender.description", "tender.service_type", "tender.status", "tender.organization_id", "tender.version", "tender.created_at", "tender.updated_at").
			From("tender").
			Join(`organization_responsible "or" ON tender.organization_id = "or".organization_id`).
			Join("employee e ON \"or\".user_id = e.id").
			Where(squirrel.Eq{"e.username": username}).
			Limit(uint64(limit)).
			Offset(uint64(offset)).
			PlaceholderFormat(squirrel.Dollar)

		sql, args, err := query.ToSql()
		if err != nil {
			return nil, err
		}
		rows, err := db.DB.QueryContext(ctx, sql, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var tender models.Tender
			if err = rows.Scan(
				&tender.ID,
				&tender.Name,
				&tender.Description,
				&tender.ServiceType,
				&tender.Status,
				&tender.OrganizationID,
				&tender.Version,
				&tender.CreatedAt,
				&tender.UpdatedAt); err != nil {
				return nil, err
			}
			tenders = append(tenders, tender)
		}
		fmt.Println(tenders)
		defer rows.Close()
		sort.Slice(tenders, func(i, j int) bool {
			return tenders[i].Name < tenders[j].Name
		})
		return tenders, nil
	}
	return nil, nil
}

func (db *DB) GetStatus(ctx context.Context, Id, username string, key int) (string, error) {
	userExist, _ := GetUser(ctx, db, username)
	if !userExist {
		return "", ErrNoUser
	}
	var table string
	switch key {
	case 1:
		BidExist, _ := GetBid(ctx, db, Id)
		if !BidExist {
			return "", ErrNoBid
		}
		table = "bid"
	case 2:
		TenderExist, _ := GetTender(ctx, db, Id)
		if !TenderExist {
			return "", ErrNoTender
		}
		table = "tender"
	}
	query := squirrel.Select("status").
		From(table).
		Where(squirrel.Eq{"id": Id}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return "", err
	}

	var status string
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("error executing query: %w", err)
	}

	switch key {
	case 1:
		if status == "Created" || status == "Canceled" {
			check, err := isUserResponsibleForBid(ctx, db, username, Id)
			fmt.Println("RES", check, err)
			if check {
				return status, nil
			}
			return "", ErrRights
		}
	case 2:
		if status == "Created" || status == "Closed" {
			check, _ := isUserResponsibleForTender(ctx, db, username, Id)
			if check {
				return status, nil
			}
			return "", ErrRights
		}
	}
	return status, nil
}

func (db *DB) UpdateStatus(ctx context.Context, Id, status, username string, key int) (interface{}, error) {
	userExist, _ := GetUser(ctx, db, username)
	if !userExist {
		return "", ErrNoUser
	}
	switch key {
	case 1:
		BidExist, _ := GetBid(ctx, db, Id)
		if !BidExist {
			return "", ErrNoBid
		}
		check, _ := isUserResponsibleToUpdateBid(ctx, db, username, Id)
		if !check {
			return nil, ErrRights
		}

		queryHistory := squirrel.Insert("bid_history").
			Columns("id", "bid_id", "name", "description", "status", "tender_id", "author_type", "author_id", "version", "created_at", "updated_at").
			Select(
				squirrel.Select("uuid_generate_v4()", "id", "name", "description", "status", "tender_id", "author_type", "author_id", "version", "created_at", "updated_at").
					From("bid").
					Where(squirrel.Eq{"id": Id}),
			).
			PlaceholderFormat(squirrel.Dollar)

		sqlHistoryQuery, historyArgs, err := queryHistory.ToSql()
		if err != nil {
			return nil, err
		}

		_, err = db.DB.ExecContext(ctx, sqlHistoryQuery, historyArgs...)
		if err != nil {
			return nil, err
		}

		query := squirrel.Update("bid").
			Set("status", status).
			Set("version", squirrel.Expr("version + 1")).
			Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
			Where(squirrel.Eq{"id": Id}).
			PlaceholderFormat(squirrel.Dollar)

		sql, args, err := query.ToSql()
		if err != nil {
			return nil, err
		}
		_, err = db.DB.ExecContext(ctx, sql, args...)
		if err != nil {
			return nil, err
		}
		updatedBid := BidByID(ctx, db, Id)
		return updatedBid, nil
	case 2:
		TenderExist, _ := GetTender(ctx, db, Id)
		if !TenderExist {
			return "", ErrNoTender
		}
		check, _ := isUserResponsibleForTender(ctx, db, username, Id)
		if !check {
			return nil, ErrRights
		}

		queryHistory := squirrel.Insert("tender_history").
			Columns("id", "tender_id", "name", "description", "service_type", "status", "organization_id", "version", "created_at", "updated_at").
			Select(
				squirrel.Select("uuid_generate_v4()", "id", "name", "description", "service_type", "status", "organization_id", "version", "created_at", "updated_at").
					From("tender").
					Where(squirrel.Eq{"id": Id}),
			).
			PlaceholderFormat(squirrel.Dollar)

		sqlHistoryQuery, historyArgs, err := queryHistory.ToSql()
		if err != nil {
			return nil, err
		}

		_, err = db.DB.ExecContext(ctx, sqlHistoryQuery, historyArgs...)
		if err != nil {
			return nil, err
		}

		query := squirrel.Update("tender").
			Set("status", status).
			Set("version", squirrel.Expr("version + 1")).
			Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
			Where(squirrel.Eq{"id": Id}).
			PlaceholderFormat(squirrel.Dollar)

		sql, args, err := query.ToSql()
		if err != nil {
			return nil, err
		}

		_, err = db.DB.ExecContext(ctx, sql, args...)
		if err != nil {
			return nil, err
		}
		updatedTender := TenderByID(ctx, db, Id)
		return updatedTender, nil
	}
	return nil, nil
}

func (db *DB) RollbackVersion(ctx context.Context, Id, version, username string, key int) (interface{}, error) {
	userExist, _ := GetUser(ctx, db, username)
	if !userExist {
		return "", ErrNoUser
	}
	switch key {
	case 1:
		BidExist, _ := GetBid(ctx, db, Id)
		if !BidExist {
			return "", ErrNoBid
		}
		VersionExist := GetVersion(ctx, db, Id, version, 1)
		if !VersionExist {
			return nil, ErrNoVersion
		}
		check, _ := isUserResponsibleToUpdateBid(ctx, db, username, Id)
		if !check {
			return nil, ErrRights
		}

		var bid models.Bid

		query := squirrel.Select("id", "name", "description", "status", "tender_id", "author_type", "author_id", "version", "created_at", "updated_at").
			From("bid_history").
			Where(squirrel.Eq{"bid_id": Id, "version": version}).
			PlaceholderFormat(squirrel.Dollar)

		sql, args, err := query.ToSql()
		if err != nil {
			return nil, nil
		}

		err = db.DB.QueryRowContext(ctx, sql, args...).Scan(
			&bid.ID,
			&bid.Name,
			&bid.Description,
			&bid.Status,
			&bid.TenderID,
			&bid.AuthorType,
			&bid.AuthorID,
			&bid.Version,
			&bid.CreatedAt,
			&bid.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		updatedBid, _ := db.EditBid(ctx, Id, username, bid.Name, bid.Description, bid.Status)
		return updatedBid, nil
	case 2:
		TenderExist, err := GetTender(ctx, db, Id)
		if !TenderExist {
			fmt.Println(err)
			return "", ErrNoTender
		}
		VersionExist := GetVersion(ctx, db, Id, version, 2)
		if !VersionExist {
			return nil, ErrNoVersion
		}
		check, _ := isUserResponsibleForTender(ctx, db, username, Id)
		if !check {
			return nil, ErrRights
		}
		var tender models.Tender

		query := squirrel.Select("id", "name", "description", "service_type", "status", "organization_id", "version", "created_at", "updated_at").
			From("tender_history").
			Where(squirrel.Eq{"tender_id": Id, "version": version}).
			PlaceholderFormat(squirrel.Dollar)

		sql, args, err := query.ToSql()
		if err != nil {
			return nil, err
		}

		err = db.DB.QueryRowContext(ctx, sql, args...).Scan(
			&tender.ID,
			&tender.Name,
			&tender.Description,
			&tender.ServiceType,
			&tender.Status,
			&tender.OrganizationID,
			&tender.Version,
			&tender.CreatedAt,
			&tender.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		updatedTender, _ := db.EditTender(ctx, Id, username, tender.Name, tender.Description, tender.ServiceType, tender.Status)

		return updatedTender, nil
	}
	return nil, nil
}

func (db *DB) GetTenders(ctx context.Context, limit, offset int, serviceType []string) ([]models.Tender, error) {
	var tenders []models.Tender

	query := squirrel.Select("id", "name", "description", "service_type", "status", "organization_id", "version", "created_at", "updated_at").
		From("tender").
		Where(squirrel.Eq{"status": "Published"}).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar)

	if len(serviceType) > 0 {
		query = query.Where(squirrel.Eq{"service_type": serviceType})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.DB.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var tender models.Tender
		if err = rows.Scan(
			&tender.ID,
			&tender.Name,
			&tender.Description,
			&tender.ServiceType,
			&tender.Status,
			&tender.OrganizationID,
			&tender.Version,
			&tender.CreatedAt,
			&tender.UpdatedAt); err != nil {
			return nil, err
		}
		tenders = append(tenders, tender)
	}
	defer rows.Close()
	sort.Slice(tenders, func(i, j int) bool {
		return tenders[i].Name < tenders[j].Name
	})
	return tenders, nil
}

func (db *DB) EditTender(ctx context.Context, tenderId, username, tenderName, description, serviceType, status string) (models.Tender, error) {
	userExist, _ := GetUser(ctx, db, username)
	if !userExist {
		return models.Tender{}, ErrNoUser
	}
	TenderExist, _ := GetTender(ctx, db, tenderId)
	if !TenderExist {
		return models.Tender{}, ErrNoTender
	}
	check, _ := isUserResponsibleForTender(ctx, db, username, tenderId)
	if !check {
		return models.Tender{}, ErrRights
	}

	queryHistory := squirrel.Insert("tender_history").
		Columns("id", "tender_id", "name", "description", "service_type", "status", "organization_id", "version", "created_at", "updated_at").
		Select(
			squirrel.Select("uuid_generate_v4()", "id", "name", "description", "service_type", "status", "organization_id", "version", "created_at", "updated_at").
				From("tender").
				Where(squirrel.Eq{"id": tenderId}),
		).
		PlaceholderFormat(squirrel.Dollar)

	sqlHistoryQuery, historyArgs, err := queryHistory.ToSql()
	if err != nil {
		return models.Tender{}, err
	}

	_, err = db.DB.ExecContext(ctx, sqlHistoryQuery, historyArgs...)
	if err != nil {
		return models.Tender{}, err
	}
	query := squirrel.Update("tender").
		Set("version", squirrel.Expr("version + 1")).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Where(squirrel.Eq{"id": tenderId}).
		PlaceholderFormat(squirrel.Dollar)

	if status != "" {
		query = query.Set("status", status)
	}
	if tenderName != "" {
		query = query.Set("name", tenderName)
	}
	if description != "" {
		query = query.Set("description", description)
	}
	if serviceType != "" {
		query = query.Set("service_type", serviceType)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return models.Tender{}, err
	}

	_, err = db.DB.ExecContext(ctx, sql, args...)
	if err != nil {
		return models.Tender{}, err
	}
	updatedTender := TenderByID(ctx, db, tenderId)
	return updatedTender, nil
}

func (db *DB) EditBid(ctx context.Context, bidId, username, bidName, description, status string) (models.Bid, error) {
	userExist, _ := GetUser(ctx, db, username)
	if !userExist {
		return models.Bid{}, ErrNoUser
	}
	BidExist, _ := GetBid(ctx, db, bidId)
	if !BidExist {
		return models.Bid{}, ErrNoBid
	}
	check, err := isUserResponsibleToUpdateBid(ctx, db, username, bidId)
	if !check {
		return models.Bid{}, ErrRights
	}

	queryHistory := squirrel.Insert("bid_history").
		Columns("id", "bid_id", "name", "description", "status", "tender_id", "author_type", "author_id", "version", "created_at", "updated_at").
		Select(
			squirrel.Select("uuid_generate_v4()", "id", "name", "description", "status", "tender_id", "author_type", "author_id", "version", "created_at", "updated_at").
				From("bid").
				Where(squirrel.Eq{"id": bidId}),
		).
		PlaceholderFormat(squirrel.Dollar)
	sqlHistoryQuery, historyArgs, err := queryHistory.ToSql()
	if err != nil {
		return models.Bid{}, err
	}
	_, err = db.DB.ExecContext(ctx, sqlHistoryQuery, historyArgs...)
	if err != nil {
		return models.Bid{}, err
	}

	query := squirrel.Update("bid").
		Set("version", squirrel.Expr("version + 1")).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Where(squirrel.Eq{"id": bidId}).
		PlaceholderFormat(squirrel.Dollar)

	if status != "" {
		query = query.Set("status", status)
	}
	if bidName != "" {
		query = query.Set("name", bidName)
	}
	if description != "" {
		query = query.Set("description", description)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return models.Bid{}, err
	}
	_, err = db.DB.ExecContext(ctx, sql, args...)
	if err != nil {
		return models.Bid{}, err
	}
	updatedBid := BidByID(ctx, db, bidId)
	return updatedBid, nil
}

func (db *DB) SubmitDecisionBid(ctx context.Context, bidId string, decision string, username string) (models.Bid, error) { // TODO: необходима еще таблица вида bidID - decision
	userExist, _ := GetUser(ctx, db, username)
	if !userExist {
		return models.Bid{}, ErrNoUser
	}
	BidExist, _ := GetBid(ctx, db, bidId)
	if !BidExist {
		return models.Bid{}, ErrNoBid
	}
	bid := BidByID(ctx, db, bidId)
	tenderId := bid.TenderID
	check, _ := isUserResponsibleForTender(ctx, db, username, tenderId)
	if !check {
		return models.Bid{}, ErrRights
	}

	query := squirrel.Insert("decisions").
		Columns("id", "bid_id", "decision", "created_at", "created_by").
		Values(uuid.GenerateCorrelationID(), bidId, decision, time.Now(), username).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return models.Bid{}, err
	}

	_, err = db.DB.ExecContext(ctx, sql, args...)
	if err != nil {
		return models.Bid{}, fmt.Errorf("error executing query: %w", err)
	}

	if decision == "Approved" {
		_, _ = db.UpdateStatus(ctx, tenderId, "Closed", username, 2)
	}
	return bid, nil
}

func (db *DB) GetTenderBids(ctx context.Context, tenderID, username string, limit, offset int) ([]models.Bid, error) {
	userExist, _ := GetUser(ctx, db, username)
	if !userExist {
		return []models.Bid{}, ErrNoUser
	}
	BidExist, _ := GetTender(ctx, db, tenderID)
	if !BidExist {
		return []models.Bid{}, ErrNoTender
	}
	status, _ := db.GetStatus(ctx, tenderID, username, 2)
	fmt.Println("STATUS", status)
	if status != "Published" {
		check, _ := isUserResponsibleForTender(ctx, db, username, tenderID)
		if !check {
			return []models.Bid{}, ErrRights
		}
	}

	query := squirrel.Select("id", "name", "description", "status", "tender_id", "author_type", "author_id", "version", "created_at", "updated_at").
		From("bid").
		Where(squirrel.Eq{"tender_id": tenderID}).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	rows, err := db.DB.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var bids []models.Bid
	for rows.Next() {
		var bid models.Bid
		if err = rows.Scan(
			&bid.ID,
			&bid.Name,
			&bid.Description,
			&bid.Status,
			&bid.TenderID,
			&bid.AuthorType,
			&bid.AuthorID,
			&bid.Version,
			&bid.CreatedAt,
			&bid.UpdatedAt); err != nil {
			return nil, err
		}
		check, _ := isUserResponsibleForBid(ctx, db, username, bid.ID)
		if check {
			bids = append(bids, bid)
		}
	}
	if len(bids) == 0 {
		return bids, ErrNoBid
	}
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].Name < bids[j].Name
	})
	return bids, nil
}

func (db *DB) AddFeedbackBid(ctx context.Context, bidId string, bidFeedback string, username string) (models.Bid, error) {
	userExist, _ := GetUser(ctx, db, username)
	if !userExist {
		return models.Bid{}, ErrNoUser
	}
	BidExist, _ := GetBid(ctx, db, bidId)
	if !BidExist {
		return models.Bid{}, ErrNoBid
	}

	bid := BidByID(ctx, db, bidId)
	check, _ := isUserResponsibleForTender(ctx, db, username, bid.TenderID)
	if !check {
		return models.Bid{}, ErrRights
	}

	var reviewer string
	sql := "SELECT id FROM employee WHERE username = $1"
	_ = db.DB.QueryRowContext(ctx, sql, username).Scan(&reviewer)

	fmt.Println(reviewer)

	query := squirrel.Insert("bid_reviews").
		Columns("id", "bid_id", "review", "reviewer", "created_at", "bid_author_id").
		Values(uuid.GenerateCorrelationID(), bidId, bidFeedback, reviewer, time.Now(), bid.AuthorID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return models.Bid{}, err
	}

	_, err = db.DB.ExecContext(ctx, sql, args...)
	if err != nil {
		return models.Bid{}, fmt.Errorf("error executing query: %w", err)
	}

	return bid, nil
}

func (db *DB) GetFeedback(ctx context.Context, tenderId, authorUsername, requesterUsername string, limit, offset int) ([]models.FeedBack, error) {
	authorExist, _ := GetUser(ctx, db, authorUsername)
	if !authorExist {
		return nil, ErrNoUser
	}
	requesterExist, _ := GetUser(ctx, db, requesterUsername)
	if !requesterExist {
		return nil, ErrNoUser
	}
	BidExist, _ := GetTender(ctx, db, tenderId)
	if !BidExist {
		return nil, ErrNoTender
	}
	check, _ := isUserResponsibleForTender(ctx, db, authorUsername, tenderId)
	if !check {
		return nil, ErrRights
	}

	requesterBids, _ := db.GetMy(ctx, math.MaxInt32, 0, requesterUsername, 1)
	bids, _ := requesterBids.([]models.Bid)

	ok := false
	for _, bid := range bids {
		if bid.TenderID == tenderId {
			ok = true
		}
	}
	if !ok {
		return nil, ErrNoReviews
	}

	var reviews []models.FeedBack

	query := squirrel.Select("br.id", "br.review", "br.created_at").
		From("bid_reviews br").
		Join("employee e ON br.bid_author_id = e.id").
		Where(squirrel.Eq{"e.username": requesterUsername}).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %v", err)
	}

	rows, err := db.DB.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var review models.FeedBack
		if err = rows.Scan(&review.ID, &review.Description, &review.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}

func IsUserResponsibleForOrganization(ctx context.Context, db *DB, userName, organizationID string) (bool, error) {
	query := squirrel.Select("id").
		From("employee").
		Where(squirrel.Eq{"username": userName}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var userID string
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&userID)
	if err != nil {
		return false, err
	}

	queryOrganization := squirrel.Select("user_id").
		From("organization_responsible").
		Where(squirrel.Eq{"organization_id": organizationID, "user_id": userID}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err = queryOrganization.ToSql()
	if err != nil {
		return false, err
	}

	var realID string
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&realID)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	return realID == userID, nil
}

func isUserResponsibleForTender(ctx context.Context, db *DB, username, tenderID string) (bool, error) {
	var TenderOrganizationID string
	query := squirrel.Select("organization_id").
		From("tender").
		Where(squirrel.Eq{"id": tenderID}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&TenderOrganizationID)
	if err != nil {
		return false, fmt.Errorf("error executing query: %w", err)
	}

	var UserOrganizationID string
	query = squirrel.Select(`"or".organization_id`).
		From(`organization_responsible "or"`).
		Join("employee e ON \"or\".user_id = e.id").
		Where(squirrel.Eq{"e.username": username}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err = query.ToSql()
	if err != nil {
		return false, err
	}

	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&UserOrganizationID)
	if err != nil {
		return false, fmt.Errorf("error executing query: %w", err)
	}

	return UserOrganizationID == TenderOrganizationID, nil
}

func isUserResponsibleToUpdateBid(ctx context.Context, db *DB, username, bidID string) (bool, error) {
	query := squirrel.Select("1").
		From("bid").
		Join(`organization_responsible "or_author" ON bid.author_id = "or_author".user_id`).
		Join(`organization_responsible "or_user" ON "or_user".organization_id = "or_author".organization_id`).
		Join(`employee e ON "or_user".user_id = e.id`).
		Where(squirrel.Eq{"e.username": username}).
		Where(squirrel.Eq{"bid.id": bidID}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var author bool
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&author)
	if err != nil {
		return false, fmt.Errorf("error executing query: %w", err)
	}

	return author, nil
}

func isUserResponsibleForBid(ctx context.Context, db *DB, username, bidID string) (bool, error) {
	query := squirrel.Select("bid.tender_id", "employee.username").
		From("bid").
		Join("employee ON bid.author_id = employee.id").
		Where(squirrel.Eq{"bid.id": bidID}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var tenderID, author string
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&tenderID, &author)
	if err != nil {
		return false, fmt.Errorf("error executing query: %w", err)
	}

	if author == username {
		return true, nil
	}
	check, _ := isUserResponsibleToUpdateBid(ctx, db, username, bidID)
	if check {
		return true, nil
	}
	ok, _ := isUserResponsibleForTender(ctx, db, username, tenderID)
	if ok {
		return true, nil
	}
	return false, nil
}

func GetBid(ctx context.Context, db *DB, bidID string) (bool, error) {
	query := squirrel.Select("COUNT(*)").
		From("bid").
		Where(squirrel.Eq{"id": bidID}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var count int
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error executing query: %w", err)
	}

	return count > 0, nil
}

func GetTender(ctx context.Context, db *DB, tenderID string) (bool, error) {
	query := squirrel.Select("COUNT(*)").
		From("tender").
		Where(squirrel.Eq{"id": tenderID}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var count int
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error executing query: %w", err)
	}

	return count > 0, nil
}

func GetUser(ctx context.Context, db *DB, userName string) (bool, error) {
	query := squirrel.Select("COUNT(*)").
		From("employee").
		Where(squirrel.Eq{"username": userName}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var count int
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error executing query: %w", err)
	}

	return count > 0, nil
}

func GetVersion(ctx context.Context, db *DB, id, version string, key int) bool {
	var tableName, temp string
	switch key {
	case 1:
		tableName = "bid_history"
		temp = "bid_id"
	case 2:
		tableName = "tender_history"
		temp = "tender_id"
	}
	query := squirrel.Select("COUNT(*)").
		From(tableName).
		Where(squirrel.Eq{temp: id}).
		Where(squirrel.Eq{"version": version}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return false
	}

	var count int
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

func GetUserByID(ctx context.Context, db *DB, id string) (bool, error) {
	query := squirrel.Select("COUNT(*)").
		From("employee").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var count int
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error executing query: %w", err)
	}

	return count > 0, nil
}

func TenderByID(ctx context.Context, db *DB, tenderID string) models.Tender {
	var tender models.Tender

	query := squirrel.Select("id", "name", "description", "service_type", "status", "organization_id", "version", "created_at", "updated_at").
		From("tender").
		Where(squirrel.Eq{"id": tenderID}).
		OrderBy("version DESC").
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return tender
	}

	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(
		&tender.ID,
		&tender.Name,
		&tender.Description,
		&tender.ServiceType,
		&tender.Status,
		&tender.OrganizationID,
		&tender.Version,
		&tender.CreatedAt,
		&tender.UpdatedAt,
	)
	if err != nil {
		return tender
	}
	return tender
}

func BidByID(ctx context.Context, db *DB, bidID string) models.Bid {
	var bid models.Bid
	query := squirrel.Select("id", "name", "description", "status", "tender_id", "author_type", "author_id", "version", "created_at", "updated_at").
		From("bid").
		Where(squirrel.Eq{"id": bidID}).
		OrderBy("version DESC").
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return bid
	}

	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(
		&bid.ID,
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.TenderID,
		&bid.AuthorType,
		&bid.AuthorID,
		&bid.Version,
		&bid.CreatedAt,
		&bid.UpdatedAt,
	)
	if err != nil {
		return bid
	}
	return bid
}

func GetUsernameByID(ctx context.Context, db *DB, id string) string {
	query := squirrel.Select("username").
		From("employee").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return ""
	}

	var username string
	err = db.DB.QueryRowContext(ctx, sql, args...).Scan(&username)
	if err != nil {
		return ""
	}

	return username
}
