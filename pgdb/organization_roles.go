package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/netbill/pgx"

	sq "github.com/Masterminds/squirrel"
)

const OrganizationRoleTable = "organization_roles"

const OrganizationRoleColumns = "id, organization_id, head, rank, name, description, color, created_at, updated_at"
const OrganizationRoleColumnsR = "r.id, r.organization_id, r.head, r.rank, r.name, r.description, r.color, r.created_at, r.updated_at"

type OrganizationRole struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Head           bool      `json:"head"`
	Rank           uint      `json:"rank"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Color          string    `json:"color"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *OrganizationRole) scan(row sq.RowScanner) error {
	err := row.Scan(
		&r.ID,
		&r.OrganizationID,
		&r.Head,
		&r.Rank,
		&r.Name,
		&r.Description,
		&r.Color,
		&r.CreatedAt,
		&r.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("scanning role: %w", err)
	}
	return nil
}

type OrgRolesQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewOrgRolesQ(db pgx.DBTX) OrgRolesQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return OrgRolesQ{
		db:       db,
		selector: builder.Select(OrganizationRoleColumnsR).From(OrganizationRoleTable + " r"),
		inserter: builder.Insert(OrganizationRoleTable),
		updater:  builder.Update(OrganizationRoleTable + " r"),
		deleter:  builder.Delete(OrganizationRoleTable + " r"),
		counter:  builder.Select("COUNT(*)").From(OrganizationRoleTable + " r"),
	}
}

type InsertRoleParams struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	Head           bool      `json:"head"`
	Rank           uint      `json:"rank"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Color          string    `json:"color"`
}

func (q OrgRolesQ) Insert(ctx context.Context, data InsertRoleParams) (OrganizationRole, error) {
	const sqlInsertAtRank = `
		WITH bumped AS (
			UPDATE organization_roles
			SET
				rank = rank + 1,
				updated_at = now()
			WHERE organization_id = $1
			  AND rank >= $2
			RETURNING 1
		),
		ins AS (
			INSERT INTO organization_roles (organization_id, head, rank, name, description, color)
			VALUES ($1, $3, $2, $4, $5, $6)
			RETURNING id, organization_id, head, rank, name, description, color, created_at, updated_at
		)
		SELECT id, organization_id, head, rank, name, description, color, created_at, updated_at
		FROM ins;
	`

	args := []any{
		data.OrganizationID,
		data.Rank,
		data.Head,
		data.Name,
		data.Description,
		data.Color,
	}

	var inserted OrganizationRole
	if err := inserted.scan(q.db.QueryRowContext(ctx, sqlInsertAtRank, args...)); err != nil {
		return OrganizationRole{}, fmt.Errorf("insert role at rank: %w", err)
	}

	return inserted, nil
}

func (q OrgRolesQ) Get(ctx context.Context) (OrganizationRole, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return OrganizationRole{}, fmt.Errorf("building select query for %s: %w", OrganizationRoleTable, err)
	}

	var r OrganizationRole
	if err = r.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return OrganizationRole{}, nil
		default:
			return OrganizationRole{}, err
		}
	}

	return r, nil
}

func (q OrgRolesQ) Select(ctx context.Context) ([]OrganizationRole, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", OrganizationRoleTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", OrganizationRoleTable, err)
	}
	defer rows.Close()

	var out []OrganizationRole
	for rows.Next() {
		var r OrganizationRole
		if err = r.scan(rows); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q OrgRolesQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", OrganizationRoleTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing delete query for %s: %w", OrganizationRoleTable, err)
	}

	return nil
}

func (q OrgRolesQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", OrganizationRoleTable, err)
	}

	var count uint
	if err = q.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("scanning count for %s: %w", OrganizationRoleTable, err)
	}

	return count, nil
}

func (q OrgRolesQ) UpdateOne(ctx context.Context) (OrganizationRole, error) {
	q.updater = q.updater.Set("updated_at", time.Now().UTC())

	query, args, err := q.updater.Suffix("RETURNING " + OrganizationRoleColumns).ToSql()
	if err != nil {
		return OrganizationRole{}, fmt.Errorf("building update query for %s: %w", OrganizationRoleTable, err)
	}

	var updated OrganizationRole
	if err = updated.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return OrganizationRole{}, err
	}

	return updated, nil
}

func (q OrgRolesQ) UpdateMany(ctx context.Context) (int64, error) {
	q.updater = q.updater.Set("updated_at", time.Now().UTC())

	query, args, err := q.updater.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building update query for %s: %w", OrganizationRoleTable, err)
	}

	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("executing update query for %s: %w", OrganizationRoleTable, err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", OrganizationRoleTable, err)
	}

	return aff, nil
}

func (q OrgRolesQ) UpdateName(name string) OrgRolesQ {
	q.updater = q.updater.Set("name", name)
	return q
}

func (q OrgRolesQ) UpdateDescription(description string) OrgRolesQ {
	q.updater = q.updater.Set("description", description)
	return q
}

func (q OrgRolesQ) UpdateColor(color string) OrgRolesQ {
	q.updater = q.updater.Set("color", color)
	return q
}

func (q OrgRolesQ) FilterByID(id ...uuid.UUID) OrgRolesQ {
	q.selector = q.selector.Where(sq.Eq{"r.id": id})
	q.counter = q.counter.Where(sq.Eq{"r.id": id})
	q.updater = q.updater.Where(sq.Eq{"r.id": id})
	q.deleter = q.deleter.Where(sq.Eq{"r.id": id})
	return q
}

func (q OrgRolesQ) FilterByOrganizationID(id uuid.UUID) OrgRolesQ {
	q.selector = q.selector.Where(sq.Eq{"r.organization_id": id})
	q.counter = q.counter.Where(sq.Eq{"r.organization_id": id})
	q.updater = q.updater.Where(sq.Eq{"r.organization_id": id})
	q.deleter = q.deleter.Where(sq.Eq{"r.organization_id": id})
	return q
}

func (q OrgRolesQ) FilterByAccountID(accountID uuid.UUID) OrgRolesQ {
	sub := sq.
		Select("DISTINCT mr.role_id").
		From("organization_members m").
		Join("organization_member_roles mr ON mr.member_id = m.id").
		Where(sq.Eq{"m.account_id": accountID})

	subSQL, subArgs, err := sub.ToSql()
	if err != nil {
		q.selector = q.selector.Where(sq.Expr("1=0"))
		q.counter = q.counter.Where(sq.Expr("1=0"))
		q.updater = q.updater.Where(sq.Expr("1=0"))
		q.deleter = q.deleter.Where(sq.Expr("1=0"))
		return q
	}

	expr := sq.Expr("r.id IN ("+subSQL+")", subArgs...)

	q.selector = q.selector.Where(expr)
	q.counter = q.counter.Where(expr)
	q.updater = q.updater.Where(expr)
	q.deleter = q.deleter.Where(expr)

	return q
}

func (q OrgRolesQ) FilterByMemberID(memberID uuid.UUID) OrgRolesQ {
	sub := sq.
		Select("mr.role_id").
		From("organization_member_roles mr").
		Where(sq.Eq{"mr.member_id": memberID})

	subSQL, subArgs, err := sub.ToSql()
	if err != nil {
		q.selector = q.selector.Where(sq.Expr("1=0"))
		q.counter = q.counter.Where(sq.Expr("1=0"))
		q.updater = q.updater.Where(sq.Expr("1=0"))
		q.deleter = q.deleter.Where(sq.Expr("1=0"))
		return q
	}

	expr := sq.Expr("r.id IN ("+subSQL+")", subArgs...)

	q.selector = q.selector.Where(expr)
	q.counter = q.counter.Where(expr)
	q.updater = q.updater.Where(expr)
	q.deleter = q.deleter.Where(expr)

	return q
}

func (q OrgRolesQ) FilterHead(head bool) OrgRolesQ {
	q.selector = q.selector.Where(sq.Eq{"r.head": head})
	q.counter = q.counter.Where(sq.Eq{"r.head": head})
	q.updater = q.updater.Where(sq.Eq{"r.head": head})
	q.deleter = q.deleter.Where(sq.Eq{"r.head": head})
	return q
}

func (q OrgRolesQ) FilterByRank(rank int) OrgRolesQ {
	q.selector = q.selector.Where(sq.Eq{"r.rank": rank})
	q.counter = q.counter.Where(sq.Eq{"r.rank": rank})
	q.updater = q.updater.Where(sq.Eq{"r.rank": rank})
	q.deleter = q.deleter.Where(sq.Eq{"r.rank": rank})
	return q
}

func (q OrgRolesQ) FilterLikeName(name string) OrgRolesQ {
	q.selector = q.selector.Where(sq.ILike{"r.name": "%" + name + "%"})
	q.counter = q.counter.Where(sq.ILike{"r.name": "%" + name + "%"})
	q.updater = q.updater.Where(sq.ILike{"r.name": "%" + name + "%"})
	q.deleter = q.deleter.Where(sq.ILike{"r.name": "%" + name + "%"})
	return q
}

func (q OrgRolesQ) OrderByRoleRank(asc bool) OrgRolesQ {
	if asc {
		q.selector = q.selector.OrderBy("r.rank ASC", "r.id ASC")
	} else {
		q.selector = q.selector.OrderBy("r.rank DESC", "r.id DESC")
	}
	return q
}

func (q OrgRolesQ) Page(limit, offset uint) OrgRolesQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}

//Special methods to interact with role ranks in organization

func (q OrgRolesQ) DeleteAndShiftRanks(ctx context.Context, roleID uuid.UUID) error {
	const sqlq = `
		WITH del AS (
			DELETE FROM organization_roles
			WHERE id = $1
			RETURNING organization_id, rank
		)
		UPDATE organization_roles r
		SET rank = r.rank - 1,
		    updated_at = now()
		FROM del
		WHERE r.organization_id = del.organization_id
		  AND r.rank > del.rank
	`

	if _, err := q.db.ExecContext(ctx, sqlq, roleID); err != nil {
		return fmt.Errorf("executing delete+shift for %s: %w", OrganizationRoleTable, err)
	}

	return nil
}

func (q OrgRolesQ) UpdateRoleRank(ctx context.Context, roleID uuid.UUID, newRank uint) (OrganizationRole, error) {
	var aggID uuid.UUID
	var oldRank int

	{
		const sqlGet = `SELECT organization_id, rank FROM roles WHERE id = $1 LIMIT 1`
		if err := q.db.QueryRowContext(ctx, sqlGet, roleID).Scan(&aggID, &oldRank); err != nil {
			return OrganizationRole{}, fmt.Errorf("scanning role rank: %w", err)
		}
	}

	if oldRank == int(newRank) {
		return NewOrgRolesQ(q.db).FilterByID(roleID).Get(ctx)
	}

	const sqlMove = `
		WITH upd AS (
			UPDATE organization_roles
			SET
				rank = CASE
					WHEN id = $1 THEN $2
					WHEN $2 < $3 AND rank >= $2 AND rank < $3 THEN rank + 1
					WHEN $2 > $3 AND rank <= $2 AND rank > $3 THEN rank - 1
					ELSE rank
				END,
				updated_at = now()
			WHERE organization_id = $4
			RETURNING id, organization_id, head, rank, name, description, color, created_at, updated_at
		)
		SELECT id, organization_id, head, rank, name, description, color, created_at, updated_at
		FROM upd
		WHERE id = $1
	`

	args := []any{roleID, int(newRank), oldRank, aggID}

	var out OrganizationRole
	if err := out.scan(q.db.QueryRowContext(ctx, sqlMove, args...)); err != nil {
		return OrganizationRole{}, err
	}

	return out, nil
}

func (q OrgRolesQ) UpdateRolesRanks(
	ctx context.Context,
	organizationID uuid.UUID,
	order map[uuid.UUID]uint,
) ([]OrganizationRole, error) {
	roles, err := NewOrgRolesQ(q.db).
		FilterByOrganizationID(organizationID).
		OrderByRoleRank(true).
		Select(ctx)
	if err != nil {
		return nil, fmt.Errorf("select roles by organization: %w", err)
	}
	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles in organization %s", organizationID)
	}

	n := uint(len(roles))

	idToRole := make(map[uuid.UUID]OrganizationRole, n)
	for i := range roles {
		idToRole[roles[i].ID] = roles[i]
	}

	usedRank := make(map[uint]uuid.UUID, len(order))
	for roleID, newRank := range order {
		if newRank < 0 || newRank >= n {
			return nil, fmt.Errorf("rank %d out of range [0..%d]", newRank, n-1)
		}
		if _, ok := idToRole[roleID]; !ok {
			return nil, fmt.Errorf("role %s not in organization %s", roleID, organizationID)
		}
		if prev, ok := usedRank[newRank]; ok && prev != roleID {
			return nil, fmt.Errorf("duplicate rank %d for roles %s and %s", newRank, prev, roleID)
		}
		usedRank[newRank] = roleID
	}

	target := make([]uuid.UUID, n)
	filled := make([]bool, n)

	for r, id := range usedRank {
		target[r] = id
		filled[r] = true
	}

	rest := make([]uuid.UUID, 0, n-uint(len(order)))
	for i := range roles {
		id := roles[i].ID
		if _, ok := order[id]; ok {
			continue
		}
		rest = append(rest, id)
	}

	j := 0
	for i := 0; uint(i) < n; i++ {
		if filled[i] {
			continue
		}
		target[i] = rest[j]
		j++
	}

	changed := make([]uuid.UUID, 0, n)
	newRanks := make([]int, 0, n)

	for newRank, id := range target {
		if roles[newRank].ID != id {
			changed = append(changed, id)
			newRanks = append(newRanks, newRank)
		}
	}

	if len(changed) == 0 {
		return roles, nil
	}

	const sqlUpdate = `
		UPDATE organization_roles r
		SET
			rank = v.rank,
			updated_at = now()
		FROM (
			SELECT UNNEST($1::uuid[]) AS id, UNNEST($2::int[]) AS rank
		) v
		WHERE r.id = v.id
		  AND r.organization_id = $3
		RETURNING r.id, r.organization_id, r.head, r.rank, r.name, r.description, r.color, r.created_at, r.updated_at
	`

	ids := make([]string, len(changed))
	for i, id := range changed {
		ids[i] = id.String()
	}

	rows, err := q.db.QueryContext(ctx, sqlUpdate, pq.Array(ids), pq.Array(newRanks), organizationID)
	if err != nil {
		return nil, fmt.Errorf("updating roles ranks: %w", err)
	}
	defer rows.Close()

	out := make([]OrganizationRole, 0, len(changed))
	for rows.Next() {
		var r OrganizationRole
		if err = r.scan(rows); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
