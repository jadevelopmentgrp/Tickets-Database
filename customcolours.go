package database

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type CustomColours struct {
	*pgxpool.Pool
}

func newCustomColours(db *pgxpool.Pool) *CustomColours {
	return &CustomColours{
		db,
	}
}

func (c CustomColours) Schema() string {
	return `
CREATE TABLE IF NOT EXISTS custom_colours(
	"guild_id" int8 NOT NULL,
	"colour_id" int2 NOT NULL,
	"colour_code" int4 NOT NULL,
	PRIMARY KEY("guild_id", "colour_id")
);`
}

func (c *CustomColours) Get(ctx context.Context, guildId uint64, colourId int16) (colourCode int, ok bool, e error) {
	query := `SELECT "colour_code" FROM custom_colours WHERE "guild_id" = $1 AND "colour_id" = $2;`

	if err := c.QueryRow(ctx, query, guildId, colourId).Scan(&colourCode); err == nil {
		ok = true
	} else {
		if err != pgx.ErrNoRows {
			e = err
		}
	}

	return
}

func (c *CustomColours) GetAll(ctx context.Context, guildId uint64) (map[int16]int, error) {
	query := `SELECT "colour_id", "colour_code" FROM custom_colours WHERE "guild_id" = $1;`

	rows, err := c.Query(ctx, query, guildId)
	if err != nil {
		return nil, err
	}

	colours := make(map[int16]int)
	for rows.Next() {
		var colourId int16
		var colourCode int

		if err = rows.Scan(&colourId, &colourCode); err != nil {
			return nil, err
		}

		colours[colourId] = colourCode
	}

	return colours, nil
}

func (c *CustomColours) Set(ctx context.Context, guildId uint64, colourId int16, colourCode int) (err error) {
	query := `
INSERT INTO custom_colours("guild_id", "colour_id", "colour_code")
VALUES($1, $2, $3) ON CONFLICT("guild_id", "colour_id")
DO UPDATE SET "colour_code" = $3;`

	_, err = c.Exec(ctx, query, guildId, colourId, colourCode)
	return
}

// BatchSet colours = map[colour_id]value
func (c *CustomColours) BatchSet(ctx context.Context, guildId uint64, colours map[int16]int) (err error) {
	query := `
INSERT INTO custom_colours("guild_id", "colour_id", "colour_code")
VALUES($1, $2, $3) ON CONFLICT("guild_id", "colour_id")
DO UPDATE SET "colour_code" = $3;`

	batch := &pgx.Batch{}
	for colourId, colourCode := range colours {
		batch.Queue(query, guildId, colourId, colourCode)
	}

	res := c.SendBatch(ctx, batch)
	defer res.Close()

	_, err = res.Exec()
	return
}
