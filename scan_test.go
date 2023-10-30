package brows

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func Test_mapping(t *testing.T) {
	// 比较 f1,f2 是否一致，这里不需要指针地址完全一致
	fnStructFieldCompare := func(f1, f2 structField) error {
		if f1.ignore != f2.ignore {
			return errors.New("field: ignore")
		}

		if !reflect.DeepEqual(f1.index, f2.index) {
			return errors.New("field: index")
		}

		if f1.field.Name != f2.field.Name {
			return errors.New("field: field.Name")
		}

		if f1.field.PkgPath != f2.field.PkgPath {
			return errors.New("field: field.PkgPath")
		}

		if !reflect.DeepEqual(f1.field.Type, f2.field.Type) {
			return errors.New("field: field.Type")
		}

		if f1.field.Tag != f2.field.Tag {
			return errors.New("field: field.Tag")
		}

		if f1.field.Offset != f2.field.Offset {
			return errors.New("field: field.Offset")
		}

		if !reflect.DeepEqual(f1.field.Index, f2.field.Index) {
			return errors.New("field: field.Index")
		}

		if f1.field.Anonymous != f2.field.Anonymous {
			return errors.New("field: field.Anonymous")
		}

		return nil
	}

	type Inner1 struct {
		F1 string  `db:"inner1.f1"`
		F2 *string `db:"inner1.f2"`
	}

	type Inner2 struct {
		F1 int  `db:"inner2.f1"`
		F2 *int `db:"inner2.f2"`
	}

	type T struct {
		F1 string  `db:"f1"`
		F2 *string `db:"f2"`

		// ignore by tag
		F3 int `db:"-"`
		F4 int

		// ignore by IsExported
		f5 string
		f6 string `db:"f6"`

		// struct
		F7 struct {
			Inner1 string `db:"f7.i1"`
		}

		Inner1

		*Inner2
	}

	test := []struct {
		rt   reflect.Type
		want map[string]structField
	}{
		// T
		{
			rt: reflect.TypeOf(T{}),
			want: map[string]structField{
				"f1": {
					index: []int{0},
				},
				"f2": {
					index: []int{1},
				},
				"f7.i1": {
					index: []int{6, 0},
				},
				"inner1.f1": {
					index: []int{7, 0},
				},
				"inner1.f2": {
					index: []int{7, 1},
				},

				"inner2.f1": {
					index: []int{8, 0},
				},
				"inner2.f2": {
					index: []int{8, 1},
				},
			},
		},
	}

	for _, tt := range test {
		t.Run("", func(t *testing.T) {
			got := mapping(tt.rt, "db")

			for k, v := range got {
				want, ok := tt.want[k]
				t.Logf("Test_mapping  got[%s]:%#v", k, v)
				want.field = tt.rt.FieldByIndex(want.index)
				t.Logf("Test_mapping want[%s]:%#v", k, want)

				if !ok {
					t.Errorf("Test_mapping got unexpected tag: %s", k)
					continue
				}
				if err := fnStructFieldCompare(got[k], want); err != nil {
					t.Errorf("Test_mapping tag struct not matched. tag:%s, err:%v", k, err)
					continue
				}

				delete(tt.want, k)
			}

			// want 还有数据
			for k := range tt.want {
				t.Errorf("Test_mapping got miss tag: %s in want", k)
			}
		})
	}
}

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

		"grade",
		"class",
		"entry_at",
		"graduated_at",

		// "head_teacher",

		"math",
		"english",
	}

	dest := &Student{}
	t.Logf("before mapping dest:%#v", dest)
	// got := mapping(reflect.TypeOf(dest), "db")
	// i := 0
	// for k, v := range got {
	// 	t.Logf("mapping column:%16s, idx:%2d, field: %#v", k, i, v)
	// 	i++
	// }
	// t.Logf("after mapping dest:%#v", dest)

	fs := mappingByColumns(columns, reflect.ValueOf(dest))
	for i, v := range fs {
		t.Logf("mappingByColumns column:%16s, idx:%2d, field: %#v", columns[i], i, v)
	}
	t.Logf("after mapping dest:%#v", dest)
}
