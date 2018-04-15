package brows

import (
	"database/sql"
	"log"
	"reflect"
)

// Brows db
type Brows struct {
	db *sql.DB
}

// New new Brows
func New(db *sql.DB) *Brows {
	return &Brows{db}
}

// DB return Brows.db
func (b *Brows) DB() *sql.DB {
	return b.db
}

// QueryRow prepare QueryRow
func (b *Brows) QueryRow(dest interface{}, query string, args ...interface{}) error {
	stmt, err := b.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return b.scanOne(rows, dest)
}

// Query prepare Query, dest 必须是切片类型
func (b *Brows) Query(dest interface{}, query string, args ...interface{}) error {
	stmt, err := b.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return b.scanSlice(rows, dest)
}

// Exec prepare Exec
func (b *Brows) Exec(query string, args ...interface{}) (sql.Result, error) {
	stmt, err := b.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return stmt.Exec(args...)
}

func (b *Brows) scanSlice(rows *sql.Rows, dest interface{}) error {
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

	deType := de.Type().Elem() // slice element is struct
	// 切片数据类型检查
	if deType.Kind() != reflect.Struct { // 必须是结构体
		return errScanSliceElement
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// args scan 时使用
	argslice := make([][]interface{}, 0)

	// 标签
	fieldTagMap := compileStructTag(deType)

	// 处理 rows
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
				arg = fieldV.Addr().Interface()
			} else {
				// nothing
			}
			args[k] = arg
		}

		argslice = append(argslice, args)
		if err := rows.Scan((argslice[i])...); err != nil {
			return err
		}

		i++
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// scanOne 只解析一条记录
func (b *Brows) scanOne(rows *sql.Rows, dest interface{}) error {
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
		fieldTagMap := compileStructTag(deType)
		columns, err := rows.Columns()
		if err != nil {
			log.Panic(err)
			return err
		}

		args := scanArgs(columns, de, fieldTagMap)
		// 处理 rows, 对应 queryRow ，所以只取结果的第一条
		if err := rows.Scan(args...); err != nil {
			return err
		}

		if err := rows.Err(); err != nil {
			return err
		}
	} else { // 普通类型
		if err := rows.Scan(dest); err != nil {
			return err
		}
	}

	return nil
}
