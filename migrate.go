package goose

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
)

var (
	// ErrNoCurrentVersion when a current migration version is not found.
	ErrNoCurrentVersion = errors.New("no current version found")
	// ErrNoNextVersion when the next migration version is not found.
	ErrNoNextVersion = errors.New("no next version found")
	// MaxVersion is the maximum allowed version.
	MaxVersion int64 = 9223372036854775807 // max(int64)

	registeredGoMigrations = map[int64]*Migration{}
)

// Migrations slice.
type Migrations []*Migration

// helpers so we can use pkg sort
func (ms Migrations) Len() int      { return len(ms) }
func (ms Migrations) Swap(i, j int) { ms[i], ms[j] = ms[j], ms[i] }
func (ms Migrations) Less(i, j int) bool {
	if ms[i].Version == ms[j].Version {
		log.Fatalf("goose: duplicate version %v detected:\n%v\n%v", ms[i].Version, ms[i].Name, ms[j].Name)
	}
	return ms[i].Version < ms[j].Version
}

// Current gets the current migration.
func (ms Migrations) Current(current int64) (*Migration, error) {
	for i, migration := range ms {
		if migration.Version == current {
			return ms[i], nil
		}
	}

	return nil, ErrNoCurrentVersion
}

// Next gets the next migration.
func (ms Migrations) Next(current int64) (*Migration, error) {
	for i, migration := range ms {
		if migration.Version > current {
			return ms[i], nil
		}
	}

	return nil, ErrNoNextVersion
}

// Previous : Get the previous migration.
func (ms Migrations) Previous(current int64) (*Migration, error) {
	for i := len(ms) - 1; i >= 0; i-- {
		if ms[i].Version < current {
			return ms[i], nil
		}
	}

	return nil, ErrNoNextVersion
}

// Last gets the last migration.
func (ms Migrations) Last() (*Migration, error) {
	if len(ms) == 0 {
		return nil, ErrNoNextVersion
	}

	return ms[len(ms)-1], nil
}

func (ms Migrations) String() string {
	str := ""
	for _, m := range ms {
		str += fmt.Sprintln(m)
	}
	return str
}

func AddMigration(version int64, name string, up func(*sql.Tx) error, down func(*sql.Tx) error) {
	migration := &Migration{
		Version:    version,
		Next:       -1,
		Previous:   -1,
		Registered: true,
		UpFn:       up,
		DownFn:     down,
		Name:       name,
	}

	if existing, ok := registeredGoMigrations[version]; ok {
		panic(fmt.Sprintf("failed to add migration %q: version conflicts with %q", name, existing.Name))
	}

	registeredGoMigrations[version] = migration
}

// CollectMigrations returns all the valid looking migration scripts in the
// migrations folder and go func registry, and key them by version.
func CollectMigrations(current, target int64) (Migrations, error) {
	var migrations Migrations

	// Go migrations registered via goose.AddMigration().
	for _, migration := range registeredGoMigrations {
		if versionFilter(migration.Version, current, target) {
			migrations = append(migrations, migration)
		}
	}

	migrations = sortAndConnectMigrations(migrations)

	return migrations, nil
}

func sortAndConnectMigrations(migrations Migrations) Migrations {
	sort.Sort(migrations)

	// now that we're sorted in the appropriate direction,
	// populate next and previous for each migration
	for i, m := range migrations {
		prev := int64(-1)
		if i > 0 {
			prev = migrations[i-1].Version
			migrations[i-1].Next = m.Version
		}
		migrations[i].Previous = prev
	}

	return migrations
}

func versionFilter(v, current, target int64) bool {

	if target > current {
		return v > current && v <= target
	}

	if target < current {
		return v <= current && v > target
	}

	return false
}

// EnsureDBVersion retrieves the current version for this DB.
// Create and initialize the DB version table if it doesn't exist.
func EnsureDBVersion(db *sql.DB) (int64, error) {
	rows, err := GetDialect().dbVersionQuery(db)
	if err != nil {
		return 0, createVersionTable(db)
	}
	defer rows.Close()

	// The most recent record for each migration specifies
	// whether it has been applied or rolled back.
	// The first version we find that has been applied is the current version.

	toSkip := make([]int64, 0)

	for rows.Next() {
		var row MigrationRecord
		if err = rows.Scan(&row.VersionID, &row.IsApplied); err != nil {
			log.Fatal("error scanning rows:", err)
		}

		// have we already marked this version to be skipped?
		skip := false
		for _, v := range toSkip {
			if v == row.VersionID {
				skip = true
				break
			}
		}

		if skip {
			continue
		}

		// if version has been applied we're done
		if row.IsApplied {
			return row.VersionID, nil
		}

		// latest version of migration has not been applied.
		toSkip = append(toSkip, row.VersionID)
	}

	return 0, ErrNoNextVersion
}

// Create the db version table
// and insert the initial 0 value into it
func createVersionTable(db *sql.DB) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}

	d := GetDialect()

	if _, err := txn.Exec(d.createVersionTableSQL()); err != nil {
		txn.Rollback()
		return err
	}

	version := 0
	applied := true
	if _, err := txn.Exec(d.insertVersionSQL(), version, applied); err != nil {
		txn.Rollback()
		return err
	}

	return txn.Commit()
}

// GetDBVersion is an alias for EnsureDBVersion, but returns -1 in error.
func GetDBVersion(db *sql.DB) (int64, error) {
	version, err := EnsureDBVersion(db)
	if err != nil {
		return -1, err
	}

	return version, nil
}
