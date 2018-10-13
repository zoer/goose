# goose

This fork has only Go registered migrations. Its main purpose to be embedded
into other apps.

# Install

    $ go get -u github.com/zoer/goose/cmd/goose

This will install the `goose` binary to your `$GOPATH/bin` directory.

# Usage

```
Usage: goose [OPTIONS] DRIVER DBSTRING COMMAND

Drivers:
    postgres
    mysql
    sqlite3
    redshift

Commands:
    create NAME [sql|go] Creates new migration file with next version

Options:
    -dir string
        directory with migration files (default ".")

Examples:
    goose sqlite3 ./foo.db create init
    goose sqlite3 ./foo.db create add_some_column
    goose sqlite3 ./foo.db create fetch_user_data

    goose postgres "user=postgres dbname=postgres sslmode=disable" create test
    goose mysql "user:password@/dbname?parseTime=true" create test
    goose redshift "postgres://user:password@qwerty.us-east-1.redshift.amazonaws.com:5439/db" create test
    goose tidb "user:password@/dbname?parseTime=true" create test
```
## Go Migrations

```go
package migrations

import (
	"database/sql"

	"github.com/zoer/goose"
)

const sqlUp20181013102322 = `
CREATE TABLE users(
  id serial PRIMARY KEY,
  name text NOT NULL,
  email text NOT NULL
);
`

const sqlDown20181013102322 = `
DROP TABLE users;
`

func init() {
	goose.AddMigration(20181013102322, "some_name", Up20181013102322, Down20181013102322)
}

func Up20181013102322(tx *sql.Tx) error {
	_, err := tx.Exec(sqlUp20181013102322)
	return err
}

func Down20181013102322(tx *sql.Tx) error {
	_, err := tx.Exec(sqlDown20181013102322)
	return err
}
```

## License

Licensed under [MIT License](./LICENSE)

[GoDoc]: https://godoc.org/github.com/zoer/goose
[GoDoc Widget]: https://godoc.org/github.com/zoer/goose?status.svg
[Travis]: https://travis-ci.org/zoer/goose
[Travis Widget]: https://travis-ci.org/zoer/goose.svg?branch=master
