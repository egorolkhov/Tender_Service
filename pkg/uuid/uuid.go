package uuid

import "github.com/google/uuid"

func GenerateCorrelationID() string {
	id := uuid.New()
	return id.String()
}
