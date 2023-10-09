package brows

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	errScanPtr      = errors.New("brows: value must be non-nil pointer to a struct")
	errScanPtrSlice = errors.New("brows: value must be non-nil pointer to a slice")
)

// Scan scan row
// desc must be *struct
func Scan(rows *sql.Rows, dest any) error {
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errScanPtr
	}

	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return errScanPtr
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 映射查询字段和结构体字段
	args := mapColumns(columns, e)
	if err := rows.Scan(args...); err != nil {
		return err
	}

	return rows.Close()
}

// ScanSlice 将多个返回结果值赋值给 dest 数组，对应 query
// dest 必须是 []T or []*T 的指针类型
// example:
//
//	type User struct {
//		Name string `db:"name"`
//		Age uint8 `db:"age"`
//	}
//
// var users []User
// // or var users []*User
// ScanSlice(rows, &users)
func ScanSlice(rows *sql.Rows, dest interface{}) error {
	// close rows
	defer rows.Close()

	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return errScanPtr
	}

	// must slice
	slice := value.Elem()
	if slice.Kind() != reflect.Slice {
		return errScanPtrSlice
	}

	sliceElemType := slice.Type().Elem() // slice element
	sliceElemInnerType := sliceElemType
	switch sliceElemType.Kind() {
	case reflect.Ptr:
		sliceElemInnerType = sliceElemInnerType.Elem()
		if sliceElemInnerType.Kind() != reflect.Struct {
			return errScanPtrSlice
		}
	case reflect.Struct:
	default:
		return errScanPtrSlice
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		one := reflect.New(sliceElemInnerType)
		args := mapColumns(columns, one)
		if err := rows.Scan(args...); err != nil {
			return err
		}
		if reflect.Ptr != sliceElemType.Kind() {
			one = one.Elem()
		}
		slice = reflect.Append(slice, one)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	value.Elem().Set(slice)
	return rows.Close()
}

// mapColumns 根据 columns 字段名称，在 e 中按 tag 找到对应 structField,
func mapColumns(columns []string, e reflect.Value) []any {
	if reflect.Ptr == e.Kind() {
		e = e.Elem()
	}

	m := mapping(e, _tagLabel)
	out := make([]any, 0, len(columns))
	for _, v := range columns {
		f, ok := m[v]
		if !ok {
			// 忽略这个字段的 scan
			out = append(out, _ignoreScan)
			continue
		}

		fv := e.FieldByIndex(f.index)
		out = append(out, fv.Addr().Interface())
	}

	// check
	if len(columns) != len(out) {
		panic("brows: columns not all matched")
	}

	return out
}

// _ignoreScan 忽略 scan
var _ignoreScan = &ignoreScan{}

type ignoreScan struct{}

func (s *ignoreScan) Scan(v any) error {
	return nil
}

// tag 标签
var _tagLabel = "db"

type structFiledIndex struct {
	// index 位置
	// 若是内嵌类型/ 则, index 切片元素依次表示每个层级的索引位置
	index []int
	field reflect.StructField
}

func mapping(value reflect.Value, tag string) map[string]structFiledIndex {
	vKind := value.Kind()
	if reflect.Ptr == vKind {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return mapping(value.Elem(), tag)
	}

	vType := value.Type()
	out := make(map[string]structFiledIndex)
	if reflect.Struct == vKind {
		for i := 0; i < value.NumField(); i++ {
			field := vType.Field(i)
			fieldValue := value.Field(i)
			switch {
			case !field.IsExported():
				// 不可导出
				continue
			case field.Anonymous:
				// 内嵌
				mappingMerge(out, field.Index, mapping(fieldValue, tag))
				continue
			case reflect.Struct == field.Type.Kind():
				if "time.Time" != field.Type.String() {
					mappingMerge(out, field.Index, mapping(fieldValue, tag))
					continue
				}
			case reflect.Ptr == field.Type.Kind():
				if fieldValue.IsNil() {
					// init
					value.Field(i).Set(reflect.New(fieldValue.Type().Elem()))
				}
				mappingMerge(out, field.Index, mapping(fieldValue, tag))
				continue
			}

			tagValue := field.Tag.Get(tag)
			tagValue, _ = head(tagValue, ",")
			if "-" == tagValue {
				continue
			}

			if tagValue == "" {
				// default FieldName
				tagValue = field.Name
			}
			if tagValue == "" {
				continue
			}

			mappingConflict(out, tagValue, field.Name)

			out[tagValue] = structFiledIndex{
				index: []int{i},
				field: field,
			}
		}
	}

	return out
}

// mappingConflict tag 冲突检查
func mappingConflict(m map[string]structFiledIndex, tag string, field string) {
	if v, ok := m[tag]; ok {
		panic(fmt.Sprintf("brows: tag[%s] conflict. field %s vs %s", tag, v.field.Name, field))
	}
}

// mappingMerge 合并
func mappingMerge(dest map[string]structFiledIndex, parentIndex []int, source map[string]structFiledIndex) {
	for k, v := range source {
		mappingConflict(dest, k, v.field.Name)
		dest[k] = structFiledIndex{
			index: append(parentIndex, v.index...),
			field: v.field,
		}
	}
}

func head(str, sep string) (head string, tail string) {
	idx := strings.Index(str, sep)
	if idx < 0 {
		return str, ""
	}
	return str[:idx], str[idx+len(sep):]
}
