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

	http.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			DefaultMiddleware(handlerDeps.Login)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			DefaultMiddleware(handlerDeps.Register)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			AuthMiddleware(handlerDeps.CreatePostHandler, []string{"user"})(w, r)
		case http.MethodGet:
			AuthMiddleware(handlerDeps.SearchPostByTag, []string{"user", "admin"})(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			AuthMiddleware(handlerDeps.UpdatePost, []string{"user", "admin"})(w, r)
		case http.MethodDelete:
			AuthMiddleware(handlerDeps.DeletePost, []string{"user", "admin"})(w, r)
		case http.MethodGet:
			AuthMiddleware(handlerDeps.GetPost, []string{"user", "admin"})(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/posts/publish/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			AuthMiddleware(handlerDeps.PublishPost, []string{"admin"})(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./blogs.html")
	})

	http.HandleFunc("/posts/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./blog-detail.html")
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./login.html")
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./register.html")
	})

	fmt.Println("Listening on : http://127.0.0.1:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
