package db

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
)

// listTablesAndDump dumps each table in the SQLite database to a separate CSV file.
func ListTablesAndDump(dbPath string, exportDir string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// List all tables
	query := `SELECT name FROM sqlite_master WHERE type='table';`
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tables = append(tables, name)
	}

	_ = os.RemoveAll(exportDir)

	// Create a fresh export directory
	if err := os.Mkdir(exportDir, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %v", err)
	}

	// Dump each table to a CSV file
	for _, table := range tables {
		fmt.Printf("Dumping table: %s\n", table)
		outputPath := fmt.Sprintf("%s/%s.csv", exportDir, table)
		if err := dumpTableToCSV(db, table, outputPath); err != nil {
			return err
		}
	}

	return nil
}

// dumpTableToCSV queries a table and writes its content to a CSV file.
func dumpTableToCSV(db *sql.DB, tableName, outputPath string) error {
	// SQLite (and most SQL databases) don't support parameterized table names or column names.
	// Parameters can only be used where you would otherwise place a value, such as in the WHERE clause.
	rows, err := db.Query("SELECT * FROM " + tableName)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header row
	if err := writer.Write(cols); err != nil {
		return err
	}

	// Write the data rows
	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for rows.Next() {
		for i := range cols {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		record := make([]string, len(cols))
		for i, col := range values {
			if col != nil {
				record[i] = fmt.Sprintf("%v", col)
			}
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
