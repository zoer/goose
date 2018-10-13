package goose

import (
	"database/sql"
	"fmt"
	"sync"
)

var (
	duplicateCheckOnce sync.Once
	minVersion         = int64(0)
	maxVersion         = int64((1 << 63) - 1)
)

// Run runs a goose command.
func Run(command string, db *sql.DB, dir string, args ...string) error {
	switch command {
	case "create":
		if len(args) == 0 {
			return fmt.Errorf("create must be of form: goose [OPTIONS] DRIVER DBSTRING create NAME")
		}

		if err := Create(db, dir, args[0]); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%q: no such command", command)
	}
	return nil
}
