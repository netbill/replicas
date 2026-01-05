package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/netbill/pgx"
)

const OrganizationInviteTable = "organization_invites"
const OrganizationInviteColumns = "id, organization_id, account_id, status, expires_at, created_at"

type OrganizationInvite struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	AccountID      uuid.UUID `json:"account_id,omitempty"`
	Status         string    `json:"status"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

func (i *OrganizationInvite) scan(row sq.RowScanner) error {
	if err := row.Scan(
		&i.ID,
		&i.OrganizationID,
		&i.AccountID,
		&i.Status,
		&i.ExpiresAt,
		&i.CreatedAt,
	); err != nil {
		return fmt.Errorf("scanning invite: %w", err)
	}
	return nil
}

type OrgInvitesQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewOrgInvitesQ(db pgx.DBTX) OrgInvitesQ {
	b := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return OrgInvitesQ{
		db:       db,
		selector: b.Select(OrganizationInviteColumns).From(OrganizationInviteTable),
		inserter: b.Insert(OrganizationInviteTable),
		updater:  b.Update(OrganizationInviteTable),
		deleter:  b.Delete(OrganizationInviteTable),
		counter:  b.Select("COUNT(*)").From(OrganizationInviteTable),
	}
}

type InsertInviteParams struct {
	OrganizationID uuid.UUID
	AccountID      uuid.UUID
	ExpiresAt      time.Time
}

func (q OrgInvitesQ) Insert(ctx context.Context, data InsertInviteParams) (OrganizationInvite, error) {
	query, args, err := q.inserter.SetMap(map[string]any{
		"organization_id": data.OrganizationID,
		"account_id":      data.AccountID,
		"expires_at":      data.ExpiresAt,
	}).Suffix("RETURNING " + OrganizationInviteColumns).ToSql()
	if err != nil {
		return OrganizationInvite{}, fmt.Errorf("building insert query for %s: %w", OrganizationInviteTable, err)
	}

	var out OrganizationInvite
	if err = out.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return OrganizationInvite{}, err
	}
	return out, nil
}

func (q OrgInvitesQ) Get(ctx context.Context) (OrganizationInvite, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return OrganizationInvite{}, fmt.Errorf("building select query for %s: %w", OrganizationInviteTable, err)
	}

	var out OrganizationInvite
	if err = out.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return OrganizationInvite{}, nil
		default:
			return OrganizationInvite{}, err
		}
	}
	return out, nil
}

func (q OrgInvitesQ) Select(ctx context.Context) ([]OrganizationInvite, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", OrganizationInviteTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", OrganizationInviteTable, err)
	}
	defer rows.Close()

	var out []OrganizationInvite
	for rows.Next() {
		var i OrganizationInvite
		if err = i.scan(rows); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q OrgInvitesQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", OrganizationInviteTable, err)
	}

	if _, err = q.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("executing delete query for %s: %w", OrganizationInviteTable, err)
	}
	return nil
}

func (q OrgInvitesQ) UpdateOne(ctx context.Context) (OrganizationInvite, error) {
	query, args, err := q.updater.Suffix("RETURNING " + OrganizationInviteColumns).ToSql()
	if err != nil {
		return OrganizationInvite{}, fmt.Errorf("building update query for %s: %w", OrganizationInviteTable, err)
	}

	var out OrganizationInvite
	if err = out.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return OrganizationInvite{}, err
	}
	return out, nil
}

func (q OrgInvitesQ) UpdateMany(ctx context.Context) (int64, error) {
	query, args, err := q.updater.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building update query for %s: %w", OrganizationInviteTable, err)
	}

	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("executing update query for %s: %w", OrganizationInviteTable, err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", OrganizationInviteTable, err)
	}

	return aff, nil
}

func (q OrgInvitesQ) FilterByID(id uuid.UUID) OrgInvitesQ {
	q.selector = q.selector.Where(sq.Eq{"id": id})
	q.counter = q.counter.Where(sq.Eq{"id": id})
	q.updater = q.updater.Where(sq.Eq{"id": id})
	q.deleter = q.deleter.Where(sq.Eq{"id": id})
	return q
}

func (q OrgInvitesQ) FilterByOrganizationID(id uuid.UUID) OrgInvitesQ {
	q.selector = q.selector.Where(sq.Eq{"organization_id": id})
	q.counter = q.counter.Where(sq.Eq{"organization_id": id})
	q.updater = q.updater.Where(sq.Eq{"organization_id": id})
	q.deleter = q.deleter.Where(sq.Eq{"organization_id": id})
	return q
}

func (q OrgInvitesQ) FilterByAccountID(id uuid.UUID) OrgInvitesQ {
	q.selector = q.selector.Where(sq.Eq{"account_id": id})
	q.counter = q.counter.Where(sq.Eq{"account_id": id})
	q.updater = q.updater.Where(sq.Eq{"account_id": id})
	q.deleter = q.deleter.Where(sq.Eq{"account_id": id})
	return q
}

func (q OrgInvitesQ) FilterByStatus(status string) OrgInvitesQ {
	q.selector = q.selector.Where(sq.Eq{"status": status})
	q.counter = q.counter.Where(sq.Eq{"status": status})
	q.updater = q.updater.Where(sq.Eq{"status": status})
	q.deleter = q.deleter.Where(sq.Eq{"status": status})
	return q
}

func (q OrgInvitesQ) FilterExpiresBefore(t time.Time) OrgInvitesQ {
	q.selector = q.selector.Where(sq.Lt{"expires_at": t})
	q.counter = q.counter.Where(sq.Lt{"expires_at": t})
	q.updater = q.updater.Where(sq.Lt{"expires_at": t})
	q.deleter = q.deleter.Where(sq.Lt{"expires_at": t})
	return q
}

func (q OrgInvitesQ) FilterExpiresAfter(t time.Time) OrgInvitesQ {
	q.selector = q.selector.Where(sq.GtOrEq{"expires_at": t})
	q.counter = q.counter.Where(sq.GtOrEq{"expires_at": t})
	q.updater = q.updater.Where(sq.GtOrEq{"expires_at": t})
	q.deleter = q.deleter.Where(sq.GtOrEq{"expires_at": t})
	return q
}

func (q OrgInvitesQ) UpdateStatus(status string) OrgInvitesQ {
	q.updater = q.updater.Set("status", status)
	return q
}

func (q OrgInvitesQ) UpdateExpiresAt(t time.Time) OrgInvitesQ {
	q.updater = q.updater.Set("expires_at", t)
	return q
}

func (q OrgInvitesQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", OrganizationInviteTable, err)
	}

	var n uint
	if err = q.db.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("scanning count for %s: %w", OrganizationInviteTable, err)
	}
	return n, nil
}

func (q OrgInvitesQ) Page(limit uint, offset uint) OrgInvitesQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
