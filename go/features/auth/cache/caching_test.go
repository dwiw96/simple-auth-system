package chache

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	cfg "github.com/dwiw96/simple-auth-system/config"
	auth "github.com/dwiw96/simple-auth-system/features/auth"
	middleware "github.com/dwiw96/simple-auth-system/middleware"
	conv "github.com/dwiw96/simple-auth-system/utils/converter"
	pg "github.com/dwiw96/simple-auth-system/utils/driver/postgresql"
	rd "github.com/dwiw96/simple-auth-system/utils/driver/redis"
	generator "github.com/dwiw96/simple-auth-system/utils/generator"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	chacheTest auth.CacheInterface
	pool       *pgxpool.Pool
	client     *redis.Client
	ctx        context.Context
)

func TestMain(m *testing.M) {
	os.Setenv("DB_USERNAME", "dwiw")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "auth_system_go")
	os.Setenv("REDIS_HOST", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "secret")
	os.Setenv("REDIS_DB", "0")

	redis_db, err := conv.ConvertStrToInt(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal(err)
	}

	envConfig := &cfg.EnvConfig{
		DB_USERNAME:    os.Getenv("DB_USERNAME"),
		DB_PASSWORD:    os.Getenv("DB_PASSWORD"),
		DB_HOST:        os.Getenv("DB_HOST"),
		DB_PORT:        os.Getenv("DB_PORT"),
		DB_NAME:        os.Getenv("DB_NAME"),
		REDIS_HOST:     os.Getenv("REDIS_HOST"),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
		REDIS_DB:       redis_db,
	}

	pool = pg.ConnectToPg(envConfig)

	client = rd.ConnectToRedis(envConfig)
	defer client.Close()

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	chacheTest = NewAuthCache(client, ctx)

	os.Exit(m.Run())
}

func TestCachingBlockedToken(t *testing.T) {
	key, err := middleware.LoadKey(ctx, pool)
	require.NoError(t, err)
	require.NotNil(t, key)

	user := auth.User{
		ID:              int64(generator.RandomInt(1, 100)),
		Fullname:        generator.CreateRandomString(5) + " " + generator.CreateRandomString(7),
		Email:           generator.CreateRandomEmail(generator.CreateRandomString(5)),
		Address:         generator.CreateRandomString(20),
		Gender:          generator.CreateRandomGender(),
		MaritalStatusID: generator.CreateRandomMaritalStatusID(),
	}
	token, err := middleware.CreateToken(user, 5, key)
	require.NoError(t, err)
	require.NotZero(t, len(token))

	payload, err := middleware.ReadToken(token, key)
	require.NoError(t, err)

	err = chacheTest.CachingBlockedToken(*payload)
	require.NoError(t, err)

	res, err := client.Get(ctx, fmt.Sprint("block ", payload.ID)).Result()
	fmt.Println(fmt.Sprint("block ", payload.ID))
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprint(payload.UserID), res)
}
