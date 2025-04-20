package user

import (
	"database/sql"
	"time"

	"github.com/Dobefu/go-web-starter/internal/database"
)

const (
	insertUserQuery = `INSERT INTO users (id, username, email, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username, email = EXCLUDED.email, status = EXCLUDED.status, updated_at = NOW() RETURNING id`
)

type User struct {
	id        int
	username  string
	email     string
	status    bool
	createdAt time.Time
	updatedAt time.Time
}

func (user *User) GetID() (id int) {
	return user.id
}

func (user *User) GetUsername() (username string) {
	return user.username
}

func (user *User) GetEmail() (email string) {
	return user.email
}

func (user *User) GetStatus() (status bool) {
	return user.status
}

func (user *User) GetCreatedAt() (createdAt time.Time) {
	return user.createdAt
}

func (user *User) GetUpdatedAt() (updatedAt time.Time) {
	return user.updatedAt
}

func (user *User) Save(db database.DatabaseInterface) (err error) {
	row := db.QueryRow(insertUserQuery,
		user.id,
		user.username,
		user.email,
		user.status,
		user.createdAt,
		user.updatedAt,
	)

	if row == nil {
		return sql.ErrConnDone
	}

	err = row.Scan(&user.id)

	if err != nil {
		return err
	}

	return nil
}
