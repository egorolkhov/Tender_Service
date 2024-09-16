package storage

import (
	"avito.go/internal/config"
	"database/sql"
	"fmt"
	"strings"
)

func GetDatabaseDSN(config config.Config) string {
	if config.PostgresConn != "" {
		return config.PostgresConn
	} else if config.PostgresJDBCUrl != "" {
		result, _ := jdbcToGoConnectionString(config.PostgresJDBCUrl, config.PostgresUser, config.PostgresPass)
		return result
	} else {
		var sb strings.Builder
		sb.WriteString("postgres://")
		sb.WriteString(config.PostgresUser)
		sb.WriteString(":")
		sb.WriteString(config.PostgresPass)
		sb.WriteString("@")
		sb.WriteString(config.PostgresHost)
		sb.WriteString(":")
		sb.WriteString(config.PostgresPort)
		sb.WriteString("/")
		sb.WriteString(config.PostgresDB)
		sb.WriteString("?sslmode=disable")
		result := sb.String()
		return result
	}
}

func jdbcToGoConnectionString(jdbc string, username, password string) (string, error) {
	if !strings.HasPrefix(jdbc, "jdbc:postgresql://") {
		return "", fmt.Errorf("invalid JDBC string: must start with jdbc:postgresql://")
	}

	jdbc = strings.TrimPrefix(jdbc, "jdbc:postgresql://")

	goConnStr := fmt.Sprintf("postgres://%s:%s@%s", username, password, jdbc)

	if !strings.Contains(goConnStr, "?") {
		goConnStr += "?sslmode=disable"
	} else if !strings.Contains(goConnStr, "sslmode=") {
		goConnStr += "&sslmode=disable"
	}

	return goConnStr, nil
}

func CreateTable(db *sql.DB) error {
	TableEmployee := `
		CREATE TABLE IF NOT EXISTS employee (
    		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(50) UNIQUE NOT NULL,
			first_name VARCHAR(50),
			last_name VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
	`
	_, err := db.Exec(TableEmployee)
	if err != nil {
		return err
	}

	TableOrganization := `
		CREATE TABLE IF NOT EXISTS organization (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(100) NOT NULL,
			description TEXT,
			type organization_type,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
    `

	_, err = db.Exec(TableOrganization)
	if err != nil {
		return err
	}

	TableOrganizationResponsible := `
		CREATE TABLE IF NOT EXISTS organization_responsible (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
			user_id UUID REFERENCES employee(id) ON DELETE CASCADE
    )`

	_, err = db.Exec(TableOrganizationResponsible)
	if err != nil {
		return err
	}

	TenderOrganization := `
		CREATE TABLE IF NOT EXISTS tender (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(100) NOT NULL,
			description TEXT NOT NULL,
			service_type service_type NOT NULL,
			status tender_status NOT NULL,
			organization_id UUID NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
			version INT NOT NULL DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
`

	_, err = db.Exec(TenderOrganization)
	if err != nil {
		return err
	}

	TenderHistory := `
		CREATE TABLE IF NOT EXISTS tender_history (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tender_id UUID NOT NULL REFERENCES tender(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			description TEXT NOT NULL,
			service_type service_type NOT NULL,
			status tender_status NOT NULL,
			organization_id UUID NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
			version INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
`

	_, err = db.Exec(TenderHistory)
	if err != nil {
		return err
	}

	Bid := `
		CREATE TABLE IF NOT EXISTS bid (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(100) NOT NULL,
			description TEXT NOT NULL,
			status bid_status NOT NULL,
			tender_id UUID NOT NULL REFERENCES tender(id) ON DELETE CASCADE,
			author_type author_type NOT NULL,
			author_id UUID NOT NULL REFERENCES employee(id),
			version INT NOT NULL DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
`
	_, err = db.Exec(Bid)
	if err != nil {
		return err
	}

	BidHistory := `
		CREATE TABLE IF NOT EXISTS bid_history (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			bid_id UUID NOT NULL REFERENCES bid(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			description TEXT NOT NULL,
			status bid_status NOT NULL,
			tender_id UUID NOT NULL REFERENCES tender(id) ON DELETE CASCADE,
			author_type author_type NOT NULL,
			author_id UUID NOT NULL REFERENCES employee(id),
			version INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
`

	_, err = db.Exec(BidHistory)
	if err != nil {
		return err
	}

	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_bid_history_bid_id ON bid_history (bid_id);",
		"CREATE INDEX IF NOT EXISTS idx_bid_history_version ON bid_history (bid_id, version);",
	}

	for _, query := range createIndexes {
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	decisions := `
	CREATE TABLE IF NOT EXISTS decisions (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		bid_id UUID NOT NULL REFERENCES bid(id) ON DELETE CASCADE,
		decision VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    	created_by VARCHAR(100) NOT NULL REFERENCES employee(username)
	)
`

	_, err = db.Exec(decisions)
	if err != nil {
		return err
	}

	feedback := `
		CREATE TABLE IF NOT EXISTS bid_reviews (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			bid_id UUID NOT NULL REFERENCES bid(id) ON DELETE CASCADE,
			review TEXT NOT NULL,
			reviewer UUID NOT NULL REFERENCES employee(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			bid_author_id UUID NOT NULL REFERENCES employee(id)
		    )
`

	_, err = db.Exec(feedback)
	if err != nil {
		return err
	}

	return nil
}
