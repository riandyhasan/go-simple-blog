package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

	post := Post{
		Title:   createPost.Title,
		Content: createPost.Content,
		Tags:    createPost.Tags,
		Status:  "draft",
	}

	err := d.DB.QueryRow("INSERT INTO posts (title, content, tags, status) VALUES ($1, $2, $3, $4) RETURNING id", post.Title, post.Content, post.Tags, post.Status).Scan(&post.PostID)
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal membuat post karena kesalahan dari sistem")
	}

	return md.ReturnSuccess(post)
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
	post := Post{
		PostID:  postID,
		Title:   updatePost.Title,
		Content: updatePost.Content,
		Tags:    updatePost.Tags,
	}
	res, err := d.DB.Exec("UPDATE posts SET title = $1, content = $2, tags = $3 WHERE id = $4", post.Title, post.Content, post.Tags, postID)
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

	return md.ReturnSuccess(post)
}

func (d *HandlerDependencies) PublishPost(md *MiddlewareContext) error {

	paths := strings.Split(md.Request.URL.Path, "/")
	if len(paths) < 5 {
		return md.ReturnError(http.StatusBadRequest, "Post ID tidak ditemukan")
	}
	postID := paths[4]
	res, err := d.DB.Exec("UPDATE posts SET publish_date = NOW(), status = 'publish' WHERE id = $1", postID)
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
	err := d.DB.QueryRow("SELECT id, title, content, tags, publish_date, status FROM posts WHERE id = $1 LIMIT 1", postID).Scan(&post.PostID, &post.Title, &post.Content, &post.Tags, &post.PublishDate, &post.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return md.ReturnError(http.StatusNotFound, "Post tidak ditemukan")
		}
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal mengambil post karena kesalahan dari sistem")
	}

	return md.ReturnSuccess(post)
}

func (d *HandlerDependencies) SearchPostByTag(md *MiddlewareContext) error {
	tag := md.Request.URL.Query().Get("tag")
	pageStr := md.Request.URL.Query().Get("page")
	limitStr := md.Request.URL.Query().Get("limit")

	var err error

	page := 1
	limit := 50

	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 50
		}
	}

	offset := (page - 1) * limit

	query := `
		SELECT id, title, content, tags, status, publish_date
		FROM posts
		WHERE $1 = ANY(tags)
		ORDER BY publish_date DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := d.DB.Query(query, tag, limit, offset)
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal memproses data post karena kesalahan dari sistem")
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.PostID, &post.Title, &post.Content, &post.Tags, &post.Status, &post.PublishDate); err != nil {
			log.Println(err)
			return md.ReturnError(http.StatusInternalServerError, "Gagal memproses hasil pencarian post")
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal memproses hasil pencarian post")
	}

	if posts == nil {
		posts = make([]Post, 0)
	}

	countQuery := `
		SELECT COUNT(*)
		FROM posts
		WHERE $1 = ANY(tags)
	`
	var total int
	err = d.DB.QueryRow(countQuery, tag).Scan(&total)
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal menghitung total post")
	}

	var paginationResponse struct {
		Data     interface{} `json:"data"`
		Page     int         `json:"page"`
		PageSize int         `json:"pageSize"`
		Total    int         `json:"total"`
	}

	paginationResponse.Data = posts
	paginationResponse.Page = page
	paginationResponse.PageSize = limit
	paginationResponse.Total = total

	return md.ReturnSuccess(paginationResponse)
}

func (d *HandlerDependencies) Register(md *MiddlewareContext) error {
	createAccount := CreateAccount{}

	if err := json.NewDecoder(md.Request.Body).Decode(&createAccount); err != nil {
		return md.ReturnError(http.StatusBadRequest, "Request tidak sesuai format")
	}

	account := Account{
		Username: createAccount.Username,
		Name:     createAccount.Name,
		Password: HashPassword(createAccount.Password),
		Role:     createAccount.Role,
	}

	err := d.DB.QueryRow("INSERT INTO accounts (username, password, name, role) VALUES ($1, $2, $3, $4) RETURNING id", account.Username, account.Password, account.Name, account.Role).Scan(&account.AccountID)
	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal membuat akun karena kesalahan dari sistem")
	}

	return md.ReturnSuccess("Berhasil membuat akun")
}

func (d *HandlerDependencies) Login(md *MiddlewareContext) error {
	loginAccount := LoginAccount{}

	if err := json.NewDecoder(md.Request.Body).Decode(&loginAccount); err != nil {
		return md.ReturnError(http.StatusBadRequest, "Request tidak sesuai format")
	}
	account := Account{}
	err := d.DB.QueryRow("SELECT id, username, password, name, role FROM accounts WHERE username = $1 LIMIT 1", loginAccount.Username).Scan(&account.AccountID, &account.Username, &account.Password, &account.Name, &account.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return md.ReturnError(http.StatusUnauthorized, "Username atau password salah")
		}
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal login karena kesalahan dari sistem")
	}

	hashedPassword := HashPassword(loginAccount.Password)
	if hashedPassword != account.Password {
		return md.ReturnError(http.StatusUnauthorized, "Username atau password salah")
	}

	var loginResponse struct {
		Account Account `json:"account"`
		Token   string  `json:"token"`
	}

	token, err := GenerateJWT(account.AccountID, account.Role)

	if err != nil {
		log.Println(err)
		return md.ReturnError(http.StatusInternalServerError, "Gagal login karena kesalahan dari sistem")
	}

	loginResponse.Account = account
	loginResponse.Token = token

	return md.ReturnSuccess(loginResponse)
}
