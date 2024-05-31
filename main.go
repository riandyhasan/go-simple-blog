package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func migrateDb(db *sql.DB) error {
	filePath := "./migration.sql"
	migrationFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer migrationFile.Close()

	migration, err := io.ReadAll(migrationFile)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(migration))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName)

	var db *sql.DB
	var err error

	// Wait db until ready
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Println("Error connecting to database:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		err = db.Ping()
		if err == nil {
			break
		}
		log.Println("Database not ready, retrying in 2 seconds...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	err = migrateDb(db)
	if err != nil {
		log.Fatal(err)
	}

	handlerDeps := NewHandlerDependencies(db)
	http.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			DefaultMiddleware(handlerDeps.CreatePostHandler)(w, r)
		case http.MethodGet:
			DefaultMiddleware(handlerDeps.SearchPostByTag)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			DefaultMiddleware(handlerDeps.UpdatePost)(w, r)
		case http.MethodDelete:
			DefaultMiddleware(handlerDeps.DeletePost)(w, r)
		case http.MethodGet:
			DefaultMiddleware(handlerDeps.GetPost)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/posts/publish", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			DefaultMiddleware(handlerDeps.PublishPost)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Listening on : http://127.0.0.1:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
