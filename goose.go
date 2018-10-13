package goose

import (
	"database/sql"
	"fmt"
	"strconv"
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
	case "up":
		if err := Up(db, dir); err != nil {
			return err
		}
	case "up-by-one":
		if err := UpByOne(db, dir); err != nil {
			return err
		}
	case "up-to":
		if len(args) == 0 {
			return fmt.Errorf("up-to must be of form: goose [OPTIONS] DRIVER DBSTRING up-to VERSION")
		}

		version, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("version must be a number (got '%s')", args[0])
		}
		if err := UpTo(db, version); err != nil {
			return err
		}
	case "create":
		if len(args) == 0 {
			return fmt.Errorf("create must be of form: goose [OPTIONS] DRIVER DBSTRING create NAME")
		}

		if err := Create(db, dir, args[0]); err != nil {
			return err
		}
	case "down":
		if err := Down(db); err != nil {
			return err
		}
	case "down-to":
		if len(args) == 0 {
			return fmt.Errorf("down-to must be of form: goose [OPTIONS] DRIVER DBSTRING down-to VERSION")
		}

		version, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("version must be a number (got '%s')", args[0])
		}
		if err := DownTo(db, version); err != nil {
			return err
		}
	case "redo":
		if err := Redo(db); err != nil {
			return err
		}
	case "reset":
		if err := Reset(db); err != nil {
			return err
		}
	case "status":
		if err := Status(db); err != nil {
			return err
		}
	case "version":
		if err := Version(db); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%q: no such command", command)
	}
	return nil
}
