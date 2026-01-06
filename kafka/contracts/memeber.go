package contracts

import (
	"time"

	"github.com/google/uuid"
)

const MemberCreatedEvent = "member.created"

type MemberCreatedPayload struct {
	Member struct {
		ID             uuid.UUID `json:"id"`
		AccountID      uuid.UUID `json:"account_id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		Position       *string   `json:"position,omitempty"`
		Label          *string   `json:"label,omitempty"`

		Username  string  `json:"username"`
		Pseudonym *string `json:"pseudonym,omitempty"`
		Official  bool    `json:"official"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"member"`
}

const MemberUpdatedEvent = "member.updated"

type MemberUpdatedPayload struct {
	Member struct {
		ID             uuid.UUID `json:"id"`
		AccountID      uuid.UUID `json:"account_id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		Position       *string   `json:"position,omitempty"`
		Label          *string   `json:"label,omitempty"`

		Username  string  `json:"username"`
		Pseudonym *string `json:"pseudonym,omitempty"`
		Official  bool    `json:"official"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"member"`
}

const MemberDeletedEvent = "member.deleted"

type MemberDeletedPayload struct {
	Member struct {
		ID             uuid.UUID `json:"id"`
		AccountID      uuid.UUID `json:"account_id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		Position       *string   `json:"position,omitempty"`
		Label          *string   `json:"label,omitempty"`

		Username  string  `json:"username"`
		Pseudonym *string `json:"pseudonym,omitempty"`
		Official  bool    `json:"official"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"member"`
}
