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

const OrganizationPermissionTable = "organization_role_permission"
const OrganizationPermissionColumns = "id, code"

type OrganizationRolePermission struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
}

func (p *OrganizationRolePermission) scan(row sq.RowScanner) error {
	if err := row.Scan(&p.ID, &p.Code); err != nil {
		return fmt.Errorf("scanning permission: %w", err)
	}
	return nil
}

type OrgRolePermissionsQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewOrgPermissionsQ(db pgx.DBTX) OrgRolePermissionsQ {
	b := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return OrgRolePermissionsQ{
		db:       db,
		selector: b.Select(OrganizationPermissionColumns).From(OrganizationPermissionTable),
		inserter: b.Insert(OrganizationPermissionTable),
		updater:  b.Update(OrganizationPermissionTable),
		deleter:  b.Delete(OrganizationPermissionTable),
		counter:  b.Select("COUNT(*)").From(OrganizationPermissionTable),
	}
}

func (q OrgRolePermissionsQ) Insert(ctx context.Context, data OrganizationRolePermission) (OrganizationRolePermission, error) {
	query, args, err := q.inserter.SetMap(map[string]any{
		"id":   data.ID,
		"code": data.Code,
	}).Suffix("RETURNING " + OrganizationPermissionColumns).ToSql()
	if err != nil {
		return OrganizationRolePermission{}, fmt.Errorf("building insert query for %s: %w", OrganizationPermissionTable, err)
	}

	var out OrganizationRolePermission
	if err = out.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return OrganizationRolePermission{}, err
	}
	return out, nil
}

func (q OrgRolePermissionsQ) Get(ctx context.Context) (OrganizationRolePermission, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return OrganizationRolePermission{}, fmt.Errorf("building select query for %s: %w", OrganizationPermissionTable, err)
	}

	var out OrganizationRolePermission
	if err = out.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return OrganizationRolePermission{}, nil
		default:
			return OrganizationRolePermission{}, err
		}
	}
	return out, nil
}

func (q OrgRolePermissionsQ) Select(ctx context.Context) ([]OrganizationRolePermission, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", OrganizationPermissionTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", OrganizationPermissionTable, err)
	}
	defer rows.Close()

	var out []OrganizationRolePermission
	for rows.Next() {
		var p OrganizationRolePermission
		if err = p.scan(rows); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q OrgRolePermissionsQ) UpdateOne(ctx context.Context) (OrganizationRolePermission, error) {
	query, args, err := q.updater.Suffix("RETURNING " + OrganizationPermissionColumns).ToSql()
	if err != nil {
		return OrganizationRolePermission{}, fmt.Errorf("building update query for %s: %w", OrganizationPermissionTable, err)
	}

	var out OrganizationRolePermission
	if err = out.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return OrganizationRolePermission{}, err
	}
	return out, nil
}

func (q OrgRolePermissionsQ) UpdateMany(ctx context.Context) (int64, error) {
	query, args, err := q.updater.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building update query for %s: %w", OrganizationPermissionTable, err)
	}

	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("executing update query for %s: %w", OrganizationPermissionTable, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", OrganizationPermissionTable, err)
	}
	return n, nil
}

func (q OrgRolePermissionsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", OrganizationPermissionTable, err)
	}
	if _, err = q.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("executing delete query for %s: %w", OrganizationPermissionTable, err)
	}
	return nil
}

func (q OrgRolePermissionsQ) FilterByID(id uuid.UUID) OrgRolePermissionsQ {
	q.selector = q.selector.Where(sq.Eq{"id": id})
	q.counter = q.counter.Where(sq.Eq{"id": id})
	q.updater = q.updater.Where(sq.Eq{"id": id})
	q.deleter = q.deleter.Where(sq.Eq{"id": id})
	return q
}

func (q OrgRolePermissionsQ) FilterByCode(code ...string) OrgRolePermissionsQ {
	q.selector = q.selector.Where(sq.Eq{"code": code})
	q.counter = q.counter.Where(sq.Eq{"code": code})
	q.updater = q.updater.Where(sq.Eq{"code": code})
	q.deleter = q.deleter.Where(sq.Eq{"code": code})
	return q
}

func (q OrgRolePermissionsQ) FilterByRoleID(roleID uuid.UUID) OrgRolePermissionsQ {
	q.selector = q.selector.
		Join("organization_role_permission_links rp ON rp.permission_id = role_permissions.id").
		Where(sq.Eq{"rp.role_id": roleID}).
		Distinct()

	q.counter = q.counter.
		Join("organization_role_permission_links rp ON rp.permission_id = role_permissions.id").
		Where(sq.Eq{"rp.role_id": roleID})

	return q
}

func (q OrgRolePermissionsQ) FilterLikeDescription(description string) OrgRolePermissionsQ {
	q.selector = q.selector.Where(sq.ILike{"description": "%" + description + "%"})
	q.counter = q.counter.Where(sq.ILike{"description": "%" + description + "%"})
	return q
}

func (q OrgRolePermissionsQ) GetForRole(
	ctx context.Context,
	roleID uuid.UUID,
) (map[OrganizationRolePermission]bool, error) {

	const sqlq = `
		SELECT
			p.id,
			p.code,
			(rp.permission_id IS NOT NULL) AS enabled
		FROM organization_role_permissions p
		LEFT JOIN organization_role_permission_links rp
			ON rp.permission_id = p.id
			AND rp.role_id = $1
		ORDER BY p.code
	`

	rows, err := q.db.QueryContext(ctx, sqlq, roleID)
	if err != nil {
		return nil, fmt.Errorf("query organization_role_permissions for role: %w", err)
	}
	defer rows.Close()

	out := make(map[OrganizationRolePermission]bool)

	for rows.Next() {
		var p OrganizationRolePermission
		var enabled bool

		if err = rows.Scan(
			&p.ID,
			&p.Code,
			&enabled,
		); err != nil {
			return nil, fmt.Errorf("scanning permission for role: %w", err)
		}

		out[p] = enabled
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
