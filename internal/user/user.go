package user

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Dobefu/go-web-starter/internal/database"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

const (
	insertUserQuery         = `INSERT INTO users (username, email, password, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	findUserByEmailQuery    = `SELECT id, username, email, password, status, created_at, updated_at FROM users WHERE email = $1`
	findUserByUsernameQuery = `SELECT id, username, email, password, status, created_at, updated_at FROM users WHERE username = $1`
	findUserByIDQuery       = `SELECT id, username, email, password, status, created_at, updated_at FROM users WHERE id = $1`
)

type User struct {
	id        int
	username  string
	email     string
	password  string
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

func (user *User) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password))

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}

		return fmt.Errorf("error comparing password hash: %w", err)
	}

	return nil
}

func (user *User) Save(db database.DatabaseInterface) (err error) {
	row := db.QueryRow(insertUserQuery,
		user.username,
		user.email,
		user.password,
		user.status,
		time.Now(),
		time.Now(),
	)

	err = row.Scan(&user.id, &user.createdAt, &user.updatedAt)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func FindByEmail(db database.DatabaseInterface, email string) (*User, error) {
	user := &User{}
	row := db.QueryRow(findUserByEmailQuery, email)

	err := row.Scan(
		&user.id,
		&user.username,
		&user.email,
		&user.password,
		&user.status,
		&user.createdAt,
		&user.updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("error finding user by email: %w", err)
	}

	return user, nil
}

func FindByUsername(db database.DatabaseInterface, username string) (*User, error) {
	user := &User{}
	row := db.QueryRow(findUserByUsernameQuery, username)

	err := row.Scan(
		&user.id,
		&user.username,
		&user.email,
		&user.password,
		&user.status,
		&user.createdAt,
		&user.updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("error finding user by username: %w", err)
	}

	return user, nil
}

func FindByID(db database.DatabaseInterface, id int) (*User, error) {
	user := &User{}
	row := db.QueryRow(findUserByIDQuery, id)

	err := row.Scan(
		&user.id,
		&user.username,
		&user.email,
		&user.password,
		&user.status,
		&user.createdAt,
		&user.updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}

		return nil, fmt.Errorf("error finding user by ID: %w", err)
	}

	return user, nil
}

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

func NewUser(username, email, hashedPassword string, status bool) *User {
	return &User{
		username: username,
		email:    email,
		password: hashedPassword,
		status:   status,
	}
}

type UserRepository interface {
	FindByEmail(email string) (*User, error)
	SaveUser(user *User) error
}

type DbUserRepository struct {
	DB database.DatabaseInterface
}

func (r *DbUserRepository) FindByEmail(email string) (*User, error) {
	return FindByEmail(r.DB, email)
}

func (r *DbUserRepository) SaveUser(user *User) error {
	return user.Save(r.DB)
}

func CreateWithRepo(repo UserRepository, username, email, plainPassword string) (*User, error) {
	_, findErr := repo.FindByEmail(email)

	if findErr == nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	} else if !errors.Is(findErr, ErrInvalidCredentials) {
		return nil, fmt.Errorf("database error checking for existing email: %w", findErr)
	}

	hashedPassword, hashErr := HashPassword(plainPassword)

	if hashErr != nil {
		return nil, fmt.Errorf("failed to hash password: %w", hashErr)
	}

	newUser := NewUser(username, email, hashedPassword, true)

	if saveErr := repo.SaveUser(newUser); saveErr != nil {
		return nil, fmt.Errorf("failed to save new user: %w", saveErr)
	}

	return newUser, nil
}

func Create(db database.DatabaseInterface, username, email, plainPassword string) (*User, error) {
	repo := &DbUserRepository{DB: db}
	return CreateWithRepo(repo, username, email, plainPassword)
}

func New(id int, username, email, password string, status bool, createdAt, updatedAt time.Time) *User {
	return &User{
		id:        id,
		username:  username,
		email:     email,
		password:  password,
		status:    status,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}
