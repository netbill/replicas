package contracts

import (
	"time"

	"github.com/google/uuid"
)

const AccountCreatedEvent = "account.created"

type AccountCreatedPayload struct {
	Account struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
		Role     string    `json:"role"`
		Status   string    `json:"status"`

		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
		UsernameUpdatedAt time.Time `json:"username_name_updated_at"`
	} `json:"account"`
	Email string `json:"email,omitempty"`
}

const AccountUsernameChangeEvent = "account.username.change"

type AccountUsernameChangePayload struct {
	Account struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
		Role     string    `json:"role"`
		Status   string    `json:"status"`

		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
		UsernameUpdatedAt time.Time `json:"username_name_updated_at"`
	} `json:"account"`
	Email string `json:"email,omitempty"`
}

const AccountDeletedEvent = "account.deleted"

type AccountDeletedPayload struct {
	AccountID uuid.UUID `json:"account_id"`
	Account   struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
		Role     string    `json:"role"`
		Status   string    `json:"status"`

		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
		UsernameUpdatedAt time.Time `json:"username_name_updated_at"`
	} `json:"account"`
	Email string `json:"email,omitempty"`
}
