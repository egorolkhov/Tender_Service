package models

import "time"

type Tender struct {
	ID             string    `json:"id" validate:"required,max=100"`                                 // Уникальный идентификатор тендера
	Name           string    `json:"name" validate:"required,max=100"`                               // Полное название тендера
	Description    string    `json:"description" validate:"required,max=500"`                        // Описание тендера
	ServiceType    string    `json:"serviceType" validate:"oneof=Construction Delivery Manufacture"` // Вид услуги
	Status         string    `json:"status" validate:"required,oneof=Created InProgress Completed"`  // Статус тендера
	OrganizationID string    `json:"organizationId" validate:"max=100"`                              // Уникальный идентификатор организации
	Version        int       `json:"version" validate:"required,min=1"`                              // Версия тендера (номер версии после правок)
	CreatedAt      time.Time `json:"createdAt" validate:"required"`                                  // Дата создания тендера в формате RFC3339
	UpdatedAt      time.Time
}
