package database

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

type RoleBlacklist struct {
	*pgxpool.Pool
}

func newRoleBlacklist(db *pgxpool.Pool) *RoleBlacklist {
	return &RoleBlacklist{
		db,
	}
}

func (b RoleBlacklist) Schema() string {
	return `CREATE TABLE IF NOT EXISTS role_blacklist("guild_id" int8 NOT NULL, "role_id" int8 NOT NULL, PRIMARY KEY("guild_id", "role_id"));`
}

func (b *RoleBlacklist) IsBlacklisted(ctx context.Context, guildId, roleId uint64) (blacklisted bool, e error) {
	query := `SELECT EXISTS(SELECT 1 FROM role_blacklist WHERE "guild_id"=$1 AND "role_id"=$2);`
	if err := b.QueryRow(ctx, query, guildId, roleId).Scan(&blacklisted); err != nil {
		e = err
	}

	return
}

func (b *RoleBlacklist) IsAnyBlacklisted(ctx context.Context, guildId uint64, roles []uint64) (blacklisted bool, e error) {
	query := `SELECT EXISTS(SELECT 1 FROM role_blacklist WHERE "guild_id"=$1 AND "role_id"=ANY($2));`

	array := &pgtype.Int8Array{}
	if err := array.Set(roles); err != nil {
		return false, err
	}

	if err := b.QueryRow(ctx, query, guildId, array).Scan(&blacklisted); err != nil {
		e = err
	}

	return
}

func (b *RoleBlacklist) GetBlacklistedRoles(ctx context.Context, guildId uint64) (roles []uint64, e error) {
	query := `SELECT "role_id" FROM role_blacklist WHERE "guild_id" = $1;`

	rows, err := b.Query(ctx, query, guildId)
	defer rows.Close()
	if err != nil {
		e = err
		return
	}

	for rows.Next() {
		var roleId uint64
		if err := rows.Scan(&roleId); err != nil {
			e = err
			continue
		}

		roles = append(roles, roleId)
	}

	return
}

func (b *RoleBlacklist) GetBlacklistedCount(ctx context.Context, guildId uint64) (count int, err error) {
	query := `SELECT COUNT(*) FROM role_blacklist WHERE "guild_id" = $1;`

	err = b.QueryRow(ctx, query, guildId).Scan(&count)
	return
}

func (b *RoleBlacklist) Add(ctx context.Context, guildId, roleId uint64) (err error) {
	// on conflict, role is already role_blacklist
	query := `INSERT INTO role_blacklist("guild_id", "role_id") VALUES($1, $2) ON CONFLICT DO NOTHING;`
	_, err = b.Exec(ctx, query, guildId, roleId)
	return
}

func (b *RoleBlacklist) Remove(ctx context.Context, guildId, roleId uint64) (err error) {
	query := `DELETE FROM role_blacklist WHERE "guild_id"=$1 AND "role_id"=$2;`
	_, err = b.Exec(ctx, query, guildId, roleId)
	return
}
