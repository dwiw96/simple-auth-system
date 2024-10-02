package delivery

import (
	auth "github.com/dwiw96/simple-auth-system/features/auth"
)

type signupRequest struct {
	FirstName     string `json:"first_name" validate:"required,min=2"`
	MiddleName    string `json:"middle_name"`
	LastName      string `json:"last_name" validate:"required,min=2"`
	Email         string `json:"email" validate:"email"`
	Address       string `json:"address" validate:"required,min=2"`
	Gender        string `json:"gender" validate:"required,oneof=male female"`
	MaritalStatus string `json:"marital_status" validate:"required,oneof=single married divorced"`
	Password      string `json:"password" validate:"min=7,required_with=alphanum"`
}

func toSignUpRequest(input signupRequest) auth.SignupRequest {
	return auth.SignupRequest{
		FirstName:     input.FirstName,
		MiddleName:    input.MiddleName,
		LastName:      input.LastName,
		Email:         input.Email,
		Address:       input.Address,
		Gender:        input.Gender,
		MaritalStatus: input.MaritalStatus,
		Password:      input.Password,
	}
}

type signinRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=7"`
}
