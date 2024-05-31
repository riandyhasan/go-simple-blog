package main

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Entity
type Post struct {
	PostID      string       `db:"id" json:"id"`
	Title       string       `db:"title" json:"title"`
	Content     string       `db:"content" json:"content"`
	Status      string       `db:"status" json:"status"`
	PublishDate sql.NullTime `db:"publish_date" json:"-"`
	Tags        []string     `db:"tags" json:"tags"`
}

// Special converter for time format
func (p Post) MarshalJSON() ([]byte, error) {
	type Alias Post
	return json.Marshal(&struct {
		PublishDate time.Time `json:"publish_date"`
		Alias
	}{
		PublishDate: p.PublishDate.Time,
		Alias:       (Alias)(p),
	})
}

func (p *Post) UnmarshalJSON(data []byte) error {
	type Alias Post
	aux := &struct {
		PublishDate *time.Time `json:"publish_date"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.PublishDate != nil {
		p.PublishDate = sql.NullTime{Time: *aux.PublishDate, Valid: true}
	} else {
		p.PublishDate = sql.NullTime{Valid: false}
	}
	return nil
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
