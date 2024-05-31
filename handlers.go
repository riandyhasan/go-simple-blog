package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

type HandlerDependencies struct {
	DB *sql.DB
}

func NewHandlerDependencies(db *sql.DB) *HandlerDependencies {
	return &HandlerDependencies{
		DB: db,
	}
}

func (d *HandlerDependencies) CreatePostHandler(md *MiddlewareContext) error {
	createPost := InsertOrUpdatePost{}

	if err := json.NewDecoder(md.Request.Body).Decode(&createPost); err != nil {
		return md.ReturnError(http.StatusBadRequest, "Request tidak sesuai format")
	}

	var postID string
	err := d.DB.QueryRow("INSERT INTO posts (title, content, tags) VALUES ($1, $2, $3) RETURNING id", createPost.Title, createPost.Tags, createPost.Tags).Scan(&postID)
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal membuat post karena kesalahan dari sistem")
	}

	return md.ReturnSuccess(createPost)
}

func (d *HandlerDependencies) UpdatePost(md *MiddlewareContext) error {
	updatePost := InsertOrUpdatePost{}

	if err := json.NewDecoder(md.Request.Body).Decode(&updatePost); err != nil {
		return md.ReturnError(http.StatusBadRequest, "Request tidak sesuai format")
	}

	paths := strings.Split(md.Request.URL.Path, "/")
	if len(paths) < 4 {
		return md.ReturnError(http.StatusBadRequest, "Post ID tidak ditemukan")
	}
	postID := paths[3]
	res, err := d.DB.Exec("UPDATE posts SET title = $1, content = $2, tags = $3 WHERE id = $4", updatePost.Title, updatePost.Content, updatePost.Tags, postID)
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal mengupdate post karena kesalahan dari sistem")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal mengupdate post karena kesalahan dari sistem")
	}

	if affected < 1 {
		return md.ReturnError(http.StatusNotFound, "Post dengan ID tersebut tidak ditemukan")
	}

	return md.ReturnSuccess(updatePost)
}

func (d *HandlerDependencies) PublishPost(md *MiddlewareContext) error {

	paths := strings.Split(md.Request.URL.Path, "/")
	if len(paths) < 4 {
		return md.ReturnError(http.StatusBadRequest, "Post ID tidak ditemukan")
	}
	postID := paths[3]
	res, err := d.DB.Exec("UPDATE posts SET publish_at = NOW(), status = publish WHERE id = $2", postID)
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal publish post karena kesalahan dari sistem")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal publish post karena kesalahan dari sistem")
	}

	if affected < 1 {
		return md.ReturnError(http.StatusNotFound, "Post dengan ID tersebut tidak ditemukan")
	}

	return md.ReturnSuccess(fmt.Sprintf("Post dengan ID %s berhasil dipublish", postID))
}

func (d *HandlerDependencies) DeletePost(md *MiddlewareContext) error {
	paths := strings.Split(md.Request.URL.Path, "/")
	if len(paths) < 4 {
		return md.ReturnError(http.StatusBadRequest, "Post ID tidak ditemukan")
	}
	postID := paths[3]

	res, err := d.DB.Exec("DELETE FROM posts WHERE id = $1", postID)
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal menghapus post karena kesalahan dari sistem")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal menghapus post karena kesalahan dari sistem")
	}

	if affected < 1 {
		return md.ReturnError(http.StatusNotFound, "Post dengan ID tersebut tidak ditemukan")
	}

	return md.ReturnSuccess(fmt.Sprintf("Post dengan ID %s berhasil dihapus", postID))
}

func (d *HandlerDependencies) GetPost(md *MiddlewareContext) error {
	paths := strings.Split(md.Request.URL.Path, "/")
	if len(paths) < 4 {
		return md.ReturnError(http.StatusBadRequest, "Post ID tidak ditemukan")
	}
	postID := paths[3]

	post := Post{}
	err := d.DB.QueryRow("SELECT title, content, tags, publish_date, status FROM posts WHERE id = $1", postID).Scan(&post.Title, &post.Content, &post.Tags, &post.PublishDate, &post.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return md.ReturnError(http.StatusNotFound, "Post tidak ditemukan")
		}
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal mengambil post karena kesalahan dari sistem")
	}

	jsonData, err := post.MarshalJSON()
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal mengambil post karena kesalahan dari sistem")
	}

	return md.ReturnSuccess(jsonData)
}

func (d *HandlerDependencies) SearchPostByTag(md *MiddlewareContext) error {
	tag := md.Request.URL.Query().Get("tag")
	rows, err := d.DB.Query("SELECT id, title, content, tags FROM posts WHERE $1 = ANY(tags)", tag)
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal memproses data post karena kesalahan dari sistem")
	}
	defer rows.Close()

	var posts [][]byte
	for rows.Next() {
		var post Post
		var postID string
		if err := rows.Scan(&postID, &post.Title, &post.Content, &post.Tags); err != nil {
			log.Println(err)
			return md.ReturnError(http.StatusInternalServerError, "Gagal memproses hasil pencarian post")
		}
		jsonData, err := post.MarshalJSON()
		if err != nil {
			log.Println(err)
			return md.ReturnError(http.StatusInternalServerError, "Gagal memproses hasil pencarian post")

		}
		posts = append(posts, jsonData)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal memproses hasil pencarian post")
	}

	if posts == nil {
		posts = make([][]byte, 0)
	}

	return md.ReturnSuccess(posts)
}
