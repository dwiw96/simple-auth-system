package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
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
	os.Setenv("DB_NAME", "auth_system_go")
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

	token, err := CreateToken(payload, 5, key)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	return token, payload, key
}

func TestCreateToken(t *testing.T) {
	createTokenAndKey(t)
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

func TestLoadKey(t *testing.T) {
	res, err := LoadKey(ctx, pool)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestCheckBlockedToken(t *testing.T) {
	token, _, key := createTokenAndKey(t)

	payload, err := ReadToken(token, key)
	require.NoError(t, err)

	t.Run("valid", func(t *testing.T) {
		err = CheckBlockedToken(client, ctx, payload.ID)
		require.NoError(t, err)
	})

	t.Run("blacklist", func(t *testing.T) {
		iat := time.Unix(payload.Iat, 0)
		exp := time.Unix(payload.Exp, 0)
		duration := time.Duration(exp.Sub(iat).Nanoseconds())
		err = client.Set(ctx, "block "+payload.ID.String(), payload.UserID, duration).Err()
		require.NoError(t, err)

		err = CheckBlockedToken(client, ctx, payload.ID)
		require.Error(t, err)
	})
}

func TestGetHeaderToken(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://localhost:8080/", nil)
	require.NoError(t, err)

	authHeader := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImYyNGZiYmExLTE5NDctNGNhYy05ODA4LTM2ZDY2YzQ2NzIwMCIsInVzZXJfaWQiOjc2LCJpc3MiOiIiLCJuYW1lIjoiR3JhY2UgRG9lIEp1bmlvciIsImVtYWlsIjoiZ3JhY2VAbWFpbC5jb20iLCJhZGRyZXNzIjoiQ2lyY2xlIFN0cmVldCwgTm8uMSwgQmFuZHVuZywgV2VzdCBKYXZhIiwiaWF0IjoxNzI3ODM3MjY5LCJleHAiOjE3Mjc4NDA4Njl9.mdQtJ22xRT5n8xYp5dGdVIzBo-OOocnaE6F054C0LEImf1rA_Fo0_fd3IGVa3XW5kDdpobqB8K6hDFm-XCPbkxvIfXjsjAwGqDrlzsjLiNmSvRwUj6FFWUkIpS_4Nl7Szcc2dEXe7n75LOs9yIhzNmuNjyC9Ago8BJiTYL0_jAkzxlHUwSaRj6naxbsLpiRhpjAW14-ema0wdbbHkaPkv0cj6rOQlsRTCW6R6i_2lrew5eOHIR750gBdImJ8HGtzB29yUA3A9P0-rGjITwZTanoqtOdv5d6lSMJ7eYMEACe4Lj3-k93V65e2ZJEFCnutk0H2ZPSaMBZwTx9B32S8JQ"

	r.Header.Set("Authorization", "Bearer "+authHeader)

	token, err := GetTokenHeader(r)
	require.NoError(t, err)
	assert.Equal(t, authHeader, token)
}

func TestPayloadVerification(t *testing.T) {
	var user auth.User
	user.FirstName = generator.CreateRandomString(int(generator.RandomInt(3, 13)))
	user.MiddleName = generator.CreateRandomString(int(generator.RandomInt(3, 13)))
	user.LastName = generator.CreateRandomString(int(generator.RandomInt(3, 13)))
	user.Email = generator.CreateRandomEmail(user.FirstName)
	user.Address = generator.CreateRandomString(int(generator.RandomInt(20, 50)))
	user.Gender = generator.CreateRandomGender()
	user.MaritalStatusID = generator.CreateRandomMaritalStatusID()
	user.HashedPassword = generator.CreateRandomString(int(generator.RandomInt(5, 10)))

	assert.NotEmpty(t, user.FirstName)
	assert.NotEmpty(t, user.LastName)
	assert.NotEmpty(t, user.LastName)
	assert.NotEmpty(t, user.Email)
	assert.NotEmpty(t, user.Address)
	assert.NotEmpty(t, user.Gender)
	assert.LessOrEqual(t, user.MaritalStatusID, int64(3))
	assert.GreaterOrEqual(t, user.MaritalStatusID, int64(1))
	assert.NotEmpty(t, user.HashedPassword)

	query := `
	INSERT INTO users(
		email,
		first_name,
		middle_name,
		last_name,
		address,
		gender,
		marital_status_id,
		hashed_password
	) VALUES 
		($1, $2, $3, $4, $5, $6, $7, $8) 
	RETURNING id;`

	row := pool.QueryRow(ctx, query, user.Email, user.FirstName, user.MiddleName, user.LastName, user.Address, user.Gender, user.MaritalStatusID, user.HashedPassword)
	err := row.Scan(&user.ID)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)

	if user.MiddleName != "" {
		user.Fullname = user.FirstName + " " + user.MiddleName + " " + user.LastName
	} else {
		user.Fullname = user.FirstName + " " + user.LastName
	}

	err = PayloadVerification(ctx, pool, user.Email, user.Fullname)
	require.NoError(t, err)
}
