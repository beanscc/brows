package brows

import (
	"reflect"
	"testing"
	"time"
)

func Test_mapColumns(t *testing.T) {
	type testAppStatus int

	type testAppTime struct {
		EndTime   int64 `db:"end_time"`
		StartTime int64 `db:"start_time"`
	}

	type testApp struct {
		ID     int64         `db:"id"`
		Name   string        `db:"name"`
		AppID  string        `db:"app_id"`
		Secret string        `db:"secret"`
		Sign   string        `db:"sign"`
		Status testAppStatus `db:"status"`
		// EndTime   int64     `db:"end_time"`
		// StartTime int64     `db:"start_time"`

		// 内嵌类型
		*testAppTime

		Ctime    time.Time `db:"ctime"`
		Utime    time.Time `db:"utime"`
		Operator string    `db:"operator"`
	}

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

	dest := &testApp{}
	e := reflect.ValueOf(dest)
	got := mapColumns(columns, e)
	i := 0
	for k, v := range got {
		t.Logf("Test_mapColumns idx:%2d, k:%v, field:%#v", i, k, v)
		i++
	}
}
