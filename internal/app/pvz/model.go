package pvz

import (
	"time"

	"github.com/google/uuid"

	"pvz_service/internal/app/product"
	"pvz_service/internal/app/reception"
)

type PVZ struct {
	ID               uuid.UUID `json:"id,omitempty"`
	RegistrationDate time.Time `json:"registrationDate,omitempty"`
	City             string    `json:"city"`
}

type PVZWithReceptions struct {
	PVZ        *PVZ                     `json:"pvz"`
	Receptions []*ReceptionWithProducts `json:"receptions"`
}

type ReceptionWithProducts struct {
	Reception *reception.Reception `json:"reception"`
	Products  []*product.Product   `json:"products"`
}

type CreatePVZRequest struct {
	City string `json:"city"`
}
