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

const ProfileTable = "profiles"

const ProfileColumns = "account_id, username, official, pseudonym, updated_at, created_at"
const ProfileColumnsP = "p.account_id, p.username, p.official, p.pseudonym, p.updated_at, p.created_at"

type Profile struct {
	AccountID uuid.UUID `json:"account_id"`
	Username  string    `json:"username"`
	Official  bool      `json:"official"`
	Pseudonym *string   `json:"pseudonym,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Profile) scan(row sq.RowScanner) error {
	err := row.Scan(
		&p.AccountID,
		&p.Username,
		&p.Official,
		&p.Pseudonym,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("scanning profile: %w", err)
	}
	return nil
}

type ProfilesQ struct {
	db       pgx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewProfilesQ(db pgx.DBTX) ProfilesQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return ProfilesQ{
		db:       db,
		selector: builder.Select(ProfileColumnsP).From(ProfileTable + " p"),
		inserter: builder.Insert(ProfileTable),
		updater:  builder.Update(ProfileTable + " p"),
		deleter:  builder.Delete(ProfileTable + " p"),
		counter:  builder.Select("COUNT(*)").From(ProfileTable + " p"),
	}
}

type ProfileInsertInput struct {
	AccountID uuid.UUID
	Username  string
	Official  bool
	Pseudonym *string
}

func (q ProfilesQ) Insert(ctx context.Context, data ProfileInsertInput) (Profile, error) {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"account_id": data.AccountID,
		"username":   data.Username,
		"official":   data.Official,
		"pseudonym":  data.Pseudonym,
	}).Suffix("RETURNING " + ProfileColumns).ToSql()
	if err != nil {
		return Profile{}, fmt.Errorf("building insert query for %s: %w", ProfileTable, err)
	}

	var inserted Profile
	if err = inserted.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return Profile{}, err
	}
	return inserted, nil
}

type ProfileUpsertInput struct {
	AccountID uuid.UUID
	Username  string
	Official  bool
	Pseudonym *string
}

func (q ProfilesQ) Upsert(ctx context.Context, data ProfileUpsertInput) (Profile, error) {
	query, args, err := q.inserter.
		SetMap(map[string]interface{}{
			"account_id": data.AccountID,
			"username":   data.Username,
			"official":   data.Official,
			"pseudonym":  data.Pseudonym,
		}).
		Suffix(`
			ON CONFLICT (account_id) DO UPDATE SET
				username  = EXCLUDED.username,
				official  = EXCLUDED.official,
				pseudonym = EXCLUDED.pseudonym,
				updated_at = (now() at time zone 'utc')
			RETURNING ` + ProfileColumns,
		).
		ToSql()

	if err != nil {
		return Profile{}, fmt.Errorf("building upsert query for %s: %w", ProfileTable, err)
	}

	var result Profile
	if err = result.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return Profile{}, err
	}

	return result, nil
}

func (q ProfilesQ) Get(ctx context.Context) (Profile, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return Profile{}, fmt.Errorf("building select query for %s: %w", ProfileTable, err)
	}

	var p Profile
	if err = p.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return Profile{}, nil
		default:
			return Profile{}, err
		}
	}
	return p, nil
}

func (q ProfilesQ) Select(ctx context.Context) ([]Profile, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", ProfileTable, err)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", ProfileTable, err)
	}
	defer rows.Close()

	var out []Profile
	for rows.Next() {
		var p Profile
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

func (q ProfilesQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", ProfileTable, err)
	}

	_, err = q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing delete query for %s: %w", ProfileTable, err)
	}

	return nil
}

func (q ProfilesQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", ProfileTable, err)
	}

	var count uint
	if err = q.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("scanning count for %s: %w", ProfileTable, err)
	}

	return count, nil
}

func (q ProfilesQ) UpdateOne(ctx context.Context) (Profile, error) {
	q.updater = q.updater.Set("updated_at", time.Now().UTC())

	query, args, err := q.updater.Suffix("RETURNING " + ProfileColumns).ToSql()
	if err != nil {
		return Profile{}, fmt.Errorf("building update query for %s: %w", ProfileTable, err)
	}

	var updated Profile
	if err = updated.scan(q.db.QueryRowContext(ctx, query, args...)); err != nil {
		return Profile{}, err
	}
	return updated, nil
}

func (q ProfilesQ) UpdateMany(ctx context.Context) (int64, error) {
	q.updater = q.updater.Set("updated_at", time.Now().UTC())

	query, args, err := q.updater.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building update query for %s: %w", ProfileTable, err)
	}

	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("executing update query for %s: %w", ProfileTable, err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", ProfileTable, err)
	}

	return aff, nil
}

func (q ProfilesQ) FilterByAccountID(accountID uuid.UUID) ProfilesQ {
	q.selector = q.selector.Where(sq.Eq{"p.account_id": accountID})
	q.counter = q.counter.Where(sq.Eq{"p.account_id": accountID})
	q.updater = q.updater.Where(sq.Eq{"p.account_id": accountID})
	q.deleter = q.deleter.Where(sq.Eq{"p.account_id": accountID})
	return q
}

func (q ProfilesQ) FilterByUsername(username string) ProfilesQ {
	q.selector = q.selector.Where(sq.Eq{"p.username": username})
	q.counter = q.counter.Where(sq.Eq{"p.username": username})
	q.updater = q.updater.Where(sq.Eq{"p.username": username})
	q.deleter = q.deleter.Where(sq.Eq{"p.username": username})
	return q
}

func (q ProfilesQ) FilterOfficial(official bool) ProfilesQ {
	q.selector = q.selector.Where(sq.Eq{"p.official": official})
	q.counter = q.counter.Where(sq.Eq{"p.official": official})
	q.updater = q.updater.Where(sq.Eq{"p.official": official})
	q.deleter = q.deleter.Where(sq.Eq{"p.official": official})
	return q
}

func (q ProfilesQ) FilterLikeUsername(username string) ProfilesQ {
	q.selector = q.selector.Where(sq.ILike{"p.username": "%" + username + "%"})
	q.counter = q.counter.Where(sq.ILike{"p.username": "%" + username + "%"})
	q.updater = q.updater.Where(sq.ILike{"p.username": "%" + username + "%"})
	q.deleter = q.deleter.Where(sq.ILike{"p.username": "%" + username + "%"})
	return q
}

func (q ProfilesQ) FilterLikePseudonym(pseudonym string) ProfilesQ {
	q.selector = q.selector.Where(sq.ILike{"p.pseudonym": "%" + pseudonym + "%"})
	q.counter = q.counter.Where(sq.ILike{"p.pseudonym": "%" + pseudonym + "%"})
	q.updater = q.updater.Where(sq.ILike{"p.pseudonym": "%" + pseudonym + "%"})
	q.deleter = q.deleter.Where(sq.ILike{"p.pseudonym": "%" + pseudonym + "%"})
	return q
}

func (q ProfilesQ) UpdateUsername(username string) ProfilesQ {
	q.updater = q.updater.Set("username", username)
	return q
}

func (q ProfilesQ) UpdateOfficial(official bool) ProfilesQ {
	q.updater = q.updater.Set("official", official)
	return q
}

func (q ProfilesQ) UpdatePseudonym(pseudonym *string) ProfilesQ {
	q.updater = q.updater.Set("pseudonym", pseudonym)
	return q
}

func (q ProfilesQ) CursorCreatedAt(limit uint, asc bool, createdAt time.Time, accountID uuid.UUID) ProfilesQ {
	if asc {
		q.selector = q.selector.OrderBy("p.created_at ASC", "p.account_id ASC")
	} else {
		q.selector = q.selector.OrderBy("p.created_at DESC", "p.account_id DESC")
	}

	q.selector = q.selector.Limit(uint64(limit))

	if asc {
		q.selector = q.selector.Where(sq.Expr("(p.created_at, p.account_id) > (?, ?)", createdAt, accountID))
	} else {
		q.selector = q.selector.Where(sq.Expr("(p.created_at, p.account_id) < (?, ?)", createdAt, accountID))
	}

	return q
}
