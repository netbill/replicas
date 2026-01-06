package contracts

import (
	"github.com/google/uuid"
)

const ProfileUpdatedEvent = "profile.updated"

type ProfileUpdatedPayload struct {
	Profile struct {
		AccountID uuid.UUID `json:"account_id"`
		Username  string    `json:"username"`
		Official  bool      `json:"official"`
		Pseudonym *string   `json:"pseudonym"`
	} `json:"profile"`
}
