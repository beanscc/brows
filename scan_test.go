package brows

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

type app struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	AppID     string    `db:"app_id"`
	Secret    string    `db:"secret"`
	Sign      string    `db:"sign"`
	Status    bool      `db:"status"`
	EndTime   int64     `db:"end_time"`
	StartTime int64     `db:"start_time"`
	Ctime     time.Time `db:"ctime"`
	Utime     time.Time `db:"utime"`
	Operator  string    `db:"operator"`
}

/*

创建测试表：

CREATE TABLE IF NOT EXISTS `app` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'APP 名称',
  `app_id` CHAR(40) NOT NULL DEFAULT '' COMMENT '应用APPID',
  `secret` VARCHAR(45) NOT NULL DEFAULT '' COMMENT '应用APP secret',
  `sign` VARCHAR(45) NOT NULL DEFAULT '' COMMENT '应用签名key',
  `start_time` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '生效时间',
  `end_time` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '结束时间',
  `status` INT NOT NULL DEFAULT 0 COMMENT '应用状态；0-停用；1-启用',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
  `operator` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作人',
  `ctime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `utime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `id_UNIQUE` (`id` ASC),
  UNIQUE INDEX `app_id_UNIQUE` (`app_id` ASC),
  UNIQUE INDEX `name_UNIQUE` (`name` ASC))
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8
COMMENT = '应用APP'

-- 添加测试数据

INSERT INTO `test`.`app`
(
`name`,
`app_id`,
`secret`,
`sign`,
`start_time`,
`end_time`,
`status`,
`description`,
`operator`)
VALUES
('t1', "app_id_1_dfsdfsdfsdfsddf", "secret_1_sfsdfsdfsdfsd", "sign_1_dsfsdfvsdghadfg", 14811110152, 1523772288, 0, "app_id_1_desc", "yx"),
('t2', "app_id_2_dfsdfsdfsdfsddf", "secret_2_sfsdfsdfsdfsd", "sign_2_dsfsdfvsdghadfg", 1482220152, 1523772288, 1, "app_id_2_desc", "yx"),
('t3', "app_id_3_dfsdfsdfsdfsddf", "secret_3_sfsdfsdfsdfsd", "sign_4_dsfsdfvsdghadfg", 1483330152, 1523772288, 0, "app_id_3_desc", "yx");


*/
func oneTestScope(fn func(db *sql.DB)) {
	// init db
	mysqlConf := &mysql.Config{
		Addr:         "0.0.0.0:3306",
		Net:          "tcp",
		User:         "root",
		Passwd:       "",
		DBName:       "test",
		Timeout:      3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		ParseTime:    true,
	}

	dsn := mysqlConf.FormatDSN()
	fmt.Printf("\n\ndsn: %v\n\n", dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Errorf("db.Open failed. err=%v", err))
	}

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("db.Ping failed. err=%v", err))
	}

	// set log
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	fn(db)
}

func Test_ScanOne(t *testing.T) {
	oneTestScope(func(db *sql.DB) {
		queryRowFunc := func(dest interface{}, query string, args ...interface{}) error {
			stmt, err := db.Prepare(query)
			if err != nil {
				return err
			}
			defer stmt.Close()

			rows, err := stmt.Query(args...)
			if err != nil {
				return err
			}
			defer rows.Close()

			return ScanOne(rows, dest)
		}

		type args struct {
			dest  interface{}
			query string
			args  []interface{}
		}

		type want struct {
			data   interface{}
			hasErr bool
		}

		var (
			 id   int64
			name string
		)

		tests := []struct {
			name string
			args args
			want want
		}{
			{
				name: "t1",
				args: args{
					dest:  &app{},
					query: `select id, name, app_id, secret, sign, start_time, end_time, status, ctime, utime, operator from app`,
					args:  nil,
				},
				want: want{
					hasErr: false,
				},
			},
			{
				name: "t2",
				args: args{
					dest:  &app{},
					query: `select id from app where status = ?`,
					args:  []interface{}{0},
				},
				want: want{
					hasErr: false,
				},
			},
			{
				name: "t3",
				args: args{
					dest:  &name,
					query: `select name from app where id = ?`,
					args:  []interface{}{3},
				},
				want: want{
					data:   "t3",
					hasErr: false,
				},
			},
			{
				name: "t4",
				args: args{
					dest:  &id,
					query: `select id from app where id = ?`,
					args:  []interface{}{3},
				},
				want: want{
					data:   int64(3),
					hasErr: false,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queryRowFunc(tt.args.dest, tt.args.query, tt.args.args...)
				if err != nil && !tt.want.hasErr {
					t.Errorf("queryRowFunc failed. err: %v", err)
					return
				}

				t.Logf("ret: %+v, type: %T", tt.args.dest, tt.args.dest)

				if tt.name == "t3" { // 单个字段类型测试
					t3Name, ok := (tt.args.dest).(*string)

					t.Logf("ok: %v, name: %+v", ok, *t3Name)

					if !reflect.DeepEqual(*t3Name, tt.want.data) || !ok {
						t.Errorf("test name failed")
					}
				}

				if tt.name == "t4" { // 单个字段类型测试
					t4ID, ok := (tt.args.dest).(*int64)

					t.Logf("ok: %v, id: %+v", ok, *t4ID)

					if !reflect.DeepEqual(*t4ID, tt.want.data) || !ok {
						t.Errorf("test id failed")
					}
				}
			})
		}
	})
}

func Test_ScanSlice(t *testing.T) {
	oneTestScope(func(db *sql.DB) {
		queryFunc := func(dest interface{}, query string, args ...interface{}) error {
			stmt, err := db.Prepare(query)
			if err != nil {
				return err
			}
			defer stmt.Close()

			rows, err := stmt.Query(args...)
			if err != nil {
				return err
			}
			defer rows.Close()

			return ScanSlice(rows, dest)
		}

		type args struct {
			dest  interface{}
			query string
			args  []interface{}
		}

		type want struct {
			data    interface{}
			hasErr  bool
			errNote string
		}

		tests := []struct {
			name string
			args args
			want want
		}{
			{
				name: "t1",
				args: args{
					dest:  &[]app{},
					query: `select id, name, app_id, secret, sign, start_time, end_time, status, ctime, utime, operator from app`,
					args:  nil,
				},
				want: want{
					hasErr: false,
				},
			},
			{
				name: "t2",
				args: args{
					dest: &[]struct {
						ID     int64 `db:"id"`
						Status bool  `db:"status"`
					}{},
					query: `select id, status from app where status = ?`,
					args:  []interface{}{0},
				},
				want: want{
					hasErr: false,
				},
			},
			{
				name: "t3",
				args: args{
					dest:  &[]int64{},
					query: `select id from app where status = ?`,
					args:  []interface{}{0},
				},
				want: want{
					hasErr:  true,
					errNote: "dest 必须是切片, 且切片的元素必须是个结构体",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queryFunc(tt.args.dest, tt.args.query, tt.args.args...)
				if err != nil && !tt.want.hasErr {
					t.Errorf("queryFunc failed. err: %v", err)
					return
				}

				t.Logf("ret: %+v", tt.args.dest)
			})
		}
	})
}

func Test_Exec(t *testing.T) {
	oneTestScope(func(db *sql.DB) {
		execFunc := func(query string, args ...interface{}) (sql.Result, error) {
			stmt, err := db.Prepare(query)
			if err != nil {
				return nil, err
			}
			defer stmt.Close()

			return stmt.Exec(args...)
		}

		type args struct {
			query string
			args  []interface{}
		}

		type want struct {
			data    interface{}
			hasErr  bool
			errNote string
		}

		tests := []struct {
			name string
			args args
			want want
		}{
			{
				name: "t1",
				args: args{
					query: `update app set status = 1 where status = ?`,
					args:  []interface{}{0},
				},
				want: want{
					hasErr: false,
				},
			},
			{
				name: "t2",
				args: args{
					query: `update app set status = 0 where status = ?`,
					args:  []interface{}{1},
				},
				want: want{
					hasErr: false,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ret, err := execFunc(tt.args.query, tt.args.args...)
				if err != nil && !tt.want.hasErr {
					t.Errorf("execFunc failed. err: %v", err)
					return
				}

				affectedRow, err := ret.RowsAffected()

				t.Logf("ret: lastInsertID: %+v err: %v", affectedRow, err)

				lastInsertID, err := ret.LastInsertId()
				t.Logf("ret: lastInsertID: %+v, err: %v", lastInsertID, err)
			})

			time.Sleep(2 * time.Second)
		}
	})
}
