package goose

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

// Create writes a new blank migration file.
func CreateWithTemplate(db *sql.DB, dir string, name string) error {
	version := time.Now().Format("20060102150405")

	filename := fmt.Sprintf("%v_%v.go", version, name)
	fpath := filepath.Join(dir, filename)

	path, err := writeTemplateToFile(fpath, name, version)
	if err != nil {
		return err
	}

	log.Printf("Created new file: %s\n", path)

	return nil
}

// Create writes a new blank migration file.
func Create(db *sql.DB, dir, name string) error {
	return CreateWithTemplate(db, dir, name)
}

func writeTemplateToFile(path, name, version string) (string, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to create file: %v already exists", path)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	err = tmplMigration.Execute(f, map[string]interface{}{
		"version": version,
		"name":    name,
	})
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

var tmplMigration = template.Must(template.New("goose.go-migration").Parse(`package migration

import (
	"database/sql"
	"github.com/zoer/goose"
)

const sqlUp{{.version}} = ` + "``" + `

const sqlDown{{.version}} = ` + "``" + `

func init() {
	goose.AddMigration({{.version}}, "{{.name}}" Up{{.version}}, Down{{.version}})
}

func Up{{.version}}(tx *sql.Tx) error {
	_, err := tx.Exec(sqlUp{{.version}})
	return err
}

func Down{{.version}}(tx *sql.Tx) error {
	_, err := tx.Exec(sqlDown{{.version}})
	return err
}
`))
