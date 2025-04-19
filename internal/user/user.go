package user

import "time"

type User struct {
	id        int       `json:"id"`
	username  string    `json:"username"`
	email     string    `json:"email"`
	password  string    `json:"-"`
	status    bool      `json:"status"`
	createdAt time.Time `json:"created_at"`
	updatedAt time.Time `json:"updated_at"`
}
