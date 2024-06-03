# Go Simple Blog

This is go simple blog, an app for writing a simple blog.

## How to Run:

### Via Docker

- Run: `make up`

### Via Non-Docker

- Make sure you have Go and PostgreSQL installed.
- Change the environment variables accordingly.
- Run: `make run`

## Endpoints

### Authentication

- **Login:** `POST /api/auth/login`
- **Register:** `POST /api/auth/register`

### Blog Posts

- **Create Post:** `POST /api/posts` (Auth required: roles `user`)
- **Search Posts by Tag:** `GET /api/posts` (Auth required: roles `user`, `admin`)
- **Update Post:** `PUT /api/posts/{id}` (Auth required: roles `user`, `admin`)
- **Delete Post:** `DELETE /api/posts/{id}` (Auth required: roles `user`, `admin`)
- **Get Post by ID:** `GET /api/posts/{id}` (Auth required: roles `user`, `admin`)
- **Publish Post:** `PUT /api/posts/publish/{id}` (Auth required: role `admin`)

### Pages

- **Blog List:** `GET /posts`
- **Blog Detail:** `GET /posts/{id}`
- **Login Page:** `GET /login`
- **Register Page:** `GET /register`
