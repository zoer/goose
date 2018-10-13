package goose

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// MigrationRecord struct.
type MigrationRecord struct {
	VersionID int64
	TStamp    time.Time
	IsApplied bool // was this a result of up() or down()
}

// Migration struct.
type Migration struct {
	Version    int64
	Next       int64 // next version, or -1 if none
	Previous   int64 // previous version, -1 if none
	Name       string
	Registered bool
	UpFn       func(*sql.Tx) error // Up go migration function
	DownFn     func(*sql.Tx) error // Down go migration function
}

func (m *Migration) String() string {
	return fmt.Sprintf("%d_%s", m.Version, m.Name)
}

// Up runs an up migration.
func (m *Migration) Up(db *sql.DB) error {
	if err := m.run(db, true); err != nil {
		return err
	}
	log.Println("OK   ", m.String())
	return nil
}

// Down runs a down migration.
func (m *Migration) Down(db *sql.DB) error {
	if err := m.run(db, false); err != nil {
		return err
	}
	log.Println("OK   ", m.String())
	return nil
}

func (m *Migration) run(db *sql.DB, direction bool) error {
	if !m.Registered {
		log.Fatalf("failed to apply migration %q: functions must be registered and built into a custom binary ", m)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("db.Begin: ", err)
	}

	fn := m.UpFn
	if !direction {
		fn = m.DownFn
	}
	if fn != nil {
		if err := fn(tx); err != nil {
			tx.Rollback()
			log.Fatalf("FAIL %s (%v), quitting migration.", m, err)
			return err
		}
	}
	if _, err := tx.Exec(GetDialect().insertVersionSQL(), m.Version, direction); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// NumericComponent looks for migration scripts with names in the form:
// XXX_descriptivename.ext where XXX specifies the version number
// and ext specifies the type of migration
func NumericComponent(name string) (int64, error) {

	base := filepath.Base(name)

	if ext := filepath.Ext(base); ext != ".go" && ext != ".sql" {
		return 0, errors.New("not a recognized migration file type")
	}

	idx := strings.Index(base, "_")
	if idx < 0 {
		return 0, errors.New("no separator found")
	}

	n, e := strconv.ParseInt(base[:idx], 10, 64)
	if e == nil && n <= 0 {
		return 0, errors.New("migration IDs must be greater than zero")
	}

	return n, e
}
