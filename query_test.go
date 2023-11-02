package brows

import (
	"database/sql"
	"fmt"
	"testing"

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
	dsn := `root:P4m@bpet@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=UTC&readTimeout=3s&timeout=3s&writeTimeout=3s`
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
