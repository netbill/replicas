package contracts

import (
	"time"

	"github.com/google/uuid"
)

const OrganizationCreatedEvent = "organization.created"

type OrganizationCreatedPayload struct {
	Organization struct {
		ID       uuid.UUID `json:"id"`
		Status   string    `json:"status"`
		Name     string    `json:"name"`
		Icon     *string   `json:"icon"`
		MaxRoles uint      `json:"max_roles"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"organization"`
}

const OrganizationUpdatedEvent = "organization.updated"

type OrganizationUpdatedPayload struct {
	Organization struct {
		ID       uuid.UUID `json:"id"`
		Status   string    `json:"status"`
		Name     string    `json:"name"`
		Icon     *string   `json:"icon"`
		MaxRoles uint      `json:"max_roles"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"organization"`
}

const OrganizationActivatedEvent = "organization.activated"

type OrganizationActivatedPayload struct {
	Organization struct {
		ID       uuid.UUID `json:"id"`
		Status   string    `json:"status"`
		Name     string    `json:"name"`
		Icon     *string   `json:"icon"`
		MaxRoles uint      `json:"max_roles"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"organization"`
}

const OrganizationDeactivatedEvent = "organization.deactivated"

type OrganizationDeactivatedPayload struct {
	Organization struct {
		ID       uuid.UUID `json:"id"`
		Status   string    `json:"status"`
		Name     string    `json:"name"`
		Icon     *string   `json:"icon"`
		MaxRoles uint      `json:"max_roles"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"organization"`
}

const OrganizationSuspendedEvent = "organization.suspended"

type OrganizationSuspendedPayload struct {
	Organization struct {
		ID       uuid.UUID `json:"id"`
		Status   string    `json:"status"`
		Name     string    `json:"name"`
		Icon     *string   `json:"icon"`
		MaxRoles uint      `json:"max_roles"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"organization"`
}

const OrganizationDeletedEvent = "organization.deleted"

type OrganizationDeletedPayload struct {
	Organization struct {
		ID       uuid.UUID `json:"id"`
		Status   string    `json:"status"`
		Name     string    `json:"name"`
		Icon     *string   `json:"icon"`
		MaxRoles uint      `json:"max_roles"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"organization"`
}
