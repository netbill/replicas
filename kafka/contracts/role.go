package contracts

import (
	"time"

	"github.com/google/uuid"
)

const RoleCreatedEvent = "role.created"

type RoleCreatedPayload struct {
	Role struct {
		ID             uuid.UUID `json:"id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		Head           bool      `json:"head"`
		Rank           uint      `json:"rank"`
		Name           string    `json:"name"`
		Description    string    `json:"description"`
		Color          string    `json:"color"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"role"`
}

const RoleUpdatedEvent = "role.updated"

type RoleUpdatedPayload struct {
	Role struct {
		ID             uuid.UUID `json:"id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		Head           bool      `json:"head"`
		Rank           uint      `json:"rank"`
		Name           string    `json:"name"`
		Description    string    `json:"description"`
		Color          string    `json:"color"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"role"`
}

const RoleDeletedEvent = "role.deleted"

type RoleDeletedPayload struct {
	Role struct {
		ID             uuid.UUID `json:"id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		Head           bool      `json:"head"`
		Rank           uint      `json:"rank"`
		Name           string    `json:"name"`
		Description    string    `json:"description"`
		Color          string    `json:"color"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"role"`
}

const RolesRanksUpdatedEvent = "roles.ranks.updated"

type RolesRanksUpdatedPayload struct {
	OrganizationID uuid.UUID          `json:"organization_id"`
	Ranks          map[uuid.UUID]uint `json:"ranks"`
}

const RolePermissionsUpdatedEvent = "role.permissions.updated"

type RolePermissionsUpdatedPayload struct {
	RoleID      uuid.UUID          `json:"role_id"`
	Permissions map[uuid.UUID]bool `json:"permissions"`
}
