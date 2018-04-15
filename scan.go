package brows

import (
	"database/sql"
	"errors"
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

// field
type spField struct {
	originField reflect.StructField
	name        string
	index       []int
}

type fieldTagMap map[string]spField

func compileStructTag(d reflect.Type) map[string]spField {
	// 存储 以 tag 为key 的 map，同时 存储 field在 struct 中的index，后面根据index要获取field
	// var fieldMap map[string]reflect.StructField = make(map[string]reflect.StructField, 0)
	var fieldMap = make(map[string]spField, d.NumField())
	for i := 0; i < d.NumField(); i++ {
		field := d.Field(i)
		log.Printf("num field: %v, field: %#v, fieldType: %#v", i, field, field.Type)

		tag := field.Tag.Get("db")
		// tag 选项
		tagOptions := strings.Split(tag, ",")

		log.Printf("num field: %v, tag: '%v', tagOptions: %v", i, tag, tagOptions)

		fieldMap[tag] = spField{
			originField: field,
			index:       []int{i},
		}
	}

	return fieldMap
}

func scanArgs(columns []string, value reflect.Value, ftm fieldTagMap) []interface{} {
	// args scan 时使用
	args := make([]interface{}, len(columns))
	log.Printf("args: %+v\n", args)

	// 按照 column 字段名称，给结构体中 tag 对应 field 赋值
	for k, column := range columns {
		var arg interface{}
		spStructField, ok := ftm[column]
		if ok {
			fieldV := value.FieldByIndex(spStructField.index)
			fieldV.Set(reflect.New(fieldV.Type()).Elem())

			structField := spStructField.originField
			log.Printf("elem : %#v, structField: %#v", fieldV, structField)

			arg = fieldV.Addr().Interface()
		} else {
			log.Printf("get field by name failed. column=%v", column)
		}

		args[k] = arg
	}

	return args
}

// ScanStruct 将 返回结果值赋值给 dest, 对应 queryRow
func ScanStruct(rows *sql.Rows, dest interface{}) error {
	// close rows
	defer rows.Close()

	// dest 类型检查
	d := reflect.ValueOf(dest)
	log.Printf("dest kind: %v", d.Kind())

	if d.Kind() != reflect.Ptr || d.IsNil() { // 必须是指针
		return errScanStructValue
	}

	// 获取指针指向的元素
	de := d.Elem()

	log.Printf("d.Elem kind:%v, type:%v", de.Kind(), de.Type())

	// 指针指向的数据类型检查
	if de.Kind() != reflect.Struct { // 必须是结构体
		return errScanStructValue
	}

	// de type
	deType := de.Type()
	// 标签
	fieldTagMap := compileStructTag(deType)

	columns, err := rows.Columns()
	if err != nil {
		log.Panic(err)
		return err
	}

	args := scanArgs(columns, de, fieldTagMap)

	// 当没有返回值时，返回 sql.ErrNoRows
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	// 处理 rows, 对应 queryRow ，所以只取结果的第一条
	log.Printf("args before scan: %#v\n", args)
	if err := rows.Scan(args...); err != nil {
		log.Printf("rows.Scan err. err: %v, args: %+v", err, args)
		return err
	}

	log.Printf("args after scan: %#v\n", args)

	// log.Printf("de Type:%#v", de) //dest 指针指向都对象
	log.Printf("dest Type:%#v", dest) // dest 指针对象

	if err := rows.Err(); err != nil {
		log.Printf("rows.err. err: %v", err)
		return err
	}

	return nil
}

// ScanSlice 将多个返回结果值赋值给 dest 数组，对应 query
func ScanSlice(rows *sql.Rows, dest interface{}) error {
	// close rows
	defer rows.Close()

	// dest 类型检查
	d := reflect.ValueOf(dest)
	log.Printf("dest kind: %v", d.Kind())

	if d.Kind() != reflect.Ptr || d.IsNil() { // 必须是指针
		return errScanStructValue
	}

	// 获取指针指向的元素
	de := d.Elem() // slice

	log.Printf("d.Elem kind:%v, type:%v", de.Kind(), de.Type())

	if de.Kind() != reflect.Slice { // 必须是切片
		return errScanSliceValue
	}

	deType := de.Type().Elem() // slice element is struct
	log.Printf("deType.Kind:%v", deType.Kind())

	// 切片数据类型检查
	if deType.Kind() != reflect.Struct { // 必须是结构体
		return errScanSliceElement
	}

	// 标签
	fieldTagMap := compileStructTag(deType)

	columns, err := rows.Columns()
	if err != nil {
		log.Panic(err)
		return err
	}

	// args scan 时使用
	argslice := make([][]interface{}, 0)
	log.Printf("argslice: %+v\n", argslice)

	// 处理 row
	i := 0
	for rows.Next() {
		// slice 扩展容量
		if i >= de.Cap() {
			newcap := de.Cap() + de.Cap()/2
			if newcap < 4 {
				newcap = 4
			}
			newv := reflect.MakeSlice(de.Type(), de.Len(), newcap)
			reflect.Copy(newv, de)
			de.Set(newv)
		}
		// slice 扩展长度
		if i >= de.Len() {
			de.SetLen(i + 1)
		}

		// 单个 struct 结构体对应的 参数
		args := make([]interface{}, len(columns))
		// 按照 column 字段名称，给结构体中 tag 对应 field 赋值
		for k, column := range columns {
			var arg interface{}
			spStructField, ok := fieldTagMap[column]
			if ok {
				fieldV := de.Index(i).FieldByIndex(spStructField.index)
				fieldV.Set(reflect.New(fieldV.Type()).Elem())

				structField := spStructField.originField
				log.Printf("elem : %#v, structField: %#v", fieldV, structField)

				arg = fieldV.Addr().Interface()
			} else {
				log.Printf("get field by name failed. column=%v", column)
			}

			args[k] = arg
		}

		argslice = append(argslice, args)

		log.Printf("args before scan: %#v\n", args)
		if err := rows.Scan((argslice[i])...); err != nil {
			log.Printf("rows.Scan err. err: %v, args: %+v", err, args)
			return err
		}

		i++
		log.Printf("i: %v, args after scan: %#v\n", i, args)
	}

	// log.Printf("de Type:%#v", de) //dest 指针指向都对象
	log.Printf("dest Type:%#v", dest) // dest 指针对象

	if err := rows.Err(); err != nil {
		log.Printf("rows.err. err: %v", err)
		return err
	}

	return nil
}
