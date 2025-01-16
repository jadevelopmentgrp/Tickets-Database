package database

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ChannelCategory struct {
	*pgxpool.Pool
}

func newChannelCategory(db *pgxpool.Pool) *ChannelCategory {
	return &ChannelCategory{
		db,
	}
}

func (c ChannelCategory) Schema() string {
	return `
CREATE TABLE IF NOT EXISTS channel_category(
	"guild_id" int8 NOT NULL UNIQUE,
	"category_id" int8 NOT NULL UNIQUE,
	PRIMARY KEY("guild_id")
);`
}

func (c *ChannelCategory) Get(ctx context.Context, guildId uint64) (channelCategory uint64, e error) {
	if err := c.QueryRow(ctx, `SELECT "category_id" from channel_category WHERE "guild_id" = $1;`, guildId).Scan(&channelCategory); err != nil && err != pgx.ErrNoRows {
		e = err
	}

	return
}

func (c *ChannelCategory) Set(ctx context.Context, guildId, channelCategory uint64) (err error) {
	_, err = c.Exec(ctx, `INSERT INTO channel_category("guild_id", "category_id") VALUES($1, $2) ON CONFLICT("guild_id") DO UPDATE SET "category_id" = $2;`, guildId, channelCategory)
	return
}

func (c *ChannelCategory) Delete(ctx context.Context, guildId uint64) (err error) {
	_, err = c.Exec(ctx, `DELETE FROM channel_category WHERE "guild_id" = $1;`, guildId)
	return
}

func (c *ChannelCategory) DeleteByChannel(ctx context.Context, channelId uint64) (err error) {
	_, err = c.Exec(ctx, `DELETE FROM channel_category WHERE "category_id" = $1;`, channelId)
	return
}
