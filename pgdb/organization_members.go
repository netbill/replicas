package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/pgx"

	sq "github.com/Masterminds/squirrel"
)

const OrganizationMembersTable = "organization_members"

const OrganizationMemberColumns = "id, account_id, organization_id, position, label, created_at, updated_at"
const OrganizationMemberColumnsM = "m.id, m.account_id, m.organization_id, m.position, m.label, m.created_at, m.updated_at"

type OrganizationMember struct {
	ID             uuid.UUID `json:"id"`
	AccountID      uuid.UUID `json:"account_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Position       *string   `json:"position"`
	Label          *string   `json:"label"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (m *OrganizationMember) scan(row sq.RowScanner) error {
	err := row.Scan(
		&m.ID,
		&m.AccountID,
		&m.OrganizationID,
		&m.Position,
		&m.Label,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("scanning member: %w", err)
	}
	return nil
}

type OrgMembersQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewOrgMembersQ(db pgx.DBTX) OrgMembersQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return OrgMembersQ{
		db:       db,
		selector: builder.Select(OrganizationMemberColumnsM).From(OrganizationMembersTable + " m"),
		inserter: builder.Insert(OrganizationMembersTable),
		updater:  builder.Update(OrganizationMembersTable + " m"),
		deleter:  builder.Delete(OrganizationMembersTable + " m"),
		counter:  builder.Select("COUNT(*)").From(OrganizationMembersTable + " m"),
	}
}

type InsertMemberParams struct {
	AccountID      uuid.UUID
	OrganizationID uuid.UUID
	Position       *string
	Label          *string
}

func (q OrgMembersQ) Insert(ctx context.Context, data InsertMemberParams) (OrganizationMember, error) {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"account_id":      data.AccountID,
		"organization_id": data.OrganizationID,
		"position":        data.Position,
		"label":           data.Label,
	}).Suffix("RETURNING " + OrganizationMemberColumns).ToSql()
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("building insert query for %s: %w", OrganizationMembersTable, err)
	}

	var inserted OrganizationMember
	err = inserted.scan(q.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		return OrganizationMember{}, err
	}

	return inserted, nil
}

func (q OrgMembersQ) Exists(ctx context.Context) (bool, error) {
	existsQ := q.selector.
		Columns("1").
		RemoveLimit().
		RemoveOffset().
		Prefix("SELECT EXISTS (").
		Suffix(") AS exists").
		Limit(1)

	query, args, err := existsQ.ToSql()
	if err != nil {
		return false, fmt.Errorf("building exists query for %s: %w", OrganizationMembersTable, err)
	}

	var ok bool
	if err = q.db.QueryRowContext(ctx, query, args...).Scan(&ok); err != nil {
		return false, fmt.Errorf("scanning exists for %s: %w", OrganizationMembersTable, err)
	}

	return ok, nil
}

func (q OrgMembersQ) Get(ctx context.Context) (OrganizationMember, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("building select query for %s: %w", OrganizationMembersTable, err)
	}

	var m OrganizationMember
	err = m.scan(q.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return OrganizationMember{}, nil
		default:
			return OrganizationMember{}, err
		}
	}

	return m, nil
}

func (q OrgMembersQ) Select(ctx context.Context) ([]OrganizationMember, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", OrganizationMembersTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", OrganizationMembersTable, err)
	}
	defer rows.Close()

	var out []OrganizationMember
	for rows.Next() {
		var m OrganizationMember
		if err = m.scan(rows); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q OrgMembersQ) FilterByID(id uuid.UUID) OrgMembersQ {
	q.selector = q.selector.Where(sq.Eq{"m.id": id})
	q.counter = q.counter.Where(sq.Eq{"m.id": id})
	q.updater = q.updater.Where(sq.Eq{"m.id": id})
	q.deleter = q.deleter.Where(sq.Eq{"m.id": id})
	return q
}

func (q OrgMembersQ) FilterByAccountID(accountID uuid.UUID) OrgMembersQ {
	q.selector = q.selector.Where(sq.Eq{"m.account_id": accountID})
	q.counter = q.counter.Where(sq.Eq{"m.account_id": accountID})
	q.updater = q.updater.Where(sq.Eq{"m.account_id": accountID})
	q.deleter = q.deleter.Where(sq.Eq{"m.account_id": accountID})
	return q
}

func (q OrgMembersQ) FilterByOrganizationID(organizationID uuid.UUID) OrgMembersQ {
	q.selector = q.selector.Where(sq.Eq{"m.organization_id": organizationID})
	q.counter = q.counter.Where(sq.Eq{"m.organization_id": organizationID})
	q.updater = q.updater.Where(sq.Eq{"m.organization_id": organizationID})
	q.deleter = q.deleter.Where(sq.Eq{"m.organization_id": organizationID})
	return q
}

func (q OrgMembersQ) FilterLikePosition(position string) OrgMembersQ {
	q.selector = q.selector.Where(sq.ILike{"m.position": "%" + position + "%"})
	q.counter = q.counter.Where(sq.ILike{"m.position": "%" + position + "%"})
	q.updater = q.updater.Where(sq.ILike{"m.position": "%" + position + "%"})
	q.deleter = q.deleter.Where(sq.ILike{"m.position": "%" + position + "%"})
	return q
}

func (q OrgMembersQ) FilterLikeLabel(label string) OrgMembersQ {
	q.selector = q.selector.Where(sq.ILike{"m.label": "%" + label + "%"})
	q.counter = q.counter.Where(sq.ILike{"m.label": "%" + label + "%"})
	q.updater = q.updater.Where(sq.ILike{"m.label": "%" + label + "%"})
	q.deleter = q.deleter.Where(sq.ILike{"m.label": "%" + label + "%"})
	return q
}

func (q OrgMembersQ) UpdateOne(ctx context.Context) (OrganizationMember, error) {
	q.updater = q.updater.Set("updated_at", time.Now().UTC())

	query, args, err := q.updater.Suffix("RETURNING " + OrganizationMemberColumns).ToSql()
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("building update query for %s: %w", OrganizationMembersTable, err)
	}

	var updated OrganizationMember
	err = updated.scan(q.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		return OrganizationMember{}, err
	}

	return updated, nil
}

func (q OrgMembersQ) UpdateMany(ctx context.Context) (int64, error) {
	q.updater = q.updater.Set("updated_at", time.Now().UTC())

	query, args, err := q.updater.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building update query for %s: %w", OrganizationMembersTable, err)
	}

	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("executing update query for %s: %w", OrganizationMembersTable, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", OrganizationMembersTable, err)
	}

	return affected, nil
}

func (q OrgMembersQ) UpdatePosition(position sql.NullString) OrgMembersQ {
	q.updater = q.updater.Set("position", position)
	return q
}

func (q OrgMembersQ) UpdateLabel(label sql.NullString) OrgMembersQ {
	q.updater = q.updater.Set("label", label)
	return q
}

func (q OrgMembersQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", OrganizationMembersTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing delete query for %s: %w", OrganizationMembersTable, err)
	}

	return nil
}

func (q OrgMembersQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", OrganizationMembersTable, err)
	}

	var count uint
	err = q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("scanning count for %s: %w", OrganizationMembersTable, err)
	}

	return count, nil
}

func (q OrgMembersQ) Page(limit uint, offset uint) OrgMembersQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
