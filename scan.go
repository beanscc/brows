package brows

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

var (
	errScanPtr         = errors.New("brows: value must be non-nil pointer")
	errScanStructValue = errors.New("brows: value must be non-nil pointer to a struct")
	errScanSliceValue  = errors.New("brows: value must be non-nil pointer to a slice")
)

// ScanSlice 将多个返回结果值赋值给 dest 数组，对应 query
// dest 必须是 []T or []*T 的指针类型
func ScanSlice(rows *sql.Rows, dest interface{}) error {
	// close rows
	defer rows.Close()

	// dest 类型检查
	dv := reflect.ValueOf(dest)
	log.Printf("dest kind: %v", dv.Kind())

	if dv.Kind() != reflect.Ptr || dv.IsNil() { // 必须是指针
		return errScanStructValue
	}

	// 获取指针指向的元素
	dve := dv.Elem() // slice
	log.Printf("d.Elem kind:%v, type:%v", dve.Kind(), dve.Type())
	if dve.Kind() != reflect.Slice { // 必须是切片
		return errScanSliceValue
	}

	dveElem := dve.Type().Elem() // slice element
	log.Printf("deType.Kind:%v", dveElem.Kind())

	// // 切片数据类型检查
	// if dveElem.Kind() != reflect.Struct { // 必须是结构体
	// 	return errScanSliceElement
	// }

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		typeOne := dveElem
		if reflect.Ptr == dveElem.Kind() {
			// []*T
			typeOne = typeOne.Elem()
		}

		one := reflect.New(typeOne)
		args := buildScanArgs(columns, one)
		if err := rows.Scan(args...); err != nil {
			return err
		}
		log.Printf("typeOne type:%s", typeOne.Kind())
		if reflect.Ptr != dveElem.Kind() { // 非指针
			one = one.Elem()
		}
		dve = reflect.Append(dve, one)
	}

	// // debug
	// for i := 0; i < dve.Len(); i++ {
	// 	log.Printf("dve: i:%2d, v:%#v", i, dve.Index(i))
	// }

	dv.Elem().Set(dve)

	// log.Printf("dest Type:%#v", dest) // dest 指针对象

	if err := rows.Err(); err != nil {
		// log.Printf("rows.err. err: %v", err)
		return err
	}

	return nil
}

// Scan scan one row
// desc must be pointer
func Scan(rows *sql.Rows, dest any) error {
	defer rows.Close()

	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errScanPtr
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	if e := v.Elem(); e.Kind() == reflect.Struct {
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		// 映射查询字段和结构体字段
		args := buildScanArgs(columns, e)
		// log.Printf("args:%#v", args)
		if err := rows.Scan(args...); err != nil {
			return err
		}
	} else {
		if err := rows.Scan(dest); err != nil {
			return err
		}
	}

	return rows.Err()
}

func buildScanArgs(columns []string, e reflect.Value) []any {
	if reflect.Ptr == e.Kind() {
		e = e.Elem()
	}

	tagMap := mapping(e)
	// debug
	for k, v := range tagMap {
		log.Printf("buildScanArgs mapping k:%s, v:%#v", k, v)
	}

	out := make([]any, 0, len(columns))
	for i, v := range columns {
		f, ok := tagMap[v]
		if !ok {
			// 忽略这个字段的 scan
			out = append(out, _ignoreScan)
			continue
		}

		log.Printf("index:%2d, v:%s, f:%#v, e:%#v", i, v, f, e)

		fv := e.FieldByIndex(f.index)
		out = append(out, fv.Addr().Interface())
	}

	// debug
	for i, v := range out {
		log.Printf("buildScanArgs out idx:%2d, v:%#v", i, v)
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

func mapping(value reflect.Value) map[string]structFiledIndex {
	log.Printf("mapping|vaule:%#v", value)

	vt := value.Type()

	out := make(map[string]structFiledIndex)
	// struct
	if reflect.Struct == value.Kind() {
		for i := 0; i < value.NumField(); i++ {
			field := vt.Field(i)
			fieldValue := value.Field(i)
			log.Printf("mapping|field i:%2d, v:%#v, type:%v, kind:%s", i, field, field.Type, field.Type.Kind())
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
