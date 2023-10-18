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

	type T1 struct {
		Name string `db:"name"`
	}

	type T2 struct {
		Name2 *string `db:"name2"`
	}

	// 内嵌
	type T3 struct {
		ID string `db:"id"`
		T1
	}

	type T4 struct {
		ID string `db:"id"`
		*T1
		T2
	}

	type T5 struct {
		ID string `db:"id"`
		*T2
		*T1
		CreatedAt time.Time  `db:"created_at"`
		DeletedAt *time.Time `db:"deleted_at"`
	}

	test := []struct {
		rt   reflect.Type
		want map[string]structField
	}{
		// T1
		{
			rt: reflect.TypeOf(T1{}),
			want: map[string]structField{
				"name": {
					ignore: false,
					index:  []int{0},
					field:  reflect.TypeOf(T1{}).Field(0),
				},
			},
		},
		// *T1
		{
			rt: reflect.TypeOf(&T1{}),
			want: map[string]structField{
				"name": {
					ignore: false,
					index:  []int{0},
					field:  reflect.TypeOf(&T1{}).Elem().Field(0),
				},
			},
		},

		// field string
		// T2
		{
			rt: reflect.TypeOf(T2{}),
			want: map[string]structField{
				"name2": {
					ignore: false,
					index:  []int{0},
					field:  reflect.TypeOf(T2{}).Field(0),
				},
			},
		},
		// *T2
		{
			rt: reflect.TypeOf(&T2{}),
			want: map[string]structField{
				"name2": {
					ignore: false,
					index:  []int{0},
					field:  reflect.TypeOf(&T2{}).Elem().Field(0),
				},
			},
		},
		// T3
		{
			rt: reflect.TypeOf(T3{}),
			want: map[string]structField{
				"id": {
					ignore: false,
					index:  []int{0},
					field:  reflect.TypeOf(T3{}).Field(0),
				},
				"name": {
					ignore: false,
					index:  []int{1, 0},
					field:  reflect.TypeOf(T3{}).FieldByIndex([]int{1, 0}),
				},
			},
		},
		// *T3
		{
			rt: reflect.TypeOf(&T3{}),
			want: map[string]structField{
				"id": {
					ignore: false,
					index:  []int{0},
					field:  reflect.TypeOf(&T3{}).Elem().Field(0),
				},
				"name": {
					ignore: false,
					index:  []int{1, 0},
					field:  reflect.TypeOf(&T3{}).Elem().FieldByIndex([]int{1, 0}),
				},
			},
		},

		// T4
		{
			rt: reflect.TypeOf(T4{}),
			want: map[string]structField{
				"id": {
					ignore: false,
					index:  []int{0},
					field:  reflect.TypeOf(T4{}).Field(0),
				},
				"name2": {
					ignore: false,
					index:  []int{2, 0},
					field:  reflect.TypeOf(T4{}).FieldByIndex([]int{2, 0}),
				},
				"name": {
					ignore: false,
					index:  []int{1, 0},
					field:  reflect.TypeOf(T4{}).FieldByIndex([]int{1, 0}),
				},
			},
		},
		// *T4
		{
			rt: reflect.TypeOf(&T4{}),
			want: map[string]structField{
				"id": {
					ignore: false,
					index:  []int{0},
					field:  reflect.TypeOf(&T4{}).Elem().Field(0),
				},
				"name2": {
					ignore: false,
					index:  []int{2, 0},
					field:  reflect.TypeOf(&T4{}).Elem().FieldByIndex([]int{2, 0}),
				},
				"name": {
					ignore: false,
					index:  []int{1, 0},
					field:  reflect.TypeOf(&T4{}).Elem().FieldByIndex([]int{1, 0}),
				},
			},
		},

		// *T5
		{
			rt: reflect.TypeOf(&T5{}),
			want: map[string]structField{
				"id": {
					ignore: false,
					index:  []int{0},
					field:  reflect.TypeOf(&T5{}).Elem().Field(0),
				},
				"name2": {
					ignore: false,
					index:  []int{1, 0},
					field:  reflect.TypeOf(&T5{}).Elem().FieldByIndex([]int{1, 0}),
				},
				"name": {
					ignore: false,
					index:  []int{2, 0},
					field:  reflect.TypeOf(&T5{}).Elem().FieldByIndex([]int{2, 0}),
				},
				"created_at": {
					ignore: false,
					index:  []int{3},
					field:  reflect.TypeOf(&T5{}).Elem().FieldByIndex([]int{3}),
				},
				"deleted_at": {
					ignore: false,
					index:  []int{4},
					field:  reflect.TypeOf(&T5{}).Elem().FieldByIndex([]int{4}),
				},
			},
		},
	}

	for _, tt := range test {
		t.Run("", func(t *testing.T) {
			got := mapping(tt.rt, "db")
			t.Logf(" got:%#v", got)
			t.Logf("want:%#v", tt.want)

			for k := range got {
				want, ok := tt.want[k]
				if !ok {
					t.Errorf("Test_mapping got unexpected tag: %s", k)
					return
				}
				if err := fnStructFieldCompare(got[k], want); err != nil {
					t.Errorf("Test_mapping tag struct not matched. tag:%s, err:%v", k, err)
					return
				}

				delete(tt.want, k)
			}

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
