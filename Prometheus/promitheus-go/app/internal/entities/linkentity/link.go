package linkentity

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	LinkID     uuid.UUID
	OriginLink string
	ResultLink string
	LinkAt     time.Time
	Rank       int
}
