package database

import (
	"context"
	_ "embed"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DashboardUsersTable struct {
	*pgxpool.Pool
}

func newDashboardUsersTable(db *pgxpool.Pool) *DashboardUsersTable {
	return &DashboardUsersTable{
		db,
	}
}

var (
	//go:embed sql/dashboard_users/schema.sql
	dashboardUsersSchema string

	//go:embed sql/dashboard_users/upsert.sql
	dashboardUsersUpsert string

	//go:embed sql/dashboard_users/purge_old_users.sql
	dashboardPurgeOldUsers string
)

func (d *DashboardUsersTable) Schema() string {
	return dashboardUsersSchema
}

func (d *DashboardUsersTable) UpdateLastSeen(ctx context.Context, userId uint64) error {
	_, err := d.Exec(ctx, dashboardUsersUpsert, userId, time.Now())
	return err
}

func (d *DashboardUsersTable) PurgeOldUsers(ctx context.Context, threshold time.Duration) (int64, error) {
	var interval pgtype.Interval
	if err := interval.Set(threshold); err != nil {
		return 0, err
	}

	metadata, err := d.Exec(ctx, dashboardPurgeOldUsers, threshold)
	return metadata.RowsAffected(), err
}
