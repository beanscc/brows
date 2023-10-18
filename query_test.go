package brows

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func testDB(t *testing.T) *sql.DB {
	conf := &mysql.Config{
		Addr:                 "127.0.0.1:3306",
		Net:                  "tcp",
		User:                 "root",
		Passwd:               "P4m@bpet",
		DBName:               "test",
		Timeout:              3 * time.Second,
		ReadTimeout:          3 * time.Second,
		WriteTimeout:         3 * time.Second,
		ParseTime:            true,
		AllowNativePasswords: true,
	}

	dsn := conf.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}
	return db
}

type TestAppStatus int

type TestApp struct {
	ID     int64         `db:"id"`
	Name   string        `db:"name"`
	AppID  string        `db:"app_id"`
	Secret string        `db:"secret"`
	Sign   string        `db:"sign"`
	Status TestAppStatus `db:"status"`
	// EndTime   int64     `db:"end_time"`
	// StartTime int64     `db:"start_time"`

	// 内嵌类型
	TestAppTime

	Ctime    time.Time `db:"ctime"`
	Utime    time.Time `db:"utime"`
	Operator string    `db:"operator"`
}

type TestAppTime struct {
	EndTime   int64  `db:"end_time"`
	StartTime *int64 `db:"start_time"`
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
COMMENT = '应用APP';

-- 添加测试数据

INSERT INTO `test`.`app` (`name`,`app_id`,`secret`,`sign`,`start_time`,`end_time`,`status`,`description`,`operator`,`ctime`,`utime`) VALUES
('t1', "app_id_1_dfsdfsdfsdfsddf", "secret_1_sfsdfsdfsdfsd", "sign_1_dsfsdfvsdghadfg", 14811110152, 1523772288, 0, "app_id_1_desc", "yx1", '2022-10-01 01:01:01', '2022-10-11 11:11:11'),
('t2', "app_id_2_dfsdfsdfsdfsddf", "secret_2_sfsdfsdfsdfsd", "sign_2_dsfsdfvsdghadfg", 14811110152, 1523872288, 1, "app_id_2_desc", "yx2", '2022-10-02 02:02:02', '2022-10-22 02:02:02'),
('t3', "app_id_3_dfsdfsdfsdfsddf", "secret_3_sfsdfsdfsdfsd", "sign_4_dsfsdfvsdghadfg", 14811110152, 1523972288, 0, "app_id_3_desc", "yx3", '2022-10-03 03:03:03', '2022-10-23 02:02:02');
*/

func TestBrows_QueryRow(t *testing.T) {
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
				dest:  &TestApp{},
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
				dest:  &TestApp{},
				query: `select id from app where status = ?`,
				args:  []interface{}{1},
			},
			want: want{
				hasErr: false,
			},
		},
		{
			name: "t2-1",
			args: args{
				dest:  &TestApp{},
				query: `select start_time, end_time, id, name, app_id, secret, sign,  status, ctime, utime, operator from app where id = ?`,
				args:  []interface{}{2},
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
				hasErr:  true,
				errNote: "dest must be non-nil pointer to a struct",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(testDB(t)).QueryRow(tt.args.query, tt.args.args...).Scan(tt.args.dest)
			if err != nil && !tt.want.hasErr {
				t.Errorf("QueryRow failed. err: %v", err)
				return
			}

			rv := reflect.ValueOf(tt.args.dest)
			t.Logf("ret value type: %s", rv.Type())
			t.Logf("ret value elem: %#v", rv.Elem().Interface())
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
				dest:  &[]TestApp{},
				query: `select id, name, app_id, secret, sign, start_time, end_time, status, ctime, utime, operator from app`,
				args:  nil,
			},
			want: want{
				hasErr: false,
			},
		},
		{
			name: "t1.1",
			args: args{
				dest:  &[]*TestApp{},
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
				hasErr: true,
			},
		},
		{
			name: "t3.1",
			args: args{
				dest:  &[]*int64{},
				query: `select id from app where status = ?`,
				args:  []interface{}{0},
			},
			want: want{
				hasErr: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(testDB(t)).Query(tt.args.query, tt.args.args...).Scan(tt.args.dest)
			if err != nil && !tt.want.hasErr {
				t.Errorf("QueryRow failed. err: %v", err)
				return
			}
			t.Logf("ret: %#v", tt.args.dest)
			if v := reflect.ValueOf(tt.args.dest); reflect.Ptr == v.Kind() && reflect.Slice == v.Elem().Kind() {
				for i := 0; i < v.Elem().Len(); i++ {
					t.Logf("ret slice idx:%2d, elem:%#v", i, v.Elem().Index(i))
				}
			}
		})
	}
}
