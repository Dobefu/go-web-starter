package user

import "time"

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
