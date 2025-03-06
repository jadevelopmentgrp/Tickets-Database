package database

import (
	"context"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ServiceRatings struct {
	*pgxpool.Pool
}

func newServiceRatings(db *pgxpool.Pool) *ServiceRatings {
	return &ServiceRatings{
		db,
	}
}

func (ServiceRatings) Schema() string {
	return `
CREATE TABLE IF NOT EXISTS service_ratings(
	"guild_id" int8 NOT NULL,
	"ticket_id" int4 NOT NULL,
	"rating" int2 NOT NULL,
	FOREIGN KEY("guild_id", "ticket_id") REFERENCES tickets("guild_id", "id"),
	PRIMARY KEY("guild_id", "ticket_id")
);`
}

func (r *ServiceRatings) Get(ctx context.Context, guildId uint64, ticketId int) (rating uint8, ok bool, e error) {
	query := `SELECT "rating" from service_ratings WHERE "guild_id" = $1 AND "ticket_id" = $2;`

	err := r.QueryRow(ctx, query, guildId, ticketId).Scan(&rating)
	if err == nil {
		return rating, true, nil
	} else if err == pgx.ErrNoRows {
		return 0, false, nil
	} else {
		return 0, false, err
	}
}

func (r *ServiceRatings) GetCount(ctx context.Context, guildId uint64) (count int, err error) {
	query := `SELECT COUNT(*) from service_ratings WHERE "guild_id" = $1;`
	err = r.QueryRow(ctx, query, guildId).Scan(&count)
	return
}

func (r *ServiceRatings) GetCountClaimedBy(ctx context.Context, guildId, userId uint64) (count int, err error) {
	query := `
SELECT COUNT(service_ratings.rating)
FROM service_ratings
INNER JOIN ticket_claims
ON service_ratings.guild_id = ticket_claims.guild_id AND service_ratings.ticket_id = ticket_claims.ticket_id
WHERE service_ratings.guild_id = $1 AND ticket_claims.user_id = $2;
`

	err = r.QueryRow(ctx, query, guildId, userId).Scan(&count)
	return
}

// TODO: Materialized view?
func (r *ServiceRatings) GetAverage(ctx context.Context, guildId uint64) (average float32, err error) {
	// Returns NULL if no ratings
	var f *float32

	query := `SELECT AVG(rating) from service_ratings WHERE "guild_id" = $1;`
	err = r.QueryRow(ctx, query, guildId).Scan(&f)
	if f != nil {
		average = *f
	}

	return
}

// TODO: Materialized view?
func (r *ServiceRatings) GetAverageClaimedBy(ctx context.Context, guildId, userId uint64) (average float32, err error) {
	// Returns NULL if no tickets claimed
	var f *float32

	query := `
SELECT AVG(service_ratings.rating)
FROM service_ratings
INNER JOIN ticket_claims
ON service_ratings.guild_id = ticket_claims.guild_id AND service_ratings.ticket_id = ticket_claims.ticket_id
WHERE service_ratings.guild_id = $1 AND ticket_claims.user_id = $2;
`

	err = r.QueryRow(ctx, query, guildId, userId).Scan(&f)
	if f != nil {
		average = *f
	}

	return
}

func (r *ServiceRatings) GetMulti(ctx context.Context, guildId uint64, ticketIds []int) (map[int]uint8, error) {
	query := `SELECT "ticket_id", "rating" from service_ratings WHERE "guild_id" = $1 AND "ticket_id" = ANY($2);`

	idArray := &pgtype.Int4Array{}
	if err := idArray.Set(ticketIds); err != nil {
		return nil, err
	}

	ratings := make(map[int]uint8)

	rows, err := r.Query(ctx, query, guildId, idArray)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var ticketId int
		var rating uint8

		if err := rows.Scan(&ticketId, &rating); err != nil {
			return nil, err
		}

		ratings[ticketId] = rating
	}

	return ratings, nil
}

// [lower,upper]
func (r *ServiceRatings) GetRange(ctx context.Context, guildId uint64, lowerId, upperId int) (map[int]uint8, error) {
	query := `SELECT "ticket_id", "rating" from service_ratings WHERE "guild_id" = $1 AND "ticket_id" >= $2 AND "ticket_id" <= 3;`

	ratings := make(map[int]uint8)

	rows, err := r.Query(ctx, query, guildId, lowerId, upperId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var ticketId int
		var rating uint8

		if err := rows.Scan(&ticketId, &rating); err != nil {
			return nil, err
		}

		ratings[ticketId] = rating
	}

	return ratings, nil
}

func (r *ServiceRatings) ImportBulk(ctx context.Context, guildId uint64, ratings map[int]uint8) (err error) {
	rows := make([][]interface{}, 0)

	for ticketId, rating := range ratings {
		rows = append(rows, []interface{}{guildId, ticketId, rating})
	}

	_, err = r.CopyFrom(ctx, pgx.Identifier{"service_ratings"}, []string{"guild_id", "ticket_id", "rating"}, pgx.CopyFromRows(rows))

	return
}

func (r *ServiceRatings) Set(ctx context.Context, guildId uint64, ticketId int, rating uint8) (err error) {
	query := `
INSERT INTO service_ratings("guild_id", "ticket_id", "rating")
VALUES($1, $2, $3)
ON CONFLICT("guild_id", "ticket_id") DO UPDATE SET "rating" = $3;`

	_, err = r.Exec(ctx, query, guildId, ticketId, rating)
	return
}
