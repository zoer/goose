package goose

import (
	"database/sql"

	"github.com/spf13/cobra"
)

func Command(dialect, dbUri string) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate up/down/redo/status",
		Short: "Migrate Database",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := args[0]
			if err := SetDialect(dialect); err != nil {
				log.Fatal(err)
			}
			db, err := sql.Open(dialect, dbUri)
			if err != nil {
				log.Fatal(err)
			}

			if err := Run(c, db, "."); err != nil {
				log.Fatal(err)
			}
		},
	}
}
