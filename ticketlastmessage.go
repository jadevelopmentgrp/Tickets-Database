package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TicketLastMessageTable struct {
	*pgxpool.Pool
}

type TicketLastMessage struct {
	LastMessageId   *uint64    `json:"last_message_id"`
	LastMessageTime *time.Time `json:"last_message_time"`
	UserId          *uint64    `json:"last_message_user_id"`
	UserIsStaff     *bool      `json:"last_message_user_is_staff"`
}

func newTicketLastMessageTable(db *pgxpool.Pool) *TicketLastMessageTable {
	return &TicketLastMessageTable{
		db,
	}
}

func (m TicketLastMessageTable) Schema() string {
	return `
CREATE TABLE IF NOT EXISTS ticket_last_message(
	"guild_id" int8 NOT NULL,
	"ticket_id" int4 NOT NULL,
	"last_message_id" int8,
	"last_message_time" timestamptz,
    "user_id" int8,
	"user_is_staff" bool NOT NULL,
	FOREIGN KEY("guild_id", "ticket_id") REFERENCES tickets("guild_id", "id"),
	PRIMARY KEY("guild_id", "ticket_id")
);
`
}

func (m *TicketLastMessageTable) Get(ctx context.Context, guildId uint64, ticketId int) (lastMessage TicketLastMessage, e error) {
	query := `
SELECT "last_message_id", "last_message_time", "user_id", "user_is_staff"
FROM ticket_last_message
WHERE "guild_id" = $1 AND "ticket_id" = $2;`

	if err := m.QueryRow(ctx, query, guildId, ticketId).Scan(
		&lastMessage.LastMessageId,
		&lastMessage.LastMessageTime,
		&lastMessage.UserId,
		&lastMessage.UserIsStaff,
	); err != nil && err != pgx.ErrNoRows { // defaults to nil if no rows
		e = err
	}

	return
}

func (m *TicketLastMessageTable) ImportBulk(ctx context.Context, guildId uint64, lastMessages map[int]TicketLastMessage) (err error) {
	rows := make([][]interface{}, 0)

	for i, msg := range lastMessages {
		rows = append(rows, []interface{}{
			guildId,
			i,
			msg.LastMessageId,
			msg.LastMessageTime,
			msg.UserId,
			msg.UserIsStaff,
		})
	}

	_, err = m.CopyFrom(ctx, pgx.Identifier{"ticket_last_message"}, []string{"guild_id", "ticket_id", "last_message_id", "last_message_time", "user_id", "user_is_staff"}, pgx.CopyFromRows(rows))
	return
}

func (m *TicketLastMessageTable) Set(ctx context.Context, guildId uint64, ticketId int, messageId, userId uint64, userIsStaff bool) (err error) {
	query := `
INSERT INTO ticket_last_message("guild_id", "ticket_id", "last_message_id", "last_message_time", "user_id", "user_is_staff")
VALUES($1, $2, $3, NOW(), $4, $5) ON CONFLICT("guild_id", "ticket_id")
DO UPDATE SET "last_message_id" = $3, "last_message_time" = NOW(), "user_id" = $4, "user_is_staff" = $5;`

	_, err = m.Exec(ctx, query, guildId, ticketId, messageId, userId, userIsStaff)
	return
}

func (m *TicketLastMessageTable) Delete(ctx context.Context, guildId uint64, ticketId int) (err error) {
	query := `DELETE FROM ticket_last_message WHERE "guild_id"=$1 AND "ticket_id"=$2;`
	_, err = m.Exec(ctx, query, guildId, ticketId)
	return
}
