# brows

将 sql.Rows 绑定赋值给结构体

## Installing

```bash
go get github.com/beanscc/brows
```

## Quick start

```go
	type User struct {
        ID   int64  `db:"id"`
        Name string `db:"name"`
        Age  uint   `db:"age"`
    }
	
    db := getDb()
	query := `select id,name,age from test where age > ?`
	var users []User
    err := brows.New(db).QueryRow(query, 10).Scan(&user)
	if err != nil {
	    // error handle	
    }   
```