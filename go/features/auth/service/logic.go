package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	auth "github.com/dwiw96/simple-auth-system/features/auth"
	middleware "github.com/dwiw96/simple-auth-system/middleware"
	mail "github.com/dwiw96/simple-auth-system/utils/email"
	password "github.com/dwiw96/simple-auth-system/utils/password"
)

type authService struct {
	repo  auth.RepositoryInterface
	cache auth.CacheInterface
}

func NewAuthService(repo auth.RepositoryInterface, cache auth.CacheInterface) auth.ServiceInterface {
	return &authService{
		repo:  repo,
		cache: cache,
	}
}

func createLinkVerification(token string) (verifyUrl, unVerifyUrl string) {
	verifyUrl = fmt.Sprintf("http://localhost:9090/fe/email/verification?token=%s", token[7:])
	unVerifyUrl = fmt.Sprintf("http://localhost:9090/fe/email/unverification?token=%s", token[7:])
	return
}

func (s *authService) SignUp(input auth.SignupRequest) (user *auth.User, code int, err error) {
	resCheckEmail, err := s.repo.CheckEmail(input.Email)
	if err != nil {
		return nil, 500, err
	}

	if resCheckEmail != 0 {
		return nil, 409, fmt.Errorf("email is registered")
	}

	maritalStatus, err := s.repo.ReadMaritalStatus(input.MaritalStatus)
	if err != nil {
		return nil, 400, err
	}

	userInput := auth.User{
		FirstName:       input.FirstName,
		MiddleName:      input.MiddleName,
		LastName:        input.LastName,
		Email:           input.Email,
		Address:         input.Address,
		Gender:          input.Gender,
		MaritalStatusID: maritalStatus.ID,
		HashedPassword:  input.Password,
	}

	userInput.HashedPassword, err = password.HashingPassword(input.Password)
	if err != nil {
		return nil, 500, err
	}

	user, err = s.repo.InsertUser(userInput)
	if err != nil {
		return nil, 400, err
	}
	user.MaritalStatus = maritalStatus.Status

	code, err = s.SendEmailVerification(*user)
	if err != nil {
		return nil, code, err
	}

	return user, 0, nil
}

func (s *authService) SendEmailVerification(user auth.User) (code int, err error) {
	key, err := s.repo.LoadKey()
	if err != nil {
		return 500, fmt.Errorf("load key error: %w", err)
	}

	token, err := middleware.CreateToken(user, 10, key)
	if err != nil {
		return 500, errors.New("failed generate authentication token")
	}

	verifyUrl, unVerifyUrl := createLinkVerification(token)
	urlPlaceholder := map[string]interface{}{
		"verify":   verifyUrl,
		"unverify": unVerifyUrl,
	}

	err = mail.SendEmail(user.Email, "[Simple Auth System - Go]: Sign Up Verification", "assets/email_signup.html", urlPlaceholder)
	if err != nil {
		return 500, err
	}

	return
}

func (s *authService) LogIn(input auth.LoginRequest) (user *auth.User, token string, code int, err error) {
	user, err = s.repo.ReadUser(input.Email)
	if err != nil {
		code = 500
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			errMsg := fmt.Errorf("no user found with this email %s", input.Email)
			return nil, "", 401, errMsg
		}
		return nil, "", 500, err
	}

	err = password.VerifyHashPassword(input.Password, user.HashedPassword)
	if err != nil {
		errMsg := errors.New("password is wrong")
		return nil, "", 401, errMsg
	}

	key, err := s.repo.LoadKey()
	if err != nil {
		return nil, "", 500, fmt.Errorf("load key error: %w", err)
	}

	token, err = middleware.CreateToken(*user, 60, key)
	if err != nil {
		errMsg := errors.New("failed generate authentication token")
		return nil, "", 500, errMsg
	}

	return user, token, 200, nil
}

func (s *authService) LogOut(payload auth.JwtPayload) error {
	return s.cache.CachingBlockedToken(payload)
}

func (s *authService) EmailVerification(userID int64, email string) (code int, err error) {
	err = s.repo.UpdateUserIsVerified(userID, email)
	if err != nil {
		return 400, err
	}

	return 0, err
}

func (s *authService) DeleteUser(userID int64, email string) (code int, err error) {
	err = s.repo.DeleteUser(userID, email)
	if err != nil {
		return 400, err
	}

	return 0, nil
}
