package brows

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DBTest struct {
	*testing.T
	db *sql.DB
}

func (dbt *DBTest) initTable() {
	dbt.mustExec(`DROP TABLE IF EXISTS test_brows`)
}

func (dbt *DBTest) mustExec(query string, args ...any) sql.Result {
	res, err := dbt.db.Exec(query, args...)
	if err != nil {
		dbt.Fatalf("error on %s | %s: %s", "exec", query, err.Error())
	}
	return res
}

func testDBScope(t *testing.T, fn func(dbt *DBTest)) {
	// init db
	dsn := `root:P4m@bpet@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=true&loc=UTC&readTimeout=3s&timeout=3s&writeTimeout=3s`
	fmt.Println("dsn:", dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	dbt := DBTest{
		T:  t,
		db: db,
	}

	// init table
	dbt.initTable()

	fn(&dbt)
}

func TestBrows_QueryRow_String(t *testing.T) {
	testDBScope(t, func(dbt *DBTest) {
		types := []string{
			// char
			"CHAR(255)", "VARCHAR(255)",
			// text
			"TINYTEXT", "TEXT", "MEDIUMTEXT", "LONGTEXT",
			// blob
			"TINYBLOB", "BLOB", "MEDIUMBLOB", "LONGBLOB",
		}
		in := "κόσμε üöäßñóùéàâÿœ'îë Árvíztűrő いろはにほへとちりぬるを イロハニホヘト דג סקרן чащах  น่าฟังเอย"

		type Out struct {
			Value string `db:"value"`
		}

		brows := New(dbt.db)
		for _, v := range types {
			dbt.initTable()

			dbt.mustExec(`CREATE TABLE test_brows (value ` + v + `) CHARACTER SET utf8`)
			dbt.mustExec(`insert into test_brows values (?)`, in)

			var out Out
			err := brows.QueryRow(`select value from test_brows`).Scan(&out)
			if err != nil {
				t.Errorf("TestBrows_QueryRow_String err:%v", err)
				return
			}

			if out.Value != in {
				t.Errorf("%s: %s != %s", v, in, out.Value)
			}
		}
	})
}

func TestBrows_QueryRow_Integer(t *testing.T) {
	testDBScope(t, func(dbt *DBTest) {
		types := []string{"TINYINT", "SMALLINT", "MEDIUMINT", "INT", "BIGINT"}
		in := int64(40)

		type Out struct {
			Value int64 `db:"value"`
		}

		brows := New(dbt.db)
		// SIGNED
		for _, v := range types {
			dbt.initTable()
			dbt.mustExec(`CREATE TABLE test_brows (value ` + v + `)`)
			dbt.mustExec(`insert into test_brows values (?)`, in)
			var out Out
			err := brows.QueryRow(`select value from test_brows`).Scan(&out)
			if err != nil {
				t.Errorf("TestBrows_QueryRow_Integer err:%v", err)
				return
			}

			if out.Value != in {
				t.Errorf("%s: %d != %d", v, in, out.Value)
			}
		}

		// If you specify ZEROFILL for a numeric column, MySQL automatically adds the UNSIGNED attribute to the column
		for _, v := range types {
			dbt.initTable()
			dbt.mustExec(`CREATE TABLE test_brows (value ` + v + ` ZEROFILL)`)
			dbt.mustExec(`insert into test_brows values (?)`, in)
			var out Out
			err := brows.QueryRow(`select value from test_brows`).Scan(&out)
			if err != nil {
				t.Errorf("TestBrows_QueryRow_Integer err:%v", err)
				return
			}

			if out.Value != in {
				t.Errorf("%s ZEROFILL: %d != %d", v, in, out.Value)
			}
		}
	})
}

func TestBrows_QueryRow_Float(t *testing.T) {
	testDBScope(t, func(dbt *DBTest) {
		types := []string{"FLOAT", "DOUBLE", "DECIMAL(10,2)"}
		in32, in64 := float32(0.99), float64(1.99)

		type Out struct {
			Value32 float32 `db:"value32"`
			Value64 float64 `db:"value64"`
		}

		brows := New(dbt.db)
		for _, v := range types {
			dbt.initTable()
			dbt.mustExec(`CREATE TABLE test_brows (value32 ` + v + `, value64 ` + v + `)`)
			dbt.mustExec(`insert into test_brows values (?,?)`, in32, in64)
			var out Out
			err := brows.QueryRow(`select value32,value64 from test_brows`).Scan(&out)
			if err != nil {
				t.Errorf("TestBrows_QueryRow_Float err:%v", err)
				return
			}

			if out.Value32 != in32 {
				t.Errorf("%s: float32 %v != %v", v, in32, out.Value32)
			}

			if out.Value64 != in64 {
				t.Errorf("%s: float64 %v != %v", v, in64, out.Value64)
			}
		}
	})
}

func TestBrows_Query(t *testing.T) {
	testDBScope(t, func(dbt *DBTest) {
		dbt.initTable()

		const create = `
CREATE TABLE test_brows (
    id int not null primary key auto_increment,
    -- string
    val_char char(255),
    val_varchar varchar(255),
    val_tinytext tinytext,
    val_text text,
    val_mediumtext mediumtext,
    val_longtext longtext,
    -- integer
    val_tinyint tinyint,
    val_smallint smallint,
    val_mediumint mediumint,
    val_int int,
    val_bigint bigint,
	-- float
	val_float float,
	val_double double,
	val_decimal decimal(10,2),
	-- time
	val_date date,
	val_time time,
	val_datetime datetime,
	val_timestamp timestamp,
	val_year year,
	
	-- other
	status tinyint,
	status1 tinyint
) engine=innodb CHARACTER SET utf8
`
		dbt.mustExec(create)

		type Status int8
		type Created struct {
			Operator string    `db:"created_by"`
			Time     time.Time `db:"created_at"`
		}

		type Deleted struct {
			Operator string    `db:"deleted_by"`
			Time     time.Time `db:"deleted_at"`
		}

		type Row struct {
			ID int `db:"id"`

			Char       string `db:"val_char"`
			VarChar    string `db:"val_varchar"`
			Tinytext   string `db:"val_tinytext"`
			Text       string `db:"val_text"`
			Mediumtext string `db:"val_mediumtext"`
			Longtext   string `db:"val_longtext"`

			Tinyint   int8  `db:"val_tinyint"`
			Smallint  int16 `db:"val_smallint"`
			Mediumint int   `db:"val_mediumint"`
			Int       int   `db:"val_int"`
			Bigint    int64 `db:"val_bigint"`

			Float   float32 `db:"val_float"`
			Double  float64 `db:"val_double"`
			Decimal float64 `db:"val_decimal"`

			// Data Type	“Zero” Value
			// DATE	'0000-00-00'
			// TIME	'00:00:00' // 不能解析为 time.Time，按 string 处理，然后转 time
			// DATETIME	'0000-00-00 00:00:00'
			// TIMESTAMP	'0000-00-00 00:00:00'
			// YEAR	0000 // 不能解析为 time.Time 按 int 处理，然后转 time
			Date      time.Time `db:"val_date"`
			Time      string    `db:"val_time"`
			Datetime  time.Time `db:"val_datetime"`
			Timestamp time.Time `db:"val_timestamp"`
			Year      int       `db:"val_year"`

			// Anonymous
			Status  `db:"status"`
			Status1 Status `db:"status1"`

			// struct
			Created Created
			// *struct
			*Deleted
		}

		in := Row{
			Char:       "Char",
			VarChar:    "VarChar",
			Tinytext:   "Tinytext",
			Text:       "Text",
			Mediumtext: "Mediumtext",
			Longtext:   "Longtext",
			Tinyint:    1,
			Smallint:   2,
			Mediumint:  3,
			Int:        4,
			Bigint:     5,
			Float:      10.09,
			Double:     20.09,
			Decimal:    30.09,
			Date:       time.Date(2000, 1, 10, 5, 6, 7, 0, time.UTC),
			Time:       time.Date(2000, 1, 10, 5, 6, 7, 0, time.UTC).Format(`15:04:05`),
			Datetime:   time.Date(2000, 1, 10, 5, 6, 7, 0, time.UTC),
			Timestamp:  time.Date(2000, 1, 10, 5, 6, 7, 0, time.UTC),
			Year:       time.Date(2000, 1, 10, 5, 6, 7, 0, time.UTC).Year(),

			Status:  0,
			Status1: 1,
		}

		insert := fmt.Sprintf(`insert into test_brows values (null %s)`, strings.Repeat(",?", 21))

		// fmt.Println("insert:", insert, ", args:",
		// 	in.Char, in.VarChar, in.Tinytext, in.Text, in.Mediumtext, in.Text,
		// 	in.Tinyint, in.Smallint, in.Mediumint, in.Int, in.Bigint,
		// 	in.Float, in.Double, in.Decimal,
		// 	in.Date, in.Time, in.Datetime, in.Timestamp, in.Year,
		// )

		dbt.mustExec(insert,
			in.Char, in.VarChar, in.Tinytext, in.Text, in.Mediumtext, in.Text,
			in.Tinyint, in.Smallint, in.Mediumint, in.Int, in.Bigint,
			in.Float, in.Double, in.Decimal,
			in.Date, in.Time, in.Datetime, in.Timestamp, in.Year,
			in.Status, in.Status1,
		)

		// row 2
		dbt.mustExec(insert,
			in.Char+"_2", in.VarChar+"_2", in.Tinytext+"_2", in.Text+"_2", in.Mediumtext+"_2", in.Text+"_2",
			in.Tinyint, in.Smallint, in.Mediumint, in.Int, in.Bigint,
			in.Float, in.Double, in.Decimal,
			in.Date, in.Time, in.Datetime, in.Timestamp, in.Year,
			in.Status, in.Status1,
		)

		// QueryRow
		var out Row
		err := New(dbt.db).QueryRow(`select * from test_brows`).Scan(&out)
		if err != nil {
			t.Errorf("TestBrows_QueryRow err:%v", err)
			return
		}
		if reflect.DeepEqual(in, out) {
			t.Errorf("TestBrows_QueryRow in != out")
		}

		t.Logf("out:%#v", out)

		// QueryRow sql.ErrNoRows
		err = New(dbt.db).QueryRow(`select * from test_brows where id = 100`).Scan(&out)
		if err == nil || !errors.Is(err, sql.ErrNoRows) {
			t.Error("TestBrows_QueryRow want sql.ErrNoRows")
			return
		}

		// query
		var rows []Row
		res := New(dbt.db).Query(`select * from test_brows`)

		cts, err := res.ColumnTypes()
		for i, v := range cts {
			// fmt.Println("i:", i, ", col:", v)
			fmt.Printf("i:%d, col:%#v\n", i, v)
		}

		if err := res.Scan(&rows); err != nil {
			t.Errorf("TestBrows_Query err:%v", err)
			return
		}

		if len(rows) != 2 {
			t.Errorf("TestBrows_Query want len(rows) = 2")
			return
		}

		if reflect.DeepEqual(in, rows[0]) {
			t.Errorf("TestBrows_Query in != rows[0]")
		}

		var pointerRows []*Row
		err = New(dbt.db).Query(`select * from test_brows`).Scan(&pointerRows)
		if err != nil {
			t.Errorf("TestBrows_Query pointerRows err:%v", err)
			return
		}

		if len(pointerRows) != 2 {
			t.Errorf("TestBrows_Query pointerRows want len(rows) = 2")
			return
		}

		if reflect.DeepEqual(in, pointerRows[0]) {
			t.Errorf("TestBrows_Query pointerRows in != rows[0]")
		}
	})
}
