package brows_test

import (
	//"brows"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/beanscc/brows"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	// init db
	mysqlConf := &mysql.Config{
		Addr:         "0.0.0.0:3306",
		Net:          "tcp",
		User:         "root",
		Passwd:       "root",
		DBName:       "test",
		Timeout:      3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		ParseTime:    true,
	}

	if err := Conn(mysqlConf); err != nil {
		panic(err)
	}

	// 设置 brows db object
	brows.SetDB(db)

	// set log
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// Conn 根据配置连接 mysql
func Conn(c *mysql.Config) error {
	dsn := c.FormatDSN()
	fmt.Printf("\n\ndsn: %v\n\n", dsn)
	mysqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err := mysqlDB.Ping(); err != nil {
		return err
	}

	db = mysqlDB

	return nil
}

type App struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	AppID     string    `db:"app_id"`
	Secret    string    `db:"secret"`
	Sign      string    `db:"sign"`
	Status    bool      `db:"status"`
	EndTime   int64     `db:"end_time"`
	StartTime int64     `db:"start_time"`
	Ctime     time.Time `db:"ctime"`
	Utime     time.Time `db:"utime"`
}

// go test -v -run TestScanStruct
func TestScanStruct(t *testing.T) {
	// // =============  use case 1 ============
	sqlStr := "select id, name, app_id, secret, sign, start_time, end_time, status, ctime, utime from app order by id desc limit 1"

	rows, err := brows.PrepareQuery(sqlStr)
	if err != nil {
		log.Panicf("brows query err. err: %v", err)
	}

	var app App
	if err := brows.ScanStruct(rows, &app); err != nil {
		t.Errorf("brows.ScanStruct err. err: %v", err)
	}

	t.Logf("ret app: %#v", app)

}

// go test -v -run TestScanSlice
func TestScanSlice(t *testing.T) {
	sqlStr2 := "select id, name, app_id, secret, sign, start_time, end_time, status, ctime, utime from app order by id desc" // limit 1

	ormDB := brows.NewOrmDB(brows.DB())

	var apps []App
	if err := ormDB.ScanSlice(&apps, sqlStr2); err != nil {
		t.Errorf("brows.ScanSlice err. err: %v", err)
	}

	t.Logf("ret apps: %#v", apps)
}
