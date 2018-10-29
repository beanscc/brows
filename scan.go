package brows

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

func init() {
	// set log
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

var (
	errScanPtr          = errors.New("brows: value must bu non-nil pointer")
	errScanStructValue  = errors.New("brows: value must be non-nil pointer to a struct")
	errScanSliceValue   = errors.New("brows: value must be non-nil pointer to a slice")
	errScanSliceElement = errors.New("brows: the element of slice must be a struct")
)

const (
	dbTagLabel = "db"
)

// parseTagFromStruct 从结构体中解析 tag 标签
// 仅支持单层结构体，不支持复杂的多层结构体
func parseTagFromStruct(typ reflect.Type) (map[string]int, []reflect.StructField) {
	if typ.Kind() != reflect.Struct {
		panic("typ must be reflect.Type of struct")
	}

	fieldTagMap := make(map[string]int, typ.NumField())
	fieldSlice := make([]reflect.StructField, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		log.Printf("num field: %v, field: %#v", i, field)
		tag := field.Tag.Get(dbTagLabel)
		// tag 项
		tagOptions := strings.Split(tag, ",")
		log.Printf("num field: %v, tag: '%v', tagOptions: %v", i, tag, tagOptions)
		if tag != "" {
			fieldTagMap[tag] = i
		}

		fieldSlice = append(fieldSlice, field)
	}

	return fieldTagMap, fieldSlice
}

func scanArgs(columns []string, value reflect.Value, ftm map[string]int) []interface{} {
	// args scan 时使用
	args := make([]interface{}, len(columns))

	// 按照 column 字段名称顺序，获取结构体中对应字段的指针
	for k, column := range columns {
		var arg interface{}
		if fieldIndex, ok := ftm[column]; ok {
			fieldV := value.Field(fieldIndex)
			arg = fieldV.Addr().Interface() // 必须是指针
		} else {
			panic(fmt.Sprintf("Scan error on column index %v named %q: no db tag in struct field", k, column))
		}

		args[k] = arg
	}

	return args
}

// ScanOne 只解析一条记录，用于 queryRow
func ScanOne(rows *sql.Rows, dest interface{}) error {
	// dest 类型检查
	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr || d.IsNil() { // 必须是指针
		return errScanPtr
	}

	// 当没有返回值时，返回 sql.ErrNoRows
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	// 获取指针指向的元素
	de := d.Elem()
	// 指针指向的数据类型检查
	if de.Kind() == reflect.Struct { // 结构体
		// de type
		deType := de.Type()
		// 标签
		fieldTagMap, _ := parseTagFromStruct(deType)
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		args := scanArgs(columns, de, fieldTagMap)
		// queryRow，只取结果的第一条
		if err := rows.Scan(args...); err != nil {
			return err
		}

		if err := rows.Err(); err != nil {
			return err
		}
	} else { // 普通类型，如 int/float/string
		if err := rows.Scan(dest); err != nil {
			return err
		}
	}

	return nil
}

// ScanSlice 解析多条记录，用于 query
func ScanSlice(rows *sql.Rows, dest interface{}) error {
	// dest 类型检查
	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr || d.IsNil() { // 必须是指针
		return errScanPtr
	}

	// 获取指针指向的元素
	de := d.Elem()                  // slice
	if de.Kind() != reflect.Slice { // 必须是切片
		return errScanSliceValue
	}

	sliceElemType := de.Type().Elem() // slice element is struct
	// 切片数据类型检查
	if sliceElemType.Kind() != reflect.Struct { // 必须是结构体
		return errScanSliceElement
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 解析结构体标签
	fieldTagMap, fieldSlice := parseTagFromStruct(sliceElemType)
	// new resp slice
	respSlice := reflect.MakeSlice(de.Type(), 0, de.Cap())

	// 处理 rows
	for rows.Next() {
		// 构造结构体
		rowStruct := reflect.New(reflect.StructOf(fieldSlice)).Elem()
		// 获取 rowStruct 结构体 db 标签对应字段的指针引用
		args := scanArgs(columns, rowStruct, fieldTagMap)
		// 将行结果映射到结构体的引用字段中
		if err := rows.Scan(args...); err != nil {
			return err
		}

		// append to slice
		respSlice = reflect.Append(respSlice, rowStruct)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// 将结果赋值给输入变量
	de.Set(respSlice)

	return nil
}
