package contracts

import (
	"time"

	"github.com/google/uuid"
)

const InviteCreatedEvent = "invite.created"

type InviteCreatedPayload struct {
	Invite struct {
		ID             uuid.UUID `json:"id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		AccountID      uuid.UUID `json:"account_id"`
		Status         string    `json:"status"`
		ExpiresAt      time.Time `json:"expires_at"`
		CreatedAt      time.Time `json:"created_at"`
	} `json:"invite"`
}

const InviteAcceptedEvent = "invite.accepted"

type InviteAcceptedPayload struct {
	Invite struct {
		ID             uuid.UUID `json:"id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		AccountID      uuid.UUID `json:"account_id"`
		Status         string    `json:"status"`
		ExpiresAt      time.Time `json:"expires_at"`
		CreatedAt      time.Time `json:"created_at"`
	} `json:"invite"`
}

const InviteDeclinedEvent = "invite.declined"

type InviteDeclinedPayload struct {
	Invite struct {
		ID             uuid.UUID `json:"id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		AccountID      uuid.UUID `json:"account_id"`
		Status         string    `json:"status"`
		ExpiresAt      time.Time `json:"expires_at"`
		CreatedAt      time.Time `json:"created_at"`
	} `json:"invite"`
}

const InviteDeletedEvent = "invite.deleted"

type InviteDeletedPayload struct {
	Invite struct {
		ID             uuid.UUID `json:"id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		AccountID      uuid.UUID `json:"account_id"`
		Status         string    `json:"status"`
		ExpiresAt      time.Time `json:"expires_at"`
		CreatedAt      time.Time `json:"created_at"`
	} `json:"invite"`
}
