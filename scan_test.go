package brows

import (
	"reflect"
	"testing"
	"time"
)

func Test_mappingByColumns(t *testing.T) {
	type Person struct {
		Name string `db:"name"`
		Age  *int   `db:"age"`
	}

	// 课程成绩
	type CourseScore struct {
		Math    float64 `db:"math"`
		English float64 `db:"english"`
	}

	type Student struct {
		// 学号
		ID string `db:"id"`
		// 年级
		Grade int `db:"grade"`
		// 班级
		Class int `db:"class"`
		// 班主任
		HeadTeacher *string `db:"head_teacher"`
		// 入学时间
		EntryAt time.Time `db:"entry_at"`
		// 毕业时间
		GraduatedAt *time.Time `db:"graduated_at"`

		// 内嵌 struct
		// 个人信息
		Person

		// 内嵌 *struct
		// 课程成绩, 考试时候才有
		*CourseScore
	}

	columns := []string{
		"id",

		"name",
		"age",

		// "grade",
		// "class",
		// "entry_at",
		// "graduated_at",
		//
		// // "head_teacher",
		//
		// "math",
		"english",
	}

	dest := &Student{}
	t.Logf("before mapping dest:%#v", dest)
	got := mapping(reflect.TypeOf(dest), "db")
	i := 0
	for k, v := range got {
		t.Logf("mapping column:%16s, idx:%2d, field: %#v", k, i, v)
		i++
	}
	t.Logf("after mapping dest:%#v", dest)

	fs := mappingByColumns(columns, reflect.ValueOf(dest))
	for i, v := range fs {
		t.Logf("mappingByColumns column:%16s, idx:%2d, field: %#v", columns[i], i, v)
	}
	t.Logf("after mapping dest:%#v", dest)
}
