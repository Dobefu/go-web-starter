package user

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"crypto/sha256"
	"encoding/hex"

	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/gin-contrib/sessions"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrNotActive          = errors.New("the user is not active")
)

const (
	insertUserQuery         = `INSERT INTO users (username, email, password, status, created_at, updated_at, last_login) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at, last_login`
	findUserByEmailQuery    = `SELECT id, username, email, password, status, created_at, updated_at, last_login FROM users WHERE email = $1`
	findUserByUsernameQuery = `SELECT id, username, email, password, status, created_at, updated_at, last_login FROM users WHERE username = $1`
	findUserByIDQuery       = `SELECT id, username, email, password, status, created_at, updated_at, last_login FROM users WHERE id = $1`
	updateUserQuery         = `UPDATE users SET username = $1, email = $2, password = $3, status = $4, updated_at = $5, last_login = $6 WHERE id = $7 RETURNING updated_at`
)

type User struct {
	id        int
	username  string
	email     string
	password  string
	status    bool
	createdAt time.Time
	updatedAt time.Time
	lastLogin time.Time
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

func (user *User) SetStatus(status bool) {
	user.status = status
}

func (user *User) GetCreatedAt() (createdAt time.Time) {
	return user.createdAt
}

func (user *User) GetUpdatedAt() (updatedAt time.Time) {
	return user.updatedAt
}

func (user *User) GetLastLogin() (lastLogin time.Time) {
	return user.lastLogin
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
	if user.id == 0 {
		row := db.QueryRow(insertUserQuery,
			user.username,
			user.email,
			user.password,
			user.status,
			time.Now(),
			time.Now(),
			time.UnixMicro(0),
		)

		err = row.Scan(&user.id, &user.createdAt, &user.updatedAt, &user.lastLogin)

		if err != nil {
			return err
		}

		return nil
	}

	row := db.QueryRow(updateUserQuery,
		user.username,
		user.email,
		user.password,
		user.status,
		time.Now(),
		user.lastLogin,
		user.id,
	)

	var updatedAt time.Time
	err = row.Scan(&updatedAt)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	user.updatedAt = updatedAt
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
		&user.lastLogin,
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
		&user.lastLogin,
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
		&user.lastLogin,
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

func New(id int, username, email, password string, status bool, createdAt, updatedAt, lastLogin time.Time) *User {
	return &User{
		id:        id,
		username:  username,
		email:     email,
		password:  password,
		status:    status,
		createdAt: createdAt,
		updatedAt: updatedAt,
		lastLogin: lastLogin,
	}
}

func (user *User) CreateVerificationToken() string {
	data := fmt.Sprintf(
		"%d:%s:%d:%s:%t:%s",
		user.updatedAt.UnixNano(),
		user.email,
		user.id,
		user.password,
		user.status,
		user.lastLogin,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (user *User) Login(db database.DatabaseInterface, session sessions.Session) (err error) {
	session.Set("userID", user.id)
	err = session.Save()

	if err != nil {
		return err
	}

	user.lastLogin = time.Now()
	user.Save(db)

	return nil
}
