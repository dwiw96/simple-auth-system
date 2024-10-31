package repository

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"

	auth "github.com/dwiw96/simple-auth-system/features/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authRepository struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

func NewAuthRepository(pool *pgxpool.Pool, ctx context.Context) auth.RepositoryInterface {
	return &authRepository{
		pool: pool,
		ctx:  ctx,
	}
}

func (r *authRepository) CheckEmail(email string) (result int, err error) {
	query := "SELECT COUNT(email) FROM users WHERE email=$1"

	row := r.pool.QueryRow(r.ctx, query, email)

	var count int
	err = row.Scan(&count)
	if err != nil {
		errMsg := fmt.Errorf("failed check email, err: %v", err)
		return -1, errMsg
	}

	return count, nil
}

func (r *authRepository) ReadUser(email string) (result *auth.User, err error) {
	query := `
	SELECT 
		u.*,
		ms.status
	FROM 
		users u
	INNER JOIN
		marital_status ms ON u.marital_status_id = ms.id
	WHERE 
		u.email = $1;
	`

	row := r.pool.QueryRow(r.ctx, query, email)

	var user auth.User
	err = row.Scan(&user.ID, &user.FirstName, &user.MiddleName, &user.LastName, &user.Email, &user.Address, &user.Gender, &user.MaritalStatusID, &user.HashedPassword, &user.CreatedAt, &user.IsVerified, &user.MaritalStatus)
	if err != nil {
		errMsg := fmt.Errorf("failed read user, err: %v", err)
		return nil, errMsg
	}

	if user.MiddleName != "" {
		user.Fullname = user.FirstName + " " + user.MiddleName + " " + user.LastName
	} else {
		user.Fullname = user.FirstName + " " + user.LastName
	}

	return &user, nil
}

func (r *authRepository) ReadMaritalStatus(status string) (result *auth.MaritalStatus, err error) {
	query := "SELECT * FROM marital_status WHERE status = $1"

	row := r.pool.QueryRow(r.ctx, query, status)

	var res auth.MaritalStatus
	err = row.Scan(&res.ID, &res.Status)
	if err != nil {
		errMsg := fmt.Errorf("failed read marital status, err: %v", err)
		return nil, errMsg
	}

	return &res, err
}

func (r *authRepository) InsertUser(input auth.User) (result *auth.User, err error) {
	query := `INSERT INTO users(
		first_name,
		middle_name,
		last_name,
		email,
		address,
		gender,
		marital_status_id,
		hashed_password
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8
	) RETURNING *`

	row := r.pool.QueryRow(r.ctx, query, input.FirstName, input.MiddleName, input.LastName, input.Email, input.Address, input.Gender, input.MaritalStatusID, input.HashedPassword)

	var user auth.User
	err = row.Scan(&user.ID, &user.FirstName, &user.MiddleName, &user.LastName, &user.Email, &user.Address, &user.Gender, &user.MaritalStatusID, &user.HashedPassword, &user.CreatedAt, &user.IsVerified)
	if err != nil {
		errMsg := fmt.Errorf("failed to insert user err: %v", err)
		return nil, errMsg
	}

	if user.MiddleName != "" {
		user.Fullname = user.FirstName + " " + user.MiddleName + " " + user.LastName
	} else {
		user.Fullname = user.FirstName + " " + user.LastName
	}

	return &user, nil
}

func (r *authRepository) LoadKey() (key *rsa.PrivateKey, err error) {
	query := "select private_key from sec_m"
	var keyBytes []byte
	rows, err := r.pool.Query(r.ctx, query)
	if err != nil {
		errMsg := fmt.Errorf("failed to load private key, err: %v", err)
		return nil, errMsg
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&keyBytes)
		if err != nil {
			errMsg := fmt.Errorf("failed to scan private key, err: %v", err)
			return nil, errMsg
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(keyBytes)
		if err != nil {
			errMsg := fmt.Errorf("failed to parse private key, err: %v", err)
			return nil, errMsg
		}

		return privateKey, nil
	}

	return nil, errors.New("no private key found in database")
}

func (r *authRepository) UpdateUserIsVerified(id int64, email string) (err error) {
	query := "UPDATE users SET is_verified = TRUE WHERE id = $1 AND email = $2;"

	res, err := r.pool.Exec(r.ctx, query, id, email)
	if err != nil {
		return fmt.Errorf("failed to update user email verified, err: %v", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("failed to verifying, user still unverified")
	}

	return
}

func (r *authRepository) DeleteUser(id int64, email string) (err error) {
	query := "DELETE FROM users WHERE id = $1 AND email = $2;"

	res, err := r.pool.Exec(r.ctx, query, id, email)
	if err != nil {
		return fmt.Errorf("failed to delete user, err: %v", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("failed to unverifying, user data still not deleted")
	}

	return
}

func (r *authRepository) ReadRefreshToken(userID int64, refreshToken uuid.UUID) (res *auth.RefreshTokenWhitelist, err error) {
	query := "SELECT * FROM refresh_token_whitelist WHERE user_id = $1 AND refresh_token = $2;"

	var result auth.RefreshTokenWhitelist
	err = r.pool.QueryRow(r.ctx, query, userID, refreshToken).Scan(&result.ID, &result.UserID, &result.RefreshToken, &result.ExpiresAt, &result.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh token, err: %v", err)
	}

	return &result, nil
}

func (r *authRepository) InsertRefreshToken(userID int64, refreshToken uuid.UUID) (err error) {
	query := "INSERT INTO refresh_token_whitelist(user_id, refresh_token, expires_at) VALUES($1, $2, NOW() + INTERVAL '5 minute')"

	res, err := r.pool.Exec(r.ctx, query, userID, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token, err: %v", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("there are no rows affrected when insert refresh token")
	}

	return nil
}

func (r *authRepository) DeleteRefreshToken(userID int64) (err error) {
	query := "DELETE FROM refresh_token_whitelist WHERE user_id = $1;"

	res, err := r.pool.Exec(r.ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token, err: %v", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("there are no rows affrected when delete refresh token")
	}

	return nil
}

func (r *authRepository) ExecDbTx(fn func(*authRepository) error) error {
	tx, err := r.pool.Begin(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to start db transaction, err: %v", err)
	}

	err = fn(r)
	if err != nil {
		if rbErr := tx.Rollback(r.ctx); rbErr != nil {
			return fmt.Errorf("failed to rollback, err = %v", rbErr)
		}
		return fmt.Errorf("failed db transaction, err: %v", err)
	}

	return tx.Commit(r.ctx)
}

func (r *authRepository) UpdateRefreshToken(userID int64, refreshToken uuid.UUID) (err error) {
	r.ExecDbTx(func(ar *authRepository) error {
		err = ar.DeleteRefreshToken(userID)
		if err != nil {
			return err
		}

		err = ar.InsertRefreshToken(userID, refreshToken)
		if err != nil {
			return err
		}

		return nil
	})

	return nil
}
