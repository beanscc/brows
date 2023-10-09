package brows

import (
	"reflect"
	"testing"
)

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

	dest := &TestApp{}
	e := reflect.ValueOf(dest)
	got := mapColumns(columns, e)
	i := 0
	for k, v := range got {
		t.Logf("Test_mapColumns idx:%2d, k:%v, field:%#v", i, k, v)
		i++
	}
}
