package brows

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestMapping(t *testing.T) {
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

			type orderedGot struct {
				key   string
				tag   string
				field structField
			}

			_orderedGot := make([]orderedGot, 0, len(got))
			for k, v := range got {
				key := ""
				for _, v2 := range v.index {
					key += strconv.Itoa(v2)
				}
				_orderedGot = append(_orderedGot, orderedGot{
					key:   key,
					tag:   k,
					field: v,
				})
			}

			sort.Slice(_orderedGot, func(i, j int) bool {
				return _orderedGot[i].key < _orderedGot[j].key
			})

			for _, v := range _orderedGot {
				k := v.tag
				want, ok := tt.want[k]
				if !ok {
					t.Errorf("Test_mapping got unexpected tag: %s", k)
					continue
				}

				if !reflect.DeepEqual(v.field.index, want.index) {
					t.Errorf("Test_mapping field index not matched. tag:%s", k)
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

func TestMappingByColumns(t *testing.T) {
	fnCompare := func(s1, s2 structFields) error {
		if len(s1) != len(s2) {
			return errors.New("length not equal")
		}

		s1m := make(map[string]structField)
		for _, v := range s1 {
			s1m[v.column] = v
		}
		if len(s1) != len(s1m) {
			return errors.New("s1 mapping tag error")
		}

		s2m := make(map[string]structField)
		for _, v := range s2 {
			s2m[v.column] = v
		}
		if len(s2) != len(s2m) {
			return errors.New("s2 mapping tag error")
		}

		for k, v1 := range s1m {
			v2, ok := s2m[k]
			if !ok {
				return errors.New(fmt.Sprintf("key:%s, in s1 not in s2", k))
			}
			if v1.ignore != v2.ignore {
				return errors.New(fmt.Sprintf("key:%s, v1.ignore != v2.ignore", k))
			}

			if v1.column != v2.column {
				return errors.New(fmt.Sprintf("key:%s, v1.column != v2.column", k))
			}

			if !reflect.DeepEqual(v1.index, v2.index) {
				return errors.New(fmt.Sprintf("key:%s, !reflect.DeepEqual(v1.index, v2.index)", k))
			}

			// value
			// if v1.value.Kind() != v2.value.Kind() {
			// 	return errors.New(fmt.Sprintf("key:%s, v1.value.Kind() != v2.value.Kind(). v1 kind:%s, v2 kind:%s", k,
			// 		v1.value.Kind(), v2.value.Kind()))
			// }

			if v1.value.Type() != v2.value.Type() {
				return errors.New(fmt.Sprintf("key:%s, v1.value.Type() != v2.value.Type(). v1 type:%s, v2 type:%s", k,
					v1.value.Type(), v2.value.Type()))
			}

			if reflect.Pointer == v1.value.Kind() {
				isV1Nil := v1.value.IsNil()
				isV2Nil := v2.value.IsNil()
				if isV1Nil && !isV2Nil {
					return errors.New(fmt.Sprintf("key:%s, v1.value.IsNil() && !v2.value.IsNil()", k))
				}

				if !isV1Nil && isV2Nil {
					return errors.New(fmt.Sprintf("key:%s, !v1.value.IsNil() && v2.value.IsNil()", k))
				}
			}
		}

		return nil
	}

	test := []struct {
		name    string
		columns []string
		ptr     any
		want    structFields
		wantErr bool
	}{

		{
			name:    "string",
			columns: []string{"name1", "name2", "name3", "name4"},
			ptr: &struct {
				Name1 string  `db:"name1"`
				Name2 *string `db:"name2"`
				Name3 []byte  `db:"name3"`
				Name4 *[]byte `db:"name4"`
			}{},
			want: structFields{
				{column: "name1", ignore: false, index: []int{0}, value: reflect.ValueOf("")},
				{column: "name2", ignore: false, index: []int{1}, value: reflect.ValueOf(addPtr(""))},
				{column: "name3", ignore: false, index: []int{2}, value: reflect.ValueOf([]byte(""))},
				// {column: "name4", ignore: false, index: []int{3}, value: reflect.ValueOf(addPtr([]byte("")))},
				{column: "name4", ignore: false, index: []int{3}, value: reflect.ValueOf(addPtr([]byte(nil)))},
			},
			wantErr: false,
		},

		// want err
		{
			name:    "string",
			columns: []string{"name1"},
			ptr: &struct {
				Name1 string  `db:"name1"`
				Name2 *string `db:"name2"`
			}{},
			want: structFields{
				{column: "name1", ignore: false, index: []int{0}, value: reflect.ValueOf("")},
				{column: "name2", ignore: false, index: []int{1}, value: reflect.ValueOf("")},
			},
			// want 和 got 不一致
			wantErr: true,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			got := mappingByColumns(tt.columns, reflect.ValueOf(tt.ptr))
			if err := fnCompare(got, tt.want); err != nil && !tt.wantErr {
				t.Errorf("TestMappingByColumns failed. err:%v", err)
			}
		})
	}
}

func addPtr[V string | int | []byte](v V) *V {
	return &v
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
