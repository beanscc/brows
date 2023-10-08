package brows

import (
	"database/sql"
)

// QueryRow query row
func QueryRow(db *sql.DB, dest interface{}, query string, args ...interface{}) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return Scan(rows, dest)
}

// Query Query, dest 必须是切片类型的引用
func Query(db *sql.DB, dest interface{}, query string, args ...interface{}) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return ScanSlice(rows, dest)
}
