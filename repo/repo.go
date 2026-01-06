package repo

import (
	"context"
	"database/sql"

	"github.com/netbill/pgx"
	"github.com/netbill/replicas/repo/pgdb"
)

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) Service {
	return Service{db: db}
}

func (s Service) ProfilesQ(ctx context.Context) pgdb.ProfilesQ {
	return pgdb.NewProfilesQ(pgx.Exec(s.db, ctx))
}

func (s Service) OrganizationsQ(ctx context.Context) pgdb.OrganizationsQ {
	return pgdb.NewOrganizationsQ(pgx.Exec(s.db, ctx))
}

func (s Service) OrgMembersQ(ctx context.Context) pgdb.OrgMembersQ {
	return pgdb.NewOrgMembersQ(pgx.Exec(s.db, ctx))
}

func (s Service) OrgMemberRolesQ(ctx context.Context) pgdb.OrgMemberRolesQ {
	return pgdb.NewOrgMemberRolesQ(pgx.Exec(s.db, ctx))
}

func (s Service) OrgRolesQ(ctx context.Context) pgdb.OrgRolesQ {
	return pgdb.NewOrgRolesQ(pgx.Exec(s.db, ctx))
}

func (s Service) OrgRolePermissionLinksQ(ctx context.Context) pgdb.OrgRolePermissionLinksQ {
	return pgdb.NewOrgRolePermissionsQ(pgx.Exec(s.db, ctx))
}

func (s Service) OrgRolePermissionsQ(ctx context.Context) pgdb.OrgRolePermissionsQ {
	return pgdb.NewOrgPermissionsQ(pgx.Exec(s.db, ctx))
}

func (s Service) OrgInvitesQ(ctx context.Context) pgdb.OrgInvitesQ {
	return pgdb.NewOrgInvitesQ(pgx.Exec(s.db, ctx))
}

func (s Service) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return pgx.Transaction(s.db, ctx, fn)
}
