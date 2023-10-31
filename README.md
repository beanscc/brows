# brows

将 sql.Rows 绑定赋值给结构体

## Installation

- install brows
```bash
go get github.com/beanscc/brows
```

- import 
```go
import "github.com/beanscc/brows"
 ```

## Quick start

```go
package main

import (
	"database/sql"

	"github.com/beanscc/brows"
)

func main() {
	type User struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
		Age  uint   `db:"age"`
	}

	var users []User
	db := getDb()

	query := `select id,name,age from test where age > ?`
	err := brows.New(db).Query(query, 10).Scan(&users)
	if err != nil {
		// error handle
	}
}

func getDb() *sql.DB {
	var db *sql.DB
	// init db

	return db
}
```