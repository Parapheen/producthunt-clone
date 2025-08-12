package product

import (
	"time"

	"github.com/google/uuid"
)

type InviteStatus string

const (
    InvitePending  InviteStatus = "pending"
    InviteAccepted InviteStatus = "accepted"
    InviteRevoked  InviteStatus = "revoked"
)

type Invitation struct {
    ID        uuid.UUID
    ProductID uuid.UUID
    Email     string
    Role      Role
    Token     string
    Status    InviteStatus
    CreatedAt time.Time
    UpdatedAt time.Time
}


