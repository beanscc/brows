package brows

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	errScanPtr      = errors.New("brows: value must be non-nil pointer")
	errScanPtrSlice = errors.New("brows: value must be non-nil pointer to a slice")
)

// Scan scan one row
// desc must be pointer
func Scan(rows *sql.Rows, dest any) error {
	// 关于scan 部分，保持和 sql.Row Scan() 方法同步
	defer rows.Close()
	if _, ok := dest.(*sql.RawBytes); ok {
		return errors.New("brows: RawBytes isn't allowed on Scan")
	}

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

	if e := v.Elem(); e.Kind() == reflect.Struct {
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		// 映射查询字段和结构体字段
		args := buildScanArgs(columns, e)
		if err := rows.Scan(args...); err != nil {
			return err
		}
	} else {
		if err := rows.Scan(dest); err != nil {
			return err
		}
	}

	return rows.Close()
}

// ScanSlice 将多个返回结果值赋值给 dest 数组，对应 query
// dest 必须是 []T or []*T 的指针类型, T 只能是 struct or 基本数据类型
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
	if reflect.Ptr == sliceElemType.Kind() {
		sliceElemInnerType = sliceElemInnerType.Elem()
	}

	// sliceElem 只支持
	isItemStruct := false
	switch sliceElemInnerType.Kind() {
	case reflect.Struct:
		isItemStruct = true
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Interface,
		reflect.String:
	default:
		return errors.New("unsupported slice elem type")
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		one := reflect.New(sliceElemInnerType)
		var args []any
		if isItemStruct {
			args = buildScanArgs(columns, one)
		} else {
			args = []any{one.Interface()}
		}

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

func buildScanArgs(columns []string, e reflect.Value) []any {
	if reflect.Ptr == e.Kind() {
		e = e.Elem()
	}

	tagMap := mapping(e)
	// // debug
	// for k, v := range tagMap {
	// 	log.Printf("buildScanArgs mapping k:%s, v:%#v", k, v)
	// }

	out := make([]any, 0, len(columns))
	for _, v := range columns {
		f, ok := tagMap[v]
		if !ok {
			// 忽略这个字段的 scan
			out = append(out, _ignoreScan)
			continue
		}

		// log.Printf("index:%2d, v:%s, f:%#v, e:%#v", i, v, f, e)

		fv := e.FieldByIndex(f.index)
		out = append(out, fv.Addr().Interface())
	}

	// // debug
	// for i, v := range out {
	// 	log.Printf("buildScanArgs out idx:%2d, v:%#v", i, v)
	// }

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

func mapping(value reflect.Value) map[string]structFiledIndex {
	// log.Printf("mapping|vaule:%#v", value)

	vt := value.Type()

	out := make(map[string]structFiledIndex)
	// struct
	if reflect.Struct == value.Kind() {
		for i := 0; i < value.NumField(); i++ {
			field := vt.Field(i)
			fieldValue := value.Field(i)
			// log.Printf("mapping|field i:%2d, v:%#v, type:%v, kind:%s", i, field, field.Type, field.Type.Kind())
			switch {
			case !field.IsExported():
				// 不可导出
				continue
			case field.Anonymous:
				// 内嵌
				mappingMerge(out, field.Index, mapping(fieldValue))
				continue
			case reflect.Struct == field.Type.Kind():
				if "time.Time" != field.Type.String() {
					mappingMerge(out, field.Index, mapping(fieldValue))
					continue
				}
			case reflect.Ptr == field.Type.Kind():
				if fieldValue.IsNil() {
					// init
					value.Field(i).Set(reflect.New(fieldValue.Type().Elem()))
				}
				mappingMerge(out, field.Index, mapping(fieldValue))
				continue
			}

			tag, ok := field.Tag.Lookup(_tagLabel)
			if !ok || "" == tag {
				continue
			}

			tags := strings.Split(tag, ",")
			// tag name
			tagName := tags[0]
			if "-" == tagName || "" == tagName {
				continue
			}

			// tag 冲突检查
			_, ok = out[tagName]
			if ok {
				panic(fmt.Sprintf("brows: tag[%s] conflict", tagName))
			}

			out[tagName] = structFiledIndex{
				index: []int{i},
				field: field,
			}
		}
	}

	if reflect.Ptr == value.Kind() {
		if value.IsNil() {
			// init
			value.Set(reflect.New(value.Type().Elem()))
		}
		return mapping(value.Elem())
	}

	return out
}

// mappingMerge 合并
func mappingMerge(dest map[string]structFiledIndex, parentIndex []int, source map[string]structFiledIndex) {
	for k, v := range source {
		// tag 冲突检查
		if dv, ok := dest[k]; ok {
			panic(fmt.Sprintf("brows: tag[%s] conflicted. in fields:%s",
				k, strings.Join([]string{dv.field.Name, v.field.Name}, "")))
		}
		dest[k] = structFiledIndex{
			index: append(parentIndex, v.index...),
			field: v.field,
		}
	}
}
