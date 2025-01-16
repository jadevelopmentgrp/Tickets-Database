package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Blacklist struct {
	*pgxpool.Pool
}

func newBlacklist(db *pgxpool.Pool) *Blacklist {
	return &Blacklist{
		db,
	}
}

func (b Blacklist) Schema() string {
	return `CREATE TABLE IF NOT EXISTS blacklist("guild_id" int8 NOT NULL, "user_id" int8 NOT NULL, PRIMARY KEY("guild_id", "user_id"));`
}

func (b *Blacklist) IsBlacklisted(ctx context.Context, guildId, userId uint64) (exists bool, e error) {
	query := `SELECT EXISTS(SELECT 1 FROM blacklist WHERE "guild_id"=$1 AND "user_id"=$2);`
	if err := b.QueryRow(ctx, query, guildId, userId).Scan(&exists); err != nil {
		e = err
	}

	return
}

func (b *Blacklist) GetBlacklistedUsers(ctx context.Context, guildId uint64, limit, offset int) (blacklisted []uint64, e error) {
	query := `
SELECT "user_id"
FROM blacklist
WHERE "guild_id" = $1
LIMIT $2
OFFSET $3;`

	rows, err := b.Query(ctx, query, guildId, limit, offset)
	defer rows.Close()
	if err != nil {
		e = err
		return
	}

	for rows.Next() {
		var userId uint64
		if err := rows.Scan(&userId); err != nil {
			e = err
			continue
		}

		blacklisted = append(blacklisted, userId)
	}

	return
}

func (b *Blacklist) GetBlacklistedCount(ctx context.Context, guildId uint64) (count int, err error) {
	query := `SELECT COUNT(*) FROM blacklist WHERE "guild_id" = $1;`

	err = b.QueryRow(ctx, query, guildId).Scan(&count)
	return
}

func (b *Blacklist) Add(ctx context.Context, guildId, userId uint64) (err error) {
	// on conflict, user is already blacklisted
	query := `INSERT INTO blacklist("guild_id", "user_id") VALUES($1, $2) ON CONFLICT DO NOTHING;`
	_, err = b.Exec(ctx, query, guildId, userId)
	return
}

func (b *Blacklist) Remove(ctx context.Context, guildId, userId uint64) (err error) {
	query := `DELETE FROM blacklist WHERE "guild_id"=$1 AND "user_id"=$2;`
	_, err = b.Exec(ctx, query, guildId, userId)
	return
}
