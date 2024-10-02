package repository

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"

	auth "github.com/dwiw96/simple-auth-system/features/auth"

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
	err = row.Scan(&user.ID, &user.FirstName, &user.MiddleName, &user.LastName, &user.Email, &user.Address, &user.Gender, &user.MaritalStatusID, &user.HashedPassword, &user.CreatedAt, &user.MaritalStatus)
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
	err = row.Scan(&user.ID, &user.FirstName, &user.MiddleName, &user.LastName, &user.Email, &user.Address, &user.Gender, &user.MaritalStatusID, &user.HashedPassword, &user.CreatedAt)
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
