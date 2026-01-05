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

const OrganizationRolePermissionsTable = "organization_role_permission_links"
const OrganizationRolePermissionsColumns = "role_id, permission_id"

type OrgRolePermissionLinksQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

type OrganizationRolePermissionLink struct {
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
}

func NewOrgRolePermissionsQ(db pgx.DBTX) OrgRolePermissionLinksQ {
	b := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return OrgRolePermissionLinksQ{
		db:       db,
		selector: b.Select(OrganizationRolePermissionsColumns).From(OrganizationRolePermissionsTable),
		inserter: b.Insert(OrganizationRolePermissionsTable),
		deleter:  b.Delete(OrganizationRolePermissionsTable),
		counter:  b.Select("COUNT(*)").From(OrganizationRolePermissionsTable),
	}
}

func (q OrgRolePermissionLinksQ) Insert(ctx context.Context, data ...OrganizationRolePermissionLink) error {
	if len(data) == 0 {
		return nil
	}

	ins := q.inserter.Columns("role_id", "permission_id")

	for _, rp := range data {
		ins = ins.Values(rp.RoleID, rp.PermissionID)
	}

	query, args, err := ins.ToSql()
	if err != nil {
		return fmt.Errorf("building insert query for %s: %w", OrganizationRolePermissionsTable, err)
	}

	if _, err := q.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("executing insert query for %s: %w", OrganizationRolePermissionsTable, err)
	}

	return nil
}

func (q OrgRolePermissionLinksQ) Get(ctx context.Context) (OrganizationRolePermissionLink, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return OrganizationRolePermissionLink{}, fmt.Errorf("building select query for %s: %w", OrganizationRolePermissionsTable, err)
	}

	var rp OrganizationRolePermissionLink
	if err = q.db.QueryRowContext(ctx, query, args...).Scan(&rp.RoleID, &rp.PermissionID); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return OrganizationRolePermissionLink{}, nil
		default:
			return OrganizationRolePermissionLink{}, fmt.Errorf("scanning row for %s: %w", OrganizationRolePermissionsTable, err)
		}
	}

	return rp, nil
}

func (q OrgRolePermissionLinksQ) Select(ctx context.Context) ([]OrganizationRolePermissionLink, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", OrganizationRolePermissionsTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", OrganizationRolePermissionsTable, err)
	}
	defer rows.Close()

	var rps []OrganizationRolePermissionLink
	for rows.Next() {
		var rp OrganizationRolePermissionLink
		if err = rows.Scan(&rp.RoleID, &rp.PermissionID); err != nil {
			return nil, fmt.Errorf("scanning row for %s: %w", OrganizationRolePermissionsTable, err)
		}
		rps = append(rps, rp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows for %s: %w", OrganizationRolePermissionsTable, err)
	}

	return rps, nil
}

func (q OrgRolePermissionLinksQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", OrganizationRolePermissionsTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing delete query for %s: %w", OrganizationRolePermissionsTable, err)
	}

	return nil
}

func (q OrgRolePermissionLinksQ) FilterByRoleID(roleID uuid.UUID) OrgRolePermissionLinksQ {
	q.selector = q.selector.Where(sq.Eq{"role_id": roleID})
	q.deleter = q.deleter.Where(sq.Eq{"role_id": roleID})
	q.counter = q.counter.Where(sq.Eq{"role_id": roleID})
	return q
}

func (q OrgRolePermissionLinksQ) FilterByPermissionID(permissionID uuid.UUID) OrgRolePermissionLinksQ {
	q.selector = q.selector.Where(sq.Eq{"permission_id": permissionID})
	q.deleter = q.deleter.Where(sq.Eq{"permission_id": permissionID})
	q.counter = q.counter.Where(sq.Eq{"permission_id": permissionID})
	return q
}

func (q OrgRolePermissionLinksQ) FilterByPermissionCode(code ...string) OrgRolePermissionLinksQ {
	sub := sq.
		Select("id").
		From(OrganizationPermissionTable).
		Where(sq.Eq{"code": code})

	subSQL, subArgs, err := sub.ToSql()
	if err != nil {
		q.selector = q.selector.Where(sq.Expr("1=0"))
		q.deleter = q.deleter.Where(sq.Expr("1=0"))
		q.counter = q.counter.Where(sq.Expr("1=0"))
		return q
	}

	expr := sq.Expr("permission_id IN ("+subSQL+")", subArgs...)
	q.selector = q.selector.Where(expr)
	q.deleter = q.deleter.Where(expr)
	q.counter = q.counter.Where(expr)

	return q
}

func (q OrgRolePermissionLinksQ) FilterByAccountID(accountID uuid.UUID) OrgRolePermissionLinksQ {
	sub := sq.
		Select("DISTINCT mr.role_id").
		From("organization_members m").
		Join("organization_member_roles mr ON mr.member_id = m.id").
		Where(sq.Eq{"m.account_id": accountID})

	subSQL, subArgs, err := sub.ToSql()
	if err != nil {
		q.selector = q.selector.Where(sq.Expr("1=0"))
		q.deleter = q.deleter.Where(sq.Expr("1=0"))
		q.counter = q.counter.Where(sq.Expr("1=0"))
		return q
	}

	expr := sq.Expr("role_id IN ("+subSQL+")", subArgs...)
	q.selector = q.selector.Where(expr)
	q.deleter = q.deleter.Where(expr)
	q.counter = q.counter.Where(expr)

	return q
}

func (q OrgRolePermissionLinksQ) FilterByOrganizationID(organizationID uuid.UUID) OrgRolePermissionLinksQ {
	sub := sq.
		Select("id").
		From("roles").
		Where(sq.Eq{"organization_id": organizationID})

	subSQL, subArgs, err := sub.ToSql()
	if err != nil {
		q.selector = q.selector.Where(sq.Expr("1=0"))
		q.deleter = q.deleter.Where(sq.Expr("1=0"))
		q.counter = q.counter.Where(sq.Expr("1=0"))
		return q
	}

	expr := sq.Expr("role_id IN ("+subSQL+")", subArgs...)
	q.selector = q.selector.Where(expr)
	q.deleter = q.deleter.Where(expr)
	q.counter = q.counter.Where(expr)

	return q
}

func (q OrgRolePermissionLinksQ) FilterByMemberID(memberID uuid.UUID) OrgRolePermissionLinksQ {
	sub := sq.
		Select("mr.role_id").
		From("organization_member_roles mr").
		Where(sq.Eq{"mr.member_id": memberID})

	subSQL, subArgs, err := sub.ToSql()
	if err != nil {
		q.selector = q.selector.Where(sq.Expr("1=0"))
		q.deleter = q.deleter.Where(sq.Expr("1=0"))
		q.counter = q.counter.Where(sq.Expr("1=0"))
		return q
	}

	expr := sq.Expr("role_id IN ("+subSQL+")", subArgs...)
	q.selector = q.selector.Where(expr)
	q.deleter = q.deleter.Where(expr)
	q.counter = q.counter.Where(expr)

	return q
}

func (q OrgRolePermissionLinksQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", OrganizationRolePermissionsTable, err)
	}

	var n uint
	if err = q.db.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("scanning count for %s: %w", OrganizationRolePermissionsTable, err)
	}
	return n, nil
}

func (q OrgRolePermissionLinksQ) Page(limit, offset uint) OrgRolePermissionLinksQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}

func (q OrgRolePermissionLinksQ) Exists(ctx context.Context) (bool, error) {
	subSQL, subArgs, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return false, fmt.Errorf("building exists query for %s: %w", OrganizationRolePermissionsTable, err)
	}

	sqlq := "SELECT EXISTS (" + subSQL + ")"

	var ok bool
	if err = q.db.QueryRowContext(ctx, sqlq, subArgs...).Scan(&ok); err != nil {
		return false, fmt.Errorf("scanning exists for %s: %w", OrganizationRolePermissionsTable, err)
	}
	return ok, nil
}
