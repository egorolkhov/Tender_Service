package models

import "time"

type Bid struct {
	ID          string    `json:"id" validate:"required,max=100"`                                                // Уникальный идентификатор предложения
	Name        string    `json:"name" validate:"required,max=100"`                                              // Полное название предложения
	Description string    `json:"description" validate:"max=500"`                                                // Описание предложения
	Status      string    `json:"status" validate:"required,oneof=Created Published Canceled Approved Rejected"` // Статус предложения
	TenderID    string    `json:"tenderId" validate:"max=100"`                                                   // Уникальный идентификатор тендера
	AuthorType  string    `json:"authorType" validate:"required,oneof=Organization User"`                        // Тип автора
	AuthorID    string    `json:"authorId" validate:"required,max=100"`                                          // Уникальный идентификатор автора предложения
	Version     int32     `json:"version" validate:"required,min=1"`                                             // Номер версии после правок
	CreatedAt   time.Time `json:"createdAt" validate:"required"`                                                 // Серверная дата и время создания предложения
	UpdatedAt   time.Time
}
