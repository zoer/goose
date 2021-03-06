package goose

import (
	"database/sql"
)

// Redo rolls back the most recently applied migration, then runs it again.
func Redo(db *sql.DB) error {
	currentVersion, err := GetDBVersion(db)
	if err != nil {
		return err
	}

	migrations, err := CollectMigrations(minVersion, maxVersion)
	if err != nil {
		return err
	}

	current, err := migrations.Current(currentVersion)
	if err != nil {
		return err
	}

	if err := current.Down(db); err != nil {
		return err
	}

	if err := current.Up(db); err != nil {
		return err
	}

	return nil
}
