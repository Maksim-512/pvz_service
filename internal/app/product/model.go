package product

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypeElectronics = "электроника"
	TypeClothing    = "одежда"
	TypeShoes       = "обувь"
)

type Product struct {
	ID          uuid.UUID `json:"id,omitempty"`
	Type        string    `json:"type"`
	ReceptionID uuid.UUID `json:"receptionID"`
	DateTime    time.Time `json:"dateTime,omitempty"`
}

type CreateProductRequest struct {
	Type  string    `json:"type"`
	PvzID uuid.UUID `json:"pvzID"`
}
