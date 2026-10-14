package store

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/naputt/code-interview/backend/migrations"
	"gorm.io/gorm"
)

// Migrate applies every embedded *.up.sql migration that has not run yet,
// in filename order, each in its own transaction. Applied versions are
// recorded in schema_migrations.
func Migrate(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	files, err := fs.Glob(migrations.FS, "*.up.sql")
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(file, ".up.sql")

		var applied int64
		if err := db.WithContext(ctx).
			Raw(`SELECT count(*) FROM schema_migrations WHERE version = ?`, version).
			Scan(&applied).Error; err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if applied > 0 {
			continue
		}

		sqlBytes, err := migrations.FS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(sqlBytes)).Error; err != nil {
				return err
			}
			return tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version).Error
		}); err != nil {
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
	}
	return nil
}
