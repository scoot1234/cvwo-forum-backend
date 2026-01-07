# CVWO Forum Backend

## Overview

The **CVWO Forum Backend** is a REST API written in Go for a lightweight forum application.

- Visitors can browse **topics**, **posts**, and **comments** without logging in.
- Logged-in users can create topics/posts/comments.
- Content owners can edit/delete their own content.
- Privileged roles (**admin**, **moderator**) can delete inappropriate posts/comments where applicable.

---

## Tech Stack

- **Language / HTTP:** Go (`net/http`)
- **Router:** `github.com/go-chi/chi/v5`
- **CORS middleware:** `github.com/go-chi/cors`
- **Database:** PostgreSQL
- **ORM:** `gorm.io/gorm`, `gorm.io/driver/postgres`
- **Env loader:** `github.com/joho/godotenv`
- **Password hashing:** `golang.org/x/crypto/bcrypt`

> Dependencies are declared in `go.mod`.

---

## Project Structure

```text
CVWO-Backend/
├── controllers/
│   ├── auth_controller.go       # Sign-up and login
│   ├── topics_controller.go     # CRUD for topics
│   ├── posts_controller.go      # CRUD for posts
│   ├── comments_controller.go   # CRUD for comments (includes replies)
│   └── roles.go                 # Role helper (admin/moderator checks)
├── db/
│   ├── db.go                    # Database connection (GORM + Postgres)
│   └── seed.go                  # Seed default topics
├── models/
│   ├── user.go                  # User model (username, password hash, role)
│   ├── topics.go                # Topic model
│   ├── post.go                  # Post model
│   └── comment.go               # Comment model (supports parentCommentId)
├── types/
│   ├── user.go                  # Public user DTO (hides sensitive fields)
│   ├── topic.go                 # Topic response DTO + mapping helpers
│   ├── post.go                  # Post response DTO + mapping helpers
│   └── comment.go               # Comment response DTO + mapping helpers
├── utils/
│   └── http.go                  # DecodeJSON, WriteJSON, ParseUintParam
├── main.go                      # Entry point (middleware + routes + server start)
├── go.mod
└── go.sum
```

---

## Folder Responsibilities

```text
controllers/: Request handling, validation, authorization and DB operations.
models/:      GORM models and their associations.
types/:       DTOs used for API responses + mapping helpers.
utils/:       Shared HTTP helpers (JSON parsing/writing, param parsing).
db/:          Database connection logic + seeding.
```

---

## API Endpoints

All requests/responses use **JSON**.

### Auth

| Method | Endpoint        | Description |
|-------:|-----------------|-------------|
| POST   | `/auth/signup`  | Create a new user (password is hashed) |
| POST   | `/auth/login`   | Validate username/password |

**Signup body**
```json
{
  "username": "alice",
  "password": "secret"
}
```

**Login body**
```json
{
  "username": "alice",
  "password": "secret"
}
```

---

### Topics

| Method | Endpoint              | Description |
|-------:|------------------------|-------------|
| GET    | `/topics`              | List all topics |
| POST   | `/topics`              | Create a topic |
| PATCH  | `/topics/{topicId}`    | Update a topic (owner or privileged) |
| DELETE | `/topics/{topicId}`    | Delete a topic (owner or privileged) |

**Create topic body**
```json
{
  "title": "Announcements",
  "description": "Project updates and notices",
  "userId": 1
}
```

**Update topic body**
```json
{
  "userId": 1,
  "title": "Updated title",
  "description": "Updated description"
}
```

**Delete topic body**
```json
{
  "userId": 1
}
```

---

### Posts

| Method | Endpoint                       | Description |
|-------:|--------------------------------|-------------|
| GET    | `/topics/{topicId}/posts`      | List posts under a topic |
| POST   | `/topics/{topicId}/posts`      | Create a post under a topic |
| GET    | `/posts/{postId}`              | Get a single post |
| PATCH  | `/posts/{postId}`              | Update a post (owner or privileged) |
| DELETE | `/posts/{postId}`              | Delete a post (owner or privileged) |

**Create post body**
```json
{
  "userId": 1,
  "title": "Welcome to the forum",
  "body": "Feel free to post questions here!"
}
```

**Update post body**
```json
{
  "userId": 1,
  "title": "Edited title",
  "body": "Edited content"
}
```

**Delete post body**
```json
{
  "userId": 1
}
```

---

### Comments (and Replies)

| Method | Endpoint                      | Description |
|-------:|-------------------------------|-------------|
| GET    | `/posts/{postId}/comments`    | List comments for a post |
| POST   | `/posts/{postId}/comments`    | Create a comment (optionally as a reply) |
| PATCH  | `/comments/{commentId}`       | Update a comment (owner or privileged) |
| DELETE | `/comments/{commentId}`       | Delete a comment (owner or privileged) |

**Create comment body**
```json
{
  "userId": 1,
  "body": "Nice post!"
}
```

**Reply to a comment body**
```json
{
  "userId": 2,
  "body": "Agreed!",
  "parentCommentId": 10
}
```

**Update comment body**
```json
{
  "userId": 1,
  "body": "Updated comment text"
}
```

**Delete comment body**
```json
{
  "userId": 1
}
```

---

## Setup

### Prerequisites

- Go (matches `go.mod`)
- PostgreSQL

### Installation

```bash
git clone https://github.com/scoot1234/cvwo-forum-backend.git
cd cvwo-forum-backend
go mod tidy
```

### Environment Variables

Create a `.env` file in the project root:

```env
DATABASE_URL=postgres://USER:PASSWORD@localhost:5432/cvwo_forum?sslmode=disable
PORT=8080
```

### Run the Server

```bash
go run main.go
```

Server will start on `http://localhost:8080` (or the `PORT` you set).

---

## Notes

- The server uses **GORM AutoMigrate** on startup to create/update tables.
- A seed script exists to populate default topics (see `db/seed.go`).
- CORS is enabled for local development (configured in `main.go`).
- Response DTOs in `types/` ensure sensitive fields (like password hashes) are not returned by the API.

---
