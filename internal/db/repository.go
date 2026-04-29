package db

import (
	"context"
	"database/sql"
	"fmt"

	"supplygraph/internal/model"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// InsertAsset inserts a new asset into the database and returns its generated ID.
func (r *Repository) InsertAsset(ctx context.Context, asset model.Asset) (string, error) {
	const query = `
		INSERT INTO assets (name, asset_type, source)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id string
	if err := r.db.QueryRowContext(ctx, query, asset.Name, asset.AssetType, asset.Source).Scan(&id); err != nil {
		return "", fmt.Errorf("insert asset: %w", err)
	}

	return id, nil
}

// InsertScan inserts a new scan into the database and returns its generated ID.
func (r *Repository) InsertScan(ctx context.Context, scan model.Scan) (string, error) {
	const query = `
		INSERT INTO scans (asset_id, status, sbom_format, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id string
	if err := r.db.QueryRowContext(
		ctx,
		query,
		scan.AssetID,
		scan.Status,
		scan.SBOMFormat,
		scan.StartedAt,
		scan.CompletedAt,
	).Scan(&id); err != nil {
		return "", fmt.Errorf("insert scan: %w", err)
	}

	return id, nil
}

// FindOrCreateComponent attempts to find a component by its PURL, and if it doesn't exist, it creates a new one. It returns the ID of the found or created component.
func (r *Repository) FindOrCreateComponent(ctx context.Context, component model.Component) (string, error) {
	// // DO NOTHING makes it harder to get the row id back directly
	const query = `
		INSERT INTO components (name, ecosystem, purl)
		VALUES ($1, $2, $3)
		ON CONFLICT (purl)
		DO UPDATE SET
			name = EXCLUDED.name,
			ecosystem = EXCLUDED.ecosystem
		RETURNING id
	`

	var id string
	if err := r.db.QueryRowContext(ctx, query, component.Name, component.Ecosystem, component.PURL).Scan(&id); err != nil {
		return "", fmt.Errorf("find or create component: %w", err)
	}

	return id, nil
}

// FindOrCreateComponentVersion attempts to find a component version by its component ID and version, and if it doesn't exist, it creates a new one. It returns the ID of the found or created component version.
func (r *Repository) FindOrCreateComponentVersion(
	ctx context.Context,
	componentVersion model.ComponentVersion,
) (string, error) {
	// DO NOTHING makes it harder to get the row id back directly
	const query = `
		INSERT INTO component_versions (component_id, version)
		VALUES ($1, $2)
		ON CONFLICT (component_id, version)
		DO UPDATE SET
			version = EXCLUDED.version
		RETURNING id
	`

	var id string
	if err := r.db.QueryRowContext(
		ctx,
		query,
		componentVersion.ComponentID,
		componentVersion.Version,
	).Scan(&id); err != nil {
		return "", fmt.Errorf("find or create component version: %w", err)
	}

	return id, nil
}
