package repository

import (
	"context"
	"os"
	"testing"
	"time"

	cfg "github.com/dwiw96/simple-auth-system/config"
	auth "github.com/dwiw96/simple-auth-system/features/auth"
	pg "github.com/dwiw96/simple-auth-system/utils/driver/postgresql"
	generator "github.com/dwiw96/simple-auth-system/utils/generator"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	repoTest auth.RepositoryInterface
	pool     *pgxpool.Pool
	ctx      context.Context
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

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	repoTest = NewAuthRepository(pool, ctx)

	os.Exit(m.Run())
}

func createRandomUser(t *testing.T) (user auth.User) {
	require.NotNil(t, pool)
	require.NotNil(t, ctx)

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
	return
}

func TestCheckEmail(t *testing.T) {
	for i := 0; i < 5; i++ {
		t.Run("success", func(t *testing.T) {
			user := createRandomUser(t)

			res, err := repoTest.CheckEmail(user.Email)
			require.NoError(t, err)
			assert.NotZero(t, res)
		})
	}

	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "error_empty_email",
			email: "",
		}, {
			name:  "error_invalid_email",
			email: "av088@mail.com",
		}, {
			name:  "error_typo_email",
			email: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			account := createRandomUser(t)

			if test.name == "error_typo_email" {
				test.email = account.Email + "m"
			}
			res, _ := repoTest.CheckEmail(test.email)
			require.Zero(t, res)
		})
	}
}

func TestReadUser(t *testing.T) {
	for i := 0; i < 5; i++ {
		t.Run("success", func(t *testing.T) {
			user := createRandomUser(t)

			res, err := repoTest.ReadUser(user.Email)
			require.NoError(t, err)
			assert.NotZero(t, &res.ID)
			assert.Equal(t, user.Email, res.Email)
			assert.Equal(t, user.FirstName, res.FirstName)
			assert.Equal(t, user.MiddleName, res.MiddleName)
			assert.Equal(t, user.LastName, res.LastName)
			assert.Equal(t, user.Address, res.Address)
			assert.Equal(t, user.Gender, res.Gender)
			assert.Equal(t, user.MaritalStatusID, res.MaritalStatusID)
			assert.Equal(t, user.HashedPassword, res.HashedPassword)
		})
	}

	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "error_empty_email",
			email: "",
		}, {
			name:  "error_invalid_email",
			email: "av088@mail.com",
		}, {
			name:  "error_typo_email",
			email: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			account := createRandomUser(t)

			if test.name == "error_typo_email" {
				test.email = account.Email + "m"
			}
			res, err := repoTest.ReadUser(test.email)
			require.Error(t, err)
			require.Nil(t, res)
		})
	}
}

func TestReadMaritalStatus(t *testing.T) {
	tests := []struct {
		name  string
		input string
		ans   auth.MaritalStatus
		err   bool
	}{
		{
			name:  "success_single",
			input: "single",
			ans: auth.MaritalStatus{
				ID:     1,
				Status: "single",
			},
			err: false,
		}, {
			name:  "success_married",
			input: "married",
			ans: auth.MaritalStatus{
				ID:     2,
				Status: "married",
			},
			err: false,
		}, {
			name:  "success_divorced",
			input: "divorced",
			ans: auth.MaritalStatus{
				ID:     3,
				Status: "divorced",
			},
			err: false,
		}, {
			name:  "error",
			input: "double",
			err:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := repoTest.ReadMaritalStatus(test.input)
			if !test.err {
				require.NoError(t, err)
				assert.Equal(t, test.ans, *res)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestInsertUser(t *testing.T) {
	email := generator.CreateRandomEmail(generator.CreateRandomString(5))
	tests := []struct {
		name  string
		input auth.User
		err   bool
	}{
		{
			name: "success",
			input: auth.User{
				FirstName:       generator.CreateRandomString(5),
				LastName:        generator.CreateRandomString(7),
				Email:           email,
				Address:         generator.CreateRandomString(20),
				Gender:          generator.CreateRandomGender(),
				MaritalStatusID: generator.CreateRandomMaritalStatusID(),
				HashedPassword:  generator.CreateRandomString(60),
			},
			err: false,
		}, {
			name: "error_nil_first_name",
			input: auth.User{
				LastName:        generator.CreateRandomString(7),
				Email:           generator.CreateRandomEmail(generator.CreateRandomString(5)),
				Address:         generator.CreateRandomString(20),
				Gender:          generator.CreateRandomGender(),
				MaritalStatusID: generator.CreateRandomMaritalStatusID(),
				HashedPassword:  generator.CreateRandomString(60),
			},
			err: true,
		}, {
			name: "error_empty_address",
			input: auth.User{
				FirstName:       generator.CreateRandomEmail(generator.CreateRandomString(5)),
				LastName:        generator.CreateRandomString(7),
				Email:           generator.CreateRandomEmail(generator.CreateRandomString(5)),
				Address:         "",
				Gender:          generator.CreateRandomGender(),
				MaritalStatusID: generator.CreateRandomMaritalStatusID(),
				HashedPassword:  generator.CreateRandomString(60),
			},
			err: true,
		}, {
			name: "error_duplicate_email",
			input: auth.User{
				FirstName:       generator.CreateRandomString(5),
				LastName:        generator.CreateRandomString(7),
				Email:           email,
				Address:         generator.CreateRandomString(20),
				Gender:          generator.CreateRandomGender(),
				MaritalStatusID: generator.CreateRandomMaritalStatusID(),
				HashedPassword:  generator.CreateRandomString(60),
			},
			err: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := repoTest.InsertUser(test.input)
			if !test.err {
				require.NoError(t, err)
				assert.Equal(t, test.input.FirstName, res.FirstName)
				assert.Equal(t, "", res.MiddleName)
				assert.Equal(t, test.input.LastName, res.LastName)
				assert.Equal(t, test.input.Email, res.Email)
				assert.Equal(t, test.input.Address, res.Address)
				assert.Equal(t, test.input.Gender, res.Gender)
				assert.Equal(t, test.input.MaritalStatusID, res.MaritalStatusID)
				assert.Equal(t, test.input.HashedPassword, res.HashedPassword)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestLoadKey(t *testing.T) {
	res, err := repoTest.LoadKey()
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestUpdateUserIsVerified(t *testing.T) {
	for i := 0; i < 5; i++ {
		user := createRandomUser(t)

		t.Run("success", func(t *testing.T) {
			err := repoTest.UpdateUserIsVerified(user.ID, user.Email)
			require.NoError(t, err)
		})
	}

	t.Run("fail", func(t *testing.T) {
		user := createRandomUser(t)

		err := repoTest.UpdateUserIsVerified(user.ID, "a"+user.Email)
		require.Error(t, err)
	})
}

func TestDeleteUser(t *testing.T) {
	for i := 0; i < 5; i++ {
		user := createRandomUser(t)

		t.Run("success", func(t *testing.T) {
			err := repoTest.DeleteUser(user.ID, user.Email)
			require.NoError(t, err)
		})
	}

	t.Run("fail_wrong_email", func(t *testing.T) {
		user := createRandomUser(t)

		err := repoTest.UpdateUserIsVerified(user.ID, "a"+user.Email)
		require.Error(t, err)
	})

	t.Run("fail_wrong_id", func(t *testing.T) {
		user := createRandomUser(t)

		err := repoTest.UpdateUserIsVerified(0, user.Email)
		require.Error(t, err)
	})
}
