package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/netbill/pgx"
)

const OrganizationMemberRoleTable = "organization_member_roles"
const OrganizationMemberRoleColumns = "member_id, role_id"

type OrganizationMemberRole struct {
	MemberID uuid.UUID `json:"member_id"`
	RoleID   uuid.UUID `json:"role_id"`
}

func (mr *OrganizationMemberRole) scan(row sq.RowScanner) error {
	if err := row.Scan(&mr.MemberID, &mr.RoleID); err != nil {
		return fmt.Errorf("scanning member_role: %w", err)
	}
	return nil
}

type OrgMemberRolesQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewOrgMemberRolesQ(db pgx.DBTX) OrgMemberRolesQ {
	b := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return OrgMemberRolesQ{
		db:       db,
		selector: b.Select(OrganizationMemberRoleColumns).From(OrganizationMemberRoleTable),
		inserter: b.Insert(OrganizationMemberRoleTable),
		deleter:  b.Delete(OrganizationMemberRoleTable),
		counter:  b.Select("COUNT(*)").From(OrganizationMemberRoleTable),
	}
}

func (q OrgMemberRolesQ) Insert(ctx context.Context, data OrganizationMemberRole) (OrganizationMemberRole, error) {
	query, args, err := q.inserter.SetMap(map[string]any{
		"member_id": data.MemberID,
		"role_id":   data.RoleID,
	}).Suffix("RETURNING " + OrganizationMemberRoleColumns).ToSql()
	if err != nil {
		return OrganizationMemberRole{}, fmt.Errorf("building insert query for %s: %w", OrganizationMemberRoleTable, err)
	}

	var out OrganizationMemberRole
	if err = out.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return OrganizationMemberRole{}, err
	}
	return out, nil
}

func (q OrgMemberRolesQ) Get(ctx context.Context) (OrganizationMemberRole, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return OrganizationMemberRole{}, fmt.Errorf("building select query for %s: %w", OrganizationMemberRoleTable, err)
	}

	var out OrganizationMemberRole
	if err = out.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return OrganizationMemberRole{}, nil
		default:
			return OrganizationMemberRole{}, err
		}
	}
	return out, nil
}

func (q OrgMemberRolesQ) Select(ctx context.Context) ([]OrganizationMemberRole, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", OrganizationMemberRoleTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", OrganizationMemberRoleTable, err)
	}
	defer rows.Close()

	var out []OrganizationMemberRole
	for rows.Next() {
		var mr OrganizationMemberRole
		if err = mr.scan(rows); err != nil {
			return nil, err
		}
		out = append(out, mr)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q OrgMemberRolesQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", OrganizationMemberRoleTable, err)
	}
	if _, err = q.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("executing delete query for %s: %w", OrganizationMemberRoleTable, err)
	}
	return nil
}

func (q OrgMemberRolesQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", OrganizationMemberRoleTable, err)
	}

	var n uint
	if err = q.db.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("scanning count for %s: %w", OrganizationMemberRoleTable, err)
	}
	return n, nil
}

func (q OrgMemberRolesQ) FilterByMemberID(memberID uuid.UUID) OrgMemberRolesQ {
	q.selector = q.selector.Where(sq.Eq{"member_id": memberID})
	q.counter = q.counter.Where(sq.Eq{"member_id": memberID})
	q.deleter = q.deleter.Where(sq.Eq{"member_id": memberID})
	return q
}

func (q OrgMemberRolesQ) FilterByRoleID(roleID uuid.UUID) OrgMemberRolesQ {
	q.selector = q.selector.Where(sq.Eq{"role_id": roleID})
	q.counter = q.counter.Where(sq.Eq{"role_id": roleID})
	q.deleter = q.deleter.Where(sq.Eq{"role_id": roleID})
	return q
}
