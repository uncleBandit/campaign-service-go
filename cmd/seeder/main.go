//cmd/seeder/main.go
package main

import (
    "database/sql"
    "fmt"
    "io/ioutil"
    "log"
    "os"

    _ "github.com/lib/pq"
)

func main() {
    dsn := os.Getenv("DATABASE_URL")
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    seedFiles := []string{
        "seed/customers.sql",
        "seed/campaigns.sql",
    }

    for _, file := range seedFiles {
        content, err := ioutil.ReadFile(file)
        if err != nil {
            log.Fatalf("failed to read %s: %v", file, err)
        }

        _, err = db.Exec(string(content))
        if err != nil {
            log.Fatalf("failed to execute %s: %v", file, err)
        }
        fmt.Printf("Seeded: %s\n", file)
    }

    fmt.Println("Database seeding completed successfully!")
}
