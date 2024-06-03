package main

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// Entity
type Post struct {
	PostID      string         `db:"id" json:"id"`
	Title       string         `db:"title" json:"title"`
	Content     string         `db:"content" json:"content"`
	Status      string         `db:"status" json:"status"`
	PublishDate sql.NullTime   `db:"publish_date" json:"-"`
	Tags        pq.StringArray `db:"tags" json:"tags"`
}

type Account struct {
	AccountID string `db:"id" json:"id"`
	Username  string `db:"username" json:"username"`
	Password  string `db:"password" json:"-"`
	Name      string `db:"name" json:"name"`
	Role      string `db:"role" json:"role"`
}

func (p Post) MarshalJSON() ([]byte, error) {
	type Alias Post
	return json.Marshal(&struct {
		PublishDate *string `json:"publish_date"`
		Alias
	}{
		PublishDate: func() *string {
			if p.PublishDate.Valid {
				s := p.PublishDate.Time.Format(time.RFC3339)
				return &s
			}
			return nil
		}(),
		Alias: (Alias)(p),
	})
}

type CustomClaims struct {
	AccountID string `json:"account_id"`
	Role      string `json:"role"`
	Exp       int64  `json:"exp"`
}

// DTO
type DefaultResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type InsertOrUpdatePost struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

type CreateAccount struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginAccount struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
