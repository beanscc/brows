package brows

import (
	"database/sql"
	"fmt"
)

// PrepareQuery prepare query
func PrepareQuery(sqlStr string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := DB().Prepare(sqlStr)
	defer stmt.Close()
	if err != nil {
		return nil, fmt.Errorf("orm: db.Prepare err. err: %v", err)
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("orm: db.stmt.Query err. err: %v", err)
	}

	return rows, nil
}

// Query db.query
func (db *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := db.DB.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return nil, fmt.Errorf("orm: db.Prepare err. err: %v", err)
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("orm: db.stmt.Query err. err: %v", err)
	}

	return rows, nil
}
