package middleware

import (
	"context"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	cfg "github.com/dwiw96/simple-auth-system/config"
	auth "github.com/dwiw96/simple-auth-system/features/auth"
	pg "github.com/dwiw96/simple-auth-system/utils/driver/postgresql"
	rd "github.com/dwiw96/simple-auth-system/utils/driver/redis"
	generator "github.com/dwiw96/simple-auth-system/utils/generator"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	pool   *pgxpool.Pool
	ctx    context.Context
	client *redis.Client
)

func TestMain(m *testing.M) {
	os.Setenv("DB_USERNAME", "dwiw")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "nongki_db")
	os.Setenv("REDIS_HOST", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "secret")

	envConfig := &cfg.EnvConfig{
		DB_USERNAME:    os.Getenv("DB_USERNAME"),
		DB_PASSWORD:    os.Getenv("DB_PASSWORD"),
		DB_HOST:        os.Getenv("DB_HOST"),
		DB_PORT:        os.Getenv("DB_PORT"),
		DB_NAME:        os.Getenv("DB_NAME"),
		REDIS_HOST:     os.Getenv("REDIS_HOST"),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
	}

	pool = pg.ConnectToPg(envConfig)

	client = rd.ConnectToRedis(envConfig)

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	os.Exit(m.Run())
}

func createTokenAndKey(t *testing.T) (string, auth.User, *rsa.PrivateKey) {
	key, err := LoadKey(ctx, pool)
	require.NoError(t, err)
	require.NotNil(t, key)

	firstname := generator.CreateRandomString(5)
	payload := auth.User{
		Fullname:      firstname + " " + generator.CreateRandomString(7),
		Email:         generator.CreateRandomEmail(firstname),
		Address:       generator.CreateRandomString(20),
		Gender:        generator.CreateRandomGender(),
		MaritalStatus: generator.CreateRandomMaritalStatus(),
	}

	token, err := CreateToken(payload, key)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	return token, payload, key
}

func TestGetToken(t *testing.T) {
	token, _, _ := createTokenAndKey(t)

	t.Log(token)
}

func TestVerifyToken(t *testing.T) {
	token, _, key := createTokenAndKey(t)

	t.Run("success", func(t *testing.T) {
		res, err := VerifyToken(token, key)
		require.NoError(t, err)
		require.True(t, res)
	})

	t.Run("failed", func(t *testing.T) {
		res, err := VerifyToken(token+"b", key)
		require.Error(t, err)
		require.False(t, res)
	})
}

func TestReadToken(t *testing.T) {
	token, payloadInput, key := createTokenAndKey(t)

	t.Run("success", func(t *testing.T) {
		payload, err := ReadToken(token, key)
		require.NoError(t, err)
		assert.Equal(t, payloadInput.Fullname, payload.Name)
		assert.Equal(t, payloadInput.Email, payload.Email)
		assert.Equal(t, payloadInput.Address, payload.Address)
	})

	t.Run("failed", func(t *testing.T) {
		payload, err := ReadToken(token+"b", key)
		require.Error(t, err)
		assert.Nil(t, payload)
	})
}

func TestCheckBlockedToken(t *testing.T) {
	token, _, key := createTokenAndKey(t)

	payload, err := ReadToken(token, key)
	require.NoError(t, err)

	t.Run("valid", func(t *testing.T) {
		err = CheckBlockedToken(client, ctx, payload.ID, payload.UserID)
		require.NoError(t, err)
	})

	t.Run("blacklist", func(t *testing.T) {
		iat := time.Unix(payload.Iat, 0)
		exp := time.Unix(payload.Exp, 0)
		duration := time.Duration(exp.Sub(iat).Nanoseconds())
		err = client.Set(ctx, payload.ID.String(), payload.UserID, duration).Err()
		require.NoError(t, err)

		err = CheckBlockedToken(client, ctx, payload.ID, payload.UserID)
		require.Error(t, err)
	})
}
