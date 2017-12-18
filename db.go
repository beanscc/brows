package brows

import (
	"database/sql"
	"fmt"
	"sync"
)

var dbObj *sql.DB

// SetDB only set once
func SetDB(db *sql.DB) {
	var onece sync.Once
	onece.Do(func() {
		dbObj = db
	})
}

func ResetDB(db *sql.DB) {
	dbObj = db
}

func DB() *sql.DB {
	return dbObj
}

type Database struct {
	DB *sql.DB
}

func NewOrmDB(db *sql.DB) *Database {
	return &Database{
		DB: db,
	}
}

// ScanStruct db.ScanStruct 对应 queryRow
func (db *Database) ScanStruct(dest interface{}, query string, args ...interface{}) error {
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return err
	}

	if err := ScanStruct(rows, dest); err != nil {
		return fmt.Errorf("orm: ScanStruct err. err: %v", err)
	}

	return nil
}

// ScanSlice db.ScanSlice 对应 query
func (db *Database) ScanSlice(dest interface{}, query string, args ...interface{}) error {
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return err
	}

	if err := ScanSlice(rows, dest); err != nil {
		return fmt.Errorf("orm: ScanSlice err. err: %v", err)
	}

	return nil
}
