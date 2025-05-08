# User Management System

This project is a **User Management System** built with Go, featuring RESTful APIs, WebSocket real-time communication, and Kafka-based messaging.

## 🚀 Features

- 🔧 **User CRUD Operations** via REST API
- 🔄 **WebSocket** support for real-time user updates
- 📬 **Kafka Integration** for event-driven architecture
- 🗃️ **PostgreSQL** as the primary data store
- 🐳 **Docker** support for containerized setup

## 🛠️ Technologies

- **Go** – Backend language
- **PostgreSQL** – Database
- **Kafka** – Message broker
- **Docker & Docker Compose** – Service orchestration

## 📦 Prerequisites

Make sure you have the following installed:

- [Go 1.20+](https://golang.org/dl/)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [PostgreSQL](https://www.postgresql.org/)
- [Apache Kafka](https://kafka.apache.org/)

## 🛠️ Setup Instructions

1. Clone the repository:
   ```bash
   git clone https://github.com/Vishnuganb/UserManagement.git
   cd UserManagement
   ```

2. Start all required services using Docker:
   ```bash
   docker-compose up -d
   ```

3. Run the Go application:
   ```bash
   go run cmd/server/main.go
   ```

Once the server is running, the application will be accessible at:

- **REST API**: http://localhost:8080
- **WebSocket**: ws://localhost:8082/ws_users

---

## 📚 API Endpoints

### 🔸 REST API

- `GET /users` — Fetch all users
- `GET /users/{id}` — Fetch a user by ID
- `POST /users` — Create a new user
- `PATCH /users/{id}` — Update a user
- `DELETE /users/{id}` — Delete a user

### 🔸 WebSocket

- Connect to `ws://localhost:8082/ws_users` for real-time updates

---

## 🧰 Using Makefile

This project includes a `Makefile` to simplify common development tasks.

### 🔧 Commands

- Start a PostgreSQL container:
  ```bash
  make postgres
  ```

- Create the `userManagement` database:
  ```bash
  make createdb
  ```

- Drop the database (if exists):
  ```bash
  make dropdb
  ```

- Apply database migrations:
  ```bash
  make migrateup
  ```

- Roll back the last migration:
  ```bash
  make migratedown
  ```

- Generate Go code from SQL queries using `sqlc`:
  ```bash
  make sqlc
  ```

- Run tests:
  ```bash
  make test
  ```

> 💡 Note: Ensure [`migrate`](https://github.com/golang-migrate/migrate), [`sqlc`](https://docs.sqlc.dev/), and Docker are installed before using these commands.

