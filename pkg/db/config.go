package db

import (
	"database/sql"
	"embed"
	_ "github.com/mattn/go-sqlite3"
	"io/fs"
	"log"
)

//go:embed schema.sql
var schemaSQL embed.FS

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	schema, err := fs.ReadFile(schemaSQL, "schema.sql")
	if err != nil {
		log.Fatalf("Failed to read embedded schema.sql: %v", err)
		return nil, err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		log.Printf("Error executing schema: %v", err)
		return nil, err
	}

	pragmas := []string{
		"PRAGMA synchronous = FULL;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA cache_size = -80000;",
		"PRAGMA temp_store = MEMORY;",
	}

	for _, pragma := range pragmas {
		_, err = db.Exec(pragma)
		if err != nil {
			log.Fatalf("Failed to execute %s: %v", pragma, err)
		}
	}

	return db, nil
}
