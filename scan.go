package brows

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// sql: Scan error on column index 0, name "cid": destination not a pointer
var (
	ErrScanDestination      = errors.New("brows: Scan destination must be a non-nil pointer to a struct")
	ErrScanSliceDestination = errors.New("brows: ScanSlice destination must be a non-nil pointer to a slice")
	ErrSliceElement         = errors.New("brows: slice element only support *struct or struct")
)

// Scan 读取第一行记录，复制到 dest. dest 必须是 *struct.
//
// 结构体字段通过 tag 和 columns 进行唯一匹配，不依赖 columns 和结构体字段顺序.
// 内部转换复制依赖 `database/sql` 包的 Rows.Scan 方法
//
//   - 若 Rows 有多条记录，只读取第一条，丢弃其他剩余记录;
//   - 若 Rows 无记录，则返回 sql.ErrNoRows 错误;
//
// example:
//
//	type User struct {
//		Name string `db:"name"`
//		Age uint8 `db:"age"`
//	}
//
//	var user User
//	Scan(rows, &user)
func Scan(rows *sql.Rows, dest any) error {
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	rv := reflect.ValueOf(dest)
	if reflect.Pointer != rv.Kind() || rv.IsNil() {
		return ErrScanDestination
	}

	ev := rv.Elem()
	if reflect.Struct != ev.Kind() {
		return ErrScanDestination
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 映射查询字段和结构体字段
	args := mapColumns(columns, ev)
	if err := rows.Scan(args...); err != nil {
		return err
	}

	return rows.Close()
}

// ScanSlice 读取所有行记录，复制到 dest.
// dest 必须是 []struct or []*struct 的指针
//
// 结构体字段通过 tag 和 columns 进行唯一匹配，不依赖 columns 和结构体字段顺序.
// 内部转换复制依赖 `database/sql` 包的 Rows.Scan 方法
//
// example:
//
//	type User struct {
//		Name string `db:"name"`
//		Age uint8 `db:"age"`
//	}
//
//	var users []User // or []*User
//	ScanSlice(rows, &users)
func ScanSlice(rows *sql.Rows, dest any) error {
	defer rows.Close()

	rv := reflect.ValueOf(dest)
	if reflect.Pointer != rv.Kind() || rv.IsNil() {
		return ErrScanSliceDestination
	}

	// must slice
	slice := rv.Elem()
	if reflect.Slice != slice.Kind() {
		return ErrScanSliceDestination
	}

	sliceElemType := slice.Type().Elem() // slice element
	sliceElemInnerType := sliceElemType
	switch sliceElemType.Kind() {
	case reflect.Pointer:
		sliceElemInnerType = sliceElemInnerType.Elem()
		if reflect.Struct != sliceElemInnerType.Kind() {
			return ErrSliceElement
		}
	case reflect.Struct:
	default:
		return ErrSliceElement
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
		if reflect.Pointer != sliceElemType.Kind() {
			one = one.Elem()
		}
		slice = reflect.Append(slice, one)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	rv.Elem().Set(slice)
	return rows.Close()
}

// mapColumns 根据 columns 字段名称，在 e 中按 tag 找到对应 structField,
func mapColumns(columns []string, e reflect.Value) []any {
	if reflect.Pointer == e.Kind() {
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

// mapping 提取 value 对象的 tag 和 structField 的映射关系
//
// 提取规则
// - tag 需唯一，value 对象内，若 tag 重复，则 panic
// - structField 以下情况的，将被忽略
//   - 不可导
//   - tag 是 '-' 或 空
//
// - structField 以下情况的，将遍历 field 对象的内部字段
//   - 匿名内嵌对象
//   - 非 time.Time 类型的结构体
//   - 指针对象
func mapping(value reflect.Value, tag string) map[string]structFiledIndex {
	vKind := value.Kind()
	if reflect.Pointer == vKind {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return mapping(value.Elem(), tag)
	}

	vType := value.Type()
	if reflect.Struct != vKind {
		return nil
	}

	out := make(map[string]structFiledIndex)
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
		case reflect.Pointer == field.Type.Kind():
			if fieldValue.IsNil() {
				value.Field(i).Set(reflect.New(fieldValue.Type().Elem()))
			}
			mappingMerge(out, field.Index, mapping(fieldValue, tag))
			continue
		}

		tagValue := field.Tag.Get(tag)
		tagValue, _ = head(tagValue, ",")
		if "-" == tagValue || "" == tagValue {
			continue
		}

		mappingConflict(out, tagValue, field.Name)

		out[tagValue] = structFiledIndex{
			index: []int{i},
			field: field,
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
