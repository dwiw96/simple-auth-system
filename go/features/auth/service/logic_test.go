package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	cfg "github.com/dwiw96/simple-auth-system/config"
	auth "github.com/dwiw96/simple-auth-system/features/auth"
	cache "github.com/dwiw96/simple-auth-system/features/auth/cache"
	repo "github.com/dwiw96/simple-auth-system/features/auth/repository"
	middleware "github.com/dwiw96/simple-auth-system/middleware"
	pg "github.com/dwiw96/simple-auth-system/utils/driver/postgresql"
	rd "github.com/dwiw96/simple-auth-system/utils/driver/redis"
	generator "github.com/dwiw96/simple-auth-system/utils/generator"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	serviceTest auth.ServiceInterface
	pool        *pgxpool.Pool
	ctx         context.Context
)

func TestMain(m *testing.M) {
	os.Setenv("DB_USERNAME", "dwiw")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "auth_system_go")

	envConfig := &cfg.EnvConfig{
		DB_USERNAME: os.Getenv("DB_USERNAME"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_PORT:     os.Getenv("DB_PORT"),
		DB_NAME:     os.Getenv("DB_NAME"),
	}

	pool = pg.ConnectToPg(envConfig)
	defer pool.Close()

	client := rd.ConnectToRedis(envConfig)
	defer client.Close()

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	repoTest := repo.NewAuthRepository(pool, ctx)
	cacheTest := cache.NewAuthCache(client, ctx)
	serviceTest = NewAuthService(repoTest, cacheTest)

	os.Exit(m.Run())
}

func createUser(t *testing.T) (user *auth.User, signupReq auth.SignupRequest) {
	email := generator.CreateRandomEmail(generator.CreateRandomString(5))

	input := auth.SignupRequest{
		FirstName:     generator.CreateRandomString(5),
		LastName:      generator.CreateRandomString(7),
		Email:         email,
		Address:       generator.CreateRandomString(20),
		Gender:        generator.CreateRandomGender(),
		MaritalStatus: generator.CreateRandomMaritalStatus(),
		Password:      generator.CreateRandomString(10),
	}

	res, code, err := serviceTest.SignUp(input)

	require.NoError(t, err)
	require.Zero(t, code)
	assert.Equal(t, input.FirstName, res.FirstName)
	assert.Equal(t, "", res.MiddleName)
	assert.Equal(t, input.LastName, res.LastName)
	assert.Equal(t, input.Email, res.Email)
	assert.Equal(t, input.Address, res.Address)
	assert.Equal(t, input.Gender, res.Gender)
	assert.Equal(t, auth.MaritalStatusMap[input.MaritalStatus], res.MaritalStatusID)
	assert.Equal(t, input.MaritalStatus, res.MaritalStatus)
	assert.NotEqual(t, input.Password, res.HashedPassword)

	return res, input
}

func TestSignUp(t *testing.T) {
	email := generator.CreateRandomEmail(generator.CreateRandomString(5))
	tests := []struct {
		name  string
		input auth.SignupRequest
		err   bool
	}{
		{
			name: "success",
			input: auth.SignupRequest{
				FirstName:     generator.CreateRandomString(5),
				LastName:      generator.CreateRandomString(7),
				Email:         email,
				Address:       generator.CreateRandomString(20),
				Gender:        generator.CreateRandomGender(),
				MaritalStatus: generator.CreateRandomMaritalStatus(),
				Password:      generator.CreateRandomString(10),
			},
			err: false,
		}, {
			name: "error_nil_first_name",
			input: auth.SignupRequest{
				LastName:      generator.CreateRandomString(7),
				Email:         generator.CreateRandomEmail(generator.CreateRandomString(5)),
				Address:       generator.CreateRandomString(20),
				Gender:        generator.CreateRandomGender(),
				MaritalStatus: generator.CreateRandomMaritalStatus(),
				Password:      generator.CreateRandomString(10),
			},
			err: true,
		}, {
			name: "error_empty_address",
			input: auth.SignupRequest{
				FirstName:     generator.CreateRandomEmail(generator.CreateRandomString(5)),
				LastName:      generator.CreateRandomString(7),
				Email:         generator.CreateRandomEmail(generator.CreateRandomString(5)),
				Address:       "",
				Gender:        generator.CreateRandomGender(),
				MaritalStatus: generator.CreateRandomMaritalStatus(),
				Password:      generator.CreateRandomString(10),
			},
			err: true,
		}, {
			name: "error_duplicate_email",
			input: auth.SignupRequest{
				FirstName:     generator.CreateRandomString(5),
				MiddleName:    generator.CreateRandomString(5),
				LastName:      generator.CreateRandomString(7),
				Email:         email,
				Address:       generator.CreateRandomString(20),
				Gender:        generator.CreateRandomGender(),
				MaritalStatus: generator.CreateRandomMaritalStatus(),
				Password:      generator.CreateRandomString(10),
			},
			err: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, code, err := serviceTest.SignUp(test.input)

			if !test.err {
				t.Log("res id:", res.ID)
				require.NoError(t, err)
				require.Zero(t, code)
				assert.Equal(t, test.input.FirstName, res.FirstName)
				assert.Equal(t, test.input.MiddleName, res.MiddleName)
				assert.Equal(t, test.input.LastName, res.LastName)
				assert.Equal(t, test.input.Email, res.Email)
				assert.Equal(t, test.input.Address, res.Address)
				assert.Equal(t, test.input.Gender, res.Gender)
				assert.Equal(t, auth.MaritalStatusMap[test.input.MaritalStatus], res.MaritalStatusID)
				assert.Equal(t, test.input.MaritalStatus, res.MaritalStatus)
				assert.NotEqual(t, test.input.Password, res.HashedPassword)
			} else {
				require.Error(t, err)
				require.NotZero(t, code)
			}
		})
	}
}

func TestLogIn(t *testing.T) {
	user, signUpReq := createUser(t)

	tests := []struct {
		name  string
		input auth.LoginRequest
		err   bool
		code  int
	}{
		{
			name: "success",
			input: auth.LoginRequest{
				Email:    signUpReq.Email,
				Password: signUpReq.Password,
			},
			err:  false,
			code: 1,
		}, {
			name: "error_email_wrong",
			input: auth.LoginRequest{
				Email:    "err" + signUpReq.Email,
				Password: signUpReq.Password,
			},
			err:  true,
			code: 2,
		}, {
			name: "success_password_wrong",
			input: auth.LoginRequest{
				Email:    signUpReq.Email,
				Password: "err" + signUpReq.Password,
			},
			err:  true,
			code: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, token, code, err := serviceTest.LogIn(test.input)
			if !test.err {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.Equal(t, 200, code)
				user.Fullname = res.Fullname

				assert.Equal(t, user.FirstName, res.FirstName)
				assert.Equal(t, user.MiddleName, res.MiddleName)
				assert.Equal(t, user.LastName, res.LastName)
				assert.Equal(t, user.Email, res.Email)
				assert.Equal(t, user.Address, res.Address)
				assert.Equal(t, user.Gender, res.Gender)
				assert.Equal(t, auth.MaritalStatusMap[user.MaritalStatus], res.MaritalStatusID)
				assert.Equal(t, user.MaritalStatus, res.MaritalStatus)
				assert.Equal(t, user.HashedPassword, res.HashedPassword)
				assert.NotZero(t, res.CreatedAt)
			} else {
				require.Error(t, err)
				assert.Empty(t, token)
				assert.Equal(t, 401, code)
				assert.Nil(t, res)
			}

			if test.code == 2 {
				assert.Equal(t, err, fmt.Errorf("no user found with this email %s", test.input.Email))
			} else if test.code == 3 {
				assert.Equal(t, err, errors.New("password is wrong"))
			}
		})
	}
}

func TestLogOut(t *testing.T) {
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
	token, err := middleware.CreateToken(user, key)
	require.NoError(t, err)
	require.NotZero(t, len(token))
	t.Log("TOKEN:", token)

	payload, err := middleware.ReadToken(token, key)
	require.NoError(t, err)

	err = serviceTest.LogOut(*payload)
	require.NoError(t, err)
}
