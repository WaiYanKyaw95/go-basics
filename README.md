# go-basics

A bare-bones HTTP server written in Go — no frameworks, no magic.

## Why

I wanted to learn backend without black boxes. Instead of starting with a framework that handles everything, I built it from scratch to understand what is actually happening underneath.

## What it does

- Raw TCP socket server
- HTTP request parsing by hand
- CRUD API for items
- SQLite persistence with raw SQL
- Authentication — register, login, logout
- Session-based authorization — protected routes require a valid session token
- Password hashing with bcrypt

## Things learned along the way

- Started with hardcoded responses, then refactored into helper functions `writeJSON` and `writeText` to remove repetition
- `json:"-"` on a struct field blocks both marshaling AND unmarshaling — learned this the hard way when bcrypt was accepting any password. Fixed by creating a separate `LoginInput` struct for reading request bodies
- `range` in a loop gives you a copy of the element, not the original — modifying it does nothing to the slice. Use the index instead
- Go's error handling is verbose but honest — every function that can fail tells you explicitly
- `defer` runs when the function exits, not where you write it — useful for closing connections and database rows

## Architecture Mapping

| This project | Django equivalent |
|---|---|
| `net.Listen()` + accept loop | `manage.py runserver` |
| `parseRequest()` | `urls.py` |
| if/else if routing | `views.py` |
| `db.Exec()` / `db.Query()` | `models.py` |
| sessions table + bcrypt | `django.contrib.auth` |

## Stack

- Go (stdlib)
- SQLite via go-sqlite3
- bcrypt via golang.org/x/crypto

## Routes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | / | No | Hello world |
| GET | /items | No | List all items |
| GET | /items/:id | No | Get one item |
| POST | /items | Yes | Create item |
| PUT | /items/:id | Yes | Update item |
| DELETE | /items/:id | Yes | Delete item |
| POST | /register | No | Create account |
| POST | /login | No | Get session token |
| POST | /logout | Yes | Delete session |