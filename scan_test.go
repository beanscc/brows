package brows

import (
	"reflect"
	"testing"
	"time"
)

type AppStatus int

type App struct {
	ID     int64     `db:"id"`
	Name   string    `db:"name"`
	AppID  string    `db:"app_id"`
	Secret string    `db:"secret"`
	Sign   string    `db:"sign"`
	Status AppStatus `db:"status"`
	// EndTime   int64     `db:"end_time"`
	// StartTime int64     `db:"start_time"`

	// 结构体
	// AppTimeField *AppTime

	// 内嵌struct
	// AppTime

	// 内嵌 *struct
	// *AppTime

	CTime    time.Time `db:"ctime"`
	UTime    time.Time `db:"utime"`
	Operator string    `db:"operator"`
}

type AppTime struct {
	EndTime   int64 `db:"end_time"`
	StartTime int64 `db:"start_time"`
}

func Test_mapColumns(t *testing.T) {
	columns := []string{
		"id",
		"name",
		"app_id",
		"secret",
		"sign",
		"status",
		"start_time",
		"end_time",
		"ctime",
		"utime",
		"operator",
	}

	dest := &App{}
	e := reflect.ValueOf(dest)
	got := mapColumns(columns, e)
	i := 0
	for k, v := range got {
		t.Logf("Test_mapColumns idx:%2d, k:%v, field:%#v", i, k, v)
		i++
	}
}
