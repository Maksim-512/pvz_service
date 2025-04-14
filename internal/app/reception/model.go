package reception

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusInProgress = "in_progress"
	StatusClosed     = "close"
)

type Reception struct {
	ID       uuid.UUID `json:"id,omitempty"`
	DateTime time.Time `json:"dateTime,omitempty"`
	PvzID    uuid.UUID `json:"pvzID"`
	Status   string    `json:"status"`
}

type OpenReceptionRequest struct {
	PvzID uuid.UUID `json:"pvzID"`
}

type CloseReceptionRequest struct {
	PvzID uuid.UUID `json:"pvzID"`
}
