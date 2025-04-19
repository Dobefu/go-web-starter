package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupUserTests() (user User) {
	return User{
		id:        1,
		username:  "Test User",
		email:     "test@user.com",
		status:    true,
		createdAt: time.Unix(100000, 0),
		updatedAt: time.Unix(200000, 0),
	}
}

func TestUserGetID(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, 1, user.GetID())
}

func TestUserGetUsername(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, "Test User", user.GetUsername())
}

func TestUserGetEmail(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, "test@user.com", user.GetEmail())
}

func TestUserGetStatus(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, true, user.GetStatus())
}

func TestUserGetCreatedAt(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, time.Unix(100000, 0), user.GetCreatedAt())
}

func TestUserGetUpdatedAt(t *testing.T) {
	t.Parallel()

	user := setupUserTests()

	assert.Equal(t, time.Unix(200000, 0), user.GetUpdatedAt())
}
