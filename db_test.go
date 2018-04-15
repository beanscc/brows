package brows_test

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/beanscc/brows"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
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

	if err := Conn(mysqlConf); err != nil {
		panic(err)
	}

	// set log
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// Conn 根据配置连接 mysql
func Conn(c *mysql.Config) error {
	dsn := c.FormatDSN()
	fmt.Printf("\n\ndsn: %v\n\n", dsn)
	mysqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err := mysqlDB.Ping(); err != nil {
		return err
	}

	db = mysqlDB

	return nil
}

type App struct {
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

func TestBrows_QueryRow(t *testing.T) {
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
		// id   int64
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
				dest:  &App{},
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
				dest:  &App{},
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
	}
	b := brows.New(db)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := b.QueryRow(tt.args.dest, tt.args.query, tt.args.args...)
			if err != nil && !tt.want.hasErr {
				t.Errorf("QueryRow failed. err: %v", err)
				return
			}

			t.Logf("ret: %+v, type: %T", tt.args.dest, tt.args.dest)

			if tt.name == "t3" { // 单个字段类型测试
				name, ok := (tt.args.dest).(*string)

				t.Logf("ok: %v, name: %+v", ok, *name)

				if !reflect.DeepEqual(*name, tt.want.data) || !ok {
					t.Errorf("test name failed")
				}
			}
		})
	}
}

func TestBrows_Query(t *testing.T) {
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
				dest:  &[]App{},
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

	b := brows.New(db)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := b.Query(tt.args.dest, tt.args.query, tt.args.args...)
			if err != nil && !tt.want.hasErr {
				t.Errorf("QueryRow failed. err: %v", err)
				return
			}

			t.Logf("ret: %+v", tt.args.dest)
		})
	}
}

// go test -v -run TestBrows_Exec
func TestBrows_Exec(t *testing.T) {
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

	b := brows.New(db)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, err := b.Exec(tt.args.query, tt.args.args...)
			if err != nil && !tt.want.hasErr {
				t.Errorf("QueryRow failed. err: %v", err)
				return
			}

			affectedRow, err := ret.RowsAffected()

			t.Logf("ret: lastInsertID: %+v err: %v", affectedRow, err)

			lastInsertID, err := ret.LastInsertId()
			t.Logf("ret: lastInsertID: %+v, err: %v", lastInsertID, err)
		})

		time.Sleep(2 * time.Second)
	}
}
