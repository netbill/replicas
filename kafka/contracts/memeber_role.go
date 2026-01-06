package contracts

import "github.com/google/uuid"

const MemberRoleAddedEvent = "member_role.added"

const MemberRoleRemovedEvent = "member_role.remove"

type MemberRoleAddedPayload struct {
	MemberID uuid.UUID `json:"member_id"`
	RoleID   uuid.UUID `json:"role_id"`
}

type MemberRoleRemovedPayload struct {
	MemberID uuid.UUID `json:"member_id"`
	RoleID   uuid.UUID `json:"role_id"`
}
