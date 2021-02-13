package store

import (
	"fmt"
	"github.com/jypelle/mifasol/internal/srv/store/migration"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
	"net/http"
)

func (s *Store) migrateDatabase() error {
	logrus.Infof("Updating")

	// Retrieve update scripts to apply
	migrate.SetTable("migration")
	migrationSource := migrate.HttpFileSystemMigrationSource{
		FileSystem: http.FS(migration.Fs),
	}

	updateScripts, err := migrationSource.FindMigrations()
	if err != nil {
		return fmt.Errorf("Can't find update scripts: %w", err)
	}

	if len(updateScripts) == 0 {
		logrus.Info("No update scripts to apply")
	}

	// Apply update scripts
	n, err := migrate.Exec(s.db.DB, "sqlite3", migrationSource, migrate.Up)
	if err != nil {
		return fmt.Errorf("Can't run migrations: %w", err)
	}

	logrus.Infof("%d update scripts applied", n)

	return nil
}
