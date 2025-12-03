// internal/db/db.go
package db

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
    "log"
    "os"
)

var DB *sql.DB

func Init() {
    user := os.Getenv("DB_USER")
    pass := os.Getenv("DB_PASSWORD")
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    name := os.Getenv("DB_NAME")

    log.Println("DB_USER:", user)
    log.Println("DB_NAME:", name)
    log.Println("DB_HOST:", host)

    dsn := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=disable",
        user, pass, host, port, name,
    )

    var err error
    DB, err = sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("failed to connect to DB: %v", err)
    }

    if err = DB.Ping(); err != nil {
        log.Fatalf("failed to ping DB: %v", err)
    }

    log.Println("✅ Connected to database")
}
