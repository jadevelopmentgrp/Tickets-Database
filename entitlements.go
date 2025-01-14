package database

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jadevelopmentgrp/Tickets-Utilities/model"
)

type Entitlements struct {
	*pgxpool.Pool
}

var (
	//go:embed sql/entitlements/schema.sql
	entitlementsSchema string

	//go:embed sql/entitlements/list_from_source.sql
	entitlementsListFromSource string

	//go:embed sql/entitlements/create.sql
	entitlementsCreate string

	//go:embed sql/entitlements/get_by_id.sql
	entitlementsGetById string

	//go:embed sql/entitlements/delete_by_id.sql
	entitlementsDeleteById string

	//go:embed sql/entitlements/get_guild_tiers.sql
	entitlementsGetGuildTiers string

	//go:embed sql/entitlements/list_guild_subscriptions.sql
	entitlementsListGuildSubscriptions string

	//go:embed sql/entitlements/list_user_subscriptions.sql
	entitlementsListUserSubscriptions string

	//go:embed sql/entitlements/list_all_user_subscriptions.sql
	entitlementsListAllUserSubscriptions string

	//go:embed sql/entitlements/increase_expiry.sql
	entitlementsIncreaseExpiry string
)

func newEntitlementsTable(db *pgxpool.Pool) *Entitlements {
	return &Entitlements{
		db,
	}
}

func (Entitlements) Schema() string {
	return entitlementsSchema
}

func (e *Entitlements) ListFromSource(ctx context.Context, source model.EntitlementSource) ([]model.Entitlement, error) {
	rows, err := e.Query(ctx, entitlementsListFromSource, source)
	if err != nil {
		return nil, err
	}

	var entitlements []model.Entitlement
	for rows.Next() {
		var entitlement model.Entitlement
		if err := rows.Scan(
			&entitlement.Id,
			&entitlement.GuildId,
			&entitlement.UserId,
			&entitlement.SkuId,
			&entitlement.Source,
			&entitlement.ExpiresAt,
		); err != nil {
			return nil, err
		}

		entitlements = append(entitlements, entitlement)
	}

	return entitlements, nil
}

func (e *Entitlements) Create(
	ctx context.Context,
	tx pgx.Tx,
	guildId *uint64,
	userId *uint64,
	skuId uuid.UUID,
	source model.EntitlementSource,
	expiresAt *time.Time,
) (model.Entitlement, error) {
	var id uuid.UUID
	if err := tx.QueryRow(ctx, entitlementsCreate, guildId, userId, skuId, source, expiresAt).Scan(&id); err != nil {
		return model.Entitlement{}, err
	}

	return model.Entitlement{
		Id:        id,
		GuildId:   guildId,
		UserId:    userId,
		SkuId:     skuId,
		Source:    source,
		ExpiresAt: expiresAt,
	}, nil
}

func (e *Entitlements) GetById(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.Entitlement, error) {
	var entitlement model.Entitlement
	if err := tx.QueryRow(ctx, entitlementsGetById, id).Scan(
		&entitlement.Id,
		&entitlement.GuildId,
		&entitlement.UserId,
		&entitlement.SkuId,
		&entitlement.Source,
		&entitlement.ExpiresAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &entitlement, nil
}

func (e *Entitlements) DeleteById(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, entitlementsDeleteById, id)
	return err
}

func (e *Entitlements) GetGuildTiers(ctx context.Context, guildId, ownerId uint64, gracePeriod time.Duration, includeVoting bool) ([]model.EntitlementTier, error) {
	rows, err := e.Query(ctx, entitlementsGetGuildTiers, guildId, ownerId, gracePeriod, includeVoting)
	if err != nil {
		return nil, err
	}

	var tiers []model.EntitlementTier
	for rows.Next() {
		var tier model.EntitlementTier
		if err := rows.Scan(&tier); err != nil {
			return nil, err
		}

		tiers = append(tiers, tier)
	}

	return tiers, nil
}

func (e *Entitlements) GetGuildMaxTier(ctx context.Context, guildId, ownerId uint64, gracePeriod time.Duration, includeVoting bool) (*model.EntitlementTier, error) {
	tiers, err := e.GetGuildTiers(ctx, guildId, ownerId, gracePeriod, includeVoting)
	if err != nil {
		return nil, err
	}

	if len(tiers) == 0 {
		return nil, nil
	}

	// tiers returns in priority desc order
	return &tiers[0], nil
}

func (e *Entitlements) ListGuildSubscriptions(ctx context.Context, guildId, ownerId uint64, gracePeriod time.Duration) ([]model.GuildEntitlementEntry, error) {
	rows, err := e.Query(ctx, entitlementsListGuildSubscriptions, guildId, ownerId, gracePeriod)
	if err != nil {
		return nil, err
	}

	var entries []model.GuildEntitlementEntry
	for rows.Next() {
		var entry model.GuildEntitlementEntry
		if err := rows.Scan(
			&entry.Id,
			&entry.UserId,
			&entry.Source,
			&entry.ExpiresAt,
			&entry.SkuId,
			&entry.SkuLabel,
			&entry.Tier,
			&entry.SkuPriority,
		); err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (e *Entitlements) ListUserSubscriptions(ctx context.Context, userId uint64, gracePeriod time.Duration) ([]model.GuildEntitlementEntry, error) {
	rows, err := e.Query(ctx, entitlementsListUserSubscriptions, userId, gracePeriod)
	if err != nil {
		return nil, err
	}

	var entries []model.GuildEntitlementEntry
	for rows.Next() {
		var entry model.GuildEntitlementEntry
		if err := rows.Scan(
			&entry.Id,
			&entry.UserId,
			&entry.Source,
			&entry.ExpiresAt,
			&entry.SkuId,
			&entry.SkuLabel,
			&entry.Tier,
			&entry.SkuPriority,
		); err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (e *Entitlements) ListAllUserSubscriptions(ctx context.Context, gracePeriod time.Duration) ([]model.GuildEntitlementEntry, error) {
	// TODO: Pagination
	rows, err := e.Query(ctx, entitlementsListAllUserSubscriptions, gracePeriod)
	if err != nil {
		return nil, err
	}

	var entries []model.GuildEntitlementEntry
	for rows.Next() {
		var entry model.GuildEntitlementEntry
		if err := rows.Scan(
			&entry.Id,
			&entry.UserId,
			&entry.Source,
			&entry.ExpiresAt,
			&entry.SkuId,
			&entry.SkuLabel,
			&entry.Tier,
			&entry.SkuPriority,
		); err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (e *Entitlements) IncreaseExpiry(ctx context.Context, tx pgx.Tx, guildId, userId *uint64, skuId uuid.UUID, source model.EntitlementSource, duration time.Duration) error {
	_, err := tx.Exec(ctx, entitlementsIncreaseExpiry, guildId, userId, skuId, source, duration)
	return err
}
