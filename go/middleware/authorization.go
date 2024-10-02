package middleware

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	auth "github.com/dwiw96/simple-auth-system/features/auth"
	response "github.com/dwiw96/simple-auth-system/utils/responses"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ContextKey string

var PayloadKey ContextKey = "payload"

func AuthMiddleware(ctx context.Context, pool *pgxpool.Pool, client *redis.Client) gin.HandlerFunc {
	return (func(c *gin.Context) {
		if c.Request.RequestURI == "/api/signin" {
			c.Next()
			return
		}
		if c.Request.RequestURI == "/api/login" {
			c.Next()
			return
		}

		key, err := LoadKey(ctx, pool)
		if err != nil {
			log.Println(err)
			response.ErrorJSON(c, 500, []string{err.Error()}, c.Request.RemoteAddr)
			c.Abort()
			return
		}

		authHeader, err := GetTokenHeader(c.Request)
		if err != nil {
			log.Println(err)
			response.ErrorJSON(c, 401, []string{err.Error()}, c.Request.RemoteAddr)
			c.Abort()
			return
		}

		isVerified, err := VerifyToken(authHeader, key)
		if err != nil {
			response.ErrorJSON(c, 401, []string{err.Error()}, c.Request.RemoteAddr)
			c.Abort()
			return
		}
		if !isVerified {
			response.ErrorJSON(c, 401, []string{"token is not valid"}, c.Request.RemoteAddr)
			c.Abort()
			return
		}

		payload, err := ReadToken(authHeader, key)
		if err != nil {
			response.ErrorJSON(c, 401, []string{err.Error()}, c.Request.RemoteAddr)
			c.Abort()
			return
		}

		err = CheckBlockedToken(client, ctx, payload.ID, payload.UserID)
		if err != nil {
			response.ErrorJSON(c, 401, []string{err.Error()}, c.Request.RemoteAddr)
			c.Abort()
			return
		}
		c.Set("payloadKey", payload)

		c.Next()
	})
}

func GetTokenHeader(r *http.Request) (token string, err error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header found")
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	return tokenString, nil
}

func CreateToken(reqData auth.User, key *rsa.PrivateKey) (token string, err error) {
	nowTime := time.Now().UTC()
	expTime := nowTime.Add(time.Minute * 60)

	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid, err: %v", err)
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256,
		auth.JwtPayload{
			ID:      id,
			UserID:  reqData.ID,
			Name:    reqData.Fullname,
			Email:   reqData.Email,
			Address: reqData.Address,
			Iat:     nowTime.Unix(),
			Exp:     expTime.Unix(),
		})

	token, err = t.SignedString(key)

	token = "Bearer " + token

	return
}

func VerifyToken(authHeader string, key *rsa.PrivateKey) (bool, error) {
	userToken := strings.Split(authHeader, " ")

	jwtToken, err := jwt.Parse(userToken[1], func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", jwtToken.Header["alg"])
		}

		return &key.PublicKey, nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to parse token when verifying, err: %v", err)
	}

	_, ok := jwtToken.Claims.(jwt.MapClaims)

	if !ok || !jwtToken.Valid {
		return false, fmt.Errorf("token is not valid")
	}

	return true, nil
}

func ReadToken(authHeader string, key *rsa.PrivateKey) (*auth.JwtPayload, error) {
	var payload auth.JwtPayload
	userToken := strings.Split(authHeader, " ")

	jwtToken, err := jwt.ParseWithClaims(userToken[1], &payload, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", jwtToken.Header["alg"])
		}

		return &key.PublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !jwtToken.Valid {
		return nil, fmt.Errorf("failed to parse token when read the token")
	}

	return &payload, err
}

type KeyCache struct {
	key        *rsa.PrivateKey
	expiration time.Time
	mutex      sync.Mutex
}

var keyCache = &KeyCache{
	key:        nil,
	expiration: time.Now(),
}

func LoadKey(ctx context.Context, conn *pgxpool.Pool) (key *rsa.PrivateKey, err error) {
	if keyCache.key != nil && time.Now().Before(keyCache.expiration) {
		return keyCache.key, nil
	}

	query := "select private_key from sec_m"
	var keyBytes []byte
	rows, err := conn.Query(ctx, query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&keyBytes)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(keyBytes)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		keyCache.mutex.Lock()
		defer keyCache.mutex.Unlock()
		keyCache.key = key
		keyCache.expiration = time.Now().Add(time.Hour)

		return privateKey, nil
	}

	return nil, errors.New("no private key found in database")
}

func CheckBlockedToken(redis *redis.Client, ctx context.Context, tokenID uuid.UUID, userID int64) error {
	check, err := redis.Exists(ctx, tokenID.String()).Result()
	if err != nil {
		return err
	}
	if check != 0 {
		return errors.New("token is blacklist")
	}

	return nil
}
