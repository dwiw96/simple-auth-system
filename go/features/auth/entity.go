package auth

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type SignupRequest struct {
	FirstName     string `json:"first_name"`
	MiddleName    string `json:"middle_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Gender        string `json:"gender"`
	MaritalStatus string `json:"marital_status"`
	Password      string `json:"password"`
}

type User struct {
	ID              int64
	Fullname        string
	FirstName       string `json:"first_name"`
	MiddleName      string `json:"middle_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	Address         string `json:"address"`
	Gender          string `json:"gender"`
	MaritalStatusID int64  `json:"marital_status"`
	HashedPassword  string `json:"hashed_password"`
	CreatedAt       time.Time
	MaritalStatus   string
	IsVerified      bool
}

type MaritalStatus struct {
	ID     int64
	Status string
}

var MaritalStatusMap = map[string]int64{
	"single":   1,
	"married":  2,
	"divorced": 3,
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type JwtPayload struct {
	jwt.RegisteredClaims
	ID      uuid.UUID `json:"id"`
	UserID  int64     `json:"user_id"`
	Iss     string    `json:"iss"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Address string    `json:"address,omitempty"`
	Iat     int64     `json:"iat"`
	Exp     int64     `json:"exp"`
}

type RepositoryInterface interface {
	CheckEmail(email string) (result int, err error)
	ReadUser(email string) (result *User, err error)
	ReadMaritalStatus(status string) (result *MaritalStatus, err error)
	InsertUser(input User) (result *User, err error)
	LoadKey() (key *rsa.PrivateKey, err error)
	UpdateUserIsVerified(id int64, email string) (err error)
	DeleteUser(id int64, email string) (err error)
}

type ServiceInterface interface {
	SignUp(input SignupRequest) (user *User, code int, err error)
	LogIn(input LoginRequest) (user *User, token string, code int, err error)
	LogOut(payload JwtPayload) error
	SendEmailVerification(user User) (code int, err error)
	EmailVerification(userID int64, email string) (code int, err error)
	DeleteUser(userID int64, email string) (code int, err error)
}

type CacheInterface interface {
	CachingBlockedToken(payload JwtPayload) error
}
